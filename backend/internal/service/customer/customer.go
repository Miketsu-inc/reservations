package customer

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/repository"
	"github.com/miketsu-inc/reservations/backend/pkg/cursor"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type bookingCursor struct {
	Id       int       `json:"id"`
	FromDate time.Time `json:"from_date"`
}

type GetBookingsResult struct {
	Bookings    []domain.CustomerBooking
	NextCursor  *string
	HasNextPage bool
}

type Service struct {
	customerRepo domain.CustomerRepository
	bookingRepo  domain.BookingRepository
	txManager    db.TransactionManager
}

func NewService(customer domain.CustomerRepository, booking domain.BookingRepository, txManager db.TransactionManager) *Service {
	return &Service{
		customerRepo: customer,
		bookingRepo:  booking,
		txManager:    txManager,
	}
}

type NewInput struct {
	FirstName   *string
	LastName    *string
	Email       *string
	PhoneNumber *string
	Birthday    *time.Time
	Note        *string
}

func (s *Service) New(ctx context.Context, input NewInput) error {
	customerId, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("unexpected error during creating customer id: %s", err.Error())
	}

	actor := actor.MustGetFromContext(ctx)

	if err := s.customerRepo.NewCustomer(ctx, actor.MerchantId, domain.Customer{
		Id:          customerId,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
		Birthday:    input.Birthday,
		Note:        input.Note,
	}); err != nil {
		return err
	}

	return nil
}

type UpdateInput struct {
	Id          uuid.UUID
	FirstName   *string
	LastName    *string
	Email       *string
	PhoneNumber *string
	Birthday    *time.Time
	Note        *string
}

func (s *Service) Update(ctx context.Context, customerId uuid.UUID, input UpdateInput) error {
	if customerId != input.Id {
		return fmt.Errorf("invalid customer id provided")
	}

	actor := actor.MustGetFromContext(ctx)

	err := s.customerRepo.UpdateCustomer(ctx, actor.MerchantId, domain.Customer{
		Id:          input.Id,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
		Birthday:    input.Birthday,
		Note:        input.Note,
	})
	if err != nil {
		return err
	}

	return nil
}

// TODO: we should ask if they want to delete their booking history as well or not
// also letting them delete customers who are user's by just deleting their bookings
// we should also decide what to do with deleted class/event participants
func (s *Service) Delete(ctx context.Context, customerId uuid.UUID) error {
	actor := actor.MustGetFromContext(ctx)

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.bookingRepo.WithTx(tx).DeleteAppointmentsByCustomer(ctx, customerId, actor.MerchantId)
		if err != nil {
			return err
		}

		err = s.bookingRepo.WithTx(tx).DecrementEveryParticipantCountForCustomer(ctx, customerId, actor.MerchantId)
		if err != nil {
			return err
		}

		err = s.bookingRepo.WithTx(tx).DeleteParticipantByCustomer(ctx, customerId, actor.MerchantId)
		if err != nil {
			return err
		}

		err = s.customerRepo.WithTx(tx).DeleteCustomer(ctx, customerId, actor.MerchantId)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) Get(ctx context.Context, customerId uuid.UUID) (domain.CustomerInfo, error) {
	actor := actor.MustGetFromContext(ctx)

	customer, err := s.customerRepo.GetCustomerInfo(ctx, actor.MerchantId, customerId)
	if err != nil {
		return domain.CustomerInfo{}, err
	}

	return customer, nil
}

func (s *Service) GetStats(ctx context.Context, customerId uuid.UUID) (domain.CustomerStatistics, error) {
	actor := actor.MustGetFromContext(ctx)

	customerStats, err := s.customerRepo.GetCustomerStats(ctx, actor.MerchantId, customerId)
	if err != nil {
		return domain.CustomerStatistics{}, err
	}

	return customerStats, nil
}

func (s *Service) GetBookings(ctx context.Context, customerId uuid.UUID, status, cursorStr string, pageSize int) (GetBookingsResult, error) {
	actor := actor.MustGetFromContext(ctx)

	decodedCursor, err := cursor.Decode[bookingCursor](cursorStr)
	if err != nil {
		return GetBookingsResult{}, fmt.Errorf("error during cursor decoding: %s", err.Error())
	}

	if cursorStr == "" && status != "upcoming" {
		decodedCursor = bookingCursor{
			Id:       math.MaxInt32,
			FromDate: time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC),
		}
	}

	bookings, err := s.customerRepo.GetCustomerBookings(
		ctx,
		actor.MerchantId,
		customerId,
		status,
		pageSize+1,
		decodedCursor.FromDate,
		decodedCursor.Id,
	)
	if err != nil {
		return GetBookingsResult{}, err
	}

	hasNextPage := len(bookings) > pageSize
	var nextCursor *string
	if hasNextPage {
		lastBooking := bookings[pageSize-1]
		encodedCursor, err := cursor.Encode(bookingCursor{
			Id:       lastBooking.Id,
			FromDate: lastBooking.FromDate,
		})
		if err != nil {
			return GetBookingsResult{}, fmt.Errorf("error during cursor encoding: %s", err.Error())
		}

		nextCursor = &encodedCursor
		bookings = bookings[:pageSize]
	}

	return GetBookingsResult{
		Bookings:    bookings,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}

type BlacklistInput struct {
	CustomerId      uuid.UUID
	BlacklistReason *string
}

func (s *Service) Blacklist(ctx context.Context, customerId uuid.UUID, input BlacklistInput) error {
	if customerId != input.CustomerId {
		return fmt.Errorf("invalid customer id")
	}

	actor := actor.MustGetFromContext(ctx)

	err := s.customerRepo.SetBlacklistStatusForCustomer(ctx, actor.MerchantId, customerId, true, input.BlacklistReason)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UnBlacklist(ctx context.Context, customerId uuid.UUID) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.customerRepo.SetBlacklistStatusForCustomer(ctx, actor.MerchantId, customerId, false, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetAll(ctx context.Context) ([]domain.PublicCustomer, error) {
	actor := actor.MustGetFromContext(ctx)

	customers, err := s.customerRepo.GetCustomers(ctx, actor.MerchantId, false)
	if err != nil {
		return []domain.PublicCustomer{}, err
	}

	return customers, nil
}

type TransferBookingsInput struct {
	FromCustomerId uuid.UUID
	ToCustomerId   uuid.UUID
}

func (s *Service) TransferBookings(ctx context.Context, input TransferBookingsInput) error {
	actor := actor.MustGetFromContext(ctx)

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.bookingRepo.WithTx(tx).TransferBookingParticipants(ctx, actor.MerchantId, input.FromCustomerId, input.ToCustomerId)
		if err != nil {
			if db.IsUniqueConstraintViolation(err, repository.UniqueBookingParticipantConstraint) {
				return ErrCustomerTransferConflict
			}

			return err
		}

		err = s.bookingRepo.WithTx(tx).TransferBookingSeriesParticipants(ctx, actor.MerchantId, input.FromCustomerId, input.ToCustomerId)
		if err != nil {
			if db.IsUniqueConstraintViolation(err, repository.UniqueBookingSeriesParticipantConstraint) {
				return ErrCustomerTransferConflict
			}

			return err
		}

		return nil
	})
}

func (s *Service) GetAllBlacklisted(ctx context.Context) ([]domain.PublicCustomer, error) {
	actor := actor.MustGetFromContext(ctx)

	blacklistedCustomers, err := s.customerRepo.GetCustomers(ctx, actor.MerchantId, true)
	if err != nil {
		return []domain.PublicCustomer{}, err
	}

	return blacklistedCustomers, nil
}

func (s *Service) GetCustomersForCalendar(ctx context.Context) ([]domain.CustomerForCalendar, error) {
	actor := actor.MustGetFromContext(ctx)

	customers, err := s.customerRepo.GetCustomersForCalendar(ctx, actor.MerchantId)
	if err != nil {
		return []domain.CustomerForCalendar{}, err
	}

	return customers, nil
}
