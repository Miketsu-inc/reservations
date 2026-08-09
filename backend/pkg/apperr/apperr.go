package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

type Wrapped struct {
	Err   *Error
	Meta  map[string]any
	cause error
}

func (w *Wrapped) Error() string        { return w.Err.Message }
func (w *Wrapped) Unwrap() error        { return w.cause }
func (w *Wrapped) Is(target error) bool { return target == w.Err }
func (w *Wrapped) As(target any) bool {
	if t, ok := target.(**Error); ok {
		*t = w.Err
		return true
	}
	return false
}
func (w *Wrapped) With(key string, value any) *Wrapped {
	if w.Meta == nil {
		w.Meta = make(map[string]any, 2)
	}
	w.Meta[key] = value
	return w
}

func Wrap(err *Error, cause error) *Wrapped {
	return &Wrapped{Err: err, cause: cause}
}

type APIError struct {
	Status int
	Err    *Error
	// for per-instance overwrite of Err.Message
	Message string
	Meta    map[string]any
	// for internal logging only, should never reach client
	Cause error
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Message
}
func (e *APIError) Unwrap() error { return e.Cause }

// StatusMap maps errors to HTTP status codes. Define one per service covering all domain errors
type StatusMap map[*Error]int

// Convert domain errors into APIErrors. Named errors are apperr.Error's defined in the application
//
//   - Named error in StatusMap 	> APIError with status
//   - Named error not in StatusMap	> APIError with 500 status and original error message + code
//   - Not named error (unexpected)	> APIError with 500 status and generic error
func (m StatusMap) Resolve(err error, source string) error {
	var named *Error
	// unexpected error
	if !errors.As(err, &named) {
		return fmt.Errorf("%s: %w", source, err)
	}

	status, ok := m[named]
	if !ok {
		// TODO: log it here, so we can see which is missing
		status = http.StatusInternalServerError
	}

	var meta map[string]any
	var cause error
	var w *Wrapped
	if errors.As(err, &w) {
		meta = w.Meta
		cause = w.cause
	}

	return &APIError{Status: status, Err: named, Meta: meta, Cause: cause}
}
