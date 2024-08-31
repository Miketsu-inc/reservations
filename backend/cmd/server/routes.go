package server

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/cmd/handlers"
	"github.com/miketsu-inc/reservations/frontend"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/api/v1/auth", userAuthRoutes)

	staticFilesHandler(r)

	r.Route("/api/v1/reservations", reservationRoutes)

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

func reservationRoutes(r chi.Router) {
	reservationHandler := &handlers.Reservation{}

	r.Post("/", reservationHandler.Create)
}

func userAuthRoutes(r chi.Router) {
	userauthHandler := &handlers.Auth{}

	r.Post("/signup", userauthHandler.HandleSignup)
	r.Post("/login", userauthHandler.HandleLogin)
}
