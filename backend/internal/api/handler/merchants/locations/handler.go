package locations

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	locationServ "github.com/miketsu-inc/reservations/backend/internal/service/location"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service *locationServ.Service
}

func NewHandler(s *locationServ.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	// TODO: temp until signup flow is figured out?
	r.Post("/", h.New)
	r.Get("/", h.GetAll)
	r.Patch("/{id}", h.Update)

	return r
}

type newReq struct {
	Country           *string         `json:"country"`
	CountryCode       *string         `json:"country_code"`
	Region            *string         `json:"region"`
	City              *string         `json:"city"`
	PostalCode        *string         `json:"postal_code"`
	Address           *string         `json:"address"`
	GeoPoint          *types.GeoPoint `json:"geo_point" validate:"required"`
	GeocodingProvider *string         `json:"geocoding_provider"`
	ProviderPlaceId   *string         `json:"provider_place_id"`
	FormattedLocation string          `json:"formatted_location" validate:"required"`
}

func (h *Handler) New(w http.ResponseWriter, r *http.Request) error {
	var req newReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}
	if err := validateGeoPoint(*req.GeoPoint); err != nil {
		return err
	}

	err := h.service.New(r.Context(), mapToNewInput(req))
	if err != nil {
		return locationServ.ErrStatus.Resolve(err, "New")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

type getResp struct {
	Id                int            `json:"id"`
	Name              string         `json:"name"`
	Sequence          int            `json:"sequence"`
	Country           *string        `json:"country"`
	CountryCode       *string        `json:"country_code"`
	Region            *string        `json:"region"`
	City              *string        `json:"city"`
	PostalCode        *string        `json:"postal_code"`
	Address           *string        `json:"address"`
	GeoPoint          types.GeoPoint `json:"geo_point"`
	GeocodingProvider *string        `json:"geocoding_provider"`
	ProviderPlaceId   *string        `json:"provider_place_id"`
	FormattedLocation string         `json:"formatted_location"`
	ContactEmail      *string        `json:"contact_email"`
	PhoneNumber       *string        `json:"phone_number"`
	ParkingInfo       string         `json:"parking_info"`
	ArrivalInfo       string         `json:"arrival_info"`
	IsActive          bool           `json:"is_active"`
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) error {
	locations, err := h.service.GetAll(r.Context())
	if err != nil {
		return locationServ.ErrStatus.Resolve(err, "GetAll")
	}

	httputil.Success(w, http.StatusOK, mapToGetResp(locations))

	return nil
}

type updateReq struct {
	Name              string          `json:"name" validate:"required,max=100"`
	ContactEmail      *string         `json:"contact_email" validate:"omitempty,email,max=320"`
	PhoneNumber       *string         `json:"phone_number" validate:"omitempty,max=30"`
	Country           *string         `json:"country"`
	CountryCode       *string         `json:"country_code"`
	Region            *string         `json:"region"`
	City              *string         `json:"city"`
	PostalCode        *string         `json:"postal_code"`
	Address           *string         `json:"address"`
	GeoPoint          *types.GeoPoint `json:"geo_point" validate:"required"`
	GeocodingProvider *string         `json:"geocoding_provider"`
	ProviderPlaceId   *string         `json:"provider_place_id"`
	FormattedLocation string          `json:"formatted_location" validate:"required"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	var req updateReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	if err := validateGeoPoint(*req.GeoPoint); err != nil {
		return err
	}

	locationId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	err = h.service.Update(r.Context(), locationId, mapToUpdateInput(req))
	if err != nil {
		return locationServ.ErrStatus.Resolve(err, "Update")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func validateGeoPoint(point types.GeoPoint) error {
	if point.Lat < -90 || point.Lat > 90 || point.Lon < -180 || point.Lon > 180 {
		return validate.NewError("geo_point is outside the valid coordinate range")
	}

	return nil
}
