package blockedtimes

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	blockedtimeServ "github.com/miketsu-inc/reservations/backend/internal/service/blockedtime"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service *blockedtimeServ.Service
}

func NewHandler(s *blockedtimeServ.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Post("/", h.New)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

type newReq struct {
	Name          string `json:"name" validate:"required"`
	EmployeeIds   []int  `json:"employee_ids"`
	BlockedTypeId *int   `json:"blocked_type_id"`
	FromDate      string `json:"from_date"`
	ToDate        string `json:"to_date"`
	BlockedDay    string `json:"blocked_day"`
	IsAllDay      bool   `json:"is_all_day"`
}

func (h *Handler) New(w http.ResponseWriter, r *http.Request) error {
	var req newReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	input, err := mapToNewInput(req)
	if err != nil {
		return validate.NewError(err.Error())
	}

	err = h.service.New(r.Context(), input)
	if err != nil {
		return blockedtimeServ.ErrStatus.Resolve(err, "New")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

type updateReq struct {
	Id            int    `json:"id" validate:"required"`
	Name          string `json:"name" validate:"required"`
	EmployeeIds   []int  `json:"employee_ids"`
	BlockedTypeId *int   `json:"blocked_type_id"`
	FromDate      string `json:"from_date"`
	ToDate        string `json:"to_date"`
	BlockedDay    string `json:"blocked_day"`
	IsAllDay      bool   `json:"is_all_day"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	var req updateReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	UrlBlockedTimeId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid blocked time id")
	}

	if UrlBlockedTimeId != req.Id {
		return validate.NewError("invalid blocked time id")
	}

	input, err := mapToUpdateInput(req)
	if err != nil {
		return err
	}

	err = h.service.Update(r.Context(), input)
	if err != nil {
		return blockedtimeServ.ErrStatus.Resolve(err, "Update")
	}

	return nil
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	UrlBlockedTimeId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid blocked time id")
	}

	err = h.service.Delete(r.Context(), UrlBlockedTimeId)
	if err != nil {
		return blockedtimeServ.ErrStatus.Resolve(err, "Delete")
	}

	return nil
}
