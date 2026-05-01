package jwt

import (
	"context"
	"fmt"
	"net/http"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type JwtType int

const (
	RefreshToken JwtType = iota
	AccessToken

	RefreshCookieName string = "jwt-refresh"
	AccessCookieName  string = "jwt-access"
)

type contextKey struct {
	name string
}

var userIDCtxKey = &contextKey{"UserID"}

// Returns UserID from the request's context. Panics if not present!
func MustGetUserIDFromContext(ctx context.Context) uuid.UUID {
	userID, ok := ctx.Value(userIDCtxKey).(uuid.UUID)
	assert.True(ok, "Authenticated route called without jwt user id", ctx.Value(userIDCtxKey), userID)

	return userID
}

// Returns UserID from the request's context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDCtxKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}

	return userID, true
}

func SetUserIdInContext(ctx context.Context, userId uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDCtxKey, userId)
}

// Create a new jwt, with HS256 signing method
func new(secret []byte, claims jwtlib.MapClaims) (string, error) {
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err

	}

	return tokenString, nil
}

// Create a new access token
func NewAccessToken(userID uuid.UUID) (string, error) {
	expMin := config.LoadEnvVars().JWT_ACCESS_EXP_MIN
	expMinDuration := time.Minute * time.Duration(expMin)

	claims := jwtlib.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(expMinDuration).Unix(),
	}

	token, err := new([]byte(config.LoadEnvVars().JWT_ACCESS_SECRET), claims)
	if err != nil {
		return "", fmt.Errorf("unexpected error when creating jwt token: %s", err.Error())
	}

	return token, nil
}

// Create a new refresh token
func NewRefreshToken(userID uuid.UUID, refreshVersion int) (string, error) {
	expMin := config.LoadEnvVars().JWT_REFRESH_EXP_MIN
	expMinDuration := time.Minute * time.Duration(expMin)

	claims := jwtlib.MapClaims{
		"sub":             userID,
		"exp":             time.Now().Add(expMinDuration).Unix(),
		"refresh_version": refreshVersion,
	}

	token, err := new([]byte(config.LoadEnvVars().JWT_REFRESH_SECRET), claims)
	if err != nil {
		return "", fmt.Errorf("unexpected error when creating jwt token: %s", err.Error())
	}

	return token, nil
}

// Set a jwt cookie
func SetJwtCookie(w http.ResponseWriter, name JwtType, token string) {
	var cookieName string
	var expMin int

	switch name {
	case AccessToken:
		cookieName = AccessCookieName
		expMin = config.LoadEnvVars().JWT_ACCESS_EXP_MIN
	case RefreshToken:
		cookieName = RefreshCookieName
		expMin = config.LoadEnvVars().JWT_REFRESH_EXP_MIN
	}

	// expMinDuration := time.Minute * time.Duration(expMin)

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		HttpOnly: true,
		MaxAge:   expMin * 60,
		// Expires:  time.Now().UTC().Add(expMinDuration),
		Path: "/",
		// needs to be true in production
		Secure:   false,
		Domain:   ".reservations.local",
		SameSite: http.SameSiteLaxMode,
	})
}

// Deletes both the access and refresh jwt cookies
func DeleteJwts(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Now().UTC(),
		// needs to be true in production
		Secure:   false,
		Domain:   ".reservations.local",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     AccessCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Now().UTC(),
		// needs to be true in production
		Secure:   false,
		Domain:   ".reservations.local",
		SameSite: http.SameSiteLaxMode,
	})
}
