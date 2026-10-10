package location

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type Service struct {
	locationRepo domain.LocationRepository
	merchantRepo domain.MerchantRepository
	txManager    db.TransactionManager
}

func NewService(location domain.LocationRepository, merchant domain.MerchantRepository, txManager db.TransactionManager) *Service {
	return &Service{locationRepo: location, merchantRepo: merchant, txManager: txManager}
}

type NewInput struct {
	Country           *string
	CountryCode       *string
	Region            *string
	City              *string
	PostalCode        *string
	Address           *string
	GeoPoint          types.GeoPoint
	GeocodingProvider *string
	ProviderPlaceId   *string
	FormattedLocation string
}

type UpdateInput struct {
	Name              string
	ContactEmail      *string
	PhoneNumber       *string
	Country           *string
	CountryCode       *string
	Region            *string
	City              *string
	PostalCode        *string
	Address           *string
	GeoPoint          types.GeoPoint
	GeocodingProvider *string
	ProviderPlaceId   *string
	FormattedLocation string
}

func (s *Service) New(ctx context.Context, input NewInput) error {
	actor := actor.MustGetFromContext(ctx)
	locations, err := s.locationRepo.GetLocations(ctx, actor.MerchantId)
	if err != nil {
		return err
	}
	if len(locations) > 0 {
		return ErrLocationAlreadyExists
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		locationId, err := s.locationRepo.WithTx(tx).NewLocation(ctx, domain.Location{
			MerchantId:        actor.MerchantId,
			Name:              "Main location",
			Sequence:          0,
			Country:           input.Country,
			CountryCode:       input.CountryCode,
			Region:            input.Region,
			City:              input.City,
			PostalCode:        input.PostalCode,
			Address:           input.Address,
			GeoPoint:          input.GeoPoint,
			GeocodingProvider: input.GeocodingProvider,
			ProviderPlaceId:   input.ProviderPlaceId,
			FormattedLocation: input.FormattedLocation,
			ParkingInfo:       "",
			ArrivalInfo:       "",
			IsActive:          true,
		})
		if err != nil {
			return err
		}

		return s.merchantRepo.WithTx(tx).NewBusinessHours(ctx, locationId, defaultBusinessHours())
	})
}

func defaultBusinessHours() domain.BusinessHours {
	start := time.Date(0, time.January, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(0, time.January, 1, 17, 0, 0, 0, time.UTC)
	hours := make(domain.BusinessHours, 7)

	for day := 0; day <= 6; day++ {
		hours[day] = []domain.TimeSlot{{StartTime: start, EndTime: end}}
	}

	return hours
}

func (s *Service) GetAll(ctx context.Context) ([]domain.Location, error) {
	actor := actor.MustGetFromContext(ctx)

	locations, err := s.locationRepo.GetLocations(ctx, actor.MerchantId)
	if err != nil {
		return nil, err
	}

	return locations, nil
}

func (s *Service) Update(ctx context.Context, locationId int, input UpdateInput) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.locationRepo.UpdateLocation(ctx, domain.Location{
		Id:                locationId,
		MerchantId:        actor.MerchantId,
		Name:              input.Name,
		ContactEmail:      input.ContactEmail,
		PhoneNumber:       input.PhoneNumber,
		Country:           input.Country,
		CountryCode:       input.CountryCode,
		Region:            input.Region,
		City:              input.City,
		PostalCode:        input.PostalCode,
		Address:           input.Address,
		GeoPoint:          input.GeoPoint,
		GeocodingProvider: input.GeocodingProvider,
		ProviderPlaceId:   input.ProviderPlaceId,
		FormattedLocation: input.FormattedLocation,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLocationNotFound
		}

		return err
	}

	return nil
}
