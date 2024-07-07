package server

import (
	"fmt"
	"net/http"

	// "admin/cmd/web"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// r.Get("/", templ.Handler(web.LoginPage()).ServeHTTP)
	r.Post("/api/login", s.LoginHandler)

	// fileServer := http.FileServer(http.FS(web.Files))
	// r.Handle("/assets/*", fileServer)
	// r.Get("/web", templ.Handler(web.HelloForm()).ServeHTTP)
	// r.Post("/hello", web.HelloWebHandler)

	return r
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("works")
}
