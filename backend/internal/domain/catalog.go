package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type CatalogRepository interface {
	WithTx(tx db.DBTX) CatalogRepository

	NewService(ctx context.Context, service Service) (int, error)
	// returns the old category id if service was under one
	UpdateService(ctx context.Context, service Service) (*int, error)
	DeleteService(ctx context.Context, merchantId uuid.UUID, serviceId int) error

	DeactivateService(ctx context.Context, merchantId uuid.UUID, serviceId int) error
	ActivateService(ctx context.Context, merchantId uuid.UUID, serviceId int) error
	ReorderServices(ctx context.Context, merchantId uuid.UUID, categoryId *int, serviceIds []int) error
	ReorderServicesAfterUpdate(ctx context.Context, categoryId *int, merchantId uuid.UUID, exludeServiceId *int) error

	GetServicesGroupedByCategory(ctx context.Context, merchantId uuid.UUID) ([]ServicesGroupedByCategory, error)
	GetServicesForCalendar(ctx context.Context, merchantId uuid.UUID) ([]ServicesGroupedByCategoriesForCalendar, error)
	GetServiceWithPhases(ctx context.Context, serviceId int, merchantId uuid.UUID) (Service, error)
	GetServicesForMerchantPage(ctx context.Context, merchantId uuid.UUID) ([]MerchantPageServicesGroupedByCategory, error)
	GetServiceDetailsForMerchantPage(ctx context.Context, merchantId uuid.UUID, serviceId int, locationId int) (PublicServiceDetails, error)
	GetAllServicePageData(ctx context.Context, serviceId int, merchantId uuid.UUID) (ServicePageData, error)
	GetServicePageFormOptions(ctx context.Context, merchantId uuid.UUID) (ServicePageFormOptions, error)
	GetMinimalServiceInfo(ctx context.Context, merchantId uuid.UUID, serviceId int, locationId int) (MinimalServiceInfo, error)
	GetServiceCancelDeadline(ctx context.Context, merchantId uuid.UUID, serviceId int) (int, error)

	NewServicePhases(ctx context.Context, serviceId int, servicePhases []ServicePhase) error
	UpdateServicePhases(ctx context.Context, servicePhases []ServicePhase) error
	UpdateServicePhaseDuration(ctx context.Context, serviceId int, duration int) error
	DeleteServicePhases(ctx context.Context, phaseIds []int) error
	DeleteServicePhasesForService(ctx context.Context, serviceId int) error
	GetServicePhases(ctx context.Context, serviceId int) ([]ServicePhase, error)

	NewServiceCategory(ctx context.Context, merchantId uuid.UUID, serviceCategory ServiceCategory) error
	UpdateServiceCategory(ctx context.Context, merchantId uuid.UUID, serviceCategory ServiceCategory) error
	DeleteServiceCategory(ctx context.Context, merchantId uuid.UUID, serviceCategoryId int) error
	ReorderServiceCategories(ctx context.Context, merchantId uuid.UUID, categoryIds []int) error

	NewServiceProduct(ctx context.Context, merchantId uuid.UUID, connectedProducts []ConnectedProducts) error
	UpdateServiceProducts(ctx context.Context, serviceId int, connectedProducts []ConnectedProducts) error
	DeleteServiceProducts(ctx context.Context, serviceId int, productIds []int) error
	GetServiceProducts(ctx context.Context, serviceId int) ([]ConnectedProducts, error)
}

type Service struct {
	Id               int                 `db:"id" json:"id"`
	MerchantId       uuid.UUID           `db:"merchant_id" json:"merchant_id"`
	CategoryId       *int                `db:"category_id" json:"category_id"`
	BookingType      types.BookingType   `db:"booking_type" json:"booking_type"`
	Name             string              `db:"name" json:"name"`
	Description      *string             `db:"description" json:"description"`
	Color            string              `db:"color" json:"color"`
	TotalDuration    int                 `db:"total_duration" json:"total_duration"`
	Price            *currencyx.Price    `db:"price_per_person" json:"price_per_person"`
	PriceType        types.PriceType     `db:"price_type" json:"price_type"`
	IsActive         bool                `db:"is_active" json:"is_active"`
	Sequence         int                 `db:"sequence" json:"sequence"`
	MinParticipants  int                 `db:"min_participants" json:"min_participants"`
	MaxParticipants  int                 `db:"max_participants" json:"max_participants"`
	CancelDeadline   *int                `db:"cancel_deadline" json:"cancel_deadline"`
	BookingWindowMin *int                `db:"booking_window_min" json:"booking_window_min"`
	BookingWindowMax *int                `db:"booking_window_max" json:"booking_window_max"`
	BufferTime       *int                `db:"buffer_time" json:"buffer_time"`
	ApprovalPolicy   *types.ApprovalType `db:"approval_policy" json:"approval_policy"`
	DeletedOn        *time.Time          `db:"deleted_on" json:"deleted_on"`
	// for convenience we do not really query the service without the phases anyway
	Phases []ServicePhase
}

func (s *Service) IsGroupService() bool {
	return s.BookingType == types.BookingTypeClass || s.BookingType == types.BookingTypeEvent
}

func (s *Service) GetTotalDuration() time.Duration {
	return time.Duration(s.TotalDuration) * time.Minute
}

func (s *Service) CalculateNewBookingPhases(bookingId int, startTime time.Time) []BookingPhase {
	if len(s.Phases) == 0 {
		return []BookingPhase{}
	}

	bookingPhases := make([]BookingPhase, len(s.Phases))
	bookingStart := startTime

	for i, phase := range s.Phases {
		bookingEnd := bookingStart.Add(phase.GetDuration())

		bookingPhases[i] = BookingPhase{
			BookingId:      bookingId,
			ServicePhaseId: phase.Id,
			FromDate:       bookingStart,
			ToDate:         bookingEnd,
		}

		bookingStart = bookingEnd
	}

	return bookingPhases
}

type ServicePhase struct {
	Id        int                    `db:"id" json:"id"`
	ServiceId int                    `db:"service_id" json:"service_id"`
	Name      string                 `db:"name" json:"name"`
	Sequence  int                    `db:"sequence" json:"sequence"`
	Duration  int                    `db:"duration" json:"duration"`
	PhaseType types.ServicePhaseType `db:"phase_type" json:"phase_type"`
	DeletedOn *time.Time             `db:"deleted_on" json:"deleted_on"`
}

func (sp *ServicePhase) IsEqual(phase ServicePhase) bool {
	return sp.Id == phase.Id &&
		sp.ServiceId == phase.ServiceId &&
		sp.Name == phase.Name &&
		sp.Sequence == phase.Sequence &&
		sp.Duration == phase.Duration &&
		sp.PhaseType == phase.PhaseType
}

func (sp *ServicePhase) GetDuration() time.Duration {
	return time.Duration(sp.Duration) * time.Minute
}

type ServiceCategory struct {
	Id         int       `db:"id"`
	MerchantId uuid.UUID `db:"merchant_id"`
	LocationId int       `db:"location_id"`
	Name       string    `db:"name"`
	Sequence   int       `db:"sequence"`
}

type ServiceSettings struct {
	CancelDeadline   *int                `json:"cancel_deadline"`
	BookingWindowMin *int                `json:"booking_window_min"`
	BookingWindowMax *int                `json:"booking_window_max"`
	BufferTime       *int                `json:"buffer_time"`
	ApprovalPolicy   *types.ApprovalType `json:"approval_policy"`
}

type ServicesGroupedByCategory struct {
	Id       *int      `json:"id"`
	Name     *string   `json:"name"`
	Sequence *int      `json:"sequence"`
	Services []Service `json:"services"`
}

type MerchantPageService struct {
	Id              int               `json:"id"`
	CategoryId      *int              `json:"category_id"`
	Name            string            `json:"name"`
	Description     *string           `json:"description"`
	TotalDuration   int               `json:"total_duration"`
	Price           *currencyx.Price  `json:"price"`
	PriceType       types.PriceType   `json:"price_type"`
	MaxParticipants int               `json:"max_participants"`
	BookingType     types.BookingType `json:"booking_type"`
	Sequence        int               `json:"sequence"`
}

type MerchantPageServicesGroupedByCategory struct {
	Id       *int                  `json:"id"`
	Name     *string               `json:"name"`
	Sequence *int                  `json:"sequence"`
	Services []MerchantPageService `json:"services"`
}

type ServicePageData struct {
	Id              int                           `db:"id"`
	BookingType     types.BookingType             `db:"booking_type"`
	CategoryId      *int                          `db:"category_id"`
	Name            string                        `db:"name"`
	Description     *string                       `db:"description"`
	Color           string                        `db:"color"`
	TotalDuration   int                           `db:"total_duration"`
	Price           *currencyx.Price              `db:"price_per_person"`
	PriceType       types.PriceType               `db:"price_type"`
	IsActive        bool                          `db:"is_active"`
	Sequence        int                           `db:"sequence"`
	MinParicipants  int                           `db:"min_participants"`
	MaxParticipants int                           `db:"max_participants"`
	Settings        ServiceSettings               `db:"settings"`
	Phases          []ServicePhase                `db:"phases"`
	Products        []MinimalProductInfoWithUsage `db:"used_products"`
}

type ServicePageFormOptions struct {
	Products   []MinimalProductInfo `json:"products"`
	Categories []ServiceCategory    `json:"categories"`
}

type ConnectedProducts struct {
	ProductId  int `json:"product_id"`
	ServiceId  int `json:"service_id"`
	AmountUsed int `json:"amount_used"`
}

type ServiceInfoForProducts struct {
	Id    int    `json:"id" db:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type PublicServiceDetails struct {
	Id                int              `json:"id"`
	Name              string           `json:"name"`
	Description       *string          `json:"description"`
	TotalDuration     int              `json:"total_duration"`
	Price             *currencyx.Price `json:"price"`
	PriceType         types.PriceType  `json:"price_type"`
	FormattedLocation string           `json:"formatted_location"`
	GeoPoint          types.GeoPoint   `json:"geo_point"`
	Phases            []ServicePhase   `json:"phases"`
}

type MinimalServiceInfo struct {
	Name              string           `json:"name"`
	TotalDuration     int              `json:"total_duration"`
	Price             *currencyx.Price `json:"price"`
	PriceType         types.PriceType  `json:"price_type"`
	FormattedLocation string           `json:"formatted_location"`
}

type ServicesGroupedByCategoriesForCalendar struct {
	Id       *int              `json:"id"`
	Name     *string           `json:"name"`
	Services []CalendarService `json:"services"`
}

type CalendarService struct {
	Id              int               `json:"id"`
	Name            string            `json:"name"`
	Duration        int               `json:"duration"`
	Price           *currencyx.Price  `json:"price"`
	PriceType       types.PriceType   `json:"price_type"`
	Color           string            `json:"color"`
	BookingType     types.BookingType `json:"booking_type"`
	MaxParticipants int               `json:"max_participants"`
}
