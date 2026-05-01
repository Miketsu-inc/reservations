package externalcalendar

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/oauthutil"
	"github.com/miketsu-inc/reservations/backend/pkg/queue"
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
	enqueuer             queue.Enqueuer
	txManager            db.TransactionManager
}

func NewService(externalCalendar domain.ExternalCalendarRepository, blockedTime domain.BlockedTimeRepository,
	merchant domain.MerchantRepository, booking domain.BookingRepository, team domain.TeamRepository,
	enqueuer queue.Enqueuer, txManager db.TransactionManager) *Service {
	return &Service{
		externalCalendarRepo: externalCalendar,
		blockedTimeRepo:      blockedTime,
		merchantRepo:         merchant,
		bookingRepo:          booking,
		teamRepo:             team,
		enqueuer:             enqueuer,
		txManager:            txManager,
	}
}

func (s *Service) SetEnqueuer(client queue.Enqueuer) {
	s.enqueuer = client
}

type OAuthState struct {
	UserId     string `json:"user_id"`
	MerchantId string `json:"merchant_id"`
	LocationId int    `json:"location_id"`
	EmployeeId int    `json:"employee_id"`
	Role       string `json:"role"`

	ExpiresAt int64  `json:"exp"`
	Nonce     string `json:"nonce"`
}

func GenerateState(a actor.EmployeeContext) (string, error) {
	nonce, err := oauthutil.RandomString(16)
	if err != nil {
		return "", err
	}

	state := OAuthState{
		UserId:     a.UserId.String(),
		MerchantId: a.MerchantId.String(),
		LocationId: a.LocationId,
		EmployeeId: a.EmployeeId,
		Role:       a.Role.String(),
		ExpiresAt:  time.Now().Add(10 * time.Minute).Unix(),
		Nonce:      nonce,
	}

	payload, err := json.Marshal(state)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, []byte(config.LoadEnvVars().OAUTH_STATE_SECRET))
	mac.Write(payload)
	signature := mac.Sum(nil)

	final := append(payload, signature...)

	return base64.RawURLEncoding.EncodeToString(final), nil
}

func ParseState(encoded string) (*OAuthState, error) {
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	if len(data) < sha256.Size {
		return nil, fmt.Errorf("invalid state")
	}

	payload := data[:len(data)-sha256.Size]
	signature := data[len(data)-sha256.Size:]

	mac := hmac.New(sha256.New, []byte(config.LoadEnvVars().OAUTH_STATE_SECRET))
	mac.Write(payload)

	expected := mac.Sum(nil)

	if !hmac.Equal(signature, expected) {
		return nil, fmt.Errorf("invalid state signature")
	}

	var state OAuthState
	if err := json.Unmarshal(payload, &state); err != nil {
		return nil, err
	}

	if time.Now().Unix() > state.ExpiresAt {
		return nil, fmt.Errorf("state expired")
	}

	return &state, nil
}

var googleCalendarConf = &oauth2.Config{
	ClientID:     config.LoadEnvVars().GOOGLE_OAUTH_CLIENT_ID,
	ClientSecret: config.LoadEnvVars().GOOGLE_OAUTH_CLIENT_SECRET,
	RedirectURL:  "http://localhost:8080/api/v1/integrations/google/calendar/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/calendar"},
	Endpoint:     google.Endpoint,
}

func (s *Service) GoogleCalendar(ctx context.Context) (string, error) {
	actor := actor.MustGetFromContext(ctx)

	state, err := GenerateState(actor)
	if err != nil {
		return "", err
	}

	// TODO: intelligent prompt consent if annoying
	url := googleCalendarConf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))

	return url, nil
}

func (s *Service) GoogleCalendarCallback(ctx context.Context, code string, urlState string) error {
	state, err := ParseState(urlState)
	if err != nil {
		return fmt.Errorf("invalid oauth sate: %s", err.Error())
	}

	merchantId, err := uuid.Parse(state.MerchantId)
	if err != nil {
		return fmt.Errorf("error parsing merchanId from state: %s", err.Error())
	}

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

	exists := true

	externalCalendar, err := s.externalCalendarRepo.GetExternalCalendarByEmployeeId(ctx, state.EmployeeId)
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
			calendarTz, err = s.merchantRepo.GetMerchantTimezone(ctx, merchantId)
			if err != nil {
				return fmt.Errorf("error while loading merchant timezone: %s", err.Error())
			}
		}

		externalCalendar := domain.ExternalCalendar{
			EmployeeId:    state.EmployeeId,
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

		err = s.initialCalendarSync(ctx, service, externalCalendar, calendarTz, merchantId)
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

	_, err = s.enqueuer.Insert(ctx, args.IncrementalCalendarSync{
		ExternalCalendarId: extCalendar.Id,
	}, nil)
	if err != nil {
		return
	}
}
