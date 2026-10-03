package locations

import (
	"net/http"

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
	r.Get("/", h.Get)

	return r
}

type newReq struct {
	Country           *string        `json:"country"`
	City              *string        `json:"city"`
	PostalCode        *string        `json:"postal_code"`
	Address           *string        `json:"address"`
	GeoPoint          types.GeoPoint `json:"geo_point"`
	PlaceId           *string        `json:"place_id"`
	FormattedLocation string         `json:"formatted_location"`
	IsPrimary         bool           `json:"is_primary"`
	IsActive          bool           `json:"is_active"`
}

func (h *Handler) New(w http.ResponseWriter, r *http.Request) error {
	var req newReq

	if err := validate.ParseStruct(r, &req); err != nil {
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
	Id                int     `json:"id"`
	Country           *string `json:"country"`
	City              *string `json:"city"`
	PostalCode        *string `json:"postal_code"`
	Address           *string `json:"address"`
	FormattedLocation string  `json:"formatted_location"`
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) error {
	location, err := h.service.Get(r.Context())
	if err != nil {
		return locationServ.ErrStatus.Resolve(err, "Get")
	}

	httputil.Success(w, http.StatusOK, mapToGetResp(location))

	return nil
}
