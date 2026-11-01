package users

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	reviewServ "github.com/miketsu-inc/reservations/backend/internal/service/review"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

func (h *Handler) reviewRoutes(r *httputil.Router) {
	r.Post("/bookings/{bookingId}/review", h.CreateBookingReview)
	r.Put("/reviews/{reviewId}", h.UpdateReview)
	r.Delete("/reviews/{reviewId}", h.DeleteReview)
}

type createReviewReq struct {
	Rating  int16   `json:"rating" validate:"required,min=1,max=5"`
	Comment *string `json:"comment" validate:"max=300"`
}

func (h *Handler) CreateBookingReview(w http.ResponseWriter, r *http.Request) error {
	bookingId, err := strconv.Atoi(chi.URLParam(r, "bookingId"))
	if err != nil {
		return validate.NewError("invalid booking id")
	}

	var req createReviewReq
	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err = h.reviewServ.CreateReview(r.Context(), bookingId, req.Rating, req.Comment)
	if err != nil {
		return reviewServ.ErrStatus.Resolve(err, "CreateBookingReview")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

type updateReviewReq struct {
	Rating    int16   `json:"rating" validate:"required,min=1,max=5"`
	Comment   *string `json:"comment" validate:"max=300"`
	BookingId *int    `json:"booking_id"`
}

func (h *Handler) UpdateReview(w http.ResponseWriter, r *http.Request) error {
	reviewId, err := strconv.Atoi(chi.URLParam(r, "reviewId"))
	if err != nil {
		return validate.NewError("invalid review id")
	}

	var req updateReviewReq
	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err = h.reviewServ.UpdateReview(r.Context(), reviewId, req.BookingId, req.Rating, req.Comment)
	if err != nil {
		return reviewServ.ErrStatus.Resolve(err, "UpdateReview")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) error {
	reviewId, err := strconv.Atoi(chi.URLParam(r, "reviewId"))
	if err != nil {
		return validate.NewError("invalid review id")
	}

	err = h.reviewServ.DeleteReview(r.Context(), reviewId)
	if err != nil {
		return reviewServ.ErrStatus.Resolve(err, "DeleteReview")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}
