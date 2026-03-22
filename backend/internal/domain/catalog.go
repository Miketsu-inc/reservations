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

	GetServices(ctx context.Context, merchantId uuid.UUID) ([]ServicesGroupedByCategory, error)
	GetServicesForCalendar(ctx context.Context, merchantId uuid.UUID) ([]ServicesGroupedByCategoriesForCalendar, error)
	GetServiceWithPhases(ctx context.Context, serviceId int, merchantId uuid.UUID) (PublicServiceWithPhases, error)
	GetServicesForMerchantPage(ctx context.Context, merchantId uuid.UUID) ([]MerchantPageServicesGroupedByCategory, error)
	GetServiceDetailsForMerchantPage(ctx context.Context, merchantId uuid.UUID, serviceId int, locationId int) (PublicServiceDetails, error)
	GetAllServicePageData(ctx context.Context, serviceId int, merchantId uuid.UUID) (ServicePageData, error)
	GetGroupServicePageData(ctx context.Context, merchantId uuid.UUID, serviceId int) (GroupServicePageData, error)
	GetServicePageFormOptions(ctx context.Context, merchantId uuid.UUID) (ServicePageFormOptions, error)
	GetMinimalServiceInfo(ctx context.Context, merchantId uuid.UUID, serviceId int, locationId int) (MinimalServiceInfo, error)

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
	Id              int               `json:"ID"`
	MerchantId      uuid.UUID         `json:"merchant_id"`
	CategoryId      *int              `json:"category_id"`
	BookingType     types.BookingType `json:"booking_type"`
	Name            string            `json:"name"`
	Description     *string           `json:"description"`
	Color           string            `json:"color"`
	TotalDuration   int               `json:"total_duration"`
	Price           *currencyx.Price  `json:"price"`
	Cost            *currencyx.Price  `json:"cost"`
	PriceType       types.PriceType   `json:"price_type"`
	IsActive        bool              `json:"is_active"`
	Sequence        int               `json:"sequence"`
	MinParticipants int               `json:"min_participants"`
	MaxParticipants int               `json:"max_participants"`
	ServiceSettings
	DeletedOn *time.Time `json:"deleted_on"`
}

type ServicePhase struct {
	Id        int                    `json:"ID"`
	ServiceId int                    `json:"service_id"`
	Name      string                 `json:"name"`
	Sequence  int                    `json:"sequence"`
	Duration  int                    `json:"duration"`
	PhaseType types.ServicePhaseType `json:"phase_type"`
	DeletedOn *time.Time             `json:"deleted_on"`
}

type ServiceCategory struct {
	Id         int       `json:"id" db:"id"`
	MerchantId uuid.UUID `json:"merchant_id"`
	LocationId int       `json:"location_id"`
	Name       string    `json:"name" db:"name"`
	Sequence   int       `json:"sequence"`
}

type ServiceSettings struct {
	CancelDeadline   *int `json:"cancel_deadline"`
	BookingWindowMin *int `json:"booking_window_min"`
	BookingWindowMax *int `json:"booking_window_max"`
	BufferTime       *int `json:"buffer_time"`
}

type PublicServicePhase struct {
	Id        int                    `json:"id" db:"id"`
	ServiceId int                    `json:"service_id" db:"service_id"`
	Name      string                 `json:"name" db:"name"`
	Sequence  int                    `json:"sequence" db:"sequence"`
	Duration  int                    `json:"duration" db:"duration"`
	PhaseType types.ServicePhaseType `json:"phase_type" db:"phase_type"`
}

type PublicServiceWithPhases struct {
	Id              int                  `json:"id"`
	MerchantId      uuid.UUID            `json:"merchant_id"`
	BookingType     types.BookingType    `json:"booking_type"`
	CategoryId      *int                 `json:"category_id"`
	Name            string               `json:"name"`
	Description     *string              `json:"description"`
	Color           string               `json:"color"`
	TotalDuration   int                  `json:"total_duration"`
	Price           *currencyx.Price     `json:"price"`
	Cost            *currencyx.Price     `json:"cost"`
	PriceType       types.PriceType      `json:"price_type"`
	IsActive        bool                 `json:"is_active"`
	MinParticipants int                  `json:"min_participants"`
	MaxParticipants int                  `json:"max_participants"`
	Sequence        int                  `json:"sequence"`
	Phases          []PublicServicePhase `json:"phases"`
}

type ServicesGroupedByCategory struct {
	Id       *int                      `json:"id"`
	Name     *string                   `json:"name"`
	Sequence *int                      `json:"sequence"`
	Services []PublicServiceWithPhases `json:"services"`
}

type MerchantPageService struct {
	Id            int              `json:"id"`
	CategoryId    *int             `json:"category_id"`
	Name          string           `json:"name"`
	Description   *string          `json:"description"`
	TotalDuration int              `json:"total_duration"`
	Price         *currencyx.Price `json:"price"`
	PriceType     types.PriceType  `json:"price_type"`
	Sequence      int              `json:"sequence"`
}

type MerchantPageServicesGroupedByCategory struct {
	Id       *int                  `json:"id"`
	Name     *string               `json:"name"`
	Sequence *int                  `json:"sequence"`
	Services []MerchantPageService `json:"services"`
}

type ServicePageData struct {
	Id            int                           `json:"id"`
	CategoryId    *int                          `json:"category_id"`
	Name          string                        `json:"name"`
	Description   *string                       `json:"description"`
	Color         string                        `json:"color"`
	TotalDuration int                           `json:"total_duration"`
	Price         *currencyx.Price              `json:"price"`
	Cost          *currencyx.Price              `json:"cost"`
	PriceType     types.PriceType               `json:"price_type"`
	IsActive      bool                          `json:"is_active"`
	Sequence      int                           `json:"sequence"`
	Settings      ServiceSettings               `json:"settings"`
	Phases        []PublicServicePhase          `json:"phases"`
	Products      []MinimalProductInfoWithUsage `json:"used_products"`
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
	Id                int                  `json:"id"`
	Name              string               `json:"name"`
	Description       *string              `json:"description"`
	TotalDuration     int                  `json:"total_duration"`
	Price             *currencyx.Price     `json:"price"`
	PriceType         types.PriceType      `json:"price_type"`
	FormattedLocation string               `json:"formatted_location"`
	GeoPoint          types.GeoPoint       `json:"geo_point"`
	Phases            []PublicServicePhase `json:"phases"`
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

type GroupServicePageData struct {
	Id              int                           `json:"id"`
	CategoryId      *int                          `json:"category_id"`
	Name            string                        `json:"name"`
	Description     *string                       `json:"description"`
	Color           string                        `json:"color"`
	Duration        int                           `json:"duration"`
	Price           *currencyx.Price              `json:"price"`
	Cost            *currencyx.Price              `json:"cost"`
	PriceType       types.PriceType               `json:"price_type"`
	IsActive        bool                          `json:"is_active"`
	Sequence        int                           `json:"sequence"`
	MinParicipants  int                           `json:"min_participants"`
	MaxParticipants int                           `json:"max_participants"`
	Settings        ServiceSettings               `json:"settings"`
	Products        []MinimalProductInfoWithUsage `json:"used_products"`
}
