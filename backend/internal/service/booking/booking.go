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
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/lang"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/service/email"
	"github.com/miketsu-inc/reservations/backend/internal/types"
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

func (s *Service) newBooking(ctx context.Context, tx db.DBTX, booking domain.Booking, details domain.BookingDetails, participants []domain.BookingParticipant,
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

	details.BookingId = bookingId

	for i := range participants {
		participants[i].BookingId = bookingId
	}

	err = s.bookingRepo.WithTx(tx).NewBookingPhases(ctx, bookingPhases)
	if err != nil {
		return 0, err
	}

	err = s.bookingRepo.WithTx(tx).NewBookingDetails(ctx, details)
	if err != nil {
		return 0, err
	}

	err = s.bookingRepo.WithTx(tx).NewBookingParticipants(ctx, participants)
	if err != nil {
		return 0, err
	}

	return bookingId, nil
}

type CreateByCustomerInput struct {
	MerchantName string
	ServiceId    int
	LocationId   int
	TimeStamp    string
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

	duration := time.Duration(service.TotalDuration) * time.Minute

	timeStamp, err := time.Parse(time.RFC3339, input.TimeStamp)
	if err != nil {
		return fmt.Errorf("timestamp could not be converted to time: %w", err)
	}

	timeStamp = timeStamp.UTC()

	now := time.Now().In(merchantTz)

	if timeStamp.Before(now.Add(time.Duration(bookingSettings.BookingWindowMin) * time.Minute)) {
		return fmt.Errorf("appointment must be booked at least %d minutes in advance", bookingSettings.BookingWindowMin)
	}

	if timeStamp.After(now.AddDate(0, bookingSettings.BookingWindowMax, 0)) {
		return fmt.Errorf("appointment cannot be booked more than %d months in advance", bookingSettings.BookingWindowMax)
	}

	toDate := timeStamp.Add(duration)

	customerId, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("unexpected error during creating customer id: %w", err)
	}

	// prevent null booking price and cost to avoid a lot of headaches
	var price currencyx.Price
	var cost currencyx.Price
	if service.Price == nil || service.Cost == nil {
		curr, err := s.merchantRepo.GetMerchantCurrency(ctx, merchantId)
		if err != nil {
			return fmt.Errorf("error while getting merchant's currency: %w", err)
		}

		zeroAmount, err := currency.NewAmount("0", curr)
		if err != nil {
			return fmt.Errorf("error while creating new amount: %w", err)
		}

		if service.Price != nil {
			price = *service.Price
		} else {
			price = currencyx.Price{Amount: zeroAmount}
		}

		if service.Cost != nil {
			cost = *service.Cost
		} else {
			cost = currencyx.Price{Amount: zeroAmount}
		}
	} else {
		price = *service.Price
		cost = *service.Cost
	}

	var bookingId int

	// inserting new customer here to avoid a nested transaction
	customerId, isBlacklisted, err := s.customerRepo.NewCustomerFromUser(ctx, customerId, merchantId, userId)
	if err != nil {
		return err
	}

	if isBlacklisted {
		return fmt.Errorf("you are blacklisted, please contact the merchant by email or phone to make a booking")
	}

	err = s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		// means customer is booking an appointment
		if input.BookingId == nil {
			booking := domain.Booking{
				Status:      types.BookingStatusBooked,
				BookingType: types.BookingTypeAppointment,
				MerchantId:  merchantId,
				ServiceId:   input.ServiceId,
				LocationId:  input.LocationId,
				FromDate:    timeStamp,
				ToDate:      toDate,
			}

			bookingDetails := domain.BookingDetails{
				PricePerPerson:      price,
				CostPerPerson:       cost,
				TotalPrice:          price,
				TotalCost:           cost,
				MinParticipants:     1,
				MaxParticipants:     1,
				CurrentParticipants: 1,
			}

			participants := []domain.BookingParticipant{{
				Status:       types.BookingStatusBooked,
				CustomerId:   &customerId,
				CustomerNote: &input.CustomerNote,
			}}

			bookingId, err = s.newBooking(ctx, tx, booking, bookingDetails, participants, service.Phases)
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
		} else {
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
				Status:       types.BookingStatusBooked,
				BookingId:    bookingId,
				CustomerId:   &customerId,
				CustomerNote: &input.CustomerNote,
			}}

			err = s.bookingRepo.WithTx(tx).NewBookingParticipants(ctx, participants)
			if err != nil {
				return err
			}
		}

		lang := lang.LangFromContext(ctx)
		_, err = s.enqueuer.Insert(ctx, args.BookingConfirmationEmail{
			Language:   lang,
			BookingId:  bookingId,
			CustomerId: customerId,
		}, nil)
		if err != nil {
			return fmt.Errorf("could not schedule booking confirmation email job: %w", err)
		}

		reminderDate := timeStamp.Add(-24 * time.Hour)

		_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingReminderEmail{
			Language:         lang,
			BookingId:        bookingId,
			CustomerId:       customerId,
			ExpectedFromDate: timeStamp,
		}, &river.InsertOpts{
			ScheduledAt: reminderDate,
		})
		if err != nil {
			return fmt.Errorf("could not schedule booking reminder email job: %w", err)
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

func (s *Service) CancelByCustomer(ctx context.Context, bookingId int, input CancelByCustomerInput) error {
	if bookingId != input.BookingId {
		return fmt.Errorf("invalid booking id")
	}

	userId := jwt.MustGetUserIDFromContext(ctx)

	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, input.MerchantName)
	if err != nil {
		return fmt.Errorf("error while searching merchant by this name: %s", err.Error())
	}

	customerId, err := s.customerRepo.GetCustomerIdByUserIdAndMerchantId(ctx, merchantId, userId)
	if err != nil {
		return fmt.Errorf("error while getting customer id: %s", err.Error())
	}

	// TODO: write seperate query for getting only fromDate and cancel deadline
	emailData, err := s.bookingRepo.GetBookingForEmail(ctx, bookingId, customerId)
	if err != nil {
		return fmt.Errorf("error while retrieving data for email sending: %s", err.Error())
	}

	latestCancelTime := emailData.FromDate.Add(-time.Duration(emailData.CancelDeadline) * time.Minute)

	if time.Now().After(latestCancelTime) {
		return fmt.Errorf("it's too late to cancel this appointments")
	}

	err = s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		bookingDetails, err := s.bookingRepo.WithTx(tx).GetBookingDetails(ctx, bookingId)
		if err != nil {
			return err
		}

		var bookingType types.BookingType

		bookingType, err = s.bookingRepo.WithTx(tx).CancelBookingByCustomer(ctx, bookingId, customerId)
		if err != nil {
			return fmt.Errorf("error while cancelling the booking by user: %s", err.Error())
		}

		newTotalPrice, err := bookingDetails.TotalPrice.Sub(bookingDetails.PricePerPerson.Amount)
		if err != nil {
			return fmt.Errorf("failed to calculate total price: %w", err)
		}

		newTotalCost, err := bookingDetails.TotalCost.Sub(bookingDetails.CostPerPerson.Amount)
		if err != nil {
			return fmt.Errorf("failed to calculate total cost: %w", err)
		}

		err = s.bookingRepo.WithTx(tx).UpdateBookingTotalPrice(ctx, bookingId, currencyx.Price{Amount: newTotalPrice}, currencyx.Price{Amount: newTotalCost})
		if err != nil {
			return err
		}

		if bookingType == types.BookingTypeAppointment {
			err = s.bookingRepo.WithTx(tx).UpdateBookingStatus(ctx, merchantId, bookingId, types.BookingStatusCancelled)
			if err != nil {
				return err
			}

			_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncDeleteBooking{
				BookingId: bookingId,
			}, nil)
			if err != nil {
				return err
			}
		} else {
			err = s.bookingRepo.WithTx(tx).DecrementParticipantCount(ctx, bookingId)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetByCustomer(ctx context.Context, bookingId int) (domain.PublicBooking, error) {
	publicBooking, err := s.bookingRepo.GetPublicBooking(ctx, bookingId)
	if err != nil {
		return domain.PublicBooking{}, err
	}

	return publicBooking, nil
}

type CreateByMerchantInput struct {
	Customers    []CustomerInput
	ServiceId    int
	TimeStamp    string
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
	employee := jwt.MustGetEmployeeFromContext(ctx)

	service, err := s.catalogRepo.GetServiceWithPhases(ctx, input.ServiceId, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error while searching service by this id: %s", err.Error())
	}

	if service.BookingType == types.BookingTypeAppointment && len(input.Customers) > 1 {
		return fmt.Errorf("appointments cannot have more than 1 customer")
	}

	if service.BookingType != types.BookingTypeAppointment && len(input.Customers) > service.MaxParticipants {
		return fmt.Errorf("customer count (%d) exceeds class limit of %d", len(input.Customers), service.MaxParticipants)
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error while getting merchant's timezone: %s", err.Error())
	}

	bookedLocation, err := s.merchantRepo.GetLocation(ctx, employee.LocationId, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error while searching location by this id: %s", err.Error())
	}

	// TODO: this should be a separate function
	// prevent null booking price and cost to avoid a lot of headaches
	var price currencyx.Price
	var cost currencyx.Price
	if service.Price == nil || service.Cost == nil {
		curr, err := s.merchantRepo.GetMerchantCurrency(ctx, employee.MerchantId)
		if err != nil {
			return fmt.Errorf("error while getting merchant's currency: %s", err.Error())
		}

		zeroAmount, err := currency.NewAmount("0", curr)
		if err != nil {
			return fmt.Errorf("error while creating new amount: %s", err.Error())
		}

		if service.Price != nil {
			price = *service.Price
		} else {
			price = currencyx.Price{Amount: zeroAmount}
		}

		if service.Cost != nil {
			cost = *service.Cost
		} else {
			cost = currencyx.Price{Amount: zeroAmount}
		}
	} else {
		price = *service.Price
		cost = *service.Cost
	}

	timeStamp, err := time.Parse(time.RFC3339, input.TimeStamp)
	if err != nil {
		return fmt.Errorf("timestamp could not be converted to time: %s", err.Error())
	}
	timeStamp = timeStamp.UTC()

	duration := time.Duration(service.TotalDuration) * time.Minute

	fromDate := timeStamp.Truncate(time.Second)
	toDate := timeStamp.Add(duration)

	var participantIds []uuid.UUID
	var emailsToSend []uuid.UUID
	isWalkIn := false

	//maybe check if the customers without an id have first and last name
	if len(input.Customers) == 0 {
		isWalkIn = true
	} else {
		for _, customer := range input.Customers {
			if customer.CustomerId != nil {
				participantIds = append(participantIds, *customer.CustomerId)
				emailsToSend = append(emailsToSend, *customer.CustomerId)
			} else {
				if customer.FirstName != nil && customer.LastName != nil {
					newCustomerId, err := uuid.NewV7()
					if err != nil {
						return fmt.Errorf("unexpected error during creating customer id: %s", err.Error())
					}

					if err := s.customerRepo.NewCustomer(ctx, employee.MerchantId, domain.Customer{
						Id:          newCustomerId,
						FirstName:   customer.FirstName,
						LastName:    customer.LastName,
						Email:       customer.Email,
						PhoneNumber: customer.PhoneNumber,
						Birthday:    nil,
						Note:        nil,
					}); err != nil {
						return fmt.Errorf("unexpected error inserting customer for merchant: %s", err.Error())
					}

					participantIds = append(participantIds, newCustomerId)

					if customer.Email != nil {
						emailsToSend = append(emailsToSend, newCustomerId)
					}
				}
			}
		}
	}

	participantCount := len(participantIds)
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

	var bookingId int

	err = s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		if input.IsRecurring && input.Rrule != nil {
			var freq rrule.Frequency

			switch strings.ToUpper(input.Rrule.Frequency) {
			case "DAILY":
				freq = rrule.DAILY
			case "WEEKLY":
				freq = rrule.WEEKLY
			case "MONTHLY":
				freq = rrule.MONTHLY
			default:
				return fmt.Errorf("recurring rule frequency is invalid")
			}

			untilTimeStamp, err := time.Parse(time.RFC3339, input.Rrule.Until)
			if err != nil {
				return fmt.Errorf("until timestamp could not be converted to time: %s", err.Error())
			}

			var weekdays []rrule.Weekday

			for _, wkd := range input.Rrule.Weekdays {
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
					return fmt.Errorf("incorrect weekday")
				}
			}

			rrule, err := rrule.NewRRule(rrule.ROption{
				Freq:      freq,
				Dtstart:   fromDate,
				Interval:  input.Rrule.Interval,
				Byweekday: weekdays,
				Until:     untilTimeStamp,
			})
			if err != nil {
				return fmt.Errorf("error while creating rrule: %s", err.Error())
			}

			// recurring bookings have to be stored in local time and converted to UTC during generation
			fromDate = timeStamp.In(merchantTz)

			var series CompleteBookingSeries

			series.BookingSeries, err = s.bookingRepo.WithTx(tx).NewBookingSeries(ctx, domain.BookingSeries{
				BookingType: service.BookingType,
				MerchantId:  employee.MerchantId,
				EmployeeId:  employee.Id,
				ServiceId:   service.Id,
				LocationId:  bookedLocation.Id,
				Rrule:       rrule.String(),
				Dstart:      fromDate,
				Timezone:    merchantTz.String(),
				IsActive:    true,
			})
			if err != nil {
				return err
			}

			series.Details, err = s.bookingRepo.WithTx(tx).NewBookingSeriesDetails(ctx, domain.BookingSeriesDetails{
				BookingSeriesId:     series.Id,
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

			seriesParticipants := make([]domain.BookingSeriesParticipant, len(participantIds))

			for i, id := range participantIds {
				seriesParticipants[i] = domain.BookingSeriesParticipant{
					BookingSeriesId: series.Id,
					CustomerId:      &id,
					IsActive:        true,
				}
			}

			series.Participants, err = s.bookingRepo.WithTx(tx).NewBookingSeriesParticipants(ctx, seriesParticipants)
			if err != nil {
				return err
			}

			bookingId, err = s.generateRecurringBookings(ctx, tx, series, service.Phases)
			if err != nil {
				return fmt.Errorf("error while generating recurring bookings: %s", err.Error())
			}
		} else {
			booking := domain.Booking{
				Status:      types.BookingStatusBooked,
				BookingType: service.BookingType,
				MerchantId:  employee.MerchantId,
				EmployeeId:  &employee.Id,
				ServiceId:   service.Id,
				LocationId:  bookedLocation.Id,
				FromDate:    fromDate,
				ToDate:      toDate,
			}

			bookingDetails := domain.BookingDetails{
				BookingId:           bookingId,
				PricePerPerson:      price,
				CostPerPerson:       cost,
				TotalPrice:          currencyx.Price{Amount: totalPrice},
				TotalCost:           currencyx.Price{Amount: totalCost},
				MerchantNote:        input.MerchantNote,
				MinParticipants:     service.MinParticipants,
				MaxParticipants:     service.MaxParticipants,
				CurrentParticipants: participantCount,
			}

			participants := make([]domain.BookingParticipant, len(participantIds))
			for i, id := range participantIds {
				participants[i] = domain.BookingParticipant{
					Status:       types.BookingStatusBooked,
					BookingId:    bookingId,
					CustomerId:   &id,
					CustomerNote: nil,
				}
			}

			bookingId, err = s.newBooking(ctx, tx, booking, bookingDetails, participants, service.Phases)
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
			lang := lang.LangFromContext(ctx)

			for _, customerId := range emailsToSend {
				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingConfirmationEmail{
					Language:   lang,
					BookingId:  bookingId,
					CustomerId: customerId,
				}, nil)
				if err != nil {
					return fmt.Errorf("could not schedule booking confirmation email job: %w", err)
				}

				reminderDate := timeStamp.Add(-24 * time.Hour)

				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingReminderEmail{
					Language:         lang,
					BookingId:        bookingId,
					CustomerId:       customerId,
					ExpectedFromDate: timeStamp,
				}, &river.InsertOpts{
					ScheduledAt: reminderDate,
				})
				if err != nil {
					return fmt.Errorf("could not schedule booking reminder email job: %w", err)
				}
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error while creating new booking: %s", err.Error())
	}

	return nil
}

type UpdateByMerchantInput struct {
	Customers       []CustomerInput     `json:"customers"`
	ServiceId       int                 `json:"service_id" validate:"required"`
	TimeStamp       string              `json:"timestamp" validate:"required"`
	MerchantNote    *string             `json:"merchant_note"`
	BookingStatus   types.BookingStatus `json:"booking_status"`
	UpdateAllFuture bool                `json:"update_all_future"`
}

func (s *Service) UpdateByMerchant(ctx context.Context, bookingId int, input UpdateByMerchantInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	// old service, participants, time and date
	oldBooking, err := s.bookingRepo.GetBooking(ctx, bookingId)
	if err != nil {
		return fmt.Errorf("error while retrieving data for email sending: %s", err.Error())
	}

	if oldBooking.Status == types.BookingStatusCompleted {
		return fmt.Errorf("you cannot update completed bookings")
	}

	oldService, err := s.catalogRepo.GetServiceWithPhases(ctx, oldBooking.ServiceId, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error searching for old service: %s", err.Error())
	}

	newService, err := s.catalogRepo.GetServiceWithPhases(ctx, input.ServiceId, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error searching for new service: %s", err.Error())
	}

	if newService.BookingType == types.BookingTypeAppointment && len(input.Customers) > 1 {
		return fmt.Errorf("appointments cannot have more than 1 customer")
	}

	if newService.BookingType != types.BookingTypeAppointment && len(input.Customers) > newService.MaxParticipants {
		return fmt.Errorf("customer count (%d) exceeds class limit of %d", len(input.Customers), newService.MaxParticipants)
	}

	timeStamp, err := time.Parse(time.RFC3339, input.TimeStamp)
	if err != nil {
		return fmt.Errorf("timestamp could not be converted: %s", err.Error())
	}

	//for a recurring series the offset can differ for bookings if one of them was modified seperately earlier
	fromDate := timeStamp.UTC().Truncate(time.Second)
	localFromDateOffset := fromDate.Sub(oldBooking.FromDate)

	var seriesFromDateOffset time.Duration
	if oldBooking.IsRecurring && oldBooking.SeriesOriginalDate != nil {
		seriesFromDateOffset = fromDate.Sub(*oldBooking.SeriesOriginalDate)
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error getting merchant timezone: %s", err.Error())
	}

	participantIdsMap := make(map[uuid.UUID]struct{})
	var newCustomerIds []uuid.UUID

	isWalkIn := len(input.Customers) == 0

	if !isWalkIn {
		for _, customer := range input.Customers {

			if customer.CustomerId != nil {
				participantIdsMap[*customer.CustomerId] = struct{}{}

			} else if customer.FirstName != nil && customer.LastName != nil {

				newId, err := uuid.NewV7()
				if err != nil {
					return fmt.Errorf("error generating customer id: %s", err.Error())
				}

				if err := s.customerRepo.NewCustomer(ctx, employee.MerchantId, domain.Customer{
					Id:          newId,
					FirstName:   customer.FirstName,
					LastName:    customer.LastName,
					Email:       customer.Email,
					PhoneNumber: customer.PhoneNumber,
				}); err != nil {
					return fmt.Errorf("error inserting new customer: %s", err.Error())
				}

				participantIdsMap[newId] = struct{}{}
				newCustomerIds = append(newCustomerIds, newId)
			}
		}
	}

	oldParticipants, err := s.bookingRepo.GetBookingParticipants(ctx, bookingId)
	if err != nil {
		return fmt.Errorf("error getting booking participants: %w", err)
	}

	existingParticipantsMap := make(map[uuid.UUID]struct{})
	for _, p := range oldParticipants {
		// customerId is nil in case of walk-in where email cannot be sent
		if p.CustomerId != nil {
			existingParticipantsMap[*p.CustomerId] = struct{}{}
		}
	}

	var seriesParticipants []domain.BookingSeriesParticipant
	exsistingSeriesParticipantsMap := make(map[uuid.UUID]struct{})

	if input.UpdateAllFuture && oldBooking.IsRecurring {
		seriesParticipants, err = s.bookingRepo.GetBookingSeriesParticipants(ctx, *oldBooking.BookingSeriesId)
		if err != nil {
			return fmt.Errorf("error getting series participants: %w", err)
		}
		for _, sp := range seriesParticipants {
			if sp.CustomerId != nil {
				exsistingSeriesParticipantsMap[*sp.CustomerId] = struct{}{}
			}
		}
	}

	var localParticipantsToInsert []uuid.UUID
	var localParticipantsToDelete []uuid.UUID
	var localParticipantsToKeep []uuid.UUID

	for id := range participantIdsMap {
		if _, exists := existingParticipantsMap[id]; !exists {
			localParticipantsToInsert = append(localParticipantsToInsert, id)
		}
	}

	for id := range existingParticipantsMap {
		if _, exists := participantIdsMap[id]; !exists {
			localParticipantsToDelete = append(localParticipantsToDelete, id)
		} else {
			localParticipantsToKeep = append(localParticipantsToKeep, id)
		}
	}

	var seriesParticipantsToInsert []uuid.UUID
	var seriesParticipantsToDelete []uuid.UUID

	if input.UpdateAllFuture && oldBooking.IsRecurring {
		//only truly new if the person isnt in either of the participant maps
		for id := range participantIdsMap {
			_, inLocal := existingParticipantsMap[id]
			_, inSeries := existingParticipantsMap[id]
			if !inLocal && !inSeries {
				seriesParticipantsToInsert = append(seriesParticipantsToInsert, id)
			}
		}
		//delete from the series only if the person is both in the maps but not in the incoming participants
		for id := range existingParticipantsMap {
			_, inIncoming := participantIdsMap[id]
			_, inSeries := existingParticipantsMap[id]
			if inSeries && !inIncoming {
				seriesParticipantsToDelete = append(seriesParticipantsToDelete, id)
			}
		}
	}

	participantCount := len(participantIdsMap)
	// walk ins do not get a booking participant row but 1 person still attending the booking
	if isWalkIn {
		participantCount = 1
	}

	countStr := strconv.Itoa(participantCount)

	// TODO: this panics if price or cost is nil. We should fix this
	//by not allowing to insert nil prices into services
	totalPrice, err := newService.Price.Mul(countStr)
	if err != nil {
		return fmt.Errorf("failed to calculate total price: %s", err.Error())
	}

	totalCost, err := newService.Cost.Mul(countStr)
	if err != nil {
		return fmt.Errorf("failed to calculate total cost: %s", err.Error())
	}

	//since the participants can differ thus the total price and total cost will be different from the local booking as well
	var seriesParticipantCount int
	var seriesTotalPrice, seriesTotalCost currency.Amount

	if input.UpdateAllFuture && oldBooking.IsRecurring {
		seriesParticipantCount = len(existingParticipantsMap) + len(seriesParticipantsToInsert) - len(seriesParticipantsToDelete)
		if isWalkIn {
			seriesParticipantCount = 1
		}

		seriesCountStr := strconv.Itoa(seriesParticipantCount)
		seriesTotalPrice, err = newService.Price.Mul(seriesCountStr)
		if err != nil {
			return fmt.Errorf("failed to calculate series total price: %s", err.Error())
		}
		seriesTotalCost, err = newService.Cost.Mul(seriesCountStr)
		if err != nil {
			return fmt.Errorf("failed to calculate series total cost: %s", err.Error())
		}
	}

	err = s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		var bookingsToUpdate []domain.Booking
		bookingsToUpdate = append(bookingsToUpdate, oldBooking)

		if input.UpdateAllFuture && oldBooking.IsRecurring {
			bookingSeries, err := s.bookingRepo.WithTx(tx).GetBookingSeries(ctx, *oldBooking.BookingSeriesId)
			if err != nil {
				return fmt.Errorf("failed to fetch booking series: %w", err)
			}

			newRruleStr := bookingSeries.Rrule
			newDstart := bookingSeries.Dstart

			if seriesFromDateOffset != 0 {
				parsedRule, err := rrule.StrToRRule(bookingSeries.Rrule)
				if err != nil {
					return fmt.Errorf("failed to parse existing rrule: %w", err)
				}

				newDstart = bookingSeries.Dstart.Add(seriesFromDateOffset)
				parsedRule.DTStart(newDstart)
				newRruleStr = parsedRule.String()
			}

			futureBookings, err := s.bookingRepo.WithTx(tx).GetFutureSeriesBookings(ctx, *oldBooking.BookingSeriesId, oldBooking.FromDate)
			if err != nil {
				return fmt.Errorf("failed to fetch future series bookings: %w", err)
			}

			for _, fb := range futureBookings {
				if fb.Id != oldBooking.Id {
					bookingsToUpdate = append(bookingsToUpdate, fb)
				}
			}

			err = s.bookingRepo.WithTx(tx).UpdateBookingSeriesCore(ctx, *oldBooking.BookingSeriesId, newService.Id, newService.BookingType, newRruleStr, newDstart)
			if err != nil {
				return fmt.Errorf("failed to update booking series core: %w", err)
			}

			err = s.bookingRepo.WithTx(tx).UpdateBookingSeriesDetails(ctx, *oldBooking.BookingSeriesId, domain.BookingSeriesDetails{
				PricePerPerson:      *newService.Price,
				CostPerPerson:       *newService.Cost,
				TotalPrice:          currencyx.Price{Amount: seriesTotalPrice},
				TotalCost:           currencyx.Price{Amount: seriesTotalCost},
				MinParticipants:     newService.MinParticipants,
				MaxParticipants:     newService.MaxParticipants,
				CurrentParticipants: participantCount,
			})
			if err != nil {
				return fmt.Errorf("failed to update booking series details: %w", err)
			}

			if len(seriesParticipantsToDelete) > 0 {
				err = s.bookingRepo.WithTx(tx).DeleteBookingSeriesParticipants(ctx, *oldBooking.BookingSeriesId, seriesParticipantsToDelete)
				if err != nil {
					return fmt.Errorf("failed to delete series participants: %w", err)
				}
			}

			if len(seriesParticipantsToInsert) > 0 {
				var seriesParts []domain.BookingSeriesParticipant
				for _, pid := range seriesParticipantsToInsert {
					seriesParts = append(seriesParts, domain.BookingSeriesParticipant{
						BookingSeriesId: *oldBooking.BookingSeriesId,
						CustomerId:      &pid,
						IsActive:        true,
					})
				}
				_, err = s.bookingRepo.WithTx(tx).NewBookingSeriesParticipants(ctx, seriesParts)
				if err != nil {
					return fmt.Errorf("failed to insert series participants: %w", err)
				}
			}
		}

		var bookingIds []int
		var futureBookingIds []int
		var newFromDates []time.Time
		var newToDates []time.Time
		//new phases to be batch inserted containing booking ids
		var allNewPhases []domain.BookingPhase
		//new participants to be batch inserted containing booking ids
		var allNewParticipants []domain.BookingParticipant

		status := input.BookingStatus
		if newService.BookingType != types.BookingTypeAppointment {
			status = types.BookingStatusBooked
		}

		for _, b := range bookingsToUpdate {
			bookingIds = append(bookingIds, b.Id)
			if b.Id != oldBooking.Id {
				futureBookingIds = append(futureBookingIds, b.Id)
			}

			var newBookingStart time.Time

			if input.UpdateAllFuture && b.IsRecurring {
				newBookingStart = b.SeriesOriginalDate.Add(seriesFromDateOffset)
			} else {
				newBookingStart = b.FromDate.Add(localFromDateOffset)
			}

			duration := time.Duration(newService.TotalDuration) * time.Minute
			newBookingEnd := newBookingStart.Add(duration)

			newFromDates = append(newFromDates, newBookingStart)
			newToDates = append(newToDates, newBookingEnd)

			if newService.Id != b.ServiceId || !newBookingStart.Equal(b.FromDate) {
				bookingStart := newBookingStart

				//if service didnt chnage new service is the old service
				for _, phase := range newService.Phases {
					phaseDuration := time.Duration(phase.Duration) * time.Minute
					bookingEnd := bookingStart.Add(phaseDuration)

					allNewPhases = append(allNewPhases, domain.BookingPhase{
						BookingId:      b.Id,
						ServicePhaseId: phase.Id,
						FromDate:       bookingStart,
						ToDate:         bookingEnd,
					})
					bookingStart = bookingEnd
				}
			}

			if b.Id == oldBooking.Id {
				//if local booking then apply all local changes as before

				if len(localParticipantsToInsert) > 0 {
					for _, pid := range localParticipantsToInsert {
						allNewParticipants = append(allNewParticipants, domain.BookingParticipant{
							BookingId:  b.Id,
							CustomerId: &pid,
							Status:     status,
						})
					}
				}

				//since the query is an upsert for an appointment we want to have the kept participants so we can update the status
				//for group booking the status management is seperate so we don't use these participant so the batch query will only insert the new ones
				if newService.BookingType == types.BookingTypeAppointment {
					if len(localParticipantsToKeep) > 0 {
						for _, pid := range localParticipantsToKeep {
							allNewParticipants = append(allNewParticipants, domain.BookingParticipant{
								BookingId:  b.Id,
								CustomerId: &pid,
								Status:     input.BookingStatus,
							})
						}
					}
				}
			} else {
				//if future series booking then apply series changes only
				for _, pid := range seriesParticipantsToInsert {
					allNewParticipants = append(allNewParticipants, domain.BookingParticipant{
						BookingId:  b.Id,
						CustomerId: &pid,
						Status:     status,
					})
				}
			}
		}

		err = s.bookingRepo.WithTx(tx).UpdateBookingCoreBatch(ctx, employee.MerchantId, bookingIds, newService.Id, newFromDates, newToDates, newService.BookingType, input.BookingStatus)
		if err != nil {
			return fmt.Errorf("failed to batch update booking core: %s", err.Error())
		}

		err = s.bookingRepo.WithTx(tx).UpdateBookingDetailsBatch(ctx, employee.MerchantId, []int{oldBooking.Id}, domain.BookingDetails{
			PricePerPerson:      *newService.Price,
			CostPerPerson:       *newService.Cost,
			TotalPrice:          currencyx.Price{Amount: totalPrice},
			TotalCost:           currencyx.Price{Amount: totalCost},
			MerchantNote:        input.MerchantNote,
			MinParticipants:     newService.MinParticipants,
			MaxParticipants:     newService.MaxParticipants,
			CurrentParticipants: participantCount,
		})
		if err != nil {
			return fmt.Errorf("failed to batch update booking details: %s", err.Error())
		}

		if len(futureBookingIds) > 0 {
			err = s.bookingRepo.WithTx(tx).UpdateBookingDetailsBatch(ctx, employee.MerchantId, futureBookingIds, domain.BookingDetails{
				PricePerPerson:      *newService.Price,
				CostPerPerson:       *newService.Cost,
				TotalPrice:          currencyx.Price{Amount: seriesTotalPrice},
				TotalCost:           currencyx.Price{Amount: seriesTotalCost},
				MerchantNote:        input.MerchantNote,
				MinParticipants:     newService.MinParticipants,
				MaxParticipants:     newService.MaxParticipants,
				CurrentParticipants: seriesParticipantCount,
			})
			if err != nil {
				return fmt.Errorf("failed to batch update future booking details: %s", err.Error())
			}
		}

		if len(allNewPhases) > 0 {
			err := s.bookingRepo.WithTx(tx).DeleteBookingPhasesBatch(ctx, bookingIds)
			if err != nil {
				return fmt.Errorf("failed to delete booking phases: %s", err.Error())
			}

			err = s.bookingRepo.WithTx(tx).NewBookingPhases(ctx, allNewPhases)
			if err != nil {
				return fmt.Errorf("failed to insert booking phases: %s", err.Error())
			}
		}

		//delete the local chnage
		if len(localParticipantsToDelete) > 0 {
			err := s.bookingRepo.WithTx(tx).DeleteBookingParticipantsBatch(ctx, []int{oldBooking.Id}, localParticipantsToDelete)
			if err != nil {
				return fmt.Errorf("failed to remove participants: %s", err.Error())
			}
		}

		//delete the series changes in the future
		if len(futureBookingIds) > 0 && len(seriesParticipantsToDelete) > 0 {
			err := s.bookingRepo.WithTx(tx).DeleteBookingParticipantsBatch(ctx, futureBookingIds, seriesParticipantsToDelete)
			if err != nil {
				return fmt.Errorf("failed to remove participants for future bookings: %s", err.Error())
			}
		}

		//insert all the assigned participants to all the bookings
		if len(allNewParticipants) > 0 {
			err := s.bookingRepo.WithTx(tx).UpdateBookingParticipants(ctx, allNewParticipants)
			if err != nil {
				return fmt.Errorf("failed to add participants: %s", err.Error())
			}
		}

		for _, booking := range bookingsToUpdate {
			if localFromDateOffset != 0 || seriesFromDateOffset != 0 {
				// TODO: don't forget to change this when we will consider employee changes in this
				if oldBooking.EmployeeId != nil {
					_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncUpdateBooking{
						BookingId: booking.Id,
					}, nil)
					if err != nil {
						return err
					}
				}
			}
		}

		fromDateMerchantTz := fromDate.In(merchantTz)
		reminderDate := fromDateMerchantTz.Add(-24 * time.Hour)

		lang := lang.LangFromContext(ctx)
		// implement the updating of the scheduled reminder emails for future recurring bookings
		for _, id := range localParticipantsToDelete {
			_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingCancellationEmail{
				Language:           lang,
				BookingId:          bookingId,
				CustomerId:         id,
				CancellationReason: "",
			}, nil)
			if err != nil {
				return fmt.Errorf("could not schedule booking cancellation email job: %w", err)
			}
		}

		for _, id := range newCustomerIds {
			_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingConfirmationEmail{
				Language:   lang,
				BookingId:  bookingId,
				CustomerId: id,
			}, nil)
			if err != nil {
				return fmt.Errorf("could not schedule booking confirmation email job: %w", err)
			}

			_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingReminderEmail{
				Language:         lang,
				BookingId:        bookingId,
				CustomerId:       id,
				ExpectedFromDate: timeStamp,
			}, &river.InsertOpts{
				ScheduledAt: reminderDate,
			})
			if err != nil {
				return fmt.Errorf("could not schedule booking reminder email job: %w", err)
			}
		}

		for _, id := range localParticipantsToKeep {
			if localFromDateOffset != 0 || newService.Id != oldBooking.ServiceId {

				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingModificationEmail{
					Language:       lang,
					BookingId:      bookingId,
					CustomerId:     id,
					OldServiceName: oldService.Name,
					OldFromDate:    oldBooking.FromDate,
					OldToDate:      oldBooking.ToDate,
				}, nil)
				if err != nil {
					return fmt.Errorf("could not schedule booking modification email job: %w", err)
				}

			}

			if localFromDateOffset != 0 {
				// TODO: expand on conditions for sending new reminder
				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingReminderEmail{
					Language:         lang,
					BookingId:        bookingId,
					CustomerId:       id,
					ExpectedFromDate: timeStamp,
				}, &river.InsertOpts{
					ScheduledAt: reminderDate,
				})
				if err != nil {
					return fmt.Errorf("could not schedule booking reminder email job: %w", err)
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

type CancelByMerchantInput struct {
	CancellationReason string
}

// TODO: what should the booking participant statuses be here?
func (s *Service) CancelByMerchant(ctx context.Context, bookingId int, input CancelByMerchantInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	booking, err := s.bookingRepo.GetBooking(ctx, bookingId)
	if err != nil {
		return err
	}

	if booking.FromDate.Before(time.Now().UTC()) {
		return fmt.Errorf("you cannot cancel past bookings")
	}

	bookingParticipants, err := s.bookingRepo.GetBookingParticipants(ctx, bookingId)
	if err != nil {
		return err
	}

	err = s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err = s.bookingRepo.WithTx(tx).CancelBookingByMerchant(ctx, employee.MerchantId, bookingId, input.CancellationReason)
		if err != nil {
			return err
		}

		err = s.bookingRepo.WithTx(tx).UpdateBookingStatus(ctx, employee.MerchantId, bookingId, types.BookingStatusCancelled)
		if err != nil {
			return err
		}

		lang := lang.LangFromContext(ctx)

		for _, participant := range bookingParticipants {
			// if not walk-in
			if participant.CustomerId != nil {
				_, err = s.enqueuer.InsertTx(ctx, tx, args.BookingCancellationEmail{
					Language:           lang,
					BookingId:          bookingId,
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
				BookingId: bookingId,
			}, nil)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error while cancelling booking by merchant: %s", err.Error())
	}

	return nil
}

type UpdatePaticipantStatusInput struct {
	Status types.BookingStatus
}

func (s *Service) UpdateParticipantStatus(ctx context.Context, bookingId int, participantId int, input UpdatePaticipantStatusInput) error {
	err := s.bookingRepo.UpdateParticipantStatus(ctx, bookingId, participantId, input.Status)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetCalendarEvents(ctx context.Context, start string, end string) (domain.CalendarEvents, error) {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	var events domain.CalendarEvents
	var err error

	events.Bookings, err = s.bookingRepo.GetBookingsForCalendar(ctx, employee.MerchantId, start, end)
	if err != nil {
		return domain.CalendarEvents{}, err
	}

	events.BlockedTimes, err = s.blockedTimeRepo.GetBlockedTimesForCalendar(ctx, employee.MerchantId, start, end)
	if err != nil {
		return domain.CalendarEvents{}, err
	}

	return events, nil
}
