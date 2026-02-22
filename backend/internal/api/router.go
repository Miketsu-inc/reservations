package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
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
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/frontend/apps/jabulani"
	"github.com/miketsu-inc/reservations/frontend/apps/tango"
)

type Handlers struct {
	Auth              *auth.Handler
	Bookings          *bookings.Handler
	Merchants         *merchants.Handler
	Merchant          *merchant.Handler
	BlockedTimes      *blockedtimes.Handler
	BlockedTimeTypes  *blockedtimetypes.Handler
	Customers         *customers.Handler
	Integrations      *integrations.Handler
	Locations         *locations.Handler
	Products          *products.Handler
	Services          *services.Handler
	ServiceCategories *servicecategories.Handler
	Team              *team.Handler
	Middleware        *middleware.Manager
}

func NewRouter(h *Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.AllowContentType("application/json"))
	// r.Use(chiMiddleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/auth", h.Auth.Routes())
		r.Mount("/bookings", h.Bookings.Routes())
		r.Mount("/merchants/{merchantName}", h.Merchants.Routes())

		r.Route("/merchant", func(r chi.Router) {
			r.Use(h.Middleware.Authentication)
			r.Use(h.Middleware.Language)

			r.Group(func(r chi.Router) {
				r.Use(h.Middleware.RoleBasedAccessControl(types.EmployeeRoleOwner))

				r.Delete("/", h.Merchant.Delete)
				r.Patch("/name", h.Merchant.UpdateName)
			})

			r.Group(func(r chi.Router) {
				r.Use(h.Middleware.RoleBasedAccessControl(types.EmployeeRoleStaff, types.EmployeeRoleAdmin, types.EmployeeRoleOwner))

				r.Get("/dashboard", h.Merchant.GetDashboard)

				r.Post("/check-url", h.Merchant.CheckUrl)

				r.Get("/settings", h.Merchant.GetSettings)
				r.Patch("/settings", h.Merchant.UpdateSettings)
				r.Get("/settings/business-hours/normalized", h.Merchant.GetNormalizedBusinessHours)

				r.Get("/preferences", h.Merchant.GetPreferences)
				r.Patch("/preferences", h.Merchant.UpdatePreferences)

				r.Get("/calendar/team", h.Merchant.GetTeamMembersForCalendar)
				r.Get("/calendar/services", h.Merchant.GetServicesForCalendar)
				r.Get("/calendar/customers", h.Merchant.GetCustomersForCalendar)
			})

			r.Mount("/blocked-times", h.BlockedTimes.Routes())
			r.Mount("/blocked-time-types", h.BlockedTimeTypes.Routes())
			r.Mount("/customers", h.Customers.Routes())
			r.Mount("/integrations", h.Integrations.Routes())
			r.Mount("/locations", h.Locations.Routes())
			r.Mount("/products", h.Products.Routes())
			r.Mount("/services", h.Services.Routes())
			r.Mount("/service-categories", h.ServiceCategories.Routes())
			r.Mount("/team", h.Team.Routes())
		})
	})

	jabulani := jabulaniRouter()
	tango := tangoRouter()

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host

		if strings.HasPrefix(host, "app.") {
			jabulani.ServeHTTP(w, r)
		} else {
			tango.ServeHTTP(w, r)
		}
	})

	return r
}

func jabulaniRouter() chi.Router {
	r := chi.NewRouter()

	jabulaniRoutes := []string{
		"/",
		"/login",
		"/signup",
		"/calendar",
		"/settings/profile",
		"/settings/merchant",
		"/settings/billing",
		"/settings/calendar",
		"/settings/scheduling",
		"/services",
		"/services/new",
		"/services/edit/{id}",
		"/services/group/new",
		"/services/group/edit/{id}",
		"/customers",
		"/customers/blacklist",
		"/customers/new",
		"/customers/edit/{id}",
		"/customers/{customerId}",
		"/integrations",
		"/products",
		"/dashboard",
		"/signup",
	}

	dist, assets := jabulani.StaticFilesPath()

	for _, route := range jabulaniRoutes {
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

	return r
}

func tangoRouter() chi.Router {
	r := chi.NewRouter()

	tangoRoutes := []string{
		"/",
		"/login",
		"/signup",
		"/m/{merchant_url}",
		"/m/{merchant_url}/booking",
		"/m/{merchant_url}/booking/completed",
		"/m/{merchant_url}/services/{serviceId}",
		"/m/{merchant_url}/cancel/{bookingId}",
		"/m/{merchant_url}/cancel/{bookingId}/completed",
	}

	dist, assets := tango.StaticFilesPath()

	for _, route := range tangoRoutes {
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

	return r
}
