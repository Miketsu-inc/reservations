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
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/lang"
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

func (s *Service) newBooking(ctx context.Context, tx db.DBTX, booking domain.Booking, participants []domain.BookingParticipant,
	servicePhases []domain.PublicServicePhase) (int, error) {
	var bookingId int
	var err error

	bookingId, err = s.bookingRepo.WithTx(tx).NewBooking(ctx, booking)
	if err != nil {
		return 0, err
	}

	bookingPhases := make([]domain.BookingPhase, len(servicePhases))

	bookingStart := booking.FromDate
	for i, phase := range servicePhases {
		phaseDuration := time.Duration(phase.Duration) * time.Minute
		bookingEnd := bookingStart.Add(phaseDuration)

		bookingPhases[i] = domain.BookingPhase{
			BookingId:      bookingId,
			ServicePhaseId: phase.Id,
			FromDate:       bookingStart,
			ToDate:         bookingEnd,
		}

		bookingStart = bookingEnd
	}

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
		return fmt.Errorf("error while searching merchant by this name: %w", err)
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, merchantId)
	if err != nil {
		return fmt.Errorf("error while getting merchant's timezone: %w", err)
	}

	service, err := s.catalogRepo.GetServiceWithPhases(ctx, input.ServiceId, merchantId)
	if err != nil {
		return fmt.Errorf("error while searching service by this id: %w", err)
	}

	bookingSettings, err := s.merchantRepo.GetBookingSettingsByMerchantAndService(ctx, merchantId, service.Id)
	if err != nil {
		return fmt.Errorf("error while getting booking settings for merchant %w", err)
	}

	fromDate := input.TimeStamp.UTC()

	duration := time.Duration(service.TotalDuration) * time.Minute

	toDate := fromDate.Add(duration)

	err = enforceBookingWindow(fromDate, time.Now().In(merchantTz), bookingSettings.BookingWindowMin, bookingSettings.BookingWindowMax)
	if err != nil {
		return err
	}

	// TODO: we should probably just check by querying the user
	customerId, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("unexpected error during creating customer id: %w", err)
	}

	price, cost, err := s.preventNilBookingPrice(ctx, merchantId, service.Price, service.Cost)
	if err != nil {
		return err
	}

	isGroupBooking := input.BookingId != nil
	var bookingId int

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

		// TODO: if it's a group booking we should probably query for the booking
		// to see if the actions can be performed
		if isGroupBooking {
			bookingId = *input.BookingId

			totalPrice, totalCost, err := s.bookingRepo.WithTx(tx).IncrementParticipantCount(ctx, bookingId)
			if err != nil {
				return err
			}

			newTotalPrice, err := totalPrice.Add(price.Amount)
			if err != nil {
				return err
			}

			newTotalCost, err := totalCost.Add(cost.Amount)
			if err != nil {
				return err
			}

			err = s.bookingRepo.WithTx(tx).UpdateBookingTotalPrice(ctx, bookingId, currencyx.Price{Amount: newTotalPrice}, currencyx.Price{Amount: newTotalCost})
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
			booking := domain.Booking{
				Status:              bookingStatus,
				BookingType:         types.BookingTypeAppointment,
				MerchantId:          merchantId,
				ServiceId:           input.ServiceId,
				LocationId:          input.LocationId,
				FromDate:            fromDate,
				ToDate:              toDate,
				PricePerPerson:      price,
				CostPerPerson:       cost,
				TotalPrice:          price,
				TotalCost:           cost,
				MinParticipants:     1,
				MaxParticipants:     1,
				CurrentParticipants: 1,
			}

			participants := []domain.BookingParticipant{{
				Status:       bookingStatus,
				CustomerId:   &customerId,
				CustomerNote: &input.CustomerNote,
			}}

			bookingId, err = s.newBooking(ctx, tx, booking, participants, service.Phases)
			if err != nil {
				return err
			}

			// TODO: this is a no-op now as the customer cannot choose an employee
			_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncNewBooking{
				BookingId: bookingId,
			}, nil)
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
		return fmt.Errorf("error while retrieving booking: %s", err.Error())
	}

	bookingParticipant, err := s.bookingRepo.GetBookingParticipantByUser(ctx, booking.Id, userId)
	if err != nil {
		return fmt.Errorf("error while retrieving booking participant: %s", err.Error())
	}

	cancelDeadline, err := s.catalogRepo.GetServiceCancelDeadline(ctx, booking.MerchantId, booking.ServiceId)
	if err != nil {
		return fmt.Errorf("error while retrieving cancel deadline: %s", err.Error())
	}

	latestCancelTime := booking.FromDate.Add(-time.Duration(cancelDeadline) * time.Minute)

	err = booking.CanCancelWithDeadline(latestCancelTime)
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
			return fmt.Errorf("error while cancelling the booking by user: %s", err.Error())
		}

		if booking.IsGroupBooking() {
			newTotalPrice, err := booking.TotalPrice.Sub(booking.PricePerPerson.Amount)
			if err != nil {
				return fmt.Errorf("failed to calculate total price: %w", err)
			}

			newTotalCost, err := booking.TotalCost.Sub(booking.CostPerPerson.Amount)
			if err != nil {
				return fmt.Errorf("failed to calculate total cost: %w", err)
			}

			err = s.bookingRepo.WithTx(tx).UpdateBookingTotalPrice(ctx, booking.Id, currencyx.Price{Amount: newTotalPrice}, currencyx.Price{Amount: newTotalCost})
			if err != nil {
				return err
			}

			err = s.bookingRepo.WithTx(tx).DecrementParticipantCount(ctx, booking.Id)
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

// prevent null booking price and cost to avoid a lot of headaches
func (s *Service) preventNilBookingPrice(ctx context.Context, merchantId uuid.UUID, price, cost *currencyx.Price) (currencyx.Price, currencyx.Price, error) {
	var outPrice currencyx.Price
	var outCost currencyx.Price

	if price == nil || cost == nil {
		curr, err := s.merchantRepo.GetMerchantCurrency(ctx, merchantId)
		if err != nil {
			return currencyx.Price{}, currencyx.Price{}, fmt.Errorf("error while getting merchant's currency: %s", err.Error())
		}

		zeroAmount, err := currency.NewAmount("0", curr)
		if err != nil {
			return currencyx.Price{}, currencyx.Price{}, fmt.Errorf("error while creating new amount: %s", err.Error())
		}

		if price != nil {
			outPrice = *price
		} else {
			outPrice = currencyx.Price{Amount: zeroAmount}
		}

		if cost != nil {
			outCost = *cost
		} else {
			outCost = currencyx.Price{Amount: zeroAmount}
		}
	} else {
		outPrice = *price
		outCost = *cost
	}

	return outPrice, outCost, nil
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

	lang := lang.LangFromContext(ctx)

	reminderDate := fromDate.Add(-24 * time.Hour)

	var statusConfirmedParams []river.InsertManyParams
	var confirmationParams []river.InsertManyParams
	reminderParams := make([]river.InsertManyParams, len(customers))

	for i, customerId := range customers {
		if statuses[i] == types.BookingStatusConfirmed {
			statusConfirmedParams = append(statusConfirmedParams, river.InsertManyParams{
				Args: args.BookingConfirmationEmail{
					Language:   lang,
					BookingId:  bookingId,
					CustomerId: customerId,
				},
			})
		} else {
			confirmationParams = append(confirmationParams, river.InsertManyParams{
				Args: args.BookingConfirmationEmail{
					Language:   lang,
					BookingId:  bookingId,
					CustomerId: customerId,
				},
			})
		}

		reminderParams[i] = river.InsertManyParams{
			Args: args.BookingReminderEmail{
				Language:         lang,
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
		return fmt.Errorf("error while searching service by this id: %s", err.Error())
	}

	if service.BookingType == types.BookingTypeAppointment && len(input.Customers) > 1 {
		return fmt.Errorf("appointments cannot have more than 1 customer")
	}

	if service.BookingType != types.BookingTypeAppointment && len(input.Customers) > service.MaxParticipants {
		return fmt.Errorf("customer count (%d) exceeds class limit of %d", len(input.Customers), service.MaxParticipants)
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, actor.MerchantId)
	if err != nil {
		return fmt.Errorf("error while getting merchant's timezone: %s", err.Error())
	}

	bookedLocation, err := s.merchantRepo.GetLocation(ctx, actor.LocationId, actor.MerchantId)
	if err != nil {
		return fmt.Errorf("error while searching location by this id: %s", err.Error())
	}

	price, cost, err := s.preventNilBookingPrice(ctx, actor.MerchantId, service.Price, service.Cost)
	if err != nil {
		return err
	}

	var incomingCustomerIds []uuid.UUID

	isWalkIn := len(input.Customers) == 0

	if !isWalkIn {
		customerIdMap, err := s.getParticipants(ctx, actor.MerchantId, input.Customers)
		if err != nil {
			return err
		}

		incomingCustomerIds = make([]uuid.UUID, len(customerIdMap))
		for id := range customerIdMap {
			incomingCustomerIds = append(incomingCustomerIds, id)
		}
	}

	participantCount := len(incomingCustomerIds)
	// walk ins do not get a booking participant row but 1 person still attending the booking
	if isWalkIn {
		participantCount = 1
	}

	countStr := strconv.Itoa(participantCount)

	totalPrice, err := price.Mul(countStr)
	if err != nil {
		return fmt.Errorf("failed to calculate total price: %w", err)
	}

	totalCost, err := cost.Mul(countStr)
	if err != nil {
		return fmt.Errorf("failed to calculate total cost: %w", err)
	}

	fromDate := input.TimeStamp.UTC()

	duration := time.Duration(service.TotalDuration) * time.Minute

	toDate := fromDate.Add(duration)

	var bookingId int

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		if input.IsRecurring && input.Rrule != nil {
			// recurring bookings have to be stored in local time and converted to UTC during generation
			fromDateMerchantTz := fromDate.In(merchantTz)

			rrule, err := parseRrule(*input.Rrule, fromDateMerchantTz)
			if err != nil {
				return err
			}

			series, err := s.bookingRepo.WithTx(tx).NewBookingSeries(ctx, domain.BookingSeries{
				BookingType:         service.BookingType,
				MerchantId:          actor.MerchantId,
				EmployeeId:          actor.EmployeeId,
				ServiceId:           service.Id,
				LocationId:          bookedLocation.Id,
				Rrule:               rrule.String(),
				Dstart:              fromDateMerchantTz,
				Timezone:            merchantTz.String(),
				IsActive:            true,
				PricePerPerson:      price,
				CostPerPerson:       cost,
				TotalPrice:          currencyx.Price{Amount: totalPrice},
				TotalCost:           currencyx.Price{Amount: totalCost},
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

			seriesParticipants, err = s.bookingRepo.WithTx(tx).NewBookingSeriesParticipants(ctx, seriesParticipants)
			if err != nil {
				return err
			}

			bookingId, err = s.GenerateRecurringBookings(ctx, tx, series, seriesParticipants, service.Phases, time.Now().UTC())
			if err != nil {
				return fmt.Errorf("error while generating recurring bookings: %s", err.Error())
			}
		} else {
			booking := domain.Booking{
				Status:              types.BookingStatusConfirmed,
				BookingType:         service.BookingType,
				MerchantId:          actor.MerchantId,
				EmployeeId:          &actor.EmployeeId,
				ServiceId:           service.Id,
				LocationId:          bookedLocation.Id,
				FromDate:            fromDate,
				ToDate:              toDate,
				PricePerPerson:      price,
				CostPerPerson:       cost,
				TotalPrice:          currencyx.Price{Amount: totalPrice},
				TotalCost:           currencyx.Price{Amount: totalCost},
				MerchantNote:        input.MerchantNote,
				MinParticipants:     service.MinParticipants,
				MaxParticipants:     service.MaxParticipants,
				CurrentParticipants: participantCount,
			}

			participants := make([]domain.BookingParticipant, len(incomingCustomerIds))
			for i, id := range incomingCustomerIds {
				participants[i] = domain.BookingParticipant{
					Status:       types.BookingStatusConfirmed,
					BookingId:    bookingId,
					CustomerId:   &id,
					CustomerNote: nil,
				}
			}

			bookingId, err = s.newBooking(ctx, tx, booking, participants, service.Phases)
			if err != nil {
				return fmt.Errorf("error during new booking creation: %s", err.Error())
			}

			if booking.EmployeeId != nil {
				_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncNewBooking{
					BookingId: bookingId,
				}, nil)
				if err != nil {
					return err
				}
			}
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
				return map[uuid.UUID]struct{}{}, fmt.Errorf("error inserting new customer: %s", err.Error())
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
	ServiceId       int
	TimeStamp       time.Time
	MerchantNote    *string
	BookingStatus   types.BookingStatus
	UpdateAllFuture bool
}

func (s *Service) UpdateByMerchant(ctx context.Context, bookingId int, input UpdateByMerchantInput) error {
	actor := actor.MustGetFromContext(ctx)

	booking, err := s.bookingRepo.GetBooking(ctx, bookingId)
	if err != nil {
		return fmt.Errorf("error while retrieving data for email sending: %s", err.Error())
	}

	err = booking.CanModify()
	if err != nil {
		return err
	}

	participants, err := s.bookingRepo.GetBookingParticipants(ctx, booking.Id)
	if err != nil {
		return fmt.Errorf("error getting booking participants: %w", err)
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, actor.MerchantId)
	if err != nil {
		return fmt.Errorf("error getting merchant timezone: %s", err.Error())
	}

	var participantCount int
	var seriesParticipantCount int
	var participantChanges participantChanges
	var seriesParticipantChanges seriesParticipantChanges

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
			seriesParticipants, err := s.bookingRepo.GetBookingSeriesParticipants(ctx, *booking.BookingSeriesId)
			if err != nil {
				return fmt.Errorf("error getting series participants: %w", err)
			}

			seriesParticipantChanges, err = detectSeriesParticipantChanges(participants, seriesParticipants, incomingCustomerIds)
			if err != nil {
				return err
			}

			seriesParticipantCount = len(seriesParticipantChanges.ToInsert) + len(seriesParticipantChanges.ToKeep)
		}
		// walk ins do not get a booking participant row but 1 person still attending the booking
	} else {
		participantCount = 1
		seriesParticipantCount = 1
	}

	participantsChanged := len(participantChanges.ToInsert) != 0 || len(participantChanges.ToDelete) != 0
	seriesParticipantsChanged := len(seriesParticipantChanges.ToInsert) != 0 || len(seriesParticipantChanges.ToDelete) != 0
	serviceChanged := booking.ServiceId != input.ServiceId
	priceChanged := serviceChanged || participantsChanged
	timeStampChanged := !booking.FromDate.Equal(input.TimeStamp.UTC())
	merchantNoteChanged := booking.MerchantNote != input.MerchantNote
	statusChanged := booking.Status != input.BookingStatus

	var service domain.PublicServiceWithPhases

	if serviceChanged {
		service, err = s.catalogRepo.GetServiceWithPhases(ctx, input.ServiceId, actor.MerchantId)
		if err != nil {
			return fmt.Errorf("error retrieving service: %s", err.Error())
		}

		if service.BookingType == types.BookingTypeAppointment && participantCount > 1 {
			return fmt.Errorf("appointments cannot have more than 1 customer")
		}

		if service.BookingType != types.BookingTypeAppointment && participantCount > service.MaxParticipants {
			return fmt.Errorf("customer count (%d) exceeds class limit of %d", participantCount, service.MaxParticipants)
		}
	} else {
		service, err = s.catalogRepo.GetServiceWithPhases(ctx, booking.ServiceId, actor.MerchantId)
		if err != nil {
			return fmt.Errorf("error retrieving service: %s", err.Error())
		}
	}

	duration := time.Duration(service.TotalDuration) * time.Minute

	fromDate := booking.FromDate
	toDate := booking.ToDate
	fromDateOffset := time.Duration(0)

	seriesFromDate := booking.FromDate.In(merchantTz)
	seriesToDate := booking.ToDate.In(merchantTz)

	if timeStampChanged {
		fromDateOffset = input.TimeStamp.UTC().Sub(booking.FromDate)

		fromDate = fromDate.Add(fromDateOffset)
		toDate = fromDate.Add(duration)

		if booking.IsRecurring && booking.SeriesOriginalDate != nil {
			// For a recurring bookings we cannot use an offset as individual bookings which are part of the series
			// could have been modified. Also it has to be in the merchant's timezone to avoid DST problems
			seriesFromDate = input.TimeStamp.In(merchantTz)
			seriesToDate = seriesFromDate.Add(duration)
		}
	} else {
		if serviceChanged {
			toDate = fromDate.Add(duration)

			if input.UpdateAllFuture && booking.IsRecurring {
				seriesToDate = seriesFromDate.Add(duration)
			}
		}
	}

	merchantNote := booking.MerchantNote

	if merchantNoteChanged {
		merchantNote = input.MerchantNote
	}

	isGroupBooking := booking.IsGroupBooking()
	if serviceChanged {
		isGroupBooking = service.BookingType != types.BookingTypeAppointment
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

	var pricePerPerson, costPerPerson currencyx.Price
	var totalPrice, totalCost currency.Amount
	var seriesTotalPrice, seriesTotalCost currency.Amount

	if priceChanged {
		countStr := strconv.Itoa(participantCount)

		pricePerPerson, costPerPerson, err = s.preventNilBookingPrice(ctx, actor.MerchantId, service.Price, service.Cost)
		if err != nil {
			return err
		}

		totalPrice, err = pricePerPerson.Mul(countStr)
		if err != nil {
			return fmt.Errorf("failed to calculate total price: %s", err.Error())
		}

		totalCost, err = costPerPerson.Mul(countStr)
		if err != nil {
			return fmt.Errorf("failed to calculate total cost: %s", err.Error())
		}

		if input.UpdateAllFuture && booking.IsRecurring {
			seriesCountStr := strconv.Itoa(seriesParticipantCount)

			seriesTotalPrice, err = pricePerPerson.Mul(seriesCountStr)
			if err != nil {
				return fmt.Errorf("failed to calculate series total price: %s", err.Error())
			}

			seriesTotalCost, err = costPerPerson.Mul(seriesCountStr)
			if err != nil {
				return fmt.Errorf("failed to calculate series total cost: %s", err.Error())
			}
		}
	} else {
		pricePerPerson = booking.PricePerPerson
		costPerPerson = booking.CostPerPerson
		totalPrice = booking.TotalPrice.Amount
		totalCost = booking.TotalCost.Amount
	}

	var bookingsToUpdate []domain.Booking

	bookingsToUpdate = append(bookingsToUpdate, booking)

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		if input.UpdateAllFuture && booking.IsRecurring {
			// TODO: this should probably be queried higher to avoid unnecessary actions
			bookingSeries, err := s.bookingRepo.WithTx(tx).GetBookingSeries(ctx, *booking.BookingSeriesId)
			if err != nil {
				return fmt.Errorf("failed to fetch booking series: %w", err)
			}

			if !bookingSeries.IsActive {
				return fmt.Errorf("cannot update an inactive booking series")
			}

			rruleStr := bookingSeries.Rrule
			dStart := bookingSeries.Dstart

			if timeStampChanged {
				parsedRule, err := rrule.StrToRRule(bookingSeries.Rrule)
				if err != nil {
					return fmt.Errorf("failed to parse existing rrule: %w", err)
				}

				dStart = seriesFromDate

				parsedRule.DTStart(dStart)
				rruleStr = parsedRule.String()
			}

			if serviceChanged || timeStampChanged {
				err = s.bookingRepo.WithTx(tx).UpdateBookingSeriesCore(ctx, *booking.BookingSeriesId, service.Id, service.BookingType, rruleStr, dStart)
				if err != nil {
					return fmt.Errorf("failed to update booking series core: %w", err)
				}
			}

			if !priceChanged {
				seriesTotalPrice = bookingSeries.TotalPrice.Amount
				seriesTotalCost = bookingSeries.TotalCost.Amount
			}

			if seriesParticipantsChanged || serviceChanged || priceChanged {
				err = s.bookingRepo.WithTx(tx).UpdateBookingSeriesDetails(ctx, *booking.BookingSeriesId, domain.BookingDetails{
					PricePerPerson:      pricePerPerson,
					CostPerPerson:       costPerPerson,
					TotalPrice:          currencyx.Price{Amount: seriesTotalPrice},
					TotalCost:           currencyx.Price{Amount: seriesTotalCost},
					MinParticipants:     service.MinParticipants,
					MaxParticipants:     service.MaxParticipants,
					CurrentParticipants: participantCount,
				})
				if err != nil {
					return fmt.Errorf("failed to update booking series details: %w", err)
				}
			}

			if seriesParticipantsChanged {
				if len(seriesParticipantChanges.ToDelete) > 0 {
					err = s.bookingRepo.WithTx(tx).DeleteBookingSeriesParticipants(ctx, *booking.BookingSeriesId, seriesParticipantChanges.ToDelete)
					if err != nil {
						return fmt.Errorf("failed to delete series participants: %w", err)
					}
				}

				if len(seriesParticipantChanges.ToInsert) > 0 {
					var seriesParticipantsInsert []domain.BookingSeriesParticipant

					for _, cid := range seriesParticipantChanges.ToInsert {
						seriesParticipantsInsert = append(seriesParticipantsInsert, domain.BookingSeriesParticipant{
							BookingSeriesId: *booking.BookingSeriesId,
							CustomerId:      &cid,
							IsActive:        true,
						})
					}

					_, err = s.bookingRepo.WithTx(tx).NewBookingSeriesParticipants(ctx, seriesParticipantsInsert)
					if err != nil {
						return fmt.Errorf("failed to insert series participants: %w", err)
					}
				}
			}

			// get all future series occurrences excluding the current booking
			bookingsToUpdate, err = s.bookingRepo.WithTx(tx).GetFutureSeriesBookings(ctx, *booking.BookingSeriesId, booking.FromDate)
			if err != nil {
				return fmt.Errorf("failed to fetch future series bookings: %w", err)
			}

		}

		var bookingIds []int
		var futureBookingIds []int
		var newFromDates []time.Time
		var newToDates []time.Time
		var bookingPhasesToInsert []domain.BookingPhase
		var participantsToUpsert []domain.BookingParticipant

		for _, b := range bookingsToUpdate {
			bookingIds = append(bookingIds, b.Id)

			if b.IsRecurring {
				newFromDates = append(newFromDates, seriesFromDate)
				newToDates = append(newToDates, seriesToDate)
			} else {
				newFromDates = append(newFromDates, fromDate)
				newToDates = append(newToDates, toDate)
			}

			if serviceChanged || timeStampChanged {
				bookingPhaseStart := fromDate

				for _, phase := range service.Phases {
					phaseDuration := time.Duration(phase.Duration) * time.Minute
					bookingPhaseEnd := bookingPhaseStart.Add(phaseDuration)

					bookingPhasesToInsert = append(bookingPhasesToInsert, domain.BookingPhase{
						BookingId:      b.Id,
						ServicePhaseId: phase.Id,
						FromDate:       bookingPhaseStart,
						ToDate:         bookingPhaseEnd,
					})

					bookingPhaseStart = bookingPhaseEnd
				}
			}

			for _, cid := range participantChanges.ToInsert {
				participantsToUpsert = append(participantsToUpsert, domain.BookingParticipant{
					BookingId:  b.Id,
					CustomerId: &cid,
					Status:     participantBookingStatus,
				})
			}

			// TODO: do we want to overwrite all bookings statuses in a series?
			// overwrite the participant status to match the new booking status
			if booking.Id == b.Id {
				if !isGroupBooking && statusChanged {
					for _, cid := range participantChanges.ToKeep {
						participantsToUpsert = append(participantsToUpsert, domain.BookingParticipant{
							BookingId:  b.Id,
							CustomerId: &cid,
							Status:     participantBookingStatus,
						})
					}
				}
			}
		}

		if input.UpdateAllFuture && booking.IsRecurring && len(bookingIds) > 1 {
			futureBookingIds = bookingIds[1:]
		}

		seriesOriginalDate := booking.SeriesOriginalDate
		if input.UpdateAllFuture && booking.IsRecurring && timeStampChanged {
			seriesOriginalDate = &seriesFromDate
		}

		if timeStampChanged || serviceChanged || statusChanged {
			err = s.bookingRepo.WithTx(tx).UpdateBookingCoreBatch(ctx, actor.MerchantId, bookingIds, service.Id, newFromDates, newToDates, service.BookingType, input.BookingStatus, seriesOriginalDate)
			if err != nil {
				return fmt.Errorf("failed to batch update booking core: %s", err.Error())
			}
		}

		if serviceChanged || priceChanged || merchantNoteChanged || participantsChanged {
			err = s.bookingRepo.WithTx(tx).UpdateBookingDetailsBatch(ctx, actor.MerchantId, []int{booking.Id}, domain.BookingDetails{
				PricePerPerson:      pricePerPerson,
				CostPerPerson:       costPerPerson,
				TotalPrice:          currencyx.Price{Amount: totalPrice},
				TotalCost:           currencyx.Price{Amount: totalCost},
				MerchantNote:        merchantNote,
				MinParticipants:     service.MinParticipants,
				MaxParticipants:     service.MaxParticipants,
				CurrentParticipants: participantCount,
			})
			if err != nil {
				return fmt.Errorf("failed to batch update booking details: %s", err.Error())
			}
		}

		if serviceChanged || priceChanged || merchantNoteChanged || seriesParticipantsChanged {
			err = s.bookingRepo.WithTx(tx).UpdateBookingDetailsBatch(ctx, actor.MerchantId, futureBookingIds, domain.BookingDetails{
				PricePerPerson: pricePerPerson,
				CostPerPerson:  costPerPerson,
				TotalPrice:     currencyx.Price{Amount: seriesTotalPrice},
				TotalCost:      currencyx.Price{Amount: seriesTotalCost},
				// TODO: the merchant note shouldn't really be updated on future bookings
				MerchantNote:        input.MerchantNote,
				MinParticipants:     service.MinParticipants,
				MaxParticipants:     service.MaxParticipants,
				CurrentParticipants: seriesParticipantCount,
			})
			if err != nil {
				return fmt.Errorf("failed to batch update future booking details: %s", err.Error())
			}
		}

		if serviceChanged || timeStampChanged {
			if len(bookingPhasesToInsert) > 0 {
				err := s.bookingRepo.WithTx(tx).DeleteBookingPhasesBatch(ctx, bookingIds)
				if err != nil {
					return fmt.Errorf("failed to delete booking phases: %s", err.Error())
				}

				err = s.bookingRepo.WithTx(tx).NewBookingPhases(ctx, bookingPhasesToInsert)
				if err != nil {
					return fmt.Errorf("failed to insert booking phases: %s", err.Error())
				}
			}
		}

		if participantsChanged {
			if len(participantChanges.ToDelete) > 0 {
				err := s.bookingRepo.WithTx(tx).DeleteBookingParticipantsBatch(ctx, []int{booking.Id}, participantChanges.ToDelete)
				if err != nil {
					return fmt.Errorf("failed to remove participants: %s", err.Error())
				}
			}
		}

		if seriesParticipantsChanged {
			if len(futureBookingIds) > 0 && len(seriesParticipantChanges.ToDelete) > 0 {
				err := s.bookingRepo.WithTx(tx).DeleteBookingParticipantsBatch(ctx, futureBookingIds, seriesParticipantChanges.ToDelete)
				if err != nil {
					return fmt.Errorf("failed to remove participants for future bookings: %s", err.Error())
				}
			}
		}

		if participantsChanged || seriesParticipantsChanged || statusChanged {
			if len(participantsToUpsert) > 0 {
				err := s.bookingRepo.WithTx(tx).UpdateBookingParticipants(ctx, participantsToUpsert)
				if err != nil {
					return fmt.Errorf("failed to add participants: %s", err.Error())
				}
			}
		}

		if timeStampChanged {
			for _, b := range bookingsToUpdate {
				// TODO: don't forget to change this when we will consider employee changes in this
				if b.EmployeeId != nil {
					_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncUpdateBooking{
						BookingId: b.Id,
					}, nil)
					if err != nil {
						return err
					}
				}
			}
		}

		lang := lang.LangFromContext(ctx)

		if participantsChanged {
			for _, id := range participantChanges.ToDelete {
				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingCancellationEmail{
					Language:           lang,
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
			if statusChanged {
				if bookingStatus == types.BookingStatusCancelled {
					_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingCancellationEmail{
						Language:           lang,
						BookingId:          booking.Id,
						CustomerId:         id,
						CancellationReason: "",
					}, nil)
					if err != nil {
						return fmt.Errorf("could not schedule booking cancellation email job: %w", err)
					}
				}
			}

			if statusChangedToConfirmed {
				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingStatusConfirmedEmail{
					Language:   lang,
					BookingId:  booking.Id,
					CustomerId: id,
				}, nil)
				if err != nil {
					return fmt.Errorf("could not schedule booking confirmation email job: %w", err)
				}
			}

			if timeStampChanged {
				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingReminderEmail{
					Language:         lang,
					BookingId:        booking.Id,
					CustomerId:       id,
					ExpectedFromDate: fromDate,
				}, &river.InsertOpts{
					ScheduledAt: reminderDate,
				})
				if err != nil {
					return fmt.Errorf("could not schedule booking reminder email job: %w", err)
				}

				// TODO: we should send a modification when the price is changed, but the email does not handle it currently
				if serviceChanged {
					_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingModificationEmail{
						Language:   lang,
						BookingId:  booking.Id,
						CustomerId: id,
						// TODO: This works incorrectly currently
						// booking should have a name field and we should rely on that here
						OldServiceName: service.Name,
						OldFromDate:    booking.FromDate,
						OldToDate:      booking.ToDate,
					}, nil)
					if err != nil {
						return fmt.Errorf("could not schedule booking modification email job: %w", err)
					}
				}
			}
		}

		return nil
	})
}

type CancelByMerchantInput struct {
	CancellationReason string
}

// TODO: what should the booking participant statuses be here?
func (s *Service) CancelByMerchant(ctx context.Context, bookingId int, input CancelByMerchantInput) error {
	actor := actor.MustGetFromContext(ctx)

	booking, err := s.bookingRepo.GetBooking(ctx, bookingId)
	if err != nil {
		return err
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

		lang := lang.LangFromContext(ctx)

		for _, participant := range bookingParticipants {
			// if not walk-in
			if participant.CustomerId != nil {
				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingCancellationEmail{
					Language:           lang,
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
