package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/internal/api"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/auth"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/bookings"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchant"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchant/blockedtimes"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchant/blockedtimetypes"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchant/customers"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchant/integrations"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchant/locations"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchant/products"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchant/servicecategories"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchant/services"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchant/team"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchants"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	"github.com/miketsu-inc/reservations/backend/internal/repository/db"
	authSrv "github.com/miketsu-inc/reservations/backend/internal/service/auth"
	blockedtimeSrv "github.com/miketsu-inc/reservations/backend/internal/service/blockedtime"
	bookingSrv "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	"github.com/miketsu-inc/reservations/backend/internal/service/catalog"
	customerSrv "github.com/miketsu-inc/reservations/backend/internal/service/customer"
	emailSrv "github.com/miketsu-inc/reservations/backend/internal/service/email"
	externalcalendarSrv "github.com/miketsu-inc/reservations/backend/internal/service/externalcalendar"
	merchantSrv "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	productSrv "github.com/miketsu-inc/reservations/backend/internal/service/product"
	teamSrv "github.com/miketsu-inc/reservations/backend/internal/service/team"
)

type App struct {
	server *http.Server
	dbConn *pgxpool.Pool
}

func New(cfg *config.Config) *App {
	dbConn := db.New()

	blockedTimeRepo := db.NewBlockedTimeRepository(dbConn)
	bookingRepo := db.NewBookingRepository(dbConn)
	catalogRepo := db.NewCatalogRepository(dbConn)
	customerRep := db.NewCustomerRepository(dbConn)
	externalCalendarRepo := db.NewExternalCalendarRepository(dbConn)
	merchantRepo := db.NewMerchantRepository(dbConn)
	productRepo := db.NewProductRepository(dbConn)
	teamRepo := db.NewTeamRepository(dbConn)
	userRepo := db.NewUserRepository(dbConn)

	emailService := emailSrv.NewService(cfg.RESEND_API_TEST, cfg.ENABLE_EMAILS)
	authService := authSrv.NewService(merchantRepo, userRepo)
	catalogService := catalog.NewService(catalogRepo, merchantRepo)
	blockedTimeService := blockedtimeSrv.NewService(blockedTimeRepo)
	bookingService := bookingSrv.NewService(bookingRepo, catalogRepo, merchantRepo, userRepo, customerRep, emailService)
	customerService := customerSrv.NewService(customerRep, bookingRepo)
	externalCalendarService := externalcalendarSrv.NewService(externalCalendarRepo, blockedTimeRepo, merchantRepo, bookingRepo, teamRepo)
	merchantService := merchantSrv.NewService(bookingRepo, catalogRepo, merchantRepo, customerRep, blockedTimeRepo, teamRepo)
	productService := productSrv.NewService(productRepo, merchantRepo)
	teamService := teamSrv.NewService(teamRepo)

	middlewareManager := middleware.NewManager(merchantRepo, userRepo)

	router := api.NewRouter(&api.Handlers{
		Auth:              auth.NewHandler(authService, middlewareManager),
		Bookings:          bookings.NewHandler(bookingService, middlewareManager),
		Merchants:         merchants.NewHandler(merchantService, middlewareManager),
		Merchant:          merchant.NewHandler(merchantService),
		BlockedTimes:      blockedtimes.NewHandler(blockedTimeService),
		BlockedTimeTypes:  blockedtimetypes.NewHandler(blockedTimeService),
		Customers:         customers.NewHandler(customerService),
		Integrations:      integrations.NewHandler(externalCalendarService),
		Locations:         locations.NewHandler(merchantService),
		Products:          products.NewHandler(productService),
		Services:          services.NewHandler(catalogService),
		ServiceCategories: servicecategories.NewHandler(catalogService),
		Team:              team.NewHandler(teamService),
		Middleware:        middlewareManager,
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.PORT),
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return &App{
		server: srv,
		dbConn: dbConn,
	}
}

func (a *App) Start() error {
	return a.server.ListenAndServe()
}

func (a *App) Stop() {
	a.dbConn.Close()
}
