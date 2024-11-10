package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

func ParseJSON(r *http.Request, data any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}
	return json.NewDecoder(r.Body).Decode(data)
}

// Send a json response with an error message
//
// structure:
//
//	{
//		error: {
//			message: "error message",
//		}
//	}
func Error(w http.ResponseWriter, status int, err error) {
	// for debug, this sould never happen
	assert.NotNil(err, "cannot write nil error in response", err)
	writeJSON(w, status, map[string]map[string]string{"error": {"message": err.Error()}})
}

// Send a json response with data
//
// structure:
//
//	{
//		data: {
//			your data
//	}
func Success(w http.ResponseWriter, status int, v any) {
	writeJSON(w, status, map[string]any{"data": v})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(v)
	// for debug, let's see if we should handle this
	assert.Nil(err, "Could not be encoded to json", v, err)
}
