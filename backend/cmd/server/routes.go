package server

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/handlers"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares"
	"github.com/miketsu-inc/reservations/frontend"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.Recoverer)

	staticFilesHandler(r)

	var rh = RouteHandlers{&s.db}
	r.Route("/api/v1/auth", rh.AuthRoutes)

	// routes which requires auth
	r.Group(func(r chi.Router) {
		r.Use(middlewares.JWTMiddleware)

		r.Route("/api/v1/appointments", rh.appointmentRoutes)
	})

	return r
}

func staticFilesHandler(r *chi.Mux) {
	frontendRoutes := []string{
		"/",
		"/login",
		"/signup",
		"/dashboard",
		"/reservations",
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
}

type RouteHandlers struct {
	Postgresdb *database.PostgreSQL
}

func (rh *RouteHandlers) appointmentRoutes(r chi.Router) {
	appointmentHandler := &handlers.Appointment{
		Postgresdb: *rh.Postgresdb,
	}

	r.Post("/", appointmentHandler.Create)
	r.Get("/calendar", appointmentHandler.GetEvents)
}

func (rh *RouteHandlers) AuthRoutes(r chi.Router) {
	authHandler := &handlers.Auth{
		Postgresdb: *rh.Postgresdb,
	}

	r.Post("/signup", authHandler.HandleSignup)
	r.Post("/login", authHandler.HandleLogin)
}
