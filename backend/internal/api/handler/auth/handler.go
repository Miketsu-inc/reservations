package auth

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	authServ "github.com/miketsu-inc/reservations/backend/internal/service/auth"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/oauthutil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service    *authServ.Service
	middleware *middleware.Manager
}

func NewHandler(s *authServ.Service, m *middleware.Manager) *Handler {
	return &Handler{service: s, middleware: m}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(h.middleware.Language)

		r.Post("/login", h.Login)

		r.Post("/users", h.UserSignup)

		r.Get("/oauth/google", h.GoogleLogin)
		r.Get("/oauth/google/callback", h.GoogleCallback)
		r.Get("/oauth/facebook", h.FacebookLogin)
		r.Get("/oauth/facebook/callback", h.FacebookCallback)
	})

	r.Group(func(r chi.Router) {
		r.Use(h.middleware.Authentication)
		r.Use(h.middleware.Language)

		// TODO: instead of just checking if authenticated it should
		// return some basic info my user
		r.Get("/me", h.Me)

		r.Post("/logout", h.Logout)
		r.Post("/logout/all", h.LogoutAllDevices)

		r.Post("/merchants", h.MerchantSignup)
	})

	return r
}

type loginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,ascii"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	tokens, err := h.service.Login(r.Context(), mapToLoginInput(req))
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, err)
		return
	}

	jwt.SetJwtCookie(w, jwt.AccessToken, tokens.AccessToken)
	jwt.SetJwtCookie(w, jwt.RefreshToken, tokens.RefreshToken)
}

type userSignupReq struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
	Password    string `json:"password" validate:"required,ascii"`
}

func (h *Handler) UserSignup(w http.ResponseWriter, r *http.Request) {
	var req userSignupReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	tokens, err := h.service.UserSignup(r.Context(), mapToUserSignupInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	jwt.SetJwtCookie(w, jwt.AccessToken, tokens.AccessToken)
	jwt.SetJwtCookie(w, jwt.RefreshToken, tokens.RefreshToken)

	w.WriteHeader(http.StatusCreated)
}

type merchantSignupReq struct {
	Name         string `json:"name" validate:"required"`
	ContactEmail string `json:"contact_email" validate:"required,email"`
	Timezone     string `json:"timezone" validate:"required,timezone"`
}

func (h *Handler) MerchantSignup(w http.ResponseWriter, r *http.Request) {
	var req merchantSignupReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.MerchantSignup(r.Context(), mapToMerchantSignupInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	jwt.DeleteJwts(w)
}

func (h *Handler) LogoutAllDevices(w http.ResponseWriter, r *http.Request) {
	err := h.service.LogoutAllDevices(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	jwt.DeleteJwts(w)
}

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url, state, err := h.service.GoogleLogin(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	oauthutil.SetOauthStateCookie(w, state)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	if err := oauthutil.ValidateOauthState(r); err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error during oauth state validation: %s", err.Error()))
		return
	}

	tokens, err := h.service.GoogleCallback(r.Context(), code)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	jwt.SetJwtCookie(w, jwt.AccessToken, tokens.AccessToken)
	jwt.SetJwtCookie(w, jwt.RefreshToken, tokens.RefreshToken)

	http.Redirect(w, r, "http://localhost:8080/", http.StatusPermanentRedirect)
}

func (h *Handler) FacebookLogin(w http.ResponseWriter, r *http.Request) {
	url, state, err := h.service.FacebookLogin(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	oauthutil.SetOauthStateCookie(w, state)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *Handler) FacebookCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	if err := oauthutil.ValidateOauthState(r); err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error during oauth state validation: %s", err.Error()))
		return
	}

	tokens, err := h.service.FacebookCallback(r.Context(), code)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	jwt.SetJwtCookie(w, jwt.AccessToken, tokens.AccessToken)
	jwt.SetJwtCookie(w, jwt.RefreshToken, tokens.RefreshToken)

	http.Redirect(w, r, "http://localhost:8080/", http.StatusPermanentRedirect)
}
