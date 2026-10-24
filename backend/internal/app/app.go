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
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/favorites"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/integrations"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/invitations"
	mediaHandler "github.com/miketsu-inc/reservations/backend/internal/api/handler/media"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchants"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchants/blockedtimes"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchants/blockedtimetypes"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchants/bookings"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchants/customers"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchants/locations"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchants/products"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchants/servicecategories"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchants/services"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/merchants/team"
	publicBookings "github.com/miketsu-inc/reservations/backend/internal/api/handler/public/bookings"
	publicMerchants "github.com/miketsu-inc/reservations/backend/internal/api/handler/public/merchants"
	"github.com/miketsu-inc/reservations/backend/internal/api/handler/users"
	"github.com/miketsu-inc/reservations/backend/internal/api/merchantpage"
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
	favoriteSrv "github.com/miketsu-inc/reservations/backend/internal/service/favorite"
	locationSrv "github.com/miketsu-inc/reservations/backend/internal/service/location"
	mediaSrv "github.com/miketsu-inc/reservations/backend/internal/service/media"
	merchantSrv "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	productSrv "github.com/miketsu-inc/reservations/backend/internal/service/product"
	teamSrv "github.com/miketsu-inc/reservations/backend/internal/service/team"
	userSrv "github.com/miketsu-inc/reservations/backend/internal/service/user"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/imgproxy"
	"github.com/miketsu-inc/reservations/backend/pkg/kv"
	"github.com/miketsu-inc/reservations/backend/pkg/objectstore"
	"github.com/miketsu-inc/reservations/backend/pkg/queue"
	"github.com/redis/go-redis/v9"
	"github.com/riverqueue/river"
)

type App struct {
	server     *http.Server
	dbConn     *pgxpool.Pool
	jobsClient *river.Client[pgx.Tx]
	kvClient   *redis.Client
}

func New(ctx context.Context, cfg *config.Config) *App {
	dbConn := db.New(ctx, registerTypes)

	blockedTimeRepo := repos.NewBlockedTimeRepository(dbConn)
	bookingRepo := repos.NewBookingRepository(dbConn)
	catalogRepo := repos.NewCatalogRepository(dbConn)
	customerRep := repos.NewCustomerRepository(dbConn)
	externalCalendarRepo := repos.NewExternalCalendarRepository(dbConn)
	favoriteRepo := repos.NewFavoriteRepository(dbConn)
	imageRepo := repos.NewImageRepository(dbConn)
	locationRepo := repos.NewLocationRepository(dbConn)
	merchantRepo := repos.NewMerchantRepository(dbConn)
	productRepo := repos.NewProductRepository(dbConn)
	teamRepo := repos.NewTeamRepository(dbConn)
	userRepo := repos.NewUserRepository(dbConn)

	transactionManager := db.NewTransactionManager(dbConn)

	kvClient := kv.NewClient()

	r2Client := objectstore.NewS3Client(cfg.R2_ENDPOINT, cfg.R2_ACCESS_KEY_ID, cfg.R2_SECRET_ACCESS_KEY)
	uploadsBucket := objectstore.NewS3Bucket(r2Client, cfg.R2_UPLOADS_BUCKET)
	mediaBucket := objectstore.NewS3Bucket(r2Client, cfg.R2_MEDIA_BUCKET)

	imgproxyClient := imgproxy.NewClient(fmt.Sprintf("http://localhost:%s", cfg.IMGPROXY_PORT), cfg.IMGPROXY_KEY, cfg.IMGPROXY_SALT)

	emailService := emailSrv.NewService(cfg.RESEND_API_TEST, cfg.ENABLE_EMAILS)
	authService := authSrv.NewService(merchantRepo, userRepo, teamRepo, kvClient, nil, transactionManager)
	teamService := teamSrv.NewService(teamRepo, userRepo, merchantRepo, nil, transactionManager)
	catalogService := catalog.NewService(catalogRepo, merchantRepo, teamService, transactionManager)
	blockedTimeService := blockedtimeSrv.NewService(blockedTimeRepo, teamRepo, teamService, nil, transactionManager)
	bookingService := bookingSrv.NewService(bookingRepo, catalogRepo, merchantRepo, locationRepo, teamRepo, userRepo, customerRep, blockedTimeRepo, emailService, nil, transactionManager)
	customerService := customerSrv.NewService(customerRep, bookingRepo, transactionManager)
	externalCalendarService := externalcalendarSrv.NewService(externalCalendarRepo, blockedTimeRepo, merchantRepo, bookingRepo, teamRepo, nil, transactionManager)
	merchantService := merchantSrv.NewService(bookingRepo, catalogRepo, locationRepo, merchantRepo, customerRep, blockedTimeRepo, teamRepo, imageRepo, transactionManager)
	locationService := locationSrv.NewService(locationRepo, merchantRepo, transactionManager)
	favoriteService := favoriteSrv.NewService(favoriteRepo)
	productService := productSrv.NewService(productRepo, merchantRepo)
	userService := userSrv.NewService(userRepo)
	mediaService := mediaSrv.NewService(imageRepo, teamRepo, locationRepo, merchantRepo, uploadsBucket, mediaBucket, imgproxyClient, nil, transactionManager)

	enqueuer, err := queue.NewClient(dbConn, workers.Deps{
		BookingService:     bookingService,
		EmailService:       emailService,
		ExtCalendarService: externalCalendarService,
		MediaService:       mediaService,
		BookingRepo:        bookingRepo,
		CatalogRepo:        catalogRepo,
		UserRepo:           userRepo,
		ExtCalendarRepo:    externalCalendarRepo,
		TeamRepo:           teamRepo,
		MerchantRepo:       merchantRepo,
		TxManager:          transactionManager,
	}, workers.RegisterWorkers, workers.GetPeriodicJobs())
	assert.Nil(err, "Failed to create new river client")

	authService.SetEnqueuer(enqueuer)
	teamService.SetEnqueuer(enqueuer)
	bookingService.SetEnqueuer(enqueuer)
	externalCalendarService.SetEnqueuer(enqueuer)
	blockedTimeService.SetEnqueuer(enqueuer)
	mediaService.SetEnqueuer(enqueuer)

	middlewareManager := middleware.NewManager(merchantRepo, userRepo)

	router := api.NewRouter(&api.Handlers{
		Auth:              auth.NewHandler(authService, teamService, middlewareManager),
		Bookings:          bookings.NewHandler(bookingService, middlewareManager),
		PublicBookings:    publicBookings.NewHandler(bookingService, middlewareManager),
		PublicMerchants:   publicMerchants.NewHandler(merchantService, catalogService, teamService, middlewareManager),
		MerchantPage:      merchantpage.NewHandler(merchantService),
		Merchants:         merchants.NewHandler(merchantService, bookingService, teamService, catalogService, customerService, externalCalendarService, mediaService),
		BlockedTimes:      blockedtimes.NewHandler(blockedTimeService),
		BlockedTimeTypes:  blockedtimetypes.NewHandler(blockedTimeService),
		Customers:         customers.NewHandler(customerService),
		Integrations:      integrations.NewHandler(externalCalendarService),
		Invitations:       invitations.NewHandler(teamService, middlewareManager),
		Media:             mediaHandler.NewHandler(mediaService, middlewareManager),
		Users:             users.NewHandler(userService, bookingService, authService, mediaService, middlewareManager),
		Favorites:         favorites.NewHandler(favoriteService, middlewareManager),
		Locations:         locations.NewHandler(locationService, mediaService, middlewareManager),
		Products:          products.NewHandler(productService),
		Services:          services.NewHandler(catalogService),
		ServiceCategories: servicecategories.NewHandler(catalogService),
		Team:              team.NewHandler(teamService, mediaService, middlewareManager),
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
		kvClient:   kvClient,
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
	_ = a.jobsClient.Stop(ctx)
	_ = a.kvClient.Close()
	a.dbConn.Close()
}

func registerTypes(ctx context.Context, conn *pgx.Conn) error {
	types, err := conn.LoadTypes(ctx, []string{"price", "_price"})
	if err != nil {
		return err
	}

	conn.TypeMap().RegisterTypes(types)

	conn.TypeMap().RegisterDefaultPgType(currencyx.Price{}, "price")
	conn.TypeMap().RegisterDefaultPgType([]currencyx.Price{}, "_price")

	return nil
}
