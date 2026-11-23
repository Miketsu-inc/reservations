package reviews

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	reviewServ "github.com/miketsu-inc/reservations/backend/internal/service/review"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service *reviewServ.Service
}

func NewHandler(s *reviewServ.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Get("/", h.GetReviews)
	r.Get("/stats", h.GetStats)
	r.Post("/{reviewId}/reply", h.Reply)
	r.Delete("/{reviewId}/reply", h.DeleteReply)
	r.Post("/{reviewId}/report", h.Report)

	return r
}

type reviewStatsResp struct {
	AverageRating float64 `json:"average_rating"`
	TotalReviews  int     `json:"total_reviews"`
	Count5Star    int     `json:"count_5_star"`
	Count4Star    int     `json:"count_4_star"`
	Count3Star    int     `json:"count_3_star"`
	Count2Star    int     `json:"count_2_star"`
	Count1Star    int     `json:"count_1_star"`
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) error {
	stats, err := h.service.GetReviewStatsByMerchant(r.Context())
	if err != nil {
		return reviewServ.ErrStatus.Resolve(err, "GetReviewStats")
	}

	httputil.Success(w, http.StatusOK, mapToReviewStatsResp(stats))

	return nil
}

type reviewCustomerResp struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	AvatarUrl *string `json:"avatar_url"`
}

type reviewEmployeeResp struct {
	Id        *int    `json:"id,omitempty"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	AvatarUrl *string `json:"avatar_url"`
}

type reviewReplyResp struct {
	Comment   string              `json:"comment"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	Employee  *reviewEmployeeResp `json:"employee"`
}

type reviewItemResp struct {
	Id           int                 `json:"id"`
	BookingId    *int                `json:"booking_id"`
	ServiceName  string              `json:"service_name"`
	ServiceId    *int                `json:"service_id"`
	Rating       int16               `json:"rating"`
	Comment      *string             `json:"comment"`
	Status       string              `json:"status"`
	ReportReason *string             `json:"report_reason"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    *time.Time          `json:"updated_at"`
	Customer     reviewCustomerResp  `json:"customer"`
	Employee     *reviewEmployeeResp `json:"employee"`
	Reply        *reviewReplyResp    `json:"reply"`
}

type getReviewsResp struct {
	Reviews     []reviewItemResp `json:"reviews"`
	HasNextPage bool             `json:"has_next_page"`
	NextCursor  *string          `json:"next_cursor"`
}

func (h *Handler) GetReviews(w http.ResponseWriter, r *http.Request) error {
	params, err := reviewServ.ParseReviewQueryParams(r.URL.Query())
	if err != nil {
		return err
	}

	result, err := h.service.GetReviewsByMerchant(r.Context(), params)
	if err != nil {
		return reviewServ.ErrStatus.Resolve(err, "GetReviews")
	}

	httputil.Success(w, http.StatusOK, mapToGetReviewsResp(result))

	return nil
}

type replyReq struct {
	Comment string `json:"comment" validate:"required,min=1,max=300"`
}

func (h *Handler) Reply(w http.ResponseWriter, r *http.Request) error {
	reviewId, err := strconv.Atoi(chi.URLParam(r, "reviewId"))
	if err != nil {
		return validate.NewError("invalid review_id")
	}

	var req replyReq
	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err = h.service.ReplyReview(r.Context(), reviewId, req.Comment)
	if err != nil {
		return reviewServ.ErrStatus.Resolve(err, "ReplyReview")
	}

	return nil
}

func (h *Handler) DeleteReply(w http.ResponseWriter, r *http.Request) error {
	reviewId, err := strconv.Atoi(chi.URLParam(r, "reviewId"))
	if err != nil {
		return validate.NewError("invalid review_id")
	}

	err = h.service.DeleteReviewReply(r.Context(), reviewId)
	if err != nil {
		return reviewServ.ErrStatus.Resolve(err, "DeleteReviewReply")
	}

	return nil
}

type reportReq struct {
	Reason string `json:"reason" validate:"required,min=1,max=500"`
}

func (h *Handler) Report(w http.ResponseWriter, r *http.Request) error {
	reviewId, err := strconv.Atoi(chi.URLParam(r, "reviewId"))
	if err != nil {
		return validate.NewError("invalid review_id")
	}

	var req reportReq
	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err = h.service.ReportReview(r.Context(), reviewId, req.Reason)
	if err != nil {
		return reviewServ.ErrStatus.Resolve(err, "ReportReview")
	}

	return nil
}
