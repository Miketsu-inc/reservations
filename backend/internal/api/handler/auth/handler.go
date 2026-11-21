package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	authServ "github.com/miketsu-inc/reservations/backend/internal/service/auth"
	teamServ "github.com/miketsu-inc/reservations/backend/internal/service/team"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/lang"
	"github.com/miketsu-inc/reservations/backend/pkg/oauthutil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
	"golang.org/x/text/language"
)

type Handler struct {
	service     *authServ.Service
	teamService *teamServ.Service
	middleware  *middleware.Manager
}

func NewHandler(s *authServ.Service, ts *teamServ.Service, m *middleware.Manager) *Handler {
	return &Handler{service: s, teamService: ts, middleware: m}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Post("/login", h.Login)
	r.Post("/forgot-password", h.ForgotPassword)
	r.Post("/reset-password", h.ResetPassword)

	r.Post("/users", h.UserSignup)

	r.Get("/oauth/google", h.GoogleLogin)
	r.Get("/oauth/google/callback", h.GoogleCallback)
	r.Get("/oauth/facebook", h.FacebookLogin)
	r.Get("/oauth/facebook/callback", h.FacebookCallback)

	r.Group(func(r *httputil.Router) {
		r.UseFunc(h.middleware.JwtAuthentication)

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

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	var req loginReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	tokens, err := h.service.Login(r.Context(), mapToLoginInput(req))
	if err != nil {
		return authServ.ErrStatus.Resolve(err, "Login")
	}

	jwt.SetJwtCookie(w, jwt.AccessToken, tokens.AccessToken)
	jwt.SetJwtCookie(w, jwt.RefreshToken, tokens.RefreshToken)

	return nil
}

type forgotPasswordReq struct {
	Email string `json:"email" validate:"required,email"`
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) error {
	var req forgotPasswordReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err := h.service.ForgotPassword(r.Context(), mapToForgotPasswordInput(req))
	if err != nil {
		return authServ.ErrStatus.Resolve(err, "ForgotPassword")
	}

	return nil
}

type resetPasswordReq struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,ascii"`
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) error {
	var req resetPasswordReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	tokens, err := h.service.ResetPassword(r.Context(), mapToResetPassordInput(req))
	if err != nil {
		return authServ.ErrStatus.Resolve(err, "ResetPassword")
	}

	jwt.SetJwtCookie(w, jwt.AccessToken, tokens.AccessToken)
	jwt.SetJwtCookie(w, jwt.RefreshToken, tokens.RefreshToken)

	return nil
}

type userSignupReq struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
	Password    string `json:"password" validate:"required,ascii"`
}

func (h *Handler) UserSignup(w http.ResponseWriter, r *http.Request) error {
	var req userSignupReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	tokens, err := h.service.UserSignup(r.Context(), lang.GetLangFromHeader(r), mapToUserSignupInput(req))
	if err != nil {
		return authServ.ErrStatus.Resolve(err, "UserSignup")
	}

	jwt.SetJwtCookie(w, jwt.AccessToken, tokens.AccessToken)
	jwt.SetJwtCookie(w, jwt.RefreshToken, tokens.RefreshToken)

	w.WriteHeader(http.StatusCreated)

	return nil
}

type merchantSignupReq struct {
	Name         string `json:"name" validate:"required"`
	ContactEmail string `json:"contact_email" validate:"required,email"`
	Timezone     string `json:"timezone" validate:"required,timezone"`
}

func (h *Handler) MerchantSignup(w http.ResponseWriter, r *http.Request) error {
	var req merchantSignupReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err := h.service.MerchantSignup(r.Context(), mapToMerchantSignupInput(req))
	if err != nil {
		return authServ.ErrStatus.Resolve(err, "MerchantSignup")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

type meResp struct {
	UserId      uuid.UUID         `json:"user_id"`
	FirstName   string            `json:"first_name"`
	LastName    string            `json:"last_name"`
	Email       string            `json:"email"`
	PhoneNumber *string           `json:"phone_number"`
	AvatarUrl   *string           `json:"avatar_url"`
	Memberships []membershipsResp `json:"memberships"`
}

type membershipsResp struct {
	MerchantId      uuid.UUID          `json:"merchant_id"`
	MerchantName    string             `json:"merchant_name"`
	MerchantUrlName string             `json:"merchant_url_name"`
	EmployeeId      int                `json:"employee_id"`
	Role            types.EmployeeRole `json:"role"`
	AvatarUrl       *string            `json:"avatar_url"`
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) error {
	result, err := h.teamService.Me(r.Context())
	if err != nil {
		return teamServ.ErrStatus.Resolve(err, "Me")
	}

	httputil.Success(w, http.StatusOK, mapToMeResp(result))

	return nil
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) error {
	jwt.DeleteJwts(w)

	return nil
}

func (h *Handler) LogoutAllDevices(w http.ResponseWriter, r *http.Request) error {
	err := h.service.LogoutAllDevices(r.Context())
	if err != nil {
		return authServ.ErrStatus.Resolve(err, "LogoutAllDevices.LogoutAllDevices")
	}

	jwt.DeleteJwts(w)

	return nil
}

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) error {
	return startOauth(w, r, h.service.GoogleLogin, "GoogleLogin")
}

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) error {
	return finishOauth(w, r, h.service.GoogleCallback)
}

func (h *Handler) FacebookLogin(w http.ResponseWriter, r *http.Request) error {
	return startOauth(w, r, h.service.FacebookLogin, "FacebookLogin")
}

func (h *Handler) FacebookCallback(w http.ResponseWriter, r *http.Request) error {
	return finishOauth(w, r, h.service.FacebookCallback)
}

const (
	oauthStateCookie    = "oauth-state"
	oauthVerifierCookie = "oauth-verifier"
	oauthRedirectCookie = "oauth-redirect"
)

func oauthCookieScope() oauthutil.CookieScope {
	return oauthutil.CookieScope{
		Path:   "/api/v1/auth/oauth",
		Domain: fmt.Sprintf(".%s", config.LoadEnvVars().DOMAIN),
	}
}

func startOauth(w http.ResponseWriter, r *http.Request, login func(context.Context) (authServ.OauthLoginResult, error), source string) error {
	result, err := login(r.Context())
	if err != nil {
		return authServ.ErrStatus.Resolve(err, source)
	}

	scope := oauthCookieScope()
	oauthutil.SetCookie(w, scope, oauthStateCookie, result.State)
	oauthutil.SetCookie(w, scope, oauthVerifierCookie, result.Verifier)

	if redirect := safeRedirect(r.URL.Query().Get("redirect")); redirect != "" {
		oauthutil.SetCookie(w, scope, oauthRedirectCookie, redirect)
	} else {
		oauthutil.DeleteCookies(w, scope, oauthRedirectCookie)
	}

	http.Redirect(w, r, result.Url, http.StatusTemporaryRedirect)

	return nil
}

func finishOauth(w http.ResponseWriter, r *http.Request, callback func(context.Context, language.Tag, string, string) (jwt.TokenPair, error)) error {
	// the cookie is validated again, a client can set it to anything
	redirect := safeRedirect(oauthutil.CookieValue(r, oauthRedirectCookie))
	verifier := oauthutil.CookieValue(r, oauthVerifierCookie)
	oauthutil.DeleteCookies(w, oauthCookieScope(), oauthStateCookie, oauthVerifierCookie, oauthRedirectCookie)

	if err := oauthutil.ValidateState(r, oauthStateCookie); err != nil {
		return &apperr.RedirectError{URL: loginURL(redirect), Err: err}
	}

	if providerErr := r.URL.Query().Get("error"); providerErr != "" {
		// the user cancelled on the consent screen, nothing to show them
		if providerErr == "access_denied" {
			http.Redirect(w, r, loginURL(redirect), http.StatusSeeOther)
			return nil
		}

		return &apperr.RedirectError{URL: loginURL(redirect), Err: fmt.Errorf("oauth provider returned an error: %s", providerErr)}
	}

	tokens, err := callback(r.Context(), lang.GetLangFromHeader(r), r.URL.Query().Get("code"), verifier)
	if err != nil {
		return &apperr.RedirectError{URL: loginURL(redirect), Err: authServ.ErrStatus.Resolve(err, "finishOauth")}
	}

	jwt.SetJwtCookie(w, jwt.AccessToken, tokens.AccessToken)
	jwt.SetJwtCookie(w, jwt.RefreshToken, tokens.RefreshToken)

	if redirect == "" {
		redirect = config.LoadEnvVars().TANGO_URL
	}

	http.Redirect(w, r, redirect, http.StatusSeeOther)

	return nil
}

func safeRedirect(raw string) string {
	cfg := config.LoadEnvVars()

	target, err := url.Parse(raw)
	if err != nil || target.User != nil {
		return ""
	}

	for _, app := range []string{cfg.TANGO_URL, cfg.JABULANI_URL} {
		allowed, err := url.Parse(app)
		if err == nil && target.Scheme == allowed.Scheme && target.Host == allowed.Host {
			return target.String()
		}
	}

	return ""
}

// The login page of the app the user started from, keeping the redirect so a
// retry still ends up in the same place
func loginURL(redirect string) string {
	if redirect == "" {
		return config.LoadEnvVars().TANGO_URL + "/login"
	}

	target, _ := url.Parse(redirect)

	return fmt.Sprintf("%s://%s/login?%s", target.Scheme, target.Host, url.Values{"redirect": {redirect}}.Encode())
}
