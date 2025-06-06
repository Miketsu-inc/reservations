package server

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/handlers"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/frontend"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.AllowContentType("application/json"))
	// r.Use(middleware.Recoverer)

	staticFilesHandler(r)

	var rh = RouteHandlers{&s.db}
	r.Route("/api/v1/auth/user", rh.userAuthRoutes)
	r.Route("/api/v1/auth/merchant", rh.merchantAuthRoutes)

	r.Route("/api/v1/merchants", rh.merchantRoutes)
	r.Route("/api/v1/appointments", rh.appointmentRoutes)

	return r
}

func staticFilesHandler(r *chi.Mux) {
	frontendRoutes := []string{
		"/",
		"/login",
		"/signup",
		"/calendar",
		"/settings/profile",
		"/settings/merchant",
		"/settings/billing",
		"/settings/calendar",
		"/services",
		"/services/new",
		"/services/edit/{id}",
		"/customers",
		"/products",
		"/dashboard",
		"/merchantsignup",
		"/m/{merchant_url}",
		"/m/{merchant_url}/booking",
		"/m/{merchant_url}/booking/completed",
		"/m/{merchant_url}/cancel/{appointmentId}",
		"/m/{merchant_url}/cancel/{appointmentId}/completed",
	}

	dist, assets := frontend.StaticFilesPath()

	for _, route := range frontendRoutes {
		r.Get(route, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFileFS(w, r, dist, "index.html")
		})
	}

	r.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/assets/", http.FileServerFS(assets)).ServeHTTP(w, r)
	})

	r.Get("/theme.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, dist, "theme.js")
	})
}

type RouteHandlers struct {
	Postgresdb *database.PostgreSQL
}

func (rh *RouteHandlers) appointmentRoutes(r chi.Router) {
	appointmentHandler := &handlers.Appointment{
		Postgresdb: *rh.Postgresdb,
	}

	r.Group(func(r chi.Router) {
		r.Use(jwt.JwtMiddleware)

		r.Delete("/{id}", appointmentHandler.CancelAppointmentByMerchant)
		r.Patch("/{id}", appointmentHandler.UpdateAppointmentData)

		r.Post("/new", appointmentHandler.Create)
		r.Get("/all", appointmentHandler.GetAppointments)

		r.Get("/public/{id}", appointmentHandler.GetPublicAppointmentData)
		r.Delete("/public/{id}", appointmentHandler.CancelAppointmentByUser)
	})
}

func (rh *RouteHandlers) userAuthRoutes(r chi.Router) {
	userAuthHandler := &handlers.UserAuth{
		Postgresdb: *rh.Postgresdb,
	}

	r.Post("/signup", userAuthHandler.Signup)
	r.Post("/login", userAuthHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(jwt.JwtMiddleware)

		r.Get("/", userAuthHandler.IsAuthenticated)
		r.Post("/logout", userAuthHandler.Logout)
		r.Post("/logout/all", userAuthHandler.LogoutAllDevices)
	})
}

func (rh *RouteHandlers) merchantAuthRoutes(r chi.Router) {
	merchantAuthHandler := &handlers.MerchantAuth{
		Postgresdb: *rh.Postgresdb,
	}

	r.Group(func(r chi.Router) {
		r.Use(jwt.JwtMiddleware)

		r.Post("/signup", merchantAuthHandler.Signup)
	})
}

func (rh *RouteHandlers) merchantRoutes(r chi.Router) {
	merchantHandler := &handlers.Merchant{
		Postgresdb: *rh.Postgresdb,
	}

	r.Get("/info", merchantHandler.InfoByName)
	r.Get("/available-times", merchantHandler.GetHours)
	r.Get("/business-hours/closed", merchantHandler.GetClosedDays)

	r.Group(func(r chi.Router) {
		r.Use(jwt.JwtMiddleware)
		r.Get("/settings-info", merchantHandler.MerchantSettingsInfoByOwner)
		r.Patch("/reservation-fields", merchantHandler.UpdateMerchantFields)

		r.Get("/preferences", merchantHandler.GetPreferences)
		r.Patch("/preferences", merchantHandler.UpdatePreferences)

		r.Post("/location", merchantHandler.NewLocation)
		r.Post("/check-url", merchantHandler.CheckUrl)

		r.Get("/services", merchantHandler.GetServices)
		r.Post("/services", merchantHandler.NewService)
		r.Get("/services/{id}", merchantHandler.GetService)
		r.Delete("/services/{id}", merchantHandler.DeleteService)
		r.Put("/services/{id}", merchantHandler.UpdateService)

		r.Post("/services/categories", merchantHandler.NewServiceCategory)

		r.Get("/customers", merchantHandler.GetCustomers)
		r.Post("/customers", merchantHandler.NewCustomer)
		r.Delete("/customers/{id}", merchantHandler.DeleteCustomer)
		r.Put("/customers/{id}", merchantHandler.UpdateCustomer)
		r.Put("/customers/transfer", merchantHandler.TransferCustomerApps)
		r.Post("/customers/blacklist/{id}", merchantHandler.BlacklistCustomer)
		r.Delete("/customers/blacklist/{id}", merchantHandler.UnBlacklistCustomer)

		r.Get("/products", merchantHandler.GetProducts)
		r.Post("/products", merchantHandler.NewProduct)
		r.Delete("/products/{id}", merchantHandler.DeleteProduct)
		r.Put("/products/{id}", merchantHandler.UpdateProduct)
		r.Get("/business-hours", merchantHandler.GetBusinessHours)

		r.Get("/dashboard", merchantHandler.GetDashboardData)
	})
}
