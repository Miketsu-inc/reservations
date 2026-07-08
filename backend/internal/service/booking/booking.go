package booking

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bojanz/currency"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/service/email"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/internal/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/queue"
	"github.com/riverqueue/river"
	"github.com/teambition/rrule-go"
)

type Service struct {
	bookingRepo     domain.BookingRepository
	catalogRepo     domain.CatalogRepository
	merchantRepo    domain.MerchantRepository
	userRepo        domain.UserRepository
	customerRepo    domain.CustomerRepository
	blockedTimeRepo domain.BlockedTimeRepository
	mailer          *email.Service
	enqueuer        queue.Enqueuer
	txManager       db.TransactionManager
}

func NewService(booking domain.BookingRepository, catalog domain.CatalogRepository, merchant domain.MerchantRepository,
	user domain.UserRepository, customer domain.CustomerRepository, blockedTime domain.BlockedTimeRepository,
	mailer *email.Service, enqueuer queue.Enqueuer, txManager db.TransactionManager) *Service {
	return &Service{
		bookingRepo:     booking,
		catalogRepo:     catalog,
		merchantRepo:    merchant,
		userRepo:        user,
		customerRepo:    customer,
		blockedTimeRepo: blockedTime,
		mailer:          mailer,
		enqueuer:        enqueuer,
		txManager:       txManager,
	}
}

func (s *Service) SetEnqueuer(client queue.Enqueuer) {
	s.enqueuer = client
}

func (s *Service) newBooking(ctx context.Context, tx pgx.Tx, booking domain.Booking, participants []domain.BookingParticipant,
	service domain.Service) (int, error) {
	bookingId, err := s.bookingRepo.WithTx(tx).NewBooking(ctx, booking)
	if err != nil {
		return 0, err
	}

	bookingPhases := service.CalculateNewBookingPhases(bookingId, booking.FromDate)

	for i := range participants {
		participants[i].BookingId = bookingId
	}

	err = s.bookingRepo.WithTx(tx).NewBookingPhases(ctx, bookingPhases)
	if err != nil {
		return 0, err
	}

	err = s.bookingRepo.WithTx(tx).NewBookingParticipants(ctx, participants)
	if err != nil {
		return 0, err
	}

	if booking.EmployeeId != nil {
		_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncNewBooking{
			BookingId: bookingId,
		}, nil)
		if err != nil {
			return 0, err
		}
	}

	return bookingId, nil
}

func enforceBookingWindow(fromDate time.Time, now time.Time, windowMin, windowMax int) error {
	if fromDate.Before(now.Add(time.Duration(windowMin) * time.Minute)) {
		return fmt.Errorf("appointment must be booked at least %d minutes in advance", windowMin)
	}

	if fromDate.After(now.AddDate(0, windowMax, 0)) {
		return fmt.Errorf("appointment cannot be booked more than %d months in advance", windowMax)
	}

	return nil
}

func getNewBookingStatus(approvalPolicy types.ApprovalType, isNewCustomer bool) (types.BookingStatus, error) {
	var status types.BookingStatus

	switch approvalPolicy {
	case types.ApprovalTypeAuto:
		status = types.BookingStatusConfirmed
	case types.ApprovalTypeManual:
		status = types.BookingStatusBooked
	case types.ApprovalTypeManualForNew:
		if isNewCustomer {
			status = types.BookingStatusBooked
		} else {
			status = types.BookingStatusConfirmed
		}
	default:
		return types.BookingStatus{}, fmt.Errorf("invalid approval policy for merchant")
	}

	return status, nil
}

type CreateByCustomerInput struct {
	MerchantName string
	ServiceId    int
	LocationId   int
	TimeStamp    time.Time
	CustomerNote string
	// only present on group bookings
	BookingId *int
}

func (s *Service) CreateByCustomer(ctx context.Context, input CreateByCustomerInput) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, input.MerchantName)
	if err != nil {
		return err
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, merchantId)
	if err != nil {
		return err
	}

	bookingSettings, err := s.merchantRepo.GetBookingSettingsByMerchantAndService(ctx, merchantId, input.ServiceId)
	if err != nil {
		return err
	}

	fromDate := input.TimeStamp.UTC()

	err = enforceBookingWindow(fromDate, time.Now().In(merchantTz), bookingSettings.BookingWindowMin, bookingSettings.BookingWindowMax)
	if err != nil {
		return err
	}

	// TODO: we should probably just check by querying the user
	customerId, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("unexpected error during creating customer id: %w", err)
	}

	isGroupBooking := input.BookingId != nil

	err = s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		customerId, isBlacklisted, isNewCustomer, err := s.customerRepo.WithTx(tx).NewCustomerFromUser(ctx, customerId, merchantId, userId)
		if err != nil {
			return err
		}

		if isBlacklisted {
			return fmt.Errorf("you are blacklisted, please contact the merchant by email or phone to make a booking")
		}

		bookingStatus, err := getNewBookingStatus(bookingSettings.ApprovalPolicy, isNewCustomer)
		if err != nil {
			return err
		}

		var bookingId int

		if isGroupBooking {
			bookingId = *input.BookingId

			booking, err := s.bookingRepo.WithTx(tx).GetBooking(ctx, bookingId)
			if err != nil {
				return err
			}

			err = booking.CanBookGroup(bookingSettings.BookingWindowMin, bookingSettings.BookingWindowMax)
			if err != nil {
				return err
			}

			_, err = s.bookingRepo.WithTx(tx).UpdateParticipantCountBatch(ctx, []int{booking.Id}, []int{1})
			if err != nil {
				return err
			}

			newTotalPrice, err := booking.TotalPrice.Add(booking.PricePerPerson.Amount)
			if err != nil {
				return err
			}

			err = s.bookingRepo.WithTx(tx).UpdateBookingTotalPriceBatch(ctx, []int{bookingId}, []currencyx.Price{{Amount: newTotalPrice}})
			if err != nil {
				return err
			}

			participants := []domain.BookingParticipant{{
				Status:       bookingStatus,
				BookingId:    bookingId,
				CustomerId:   &customerId,
				CustomerNote: &input.CustomerNote,
			}}

			err = s.bookingRepo.WithTx(tx).NewBookingParticipants(ctx, participants)
			if err != nil {
				return err
			}

		} else {
			service, err := s.catalogRepo.GetServiceWithPhases(ctx, input.ServiceId, merchantId)
			if err != nil {
				return err
			}

			duration := time.Duration(service.TotalDuration) * time.Minute
			toDate := fromDate.Add(duration)

			price, err := s.preventNilBookingPrice(ctx, merchantId, service.Price)
			if err != nil {
				return err
			}

			location, err := s.merchantRepo.GetLocation(ctx, input.LocationId, merchantId)
			if err != nil {
				return err
			}

			booking := domain.Booking{
				Status:              bookingStatus,
				BookingType:         types.BookingTypeAppointment,
				MerchantId:          merchantId,
				EmployeeId:          nil,
				ServiceId:           &input.ServiceId,
				LocationId:          input.LocationId,
				FromDate:            fromDate,
				ToDate:              toDate,
				ServiceName:         service.Name,
				PricePerPerson:      price,
				TotalPrice:          price,
				PriceType:           service.PriceType,
				FormattedLocation:   location.FormattedLocation,
				MinParticipants:     1,
				MaxParticipants:     1,
				CurrentParticipants: 1,
			}

			participants := []domain.BookingParticipant{{
				Status:       bookingStatus,
				CustomerId:   &customerId,
				CustomerNote: &input.CustomerNote,
			}}

			bookingId, err = s.newBooking(ctx, tx, booking, participants, service)
			if err != nil {
				return err
			}
		}

		err = s.scheduleNewBookingEmails(ctx, tx, []uuid.UUID{customerId}, []types.BookingStatus{bookingStatus}, bookingId, fromDate)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error creating new booking by customer: %s", err.Error())
	}

	return nil
}

type CancelByCustomerInput struct {
	BookingId    int
	MerchantName string
}

func (s *Service) CancelByCustomer(ctx context.Context, input CancelByCustomerInput) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	booking, err := s.bookingRepo.GetBooking(ctx, input.BookingId)
	if err != nil {
		return err
	}

	cancelDeadline, err := s.bookingRepo.GetBookingCancelDeadline(ctx, booking.Id)
	if err != nil {
		return err
	}

	latestCancelTime := booking.FromDate.Add(-time.Duration(cancelDeadline) * time.Minute)

	err = booking.CanCancelWithDeadline(latestCancelTime)
	if err != nil {
		return err
	}

	bookingParticipant, err := s.bookingRepo.GetBookingParticipantByUser(ctx, booking.Id, userId)
	if err != nil {
		return err
	}

	err = bookingParticipant.CanModify()
	if err != nil {
		return err
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err = s.bookingRepo.WithTx(tx).UpdateParticipantStatus(ctx, booking.Id, bookingParticipant.Id, types.BookingStatusCancelled)
		if err != nil {
			return err
		}

		if booking.IsGroupBooking() {
			newTotalPrice, err := booking.TotalPrice.Sub(booking.PricePerPerson.Amount)
			if err != nil {
				return fmt.Errorf("failed to calculate total price: %w", err)
			}

			err = s.bookingRepo.WithTx(tx).UpdateBookingTotalPriceBatch(ctx, []int{booking.Id}, []currencyx.Price{{Amount: newTotalPrice}})
			if err != nil {
				return err
			}

			_, err = s.bookingRepo.WithTx(tx).UpdateParticipantCountBatch(ctx, []int{booking.Id}, []int{-1})
			if err != nil {
				return err
			}

		} else {
			err = s.bookingRepo.WithTx(tx).UpdateBookingStatus(ctx, booking.MerchantId, booking.Id, types.BookingStatusCancelled)
			if err != nil {
				return err
			}

			_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncDeleteBooking{
				BookingId: booking.Id,
			}, nil)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Service) GetByCustomer(ctx context.Context, bookingId int) (domain.PublicBooking, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	publicBooking, err := s.bookingRepo.GetPublicBooking(ctx, bookingId, userId)
	if err != nil {
		return domain.PublicBooking{}, err
	}

	return publicBooking, nil
}

// prevent null booking price to avoid a lot of headaches
func (s *Service) preventNilBookingPrice(ctx context.Context, merchantId uuid.UUID, price *currencyx.Price) (currencyx.Price, error) {
	var outPrice currencyx.Price

	if price == nil {
		curr, err := s.merchantRepo.GetMerchantCurrency(ctx, merchantId)
		if err != nil {
			return currencyx.Price{}, err
		}

		zeroAmount, err := currency.NewAmount("0", curr)
		if err != nil {
			return currencyx.Price{}, fmt.Errorf("error while creating new amount: %s", err.Error())
		}

		outPrice = currencyx.Price{Amount: zeroAmount}
	} else {
		outPrice = *price
	}

	return outPrice, nil
}

func parseRrule(rruleInput RecurringRuleInput, dStart time.Time) (*rrule.RRule, error) {
	var freq rrule.Frequency

	switch strings.ToUpper(rruleInput.Frequency) {
	case "DAILY":
		freq = rrule.DAILY
	case "WEEKLY":
		freq = rrule.WEEKLY
	case "MONTHLY":
		freq = rrule.MONTHLY
	default:
		return nil, fmt.Errorf("recurring rule frequency is invalid")
	}

	untilTimeStamp, err := time.Parse(time.RFC3339, rruleInput.Until)
	if err != nil {
		return nil, fmt.Errorf("until timestamp could not be converted to time: %s", err.Error())
	}

	var weekdays []rrule.Weekday

	for _, wkd := range rruleInput.Weekdays {
		switch strings.ToUpper(wkd) {
		case rrule.MO.String():
			weekdays = append(weekdays, rrule.MO)
		case rrule.TU.String():
			weekdays = append(weekdays, rrule.TU)
		case rrule.WE.String():
			weekdays = append(weekdays, rrule.WE)
		case rrule.TH.String():
			weekdays = append(weekdays, rrule.TH)
		case rrule.FR.String():
			weekdays = append(weekdays, rrule.FR)
		case rrule.SA.String():
			weekdays = append(weekdays, rrule.SA)
		case rrule.SU.String():
			weekdays = append(weekdays, rrule.SU)
		default:
			return nil, fmt.Errorf("incorrect weekday")
		}
	}

	rrule, err := rrule.NewRRule(rrule.ROption{
		Freq:      freq,
		Dtstart:   dStart,
		Interval:  rruleInput.Interval,
		Byweekday: weekdays,
		Until:     untilTimeStamp,
	})
	if err != nil {
		return nil, fmt.Errorf("error while creating rrule: %s", err.Error())
	}

	return rrule, nil
}

func (s *Service) scheduleNewBookingEmails(ctx context.Context, tx pgx.Tx, customers []uuid.UUID, statuses []types.BookingStatus, bookingId int, fromDate time.Time) error {
	assert.True(len(customers) == len(statuses), "customers and statuses length shall be the same", len(customers), len(statuses))

	reminderDate := fromDate.Add(-24 * time.Hour)

	var statusConfirmedParams []river.InsertManyParams
	var confirmationParams []river.InsertManyParams
	reminderParams := make([]river.InsertManyParams, len(customers))

	for i, customerId := range customers {
		if statuses[i] == types.BookingStatusConfirmed {
			statusConfirmedParams = append(statusConfirmedParams, river.InsertManyParams{
				Args: args.BookingConfirmationEmail{
					BookingId:  bookingId,
					CustomerId: customerId,
				},
			})
		} else {
			confirmationParams = append(confirmationParams, river.InsertManyParams{
				Args: args.BookingConfirmationEmail{
					BookingId:  bookingId,
					CustomerId: customerId,
				},
			})
		}

		reminderParams[i] = river.InsertManyParams{
			Args: args.BookingReminderEmail{
				BookingId:        bookingId,
				CustomerId:       customerId,
				ExpectedFromDate: fromDate,
			}, InsertOpts: &river.InsertOpts{
				ScheduledAt: reminderDate,
			},
		}
	}

	if len(confirmationParams) > 0 {
		_, err := s.enqueuer.InsertManyFastTx(ctx, tx, confirmationParams)
		if err != nil {
			return fmt.Errorf("could not schedule booking confirmation email job: %w", err)
		}
	}

	if len(statusConfirmedParams) > 0 {
		_, err := s.enqueuer.InsertManyFastTx(ctx, tx, statusConfirmedParams)
		if err != nil {
			return fmt.Errorf("could not schedule booking status confirmed email: %w", err)
		}
	}

	if len(reminderParams) > 0 {
		_, err := s.enqueuer.InsertManyFastTx(ctx, tx, reminderParams)
		if err != nil {
			return fmt.Errorf("could not schedule booking reminder email job: %w", err)
		}
	}

	return nil
}

type CreateByMerchantInput struct {
	Customers    []CustomerInput
	ServiceId    int
	EmployeeId   int
	TimeStamp    time.Time
	MerchantNote *string
	IsRecurring  bool
	Rrule        *RecurringRuleInput
}

type CustomerInput struct {
	CustomerId  *uuid.UUID
	FirstName   *string
	LastName    *string
	Email       *string
	PhoneNumber *string
}

type RecurringRuleInput struct {
	Frequency string
	Interval  int
	Weekdays  []string
	Until     string
}

func (s *Service) CreateByMerchant(ctx context.Context, input CreateByMerchantInput) error {
	actor := actor.MustGetFromContext(ctx)

	service, err := s.catalogRepo.GetServiceWithPhases(ctx, input.ServiceId, actor.MerchantId)
	if err != nil {
		return err
	}

	if !service.IsGroupService() && len(input.Customers) > 1 {
		return fmt.Errorf("appointments cannot have more than 1 customer")
	}

	if service.IsGroupService() && len(input.Customers) > service.MaxParticipants {
		return fmt.Errorf("customer count (%d) exceeds class limit of %d", len(input.Customers), service.MaxParticipants)
	}

	bookedLocation, err := s.merchantRepo.GetLocation(ctx, actor.LocationId, actor.MerchantId)
	if err != nil {
		return err
	}

	price, err := s.preventNilBookingPrice(ctx, actor.MerchantId, service.Price)
	if err != nil {
		return err
	}

	var incomingCustomerIds []uuid.UUID
	var participantCount int

	var totalPrice currency.Amount

	isWalkIn := len(input.Customers) == 0

	if !isWalkIn {
		customerIdMap, err := s.getParticipants(ctx, actor.MerchantId, input.Customers)
		if err != nil {
			return err
		}

		for id := range customerIdMap {
			incomingCustomerIds = append(incomingCustomerIds, id)
		}

		participantCount = len(incomingCustomerIds)

		totalPrice, err = price.Mul(strconv.Itoa(participantCount))
		if err != nil {
			return fmt.Errorf("failed to calculate total price: %w", err)
		}
	} else {
		// walk-ins do not get a booking participant row but 1 person still attending the booking
		// group bookings can't have walk-ins
		if service.IsGroupService() {
			participantCount = 0

			totalPrice, err = currency.NewAmount("0", price.CurrencyCode())
			if err != nil {
				return fmt.Errorf("failed to calculate total price: %w", err)
			}
		} else {
			participantCount = 1

			totalPrice = price.Amount
		}
	}

	fromDate := input.TimeStamp.UTC()

	duration := service.GetTotalDuration()

	toDate := fromDate.Add(duration)

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		var bookingSeriesId *int
		var seriesOriginalDate *time.Time
		var occurrenceIndex *int
		var seriesVersion *int

		if input.IsRecurring && input.Rrule != nil {
			merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, actor.MerchantId)
			if err != nil {
				return err
			}

			// recurring bookings have to be stored in local time and converted to UTC during generation
			fromDateMerchantTz := fromDate.In(merchantTz)

			rrule, err := parseRrule(*input.Rrule, fromDateMerchantTz)
			if err != nil {
				return err
			}

			series, err := s.bookingRepo.WithTx(tx).NewBookingSeries(ctx, domain.BookingSeries{
				BookingType:         service.BookingType,
				MerchantId:          actor.MerchantId,
				EmployeeId:          &input.EmployeeId,
				ServiceId:           &service.Id,
				LocationId:          bookedLocation.Id,
				Rrule:               rrule.String(),
				Dstart:              fromDateMerchantTz,
				Timezone:            merchantTz.String(),
				IsActive:            true,
				ServiceName:         service.Name,
				PricePerPerson:      price,
				TotalPrice:          currencyx.Price{Amount: totalPrice},
				PriceType:           service.PriceType,
				FormattedLocation:   bookedLocation.FormattedLocation,
				MinParticipants:     service.MinParticipants,
				MaxParticipants:     service.MaxParticipants,
				CurrentParticipants: participantCount,
			})
			if err != nil {
				return err
			}

			seriesParticipants := make([]domain.BookingSeriesParticipant, len(incomingCustomerIds))
			for i, id := range incomingCustomerIds {
				seriesParticipants[i] = domain.BookingSeriesParticipant{
					BookingSeriesId: series.Id,
					CustomerId:      &id,
					IsActive:        true,
				}
			}

			_, err = s.bookingRepo.WithTx(tx).NewBookingSeriesParticipants(ctx, seriesParticipants)
			if err != nil {
				return err
			}

			seriesPhases := make([]domain.BookingSeriesPhase, len(service.Phases))
			for i, p := range service.Phases {
				seriesPhases[i] = domain.BookingSeriesPhase{
					BookingSeriesId: series.Id,
					ServicePhaseId:  &p.Id,
					Name:            p.Name,
					Sequence:        p.Sequence,
					Duration:        p.Duration,
					PhaseType:       p.PhaseType,
				}
			}

			err = s.bookingRepo.WithTx(tx).NewBookingSeriesPhases(ctx, seriesPhases)
			if err != nil {
				return err
			}

			_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingOccurrenceGenerator{
				BookingSeriesId: series.Id,
				GenerateFrom:    toDate,
			}, &river.InsertOpts{})
			if err != nil {
				return fmt.Errorf("error scheduling recurring booking generation: %w", err)
			}

			one := 1

			bookingSeriesId = &series.Id
			seriesOriginalDate = &fromDate
			occurrenceIndex = &one
			seriesVersion = &one
		}

		booking := domain.Booking{
			Status:              types.BookingStatusConfirmed,
			BookingType:         service.BookingType,
			IsRecurring:         input.IsRecurring,
			MerchantId:          actor.MerchantId,
			EmployeeId:          &input.EmployeeId,
			ServiceId:           &service.Id,
			LocationId:          bookedLocation.Id,
			BookingSeriesId:     bookingSeriesId,
			SeriesOriginalDate:  seriesOriginalDate,
			FromDate:            fromDate,
			ToDate:              toDate,
			ServiceName:         service.Name,
			PricePerPerson:      price,
			TotalPrice:          currencyx.Price{Amount: totalPrice},
			PriceType:           service.PriceType,
			FormattedLocation:   bookedLocation.FormattedLocation,
			MerchantNote:        input.MerchantNote,
			MinParticipants:     service.MinParticipants,
			MaxParticipants:     service.MaxParticipants,
			CurrentParticipants: participantCount,
			OccurrenceIndex:     occurrenceIndex,
			SeriesVersion:       seriesVersion,
		}

		participants := make([]domain.BookingParticipant, len(incomingCustomerIds))
		for i, id := range incomingCustomerIds {
			participants[i] = domain.BookingParticipant{
				Status:       types.BookingStatusConfirmed,
				CustomerId:   &id,
				CustomerNote: nil,
			}
		}

		bookingId, err := s.newBooking(ctx, tx, booking, participants, service)
		if err != nil {
			return fmt.Errorf("error during new booking creation: %s", err.Error())
		}

		if !isWalkIn {
			statuses := utils.RepeatSlice([]types.BookingStatus{types.BookingStatusConfirmed}, len(incomingCustomerIds))

			err = s.scheduleNewBookingEmails(ctx, tx, incomingCustomerIds, statuses, bookingId, fromDate)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// TODO: rename to something more expressive
// also I do not like that there is uuid generation and db insert in the middle of this
// also this should be in a transaction
func (s *Service) getParticipants(ctx context.Context, merchantId uuid.UUID, customers []CustomerInput) (map[uuid.UUID]struct{}, error) {
	customerIds := make(map[uuid.UUID]struct{})

	for _, c := range customers {
		if c.CustomerId != nil {
			customerIds[*c.CustomerId] = struct{}{}

		} else if c.FirstName != nil && c.LastName != nil {

			newId, err := uuid.NewV7()
			if err != nil {
				return map[uuid.UUID]struct{}{}, fmt.Errorf("error generating customer id: %s", err.Error())
			}

			if err := s.customerRepo.NewCustomer(ctx, merchantId, domain.Customer{
				Id:          newId,
				FirstName:   c.FirstName,
				LastName:    c.LastName,
				Email:       c.Email,
				PhoneNumber: c.PhoneNumber,
			}); err != nil {
				return map[uuid.UUID]struct{}{}, err
			}

			customerIds[newId] = struct{}{}
		}
	}

	return customerIds, nil
}

type participantChanges struct {
	ToInsert []uuid.UUID
	ToDelete []uuid.UUID
	ToKeep   []uuid.UUID
}

func detectParticipantChanges(existing []domain.BookingParticipant, incomingCustomerIds map[uuid.UUID]struct{}) (participantChanges, error) {
	var pc participantChanges

	existingMap := make(map[uuid.UUID]struct{}, len(existing))
	for _, p := range existing {
		// customerId can be nil in case of walk-ins
		if p.CustomerId != nil {
			existingMap[*p.CustomerId] = struct{}{}
		}
	}

	for id := range incomingCustomerIds {
		if _, ok := existingMap[id]; !ok {
			pc.ToInsert = append(pc.ToInsert, id)
		}
	}

	for id := range existingMap {
		if _, ok := incomingCustomerIds[id]; !ok {
			pc.ToDelete = append(pc.ToDelete, id)
		} else {
			pc.ToKeep = append(pc.ToKeep, id)
		}
	}

	return pc, nil
}

type seriesParticipantChanges struct {
	ToInsert []uuid.UUID
	ToDelete []uuid.UUID
	ToKeep   []uuid.UUID
}

func detectSeriesParticipantChanges(existing []domain.BookingParticipant, existingSeries []domain.BookingSeriesParticipant, incomingCustomerIds map[uuid.UUID]struct{}) (seriesParticipantChanges, error) {
	var spc seriesParticipantChanges

	existingMap := make(map[uuid.UUID]struct{}, len(existing))
	for _, p := range existing {
		// customerId can be nil in case of walk-ins
		if p.CustomerId != nil {
			existingMap[*p.CustomerId] = struct{}{}
		}
	}

	existingSeriesMap := make(map[uuid.UUID]struct{}, len(existingSeries))
	for _, sp := range existingSeries {
		if sp.CustomerId != nil {
			existingSeriesMap[*sp.CustomerId] = struct{}{}
		}
	}

	for id := range incomingCustomerIds {
		_, inCurrent := existingMap[id]
		_, inSeries := existingSeriesMap[id]

		if !inCurrent && !inSeries {
			spc.ToInsert = append(spc.ToInsert, id)
		}

		if inSeries {
			spc.ToKeep = append(spc.ToKeep, id)
		}
	}

	for id := range existingSeriesMap {
		if _, exists := incomingCustomerIds[id]; !exists {
			spc.ToDelete = append(spc.ToDelete, id)
		}
	}

	return spc, nil
}

type UpdateByMerchantInput struct {
	Customers       []CustomerInput
	TimeStamp       time.Time
	EmployeeId      int
	MerchantNote    *string
	BookingStatus   types.BookingStatus
	UpdateAllFuture bool
}

func (s *Service) UpdateByMerchant(ctx context.Context, bookingId int, input UpdateByMerchantInput) error {
	actor := actor.MustGetFromContext(ctx)

	booking, err := s.bookingRepo.GetBooking(ctx, bookingId)
	if err != nil {
		return err
	}

	if !booking.IsOwnedByMerchant(actor.MerchantId) {
		return fmt.Errorf("booking not found for this merchant")
	}

	if input.UpdateAllFuture && !booking.IsRecurring {
		return fmt.Errorf("cannot update future occurrences of non-recurring booking")
	}

	err = booking.CanModify()
	if err != nil {
		return err
	}

	participants, err := s.bookingRepo.GetBookingParticipants(ctx, booking.Id)
	if err != nil {
		return err
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, actor.MerchantId)
	if err != nil {
		return err
	}

	isGroupBooking := booking.IsGroupBooking()

	var participantCount int
	var seriesParticipantCount int
	var participantChanges participantChanges
	var seriesParticipantChanges seriesParticipantChanges
	var seriesParticipants []domain.BookingSeriesParticipant

	isWalkIn := len(input.Customers) == 0

	if !isWalkIn || (isWalkIn && len(participants) != 0) {
		incomingCustomerIds, err := s.getParticipants(ctx, actor.MerchantId, input.Customers)
		if err != nil {
			return err
		}

		participantChanges, err = detectParticipantChanges(participants, incomingCustomerIds)
		if err != nil {
			return err
		}

		participantCount = len(participantChanges.ToInsert) + len(participantChanges.ToKeep)

		if input.UpdateAllFuture && booking.IsRecurring {
			seriesParticipants, err = s.bookingRepo.GetBookingSeriesParticipants(ctx, *booking.BookingSeriesId)
			if err != nil {
				return err
			}

			seriesParticipantChanges, err = detectSeriesParticipantChanges(participants, seriesParticipants, incomingCustomerIds)
			if err != nil {
				return err
			}

			seriesParticipantCount = len(seriesParticipantChanges.ToInsert) + len(seriesParticipantChanges.ToKeep)
		}
	}

	if isWalkIn {
		if isGroupBooking {
			participantCount = 0
			seriesParticipantCount = 0
		} else {
			participantCount = 1
			seriesParticipantCount = 1
		}
	}

	participantsChanged := len(participantChanges.ToInsert) != 0 || len(participantChanges.ToDelete) != 0
	seriesParticipantsChanged := len(seriesParticipantChanges.ToInsert) != 0 || len(seriesParticipantChanges.ToDelete) != 0

	if participantsChanged {
		if participantCount > booking.MaxParticipants {
			return fmt.Errorf("participant count (%d) cannot be higher than maximum (%d)", participantCount, booking.MaxParticipants)
		}
	}

	if seriesParticipantsChanged {
		if seriesParticipantCount > booking.MaxParticipants {
			return fmt.Errorf("participant count (%d) cannot be higher than maximum (%d)", seriesParticipantCount, booking.MaxParticipants)
		}
	}

	// TODO: priceChanged should only indicate if pricePerPerson changed. Change this once changing price is introduced
	priceChanged := participantsChanged

	var pricePerPerson currencyx.Price
	var totalPrice, seriesTotalPrice currency.Amount

	if priceChanged {
		pricePerPerson = booking.PricePerPerson

		if isWalkIn && isGroupBooking {
			totalPrice, err = currency.NewAmount("0", pricePerPerson.CurrencyCode())
			if err != nil {
				return fmt.Errorf("failed to calculate total price: %s", err.Error())
			}
		} else {
			countStr := strconv.Itoa(participantCount)

			totalPrice, err = pricePerPerson.Mul(countStr)
			if err != nil {
				return fmt.Errorf("failed to calculate total price: %s", err.Error())
			}
		}

		if input.UpdateAllFuture && booking.IsRecurring {
			if isWalkIn && isGroupBooking {
				seriesTotalPrice, err = currency.NewAmount("0", pricePerPerson.CurrencyCode())
				if err != nil {
					return fmt.Errorf("failed to calculate series total price: %s", err.Error())
				}
			} else {
				seriesCountStr := strconv.Itoa(seriesParticipantCount)

				seriesTotalPrice, err = pricePerPerson.Mul(seriesCountStr)
				if err != nil {
					return fmt.Errorf("failed to calculate series total price: %s", err.Error())
				}
			}
		}
	} else {
		pricePerPerson = booking.PricePerPerson
		totalPrice = booking.TotalPrice.Amount
	}

	timeStampChanged := !booking.FromDate.Equal(input.TimeStamp.UTC())
	merchantNoteChanged := booking.MerchantNote != input.MerchantNote
	statusChanged := booking.Status != input.BookingStatus
	employeeChanged := booking.EmployeeId != &input.EmployeeId

	fromDate := booking.FromDate
	toDate := booking.ToDate
	timestampOffset := time.Duration(0)
	seriesOriginalDateOffset := time.Duration(0)

	seriesFromDate := booking.FromDate.In(merchantTz)

	if timeStampChanged {
		timestampOffset = input.TimeStamp.UTC().Sub(booking.FromDate)

		fromDate = fromDate.Add(timestampOffset)
		toDate = toDate.Add(timestampOffset)

		if booking.IsRecurring && booking.SeriesOriginalDate != nil {
			// For a recurring bookings we cannot use an offset as individual bookings which are part of the series
			// could have been modified. Also it has to be in the merchant's timezone to avoid DST problems
			seriesFromDate = input.TimeStamp.In(merchantTz)
			seriesOriginalDateOffset = input.TimeStamp.UTC().Sub(*booking.SeriesOriginalDate)
		}
	}

	merchantNote := booking.MerchantNote

	if merchantNoteChanged {
		merchantNote = input.MerchantNote
	}

	bookingStatus := booking.Status

	if statusChanged {
		err = booking.CanTransition(input.BookingStatus)
		if err != nil {
			return err
		}

		bookingStatus = input.BookingStatus
	}

	participantBookingStatus := bookingStatus
	// If it's a group booking the participant status can be managed separately,
	// we deault to confirmed as the merchant is adding them to the booking
	if isGroupBooking {
		participantBookingStatus = types.BookingStatusConfirmed
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		if input.UpdateAllFuture && booking.IsRecurring {
			bookingSeries, err := s.bookingRepo.WithTx(tx).GetBookingSeries(ctx, *booking.BookingSeriesId)
			if err != nil {
				return err
			}

			if !priceChanged {
				seriesTotalPrice = bookingSeries.TotalPrice.Amount
			}

			if timeStampChanged {
				rrule, err := rrule.StrToRRule(bookingSeries.Rrule)
				if err != nil {
					return fmt.Errorf("failed to parse existing rrule: %w", err)
				}

				rrule.DTStart(seriesFromDate)
				// adjusting the until time is needed because a large timestamp offset might
				// make the last occurrences fall out of the rrule, causing occurrence calulcation problems
				rrule.Until(rrule.GetUntil().Add(seriesOriginalDateOffset))

				seriesVersion, err := s.bookingRepo.WithTx(tx).UpdateBookingSeriesRrule(ctx, bookingSeries.Id, rrule.String(), seriesFromDate)
				if err != nil {
					return err
				}

				err = s.bookingRepo.WithTx(tx).UpdateBookingSeriesOriginalDateAndVersion(ctx, booking.Id, fromDate, seriesVersion)
				if err != nil {
					return err
				}
			}

			if seriesParticipantsChanged || priceChanged {
				err = s.bookingRepo.WithTx(tx).UpdateBookingSeriesDetails(ctx, bookingSeries.Id, domain.BookingDetails{
					PricePerPerson:      pricePerPerson,
					TotalPrice:          currencyx.Price{Amount: seriesTotalPrice},
					MinParticipants:     booking.MinParticipants,
					MaxParticipants:     booking.MaxParticipants,
					CurrentParticipants: seriesParticipantCount,
				})
				if err != nil {
					return err
				}
			}

			if seriesParticipantsChanged {
				if len(seriesParticipantChanges.ToDelete) > 0 {
					err = s.bookingRepo.WithTx(tx).DeleteBookingSeriesParticipants(ctx, bookingSeries.Id, seriesParticipantChanges.ToDelete)
					if err != nil {
						return err
					}
				}

				if len(seriesParticipantChanges.ToInsert) > 0 {
					var seriesParticipantsInsert []domain.BookingSeriesParticipant

					for _, cid := range seriesParticipantChanges.ToInsert {
						seriesParticipantsInsert = append(seriesParticipantsInsert, domain.BookingSeriesParticipant{
							BookingSeriesId: bookingSeries.Id,
							CustomerId:      &cid,
							IsActive:        true,
						})
					}

					_, err = s.bookingRepo.WithTx(tx).NewBookingSeriesParticipants(ctx, seriesParticipantsInsert)
					if err != nil {
						return err
					}
				}
			}

			_, err = s.enqueuer.InsertTx(ctx, tx, args.UpdateFutureBookingOccurrences{
				BookingSeriesId:          bookingSeries.Id,
				OccurrenceIndex:          *booking.OccurrenceIndex,
				SeriesOriginalDateOffset: seriesOriginalDateOffset,
				PriceChanged:             priceChanged,
				// series cancellation is handled by CancelByMerchant()
				StatusChangedToCancelled: false,
				CancellationReason:       "",
				EmployeeChanged:          employeeChanged,
				ParticipantsToInsert:     seriesParticipantChanges.ToInsert,
				ParticipantsToDelete:     seriesParticipantChanges.ToDelete,
				ParticipantsBefore:       seriesParticipants,
			}, &river.InsertOpts{})
			if err != nil {
				return fmt.Errorf("failed to insert update future occurrences job: %w", err)
			}
		}

		if timeStampChanged || statusChanged || merchantNoteChanged || employeeChanged {
			err = s.bookingRepo.WithTx(tx).UpdateBookingCoreBatch(ctx, actor.MerchantId, []int{booking.Id}, booking.ServiceId,
				&input.EmployeeId, []time.Time{fromDate}, []time.Time{toDate}, booking.BookingType, bookingStatus, merchantNote)
			if err != nil {
				return err
			}
		}

		if priceChanged || participantsChanged {
			err = s.bookingRepo.WithTx(tx).UpdateBookingDetailsBatch(ctx, actor.MerchantId, []int{booking.Id}, []domain.BookingDetails{{
				PricePerPerson:      pricePerPerson,
				TotalPrice:          currencyx.Price{Amount: totalPrice},
				MinParticipants:     booking.MinParticipants,
				MaxParticipants:     booking.MaxParticipants,
				CurrentParticipants: participantCount,
			}})
			if err != nil {
				return err
			}
		}

		if timeStampChanged {
			bookingPhases, err := s.bookingRepo.WithTx(tx).GetBookingPhases(ctx, booking.Id)
			if err != nil {
				return err
			}

			bookingPhasesToUpdate := make([]domain.BookingPhase, len(bookingPhases))
			for i, bp := range bookingPhases {
				bookingPhasesToUpdate[i] = domain.BookingPhase{
					Id:             bp.Id,
					BookingId:      bp.BookingId,
					ServicePhaseId: bp.ServicePhaseId,
					FromDate:       bp.FromDate.Add(timestampOffset),
					ToDate:         bp.ToDate.Add(timestampOffset),
					PhaseType:      bp.PhaseType,
				}
			}

			err = s.bookingRepo.WithTx(tx).UpdateBookingPhasesBatch(ctx, bookingPhasesToUpdate)
			if err != nil {
				return err
			}
		}

		var participantsToUpsert []domain.BookingParticipant

		if participantsChanged {
			if len(participantChanges.ToDelete) > 0 {
				err := s.bookingRepo.WithTx(tx).DeleteBookingParticipantsBatch(ctx, []int{booking.Id}, participantChanges.ToDelete)
				if err != nil {
					return err
				}
			}

			for _, cid := range participantChanges.ToInsert {
				participantsToUpsert = append(participantsToUpsert, domain.BookingParticipant{
					BookingId:  booking.Id,
					CustomerId: &cid,
					Status:     participantBookingStatus,
				})
			}
		}

		// overwrite the participant status to match the new booking status in case of appointments
		// where the statuses match eachother
		if !isGroupBooking && statusChanged {
			for _, cid := range participantChanges.ToKeep {
				participantsToUpsert = append(participantsToUpsert, domain.BookingParticipant{
					BookingId:  booking.Id,
					CustomerId: &cid,
					Status:     participantBookingStatus,
				})
			}
		}

		if participantsChanged || (!isGroupBooking && statusChanged) {
			if len(participantsToUpsert) > 0 {
				err := s.bookingRepo.WithTx(tx).UpdateBookingParticipants(ctx, participantsToUpsert, true)
				if err != nil {
					return err
				}
			}
		}

		if timeStampChanged {
			// TODO: don't forget to change this when we will consider employee changes in this
			if booking.EmployeeId != nil {
				_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncUpdateBooking{
					BookingId: booking.Id,
				}, nil)
				if err != nil {
					return err
				}
			}
		}

		if participantsChanged {
			for _, id := range participantChanges.ToDelete {
				// TODO: send a modification email for the entire series for series participants if recurring
				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingCancellationEmail{
					BookingId:          booking.Id,
					CustomerId:         id,
					CancellationReason: "",
				}, nil)
				if err != nil {
					return fmt.Errorf("could not schedule booking cancellation email job: %w", err)
				}
			}

			if len(participantChanges.ToInsert) > 0 {
				statuses := utils.RepeatSlice([]types.BookingStatus{participantBookingStatus}, len(participantChanges.ToInsert))

				err := s.scheduleNewBookingEmails(ctx, tx, participantChanges.ToInsert, statuses, booking.Id, fromDate)
				if err != nil {
					return err
				}
			}
		}

		fromDateMerchantTz := fromDate.In(merchantTz)
		reminderDate := fromDateMerchantTz.Add(-24 * time.Hour)

		statusChangedToConfirmed := !isGroupBooking &&
			booking.Status == types.BookingStatusBooked &&
			input.BookingStatus == types.BookingStatusConfirmed

		for _, id := range participantChanges.ToKeep {
			if statusChangedToConfirmed {
				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingStatusConfirmedEmail{
					BookingId:  booking.Id,
					CustomerId: id,
				}, nil)
				if err != nil {
					return fmt.Errorf("could not schedule booking confirmation email job: %w", err)
				}
			}

			if timeStampChanged {
				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingReminderEmail{
					BookingId:        booking.Id,
					CustomerId:       id,
					ExpectedFromDate: fromDate,
				}, &river.InsertOpts{
					ScheduledAt: reminderDate,
				})
				if err != nil {
					return fmt.Errorf("could not schedule booking reminder email job: %w", err)
				}

				// TODO: send a modification email for the entire series for series participants if recurring
				// TODO: we should send a modification when the price is changed, but the email does not handle it currently
				// should be revisited once changing price and name is allowed
				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingModificationEmail{
					BookingId:      booking.Id,
					CustomerId:     id,
					OldServiceName: booking.ServiceName,
					OldFromDate:    booking.FromDate,
					OldToDate:      booking.ToDate,
				}, nil)
				if err != nil {
					return fmt.Errorf("could not schedule booking modification email job: %w", err)
				}
			}
		}

		return nil
	})
}

type CancelByMerchantInput struct {
	CancellationReason string
	CancelFuture       bool
}

// TODO: what should the booking participant statuses be here?
func (s *Service) CancelByMerchant(ctx context.Context, bookingId int, input CancelByMerchantInput) error {
	actor := actor.MustGetFromContext(ctx)

	booking, err := s.bookingRepo.GetBooking(ctx, bookingId)
	if err != nil {
		return err
	}

	if input.CancelFuture && !booking.IsRecurring {
		return fmt.Errorf("cannot cancel future occurrences of non-recurring booking")
	}

	err = booking.CanCancel()
	if err != nil {
		return err
	}

	bookingParticipants, err := s.bookingRepo.GetBookingParticipants(ctx, booking.Id)
	if err != nil {
		return err
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err = s.bookingRepo.WithTx(tx).CancelBookingByMerchant(ctx, actor.MerchantId, booking.Id, input.CancellationReason)
		if err != nil {
			return err
		}

		if input.CancelFuture {
			seriesParticipants, err := s.bookingRepo.WithTx(tx).GetBookingSeriesParticipants(ctx, *booking.BookingSeriesId)
			if err != nil {
				return err
			}

			_, err = s.enqueuer.InsertTx(ctx, tx, args.UpdateFutureBookingOccurrences{
				BookingSeriesId:          *booking.BookingSeriesId,
				OccurrenceIndex:          *booking.OccurrenceIndex,
				SeriesOriginalDateOffset: time.Duration(0),
				PriceChanged:             false,
				StatusChangedToCancelled: true,
				CancellationReason:       input.CancellationReason,
				ParticipantsToInsert:     nil,
				ParticipantsToDelete:     nil,
				ParticipantsBefore:       seriesParticipants,
			}, &river.InsertOpts{})
			if err != nil {
				return fmt.Errorf("failed to insert update future occurrences job: %w", err)
			}
		}

		for _, participant := range bookingParticipants {
			// if not walk-in
			if participant.CustomerId != nil {
				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingCancellationEmail{
					BookingId:          booking.Id,
					CancellationReason: input.CancellationReason,
					CustomerId:         *participant.CustomerId,
				}, nil)
				if err != nil {
					return fmt.Errorf("could not schedule booking cancellation email job: %w", err)
				}
			}
		}

		if booking.EmployeeId != nil {
			_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncDeleteBooking{
				BookingId: booking.Id,
			}, nil)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

type UpdatePaticipantStatusInput struct {
	Status types.BookingStatus
}

func (s *Service) UpdateParticipantStatus(ctx context.Context, bookingId int, participantId int, input UpdatePaticipantStatusInput) error {
	actor := actor.MustGetFromContext(ctx)

	booking, err := s.bookingRepo.GetBooking(ctx, bookingId)
	if err != nil {
		return err
	}

	if booking.IsOwnedByMerchant(actor.MerchantId) {
		return fmt.Errorf("booking could not be found for this merchant")
	}

	err = booking.CanModify()
	if err != nil {
		return err
	}

	bookingParticipant, err := s.bookingRepo.GetBookingParticipant(ctx, participantId)
	if err != nil {
		return err
	}

	err = bookingParticipant.CanTransition(input.Status)
	if err != nil {
		return err
	}

	err = s.bookingRepo.UpdateParticipantStatus(ctx, bookingId, participantId, input.Status)
	if err != nil {
		return err
	}

	return nil
}
