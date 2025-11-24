package location

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type GeoPoint struct {
	Lat float64 `json:"latitude"`
	Lon float64 `json:"longitude"`
}

func (g GeoPoint) Value() (driver.Value, error) {
	return fmt.Sprintf("POINT(%f %f)", g.Lon, g.Lat), nil
}

func (g *GeoPoint) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		b, ok := src.([]byte)
		if !ok {
			return fmt.Errorf("value cannot be converted to GeoPoint: %v", src)
		}

		s = string(b)
	}

	s = strings.TrimPrefix(s, "POINT(")
	s = strings.TrimSuffix(s, ")")
	parts := strings.Split(s, " ")
	if len(parts) != 2 {
		return fmt.Errorf("invalid POINT format: %s", s)
	}

	lon, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return err
	}

	lat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return err
	}

	g.Lat = lat
	g.Lon = lon
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
