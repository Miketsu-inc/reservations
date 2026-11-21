package oauthutil

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

func RandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type CookieScope struct {
	Path   string
	Domain string
}

const cookieMaxAge = 10 * 60

func ValidateState(r *http.Request, cookieName string) error {
	state := r.URL.Query().Get("state")
	if state == "" {
		return validate.NewError("oauth: missing state in callback")
	}

	stateCookie, err := r.Cookie(cookieName)
	if err != nil {
		return validate.NewError(fmt.Sprintf("oauth: missing %s cookie", cookieName))
	}

	if subtle.ConstantTimeCompare([]byte(state), []byte(stateCookie.Value)) != 1 {
		return validate.NewError("oauth: invalid state")
	}

	return nil
}

func CookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func SetCookie(w http.ResponseWriter, scope CookieScope, name, value string) {
	setCookie(w, scope, name, value, cookieMaxAge)
}

func DeleteCookies(w http.ResponseWriter, scope CookieScope, names ...string) {
	for _, name := range names {
		setCookie(w, scope, name, "", -1)
	}
}

func setCookie(w http.ResponseWriter, scope CookieScope, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     scope.Path,
		Domain:   scope.Domain,
		HttpOnly: true,
		MaxAge:   maxAge,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
