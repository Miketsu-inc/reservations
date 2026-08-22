package invitations

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	teamServ "github.com/miketsu-inc/reservations/backend/internal/service/team"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service    *teamServ.Service
	middleware *middleware.Manager
}

func NewHandler(s *teamServ.Service, m *middleware.Manager) *Handler {
	return &Handler{service: s, middleware: m}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Get("/{token}", h.GetInvitation)

	r.Group(func(r *httputil.Router) {
		r.UseFunc(h.middleware.JwtAuthentication)
		r.UseFunc(h.middleware.Language)

		r.Post("/{token}/accept", h.AcceptInvitation)
		r.Post("/{token}/decline", h.DeclineInvitation)
	})

	return r
}

type getInvitationResp struct {
	MerchantName string                         `json:"merchant_name"`
	InvitorName  *string                        `json:"invitor_name"`
	Role         types.EmployeeInvitationRole   `json:"role"`
	Email        string                         `json:"email"`
	Status       types.EmployeeInvitationStatus `json:"status"`
}

func (h *Handler) GetInvitation(w http.ResponseWriter, r *http.Request) error {
	urlToken := chi.URLParam(r, "token")

	if urlToken == "" {
		return validate.NewError("invalid invitation token")
	}

	result, err := h.service.GetInvitation(r.Context(), urlToken)
	if err != nil {
		return teamServ.ErrStatus.Resolve(err, "GetInvitation")
	}

	httputil.Success(w, http.StatusOK, mapToGetInvitation(result))

	return nil
}

func (h *Handler) AcceptInvitation(w http.ResponseWriter, r *http.Request) error {
	urlToken := chi.URLParam(r, "token")

	if urlToken == "" {
		return validate.NewError("invalid invitation token")
	}

	redirectUrl, err := h.service.AcceptInvitation(r.Context(), urlToken)
	if err != nil {
		return teamServ.ErrStatus.Resolve(err, "AcceptInvitation")
	}

	http.Redirect(w, r, redirectUrl, http.StatusPermanentRedirect)

	return nil
}

func (h *Handler) DeclineInvitation(w http.ResponseWriter, r *http.Request) error {
	urlToken := chi.URLParam(r, "token")

	if urlToken == "" {
		return validate.NewError("invalid invitation token")
	}

	err := h.service.DeclineInvitation(r.Context(), urlToken)
	if err != nil {
		return teamServ.ErrStatus.Resolve(err, "DeclineInvitation")
	}

	return nil
}
