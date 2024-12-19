package jwt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
)

type JwtType int

const (
	RefreshToken JwtType = iota
	AccessToken

	JwtRefreshCookieName string = "jwt-refresh"
	JwtAccessCookieName  string = "jwt-access"
)

type contextKey struct {
	name string
}

var UserIDCtxKey = &contextKey{"UserID"}

// Returns UserID from the request's context.
// Should be only called in authenticated routes!
func UserIDFromContext(ctx context.Context) uuid.UUID {
	userID, ok := ctx.Value(UserIDCtxKey).(uuid.UUID)
	assert.True(ok, "Authenticated route called without jwt user id", ctx.Value(UserIDCtxKey), userID)

	return userID
}

// Jwt authentication middleware. Uses refresh and access tokens
func JwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// try to verify request with access token
		claims, err := verifyRequest(r, AccessToken, getTokenFromCookie)
		if err != nil {
			// if access token could not be found in cookies it means it's expired or did not exist
			if !errors.Is(err, ErrJwtAccessNotFound) {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("%v", err.Error()))
				return
			}

			// try to verify request with refresh token
			claims, err = verifyRequest(r, RefreshToken, getTokenFromCookie)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("%v", err.Error()))
				return
			}

			userID := getUserIdFromClaims(claims)
			if userID == uuid.Nil {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("could not parse jwt claims"))
				return
			}

			// TODO: Is this ( database.New() ) okay? can it cause problems? race conditions?
			dbRefreshVersion, err := database.PostgreSQL.GetUserJwtRefreshVersion(database.New(), ctx, userID)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("unexpected error when reading jwt refresh version %s", err.Error()))
				return
			}

			tokenRefreshVersion, err := getRefreshVersionFromClaims(claims)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("unexpected error when parsing refresh version: %s", err.Error()))
				return
			}

			// check if refresh version matches in the resfresh token and database
			// if they match a new access token can be issued
			if dbRefreshVersion != tokenRefreshVersion {
				DeleteJwts(w)
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("refresh token version does not match"))
				return
			}

			err = NewAccessToken(w, userID)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("could not create new access jwt"))
				return
			}
		}

		userID := getUserIdFromClaims(claims)
		if userID == uuid.Nil {
			httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("could not parse jwt claims"))
			return
		}

		ctx = context.WithValue(ctx, UserIDCtxKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Create a new jwt, with HS256 signing method
func New(secret []byte, claims jwtlib.MapClaims) (string, error) {
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err

	}

	return tokenString, nil
}

var cfg *config.Config = config.LoadEnvVars()

// Create a new access token and set it in cookies
func NewAccessToken(w http.ResponseWriter, userID uuid.UUID) error {
	expMin := cfg.JWT_ACCESS_EXP_MIN

	expMinDuration := time.Minute * time.Duration(expMin)

	token, err := New([]byte(cfg.JWT_ACCESS_SECRET), jwtlib.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(expMinDuration).Unix(),
	})

	if err != nil {
		return fmt.Errorf("unexpected error when creating jwt token: %s", err.Error())
	}

	http.SetCookie(w, &http.Cookie{
		Name:     JwtAccessCookieName,
		Value:    token,
		HttpOnly: true,
		MaxAge:   expMin * 60,
		Expires:  time.Now().UTC().Add(expMinDuration),
		Path:     "/",
		// needs to be true in production
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// Create a new refresh token and set it in cookies
func NewRefreshToken(w http.ResponseWriter, userID uuid.UUID, refreshVersion int) error {
	expMin := cfg.JWT_REFRESH_EXP_MIN

	expMinDuration := time.Minute * time.Duration(expMin)

	token, err := New([]byte(cfg.JWT_REFRESH_SECRET), jwtlib.MapClaims{
		"sub":             userID,
		"exp":             time.Now().Add(expMinDuration).Unix(),
		"refresh_version": refreshVersion,
	})

	if err != nil {
		return fmt.Errorf("unexpected error when creating jwt token: %s", err.Error())
	}

	http.SetCookie(w, &http.Cookie{
		Name:     JwtRefreshCookieName,
		Value:    token,
		HttpOnly: true,
		MaxAge:   expMin * 60,
		Expires:  time.Now().UTC().Add(expMinDuration),
		Path:     "/",
		// needs to be true in production
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// Deletes both the access and refresh jwt cookies
func DeleteJwts(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     JwtRefreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Now().UTC(),
		// needs to be true in production
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     JwtAccessCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Now().UTC(),
		// needs to be true in production
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func getUserIdFromClaims(claims jwtlib.MapClaims) uuid.UUID {
	uuidStr, err := claims.GetSubject()
	if err != nil {
		return uuid.Nil
	}

	userID, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil
	}

	return userID
}

func getRefreshVersionFromClaims(claims jwtlib.MapClaims) (int, error) {
	val, ok := claims["refresh_version"]
	if !ok {
		return 0, nil
	}

	switch refreshVersion := val.(type) {
	case float64:
		return int(refreshVersion), nil

	case json.Number:
		val, _ := refreshVersion.Float64()

		return int(val), nil
	}

	return 0, nil
}

// parse and validate jwt, returning the claims if valid
func verifyToken(tokenString string, tokenType JwtType) (jwtlib.MapClaims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, jwtlib.MapClaims{}, func(token *jwtlib.Token) (interface{}, error) {
		switch tokenType {
		case AccessToken:
			return []byte(cfg.JWT_ACCESS_SECRET), nil
		case RefreshToken:
			return []byte(cfg.JWT_REFRESH_SECRET), nil
		default:
			assert.Never("Jwt token type can be either refresh or access", tokenType)
			return "", fmt.Errorf("Jwt token type can be either refresh or access")
		}
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwtlib.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

var ErrJwtAccessNotFound = errors.New("jwt access token could not be found")

// check if a token is sent with the request
func verifyRequest(r *http.Request, tokenType JwtType, findTokenFns ...func(r *http.Request, tokenType JwtType) string) (jwtlib.MapClaims, error) {
	var tokenString string

	for _, fn := range findTokenFns {
		tokenString = fn(r, tokenType)
		if tokenString != "" {
			break
		}
	}
	if tokenString == "" {
		return nil, ErrJwtAccessNotFound
	}

	return verifyToken(tokenString, tokenType)
}

func getTokenFromCookie(r *http.Request, tokenType JwtType) string {
	var cookieName string

	switch tokenType {
	case AccessToken:
		cookieName = JwtAccessCookieName
	case RefreshToken:
		cookieName = JwtRefreshCookieName
	default:
		assert.Never("Jwt token type can be either refresh or access", tokenType)
	}

	cookie, err := r.Cookie(cookieName)

	if err != nil {
		return ""
	}

	return cookie.Value
}

// Commented as idk how it should be done when having a refresh and access token

// Rerive token from "Authorization" request header: "Authorization: BEARER T".
// func getTokenFromHeader(r *http.Request) string {
// 	bearer := r.Header.Get("Authorization")
// 	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
// 		return bearer[7:]
// 	}

// 	return ""
// }
