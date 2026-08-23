package oauthutil

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

func RandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func ValidateOauthState(r *http.Request) error {
	state := r.URL.Query().Get("state")
	if state == "" {
		return validate.NewError("oauth: missing state in callback")
	}

	stateCookie, err := r.Cookie("oauth-state")
	if err != nil {
		return validate.NewError("oauth: missing oauth-sate cookie")
	}

	if subtle.ConstantTimeCompare([]byte(state), []byte(stateCookie.Value)) != 1 {
		return validate.NewError("oauth: invalid state")
	}

	return nil
}

func SetOauthStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth-state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   5 * 60,
		Expires:  time.Now().UTC().Add(time.Minute * 5),
		Secure:   config.LoadEnvVars().IsProd(),
		SameSite: http.SameSiteLaxMode,
	})
}
