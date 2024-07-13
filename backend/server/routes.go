package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	})

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {

	})

	r.Get("/calendar", func(w http.ResponseWriter, r *http.Request) {

	})

	return r
}
