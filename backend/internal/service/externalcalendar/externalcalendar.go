package externalcalendar

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/oauthutil"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Service struct {
	externalCalendarRepo domain.ExternalCalendarRepository
	blockedTimeRepo      domain.BlockedTimeRepository
	merchantRepo         domain.MerchantRepository
	bookingRepo          domain.BookingRepository
	teamRepo             domain.TeamRepository
	txManager            db.TransactionManager
}

func NewService(externalCalendar domain.ExternalCalendarRepository, blockedTime domain.BlockedTimeRepository,
	merchant domain.MerchantRepository, booking domain.BookingRepository, team domain.TeamRepository,
	txManager db.TransactionManager) *Service {
	return &Service{
		externalCalendarRepo: externalCalendar,
		blockedTimeRepo:      blockedTime,
		merchantRepo:         merchant,
		bookingRepo:          booking,
		teamRepo:             team,
		txManager:            txManager,
	}
}

var googleCalendarConf = &oauth2.Config{
	ClientID:     config.LoadEnvVars().GOOGLE_OAUTH_CLIENT_ID,
	ClientSecret: config.LoadEnvVars().GOOGLE_OAUTH_CLIENT_SECRET,
	RedirectURL:  "http://localhost:8080/api/v1/merchant/integrations/google/calendar/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/calendar"},
	Endpoint:     google.Endpoint,
}

func (s *Service) GoogleCalendar(ctx context.Context) (string, string, error) {
	state, err := oauthutil.GenerateSate()
	if err != nil {
		return "", "", err
	}

	// TODO: intelligent prompt consent if annoying
	url := googleCalendarConf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))

	return url, state, nil
}

func (s *Service) GoogleCalendarCallback(ctx context.Context, code string) error {
	token, err := googleCalendarConf.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("error during google oauth exchange: %s", err.Error())
	}

	service, err := calendar.NewService(ctx, option.WithTokenSource(googleCalendarConf.TokenSource(ctx, token)))
	if err != nil {
		return fmt.Errorf("error while creating new google calendar service: %s", err.Error())
	}

	cal, err := service.Calendars.Get("primary").Do()
	if err != nil {
		return fmt.Errorf("error while getting primary calendar: %s", err.Error())
	}

	employee := jwt.MustGetEmployeeFromContext(ctx)

	exists := true

	externalCalendar, err := s.externalCalendarRepo.GetExternalCalendarByEmployeeId(ctx, employee.Id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("error while getting existing external calendar: %s", err.Error())
		}

		exists = false
	}

	// TODO: optional resync here
	if exists {
		if token.RefreshToken != "" {
			err = s.externalCalendarRepo.UpdateExternalCalendarAuthTokens(ctx, externalCalendar.Id, token.AccessToken, token.RefreshToken, token.Expiry)
			if err != nil {
				return fmt.Errorf("error updating external calendar auth tokens: %s", err.Error())
			}
		}
	} else {

		var calendarTz *time.Location
		if cal.TimeZone != "" {
			calendarTz, err = time.LoadLocation(cal.TimeZone)
			if err != nil {
				return fmt.Errorf("error while parsing google calendar timezone: %s", err.Error())
			}
		} else {
			calendarTz, err = s.merchantRepo.GetMerchantTimezone(ctx, employee.MerchantId)
			if err != nil {
				return fmt.Errorf("error while loading merchant timezone: %s", err.Error())
			}
		}

		externalCalendar := domain.ExternalCalendar{
			EmployeeId:    employee.Id,
			CalendarId:    cal.Id,
			AccessToken:   token.AccessToken,
			RefreshToken:  token.RefreshToken,
			TokenExpiry:   token.Expiry,
			SyncToken:     nil,
			ChannelId:     nil,
			ResourceId:    nil,
			ChannelExpiry: nil,
			Timezone:      calendarTz.String(),
		}

		extCalendarId, err := s.externalCalendarRepo.NewExternalCalendar(ctx, externalCalendar)
		if err != nil {
			return fmt.Errorf("error while creating new external calendar: %s", err.Error())
		}

		externalCalendar.Id = extCalendarId

		err = s.initialCalendarSync(ctx, service, externalCalendar, calendarTz, employee.MerchantId)
		if err != nil {
			return fmt.Errorf("error during initial external calendar sync: %s", err.Error())
		}
	}

	return nil
}

func (s *Service) GoogleCalendarWatch(ctx context.Context, channelId, resourceId string) {
	extCalendar, err := s.externalCalendarRepo.GetExternalCalendarByChannel(ctx, channelId, resourceId)
	if err != nil {
		return
	}

	// TODO: call incremental sync as background job
	err = s.incrementalCalendarSync(ctx, extCalendar)
	if err != nil {
		return
	}
}
