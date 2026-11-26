package favorites

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	favoriteServ "github.com/miketsu-inc/reservations/backend/internal/service/favorite"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service    *favoriteServ.Service
	middleware *middleware.Manager
}

func NewHandler(s *favoriteServ.Service, m *middleware.Manager) *Handler {
	return &Handler{service: s, middleware: m}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Group(func(r *httputil.Router) {
		r.UseFunc(h.middleware.JwtAuthentication)

		r.Get("/", h.GetFavorites)
		r.Get("/merchants/{merchantId}/locations/{locationId}", h.IsFavorite)
		r.Post("/merchants/{merchantId}/locations/{locationId}", h.AddFavorite)
		r.Delete("/merchants/{merchantId}/locations/{locationId}", h.RemoveFavorite)
	})

	return r
}

type favoriteLocationResp struct {
	MerchantId        uuid.UUID `json:"merchant_id"`
	MerchantName      string    `json:"merchant_name"`
	MerchantUrlName   string    `json:"merchant_url_name"`
	LocationId        int       `json:"location_id"`
	FormattedLocation string    `json:"formatted_location"`
	AverageRating     float64   `json:"average_rating"`
	ReviewCount       int       `json:"review_count"`
	MainImageUrl      *string   `json:"main_image_url"`
}

func (h *Handler) GetFavorites(w http.ResponseWriter, r *http.Request) error {
	favorites, err := h.service.GetFavorites(r.Context())
	if err != nil {
		return favoriteServ.ErrStatus.Resolve(err, "GetFavorites")
	}

	httputil.Success(w, http.StatusOK, mapToGetFavoritesResp(favorites))

	return nil
}

func (h *Handler) AddFavorite(w http.ResponseWriter, r *http.Request) error {
	merchantId, err := uuid.Parse(chi.URLParam(r, "merchantId"))
	if err != nil {
		return validate.NewError("invalid merchant id")
	}

	locationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil || locationId <= 0 {
		return validate.NewError("invalid location id")
	}

	err = h.service.AddFavorite(r.Context(), merchantId, locationId)
	if err != nil {
		return favoriteServ.ErrStatus.Resolve(err, "AddFavorite")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

func (h *Handler) RemoveFavorite(w http.ResponseWriter, r *http.Request) error {
	merchantId, err := uuid.Parse(chi.URLParam(r, "merchantId"))
	if err != nil {
		return validate.NewError("invalid merchant id")
	}

	locationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil || locationId <= 0 {
		return validate.NewError("invalid location id")
	}

	err = h.service.RemoveFavorite(r.Context(), merchantId, locationId)
	if err != nil {
		return favoriteServ.ErrStatus.Resolve(err, "RemoveFavorite")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

type isFavoriteResp struct {
	IsFavorite bool `json:"is_favorite"`
}

func (h *Handler) IsFavorite(w http.ResponseWriter, r *http.Request) error {
	merchantId, err := uuid.Parse(chi.URLParam(r, "merchantId"))
	if err != nil {
		return validate.NewError("invalid merchant id")
	}

	locationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil || locationId <= 0 {
		return validate.NewError("invalid location id")
	}

	isFav, err := h.service.IsFavorite(r.Context(), merchantId, locationId)
	if err != nil {
		return favoriteServ.ErrStatus.Resolve(err, "IsFavorite")
	}

	httputil.Success(w, http.StatusOK, isFavoriteResp{IsFavorite: isFav})

	return nil
}
