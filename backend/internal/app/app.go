package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
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
	"github.com/miketsu-inc/reservations/backend/internal/jobs/workers"
	repos "github.com/miketsu-inc/reservations/backend/internal/repository/db"
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
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/queue"
	"github.com/riverqueue/river"
)

type App struct {
	server     *http.Server
	dbConn     *pgxpool.Pool
	jobsClient *river.Client[pgx.Tx]
}

func New(ctx context.Context, cfg *config.Config) *App {
	dbConn := db.New(ctx, RegisterTypes)

	blockedTimeRepo := repos.NewBlockedTimeRepository(dbConn)
	bookingRepo := repos.NewBookingRepository(dbConn)
	catalogRepo := repos.NewCatalogRepository(dbConn)
	customerRep := repos.NewCustomerRepository(dbConn)
	externalCalendarRepo := repos.NewExternalCalendarRepository(dbConn)
	merchantRepo := repos.NewMerchantRepository(dbConn)
	productRepo := repos.NewProductRepository(dbConn)
	teamRepo := repos.NewTeamRepository(dbConn)
	userRepo := repos.NewUserRepository(dbConn)

	transactionManager := db.NewTransactionManager(dbConn)

	emailService := emailSrv.NewService(cfg.RESEND_API_TEST, cfg.ENABLE_EMAILS)
	authService := authSrv.NewService(merchantRepo, userRepo, teamRepo, transactionManager)
	catalogService := catalog.NewService(catalogRepo, merchantRepo, transactionManager)
	blockedTimeService := blockedtimeSrv.NewService(blockedTimeRepo)
	bookingService := bookingSrv.NewService(bookingRepo, catalogRepo, merchantRepo, userRepo, customerRep, blockedTimeRepo, emailService, nil, transactionManager)
	customerService := customerSrv.NewService(customerRep, bookingRepo, transactionManager)
	externalCalendarService := externalcalendarSrv.NewService(externalCalendarRepo, blockedTimeRepo, merchantRepo, bookingRepo, teamRepo, transactionManager)
	merchantService := merchantSrv.NewService(bookingRepo, catalogRepo, merchantRepo, customerRep, blockedTimeRepo, teamRepo, productRepo, transactionManager)
	productService := productSrv.NewService(productRepo, merchantRepo)
	teamService := teamSrv.NewService(teamRepo)

	enqueuer, err := queue.NewClient(dbConn, workers.Deps{
		EmailService:   emailService,
		BookingService: bookingService,
		BookingRepo:    bookingRepo,
		TxManager:      transactionManager,
	}, workers.RegisterWorkers)
	assert.Nil(err, "Failed to create new river client")

	bookingService.SetEnqueuer(enqueuer)

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
		server:     srv,
		dbConn:     dbConn,
		jobsClient: enqueuer,
	}
}

func (a *App) Start(ctx context.Context) error {
	err := a.jobsClient.Start(ctx)
	if err != nil {
		return err
	}

	return a.server.ListenAndServe()
}

func (a *App) Stop(ctx context.Context) {
	a.jobsClient.Stop(ctx)
	a.dbConn.Close()
}

func RegisterTypes(ctx context.Context, conn *pgx.Conn) error {
	types, err := conn.LoadTypes(ctx, []string{"price", "_price"})
	if err != nil {
		return err
	}

	conn.TypeMap().RegisterTypes(types)

	conn.TypeMap().RegisterDefaultPgType(currencyx.Price{}, "price")
	conn.TypeMap().RegisterDefaultPgType([]currencyx.Price{}, "_price")

	return nil
}
