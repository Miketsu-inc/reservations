package oauthutil

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

func GenerateSate() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func ValidateOauthState(r *http.Request) error {
	state := r.URL.Query().Get("state")
	if state == "" {
		return fmt.Errorf("missing state in callback")
	}

	stateCookie, err := r.Cookie("oauth-state")
	if err != nil {
		return fmt.Errorf("missing oauth-sate cookie")
	}

	if subtle.ConstantTimeCompare([]byte(state), []byte(stateCookie.Value)) != 1 {
		return fmt.Errorf("invalid oauth state")
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
		// needs to be true in production
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}
