package merchants

import (
	"net/http"
	"strconv"
	"time"

	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type dashboardStatisticsResp struct {
	RevenueSum            string `json:"revenue_sum"`
	RevenueChange         int    `json:"revenue_change"`
	Bookings              int    `json:"bookings"`
	BookingsChange        int    `json:"bookings_change"`
	Cancellations         int    `json:"cancellations"`
	CancellationsChange   int    `json:"cancellations_change"`
	AverageDuration       int    `json:"average_duration"`
	AverageDurationChange int    `json:"average_duration_change"`
}

type dashboardRevenueResp struct {
	PeriodStart time.Time         `json:"period_start"`
	PeriodEnd   time.Time         `json:"period_end"`
	Revenue     []revenueStatResp `json:"revenue"`
}

// TODO: value is of numeric type so float might not be the best
// type to return here
type revenueStatResp struct {
	Value float64   `json:"value"`
	Day   time.Time `json:"day"`
}

func dashboardPeriodFromRequest(r *http.Request) (int, error) {
	period, err := strconv.Atoi(r.URL.Query().Get("period"))
	if err != nil || (period != 7 && period != 30) {
		return 0, validate.NewError("period must be either 7 or 30")
	}

	return period, nil
}

func (h *Handler) GetDashboardStatistics(w http.ResponseWriter, r *http.Request) error {
	period, err := dashboardPeriodFromRequest(r)
	if err != nil {
		return err
	}

	statistics, err := h.service.GetDashboardStatistics(r.Context(), period)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetDashboardStatistics")
	}

	httputil.Success(w, http.StatusOK, mapToDashboardStatisticsResp(statistics))

	return nil
}

func (h *Handler) GetDashboardRevenue(w http.ResponseWriter, r *http.Request) error {
	period, err := dashboardPeriodFromRequest(r)
	if err != nil {
		return err
	}

	revenue, err := h.service.GetDashboardRevenue(r.Context(), period)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetDashboardRevenue")
	}

	httputil.Success(w, http.StatusOK, mapToDashboardRevenueResp(revenue))

	return nil
}
