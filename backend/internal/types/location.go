package types

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

type GeoPoint struct {
	Lat float64 `json:"latitude"`
	Lon float64 `json:"longitude"`
}

func (g GeoPoint) Value() (driver.Value, error) {
	return fmt.Sprintf("POINT(%f %f)", g.Lon, g.Lat), nil
}

func (g *GeoPoint) Scan(src any) error {
	if src == nil {
		return nil
	}

	var data []byte

	switch v := src.(type) {

	case []byte:
		data = v

	case string:
		d, err := hex.DecodeString(v)
		if err != nil {
			return fmt.Errorf("invalid EWKB hex: %w", err)
		}
		data = d

	default:
		return fmt.Errorf("cannot scan type %T into GeoPoint", src)
	}

	geometry, err := ewkb.Unmarshal(data)
	if err != nil {
		return fmt.Errorf("EWKB decode error: %w", err)
	}

	point, ok := geometry.(*geom.Point)
	if !ok {
		return fmt.Errorf("EWKB is not a POINT")
	}

	coords := point.Coords()
	g.Lon = coords.X()
	g.Lat = coords.Y()

	return nil
}

func (g GeoPoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Lat float64 `json:"latitude"`
		Lon float64 `json:"longitude"`
	}{
		Lat: g.Lat,
		Lon: g.Lon,
	})
}

func (g *GeoPoint) UnmarshalJSON(data []byte) error {
	var point struct {
		Lat float64 `json:"latitude"`
		Lon float64 `json:"longitude"`
	}

	if err := json.Unmarshal(data, &point); err != nil {
		return err
	}

	g.Lat = point.Lat
	g.Lon = point.Lon
	return nil
}
