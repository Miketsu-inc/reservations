package reviews

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	reviewServ "github.com/miketsu-inc/reservations/backend/internal/service/review"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	merchantServ *merchantServ.Service
	reviewServ   *reviewServ.Service
}

func NewHandler(merchant *merchantServ.Service, review *reviewServ.Service) *Handler {
	return &Handler{
		merchantServ: merchant,
		reviewServ:   review,
	}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Get("/", h.GetReviews)
	r.Get("/stats", h.GetStats)

	return r
}

type reviewStatsResp struct {
	AverageRating float64     `json:"average_rating"`
	TotalReviews  int         `json:"total_reviews"`
	Count5Star    int         `json:"count_5_star"`
	Count4Star    int         `json:"count_4_star"`
	Count3Star    int         `json:"count_3_star"`
	Count2Star    int         `json:"count_2_star"`
	Count1Star    int         `json:"count_1_star"`
	ServiceCounts map[int]int `json:"service_counts"`
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) error {
	urlName := chi.URLParam(r, "merchantName")
	if urlName == "" {
		return validate.NewError("invalid merchant name")
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	merchantId, err := h.merchantServ.GetMerchantIdByUrlName(r.Context(), urlName)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetMerchantIdByUrlName")
	}

	stats, err := h.reviewServ.GetReviewStats(r.Context(), merchantId, &urlLocationId, nil)
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
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}

type reviewReplyResp struct {
	Comment string `json:"comment"`
}

type reviewItemResp struct {
	Id          int                 `json:"id"`
	ServiceName string              `json:"service_name"`
	Rating      int16               `json:"rating"`
	Comment     *string             `json:"comment"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   *time.Time          `json:"updated_at"`
	Customer    reviewCustomerResp  `json:"customer"`
	Employee    *reviewEmployeeResp `json:"employee"`
	Reply       *reviewReplyResp    `json:"reply"`
}

type getReviewsResp struct {
	Reviews     []reviewItemResp `json:"reviews"`
	HasNextPage bool             `json:"has_next_page"`
	NextCursor  *string          `json:"next_cursor"`
}

func (h *Handler) GetReviews(w http.ResponseWriter, r *http.Request) error {
	urlName := chi.URLParam(r, "merchantName")
	if urlName == "" {
		return validate.NewError("invalid merchant name")
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	params, err := reviewServ.ParseReviewQueryParams(r.URL.Query())
	if err != nil {
		return err
	}

	merchantId, err := h.merchantServ.GetMerchantIdByUrlName(r.Context(), urlName)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetMerchantIdByUrlName")
	}

	result, err := h.reviewServ.GetReviews(r.Context(), reviewServ.GetReviewsInput{
		MerchantId:        merchantId,
		LocationId:        &urlLocationId,
		ReviewQueryParams: params,
	})
	if err != nil {
		return reviewServ.ErrStatus.Resolve(err, "GetReviews")
	}

	httputil.Success(w, http.StatusOK, mapToGetReviewsResp(result))

	return nil
}
