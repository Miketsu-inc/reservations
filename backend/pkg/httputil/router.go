package httputil

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/logger"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

type Router struct {
	chi.Router
}

func NewRouter() *Router {
	return &Router{chi.NewRouter()}
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func Handle(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			renderError(w, r, err)
		}
	}
}

func (r *Router) Connect(pattern string, h HandlerFunc) { r.Router.Connect(pattern, Handle(h)) }
func (r *Router) Delete(pattern string, h HandlerFunc)  { r.Router.Delete(pattern, Handle(h)) }
func (r *Router) Get(pattern string, h HandlerFunc)     { r.Router.Get(pattern, Handle(h)) }
func (r *Router) Head(pattern string, h HandlerFunc)    { r.Router.Head(pattern, Handle(h)) }
func (r *Router) Options(pattern string, h HandlerFunc) { r.Router.Options(pattern, Handle(h)) }
func (r *Router) Patch(pattern string, h HandlerFunc)   { r.Router.Patch(pattern, Handle(h)) }
func (r *Router) Post(pattern string, h HandlerFunc)    { r.Router.Post(pattern, Handle(h)) }
func (r *Router) Put(pattern string, h HandlerFunc)     { r.Router.Put(pattern, Handle(h)) }
func (r *Router) Trace(pattern string, h HandlerFunc)   { r.Router.Trace(pattern, Handle(h)) }
func (r *Router) Query(pattern string, h HandlerFunc)   { r.Router.Query(pattern, Handle(h)) }

func (r *Router) Group(fn func(r *Router)) {
	r.Router.Group(func(inner chi.Router) {
		fn(&Router{inner})
	})
}

func (r *Router) Route(pattern string, fn func(r *Router)) {
	r.Router.Route(pattern, func(inner chi.Router) {
		fn(&Router{inner})
	})
}

type MiddlewareFunc func(http.Handler) HandlerFunc

func (r *Router) UseFunc(m MiddlewareFunc) {
	r.Use(func(next http.Handler) http.Handler {
		return Handle(m(next))
	})
}

func (r *Router) NotFound(h HandlerFunc) {
	r.Router.NotFound(Handle(h))
}

func (r *Router) MethodNotAllowed(h HandlerFunc) {
	r.Router.MethodNotAllowed(Handle(h))
}

type errorResp struct {
	Status  int            `json:"status"`
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Meta    map[string]any `json:"meta,omitempty"`
}

func renderError(w http.ResponseWriter, r *http.Request, err error) {
	logger := logger.FromContext(r.Context())

	var apiErr *apperr.APIError
	if !errors.As(err, &apiErr) {
		logger.Error("unexpected error", "error", err.Error())
		WriteJSON(w, http.StatusInternalServerError, map[string]errorResp{
			"error": {
				Status:  http.StatusInternalServerError,
				Code:    "internal_server_error",
				Message: "An unexpected error occurred",
			},
		})
		return
	}

	attrs := []any{
		"code", apiErr.Err.Code,
		"error", apiErr.Error(),
		"status", apiErr.Status,
	}
	if apiErr.Cause != nil {
		attrs = append(attrs, "cause", apiErr.Cause.Error())
	}
	if len(apiErr.Meta) > 0 {
		attrs = append(attrs, "meta", apiErr.Meta)
	}
	logger.Error("api error", attrs...)

	WriteJSON(w, apiErr.Status, map[string]errorResp{
		"error": {
			Status:  apiErr.Status,
			Code:    apiErr.Err.Code,
			Message: apiErr.Error(),
			Meta:    apiErr.Meta,
		},
	})
}
