package team

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	teamServ "github.com/miketsu-inc/reservations/backend/internal/service/team"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service *teamServ.Service
}

func NewHandler(s *teamServ.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.NewMember)
	r.Put("/{id}", h.UpdateMember)
	r.Delete("/{id}", h.DeleteMember)
	r.Get("/{id}", h.GetMember)

	r.Get("/", h.GetTeam)

	return r
}

type newMemberReq struct {
	Role        types.EmployeeRole `json:"role" validate:"required"`
	FirstName   string             `json:"first_name" validate:"required"`
	LastName    string             `json:"last_name" validate:"required"`
	Email       *string            `json:"email"`
	PhoneNumber *string            `json:"phone_number"`
	IsActive    bool               `json:"is_active"`
}

func (h *Handler) NewMember(w http.ResponseWriter, r *http.Request) {
	var req newMemberReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.NewMember(r.Context(), mapToNewMemberInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type updateMemberReq struct {
	Role        types.EmployeeRole `json:"role" validate:"required"`
	FirstName   string             `json:"first_name" validate:"required"`
	LastName    string             `json:"last_name" validate:"required"`
	Email       *string            `json:"email"`
	PhoneNumber *string            `json:"phone_number"`
	IsActive    bool               `json:"is_active"`
}

func (h *Handler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	var req updateMemberReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlMemberId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.UpdateMember(r.Context(), urlMemberId, mapToUpdateMemberInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

func (h *Handler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	urlMemberId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.DeleteMember(r.Context(), urlMemberId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type getMemberResp struct {
	Id          int                `json:"id"`
	Role        types.EmployeeRole `json:"role"`
	FirstName   *string            `json:"first_name"`
	LastName    *string            `json:"last_name"`
	Email       *string            `json:"email"`
	PhoneNumber *string            `json:"phone_number"`
	IsActive    bool               `json:"is_active"`
}

func (h *Handler) GetMember(w http.ResponseWriter, r *http.Request) {
	urlMemberId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	teamMember, err := h.service.GetMember(r.Context(), urlMemberId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetMemberResp(teamMember))
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	team, err := h.service.GetTeam(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	var result []getMemberResp
	for _, member := range team {
		result = append(result, mapToGetMemberResp(member))
	}

	httputil.Success(w, http.StatusOK, result)
}
