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

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

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

func (h *Handler) NewMember(w http.ResponseWriter, r *http.Request) error {
	var req newMemberReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err := h.service.NewMember(r.Context(), mapToNewMemberInput(req))
	if err != nil {
		return teamServ.ErrStatus.Resolve(err, "NewMember")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

type updateMemberReq struct {
	Role        types.EmployeeRole `json:"role" validate:"required"`
	FirstName   string             `json:"first_name" validate:"required"`
	LastName    string             `json:"last_name" validate:"required"`
	Email       *string            `json:"email"`
	PhoneNumber *string            `json:"phone_number"`
	IsActive    bool               `json:"is_active"`
}

func (h *Handler) UpdateMember(w http.ResponseWriter, r *http.Request) error {
	var req updateMemberReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	urlMemberId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid team member id")
	}

	err = h.service.UpdateMember(r.Context(), urlMemberId, mapToUpdateMemberInput(req))
	if err != nil {
		return teamServ.ErrStatus.Resolve(err, "UpdateMember")
	}

	return nil
}

func (h *Handler) DeleteMember(w http.ResponseWriter, r *http.Request) error {
	urlMemberId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid team member id")
	}

	err = h.service.DeleteMember(r.Context(), urlMemberId)
	if err != nil {
		return teamServ.ErrStatus.Resolve(err, "DeleteMember")
	}

	return nil
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

func (h *Handler) GetMember(w http.ResponseWriter, r *http.Request) error {
	urlMemberId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid team member id")
	}

	teamMember, err := h.service.GetMember(r.Context(), urlMemberId)
	if err != nil {
		return teamServ.ErrStatus.Resolve(err, "GetMember")
	}

	httputil.Success(w, http.StatusOK, mapToGetMemberResp(teamMember))

	return nil
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) error {
	team, err := h.service.GetTeam(r.Context())
	if err != nil {
		return teamServ.ErrStatus.Resolve(err, "GetTeam")
	}

	var result []getMemberResp
	for _, member := range team {
		result = append(result, mapToGetMemberResp(member))
	}

	httputil.Success(w, http.StatusOK, result)

	return nil
}
