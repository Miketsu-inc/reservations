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

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.New)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

type newReq struct {
	Name string `json:"name" validate:"required"`
	// EmployeeIds []int  `json:"employee_ids" validate:"required"`
	BlockedTypeId *int   `json:"blocked_type_id"`
	FromDate      string `json:"from_date" validate:"required"`
	ToDate        string `json:"to_date" validate:"required"`
	AllDay        bool   `json:"all_day"`
}

func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	var req newReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	input, err := mapToNewInput(req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.New(r.Context(), input)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// employee id not ids but its a front end issue
type updateReq struct {
	Id   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
	// EmployeeId int    `json:"employee_id" validate:"required"`
	BlockedTypeId *int   `json:"blocked_type_id"`
	FromDate      string `json:"from_date" validate:"required"`
	ToDate        string `json:"to_date" validate:"required"`
	AllDay        bool   `json:"all_day"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req updateReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	UrlBlockedTimeId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	input, err := mapToUpdateInput(req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.Update(r.Context(), UrlBlockedTimeId, input)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

// type deleteReq struct {
// 	EmployeeId int `json:"employee_id" validate:"required"`
// }

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// var req deleteReq

	// if err := validate.ParseStruct(r, &req); err != nil {
	// 	httputil.Error(w, http.StatusBadRequest, err)
	// 	return
	// }

	UrlBlockedTimeId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	// input, err := mapToDeleteInput(req)
	// if err != nil {
	// 	httputil.Error(w, http.StatusBadRequest, err)
	// 	return
	// }

	// err = h.service.Delete(r.Context(), UrlBlockedTimeId, input)
	err = h.service.Delete(r.Context(), UrlBlockedTimeId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}
