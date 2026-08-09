package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

var ErrInvalidRefreshVersion = &apperr.Error{Code: "invalid_refresh_version", Message: "invalid refresh version"}
var ErrRefreshTokenVersionMismatch = &apperr.Error{Code: "refresh_token_version_mismatch", Message: "refresh token version does not match"}
var ErrAuthenticationRequired = &apperr.Error{Code: "authentication_required", Message: "authentication is required for this resource"}

// Jwt authentication middleware. Uses refresh and access tokens
func (m *Manager) JwtAuthentication(next http.Handler) httputil.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		// try to verify request with access token
		claims, err := verifyRequest(r, jwt.AccessToken, getTokenFromCookie)
		if err != nil {
			// if access token could not be found in cookies it means it's either expired or did not exist
			// if it is found but invalid unauthorized status can be returned
			if !errors.Is(err, ErrTokenMissing) {
				if errors.Is(err, ErrInvalidAccessToken) {
					return &apperr.APIError{
						Status: http.StatusUnauthorized,
						Err:    ErrInvalidAccessToken,
						Cause:  err,
					}
				}

				return err
			}

			// try to verify request with refresh token
			claims, err = verifyRequest(r, jwt.RefreshToken, getTokenFromCookie)
			if err != nil {
				if errors.Is(err, ErrTokenMissing) {
					return &apperr.APIError{
						Status: http.StatusUnauthorized,
						Err:    ErrAuthenticationRequired,
						Cause:  err,
					}
				} else if errors.Is(err, ErrInvalidRefreshToken) {
					return &apperr.APIError{
						Status: http.StatusUnauthorized,
						Err:    ErrInvalidRefreshToken,
						Cause:  err,
					}
				}

				return err
			}

			userID, err := getUserIdFromClaims(claims)
			if err != nil {
				return err
			}

			dbRefreshVersion, err := m.userRepo.GetUserJwtRefreshVersion(ctx, userID)
			if err != nil {
				return err
			}

			tokenRefreshVersion, ok := getRefreshVersionFromClaims(claims)
			if !ok {
				return &apperr.APIError{
					Status: http.StatusUnauthorized,
					Err:    ErrInvalidRefreshVersion,
				}
			}

			// check if refresh version matches in the resfresh token and database
			// if they match a new access token can be issued
			if dbRefreshVersion != tokenRefreshVersion {
				jwt.DeleteJwts(w)

				return &apperr.APIError{
					Status: http.StatusUnauthorized,
					Err:    ErrRefreshTokenVersionMismatch,
				}
			}

			token, err := jwt.NewAccessToken(userID)
			if err != nil {
				return err
			}

			jwt.SetJwtCookie(w, jwt.AccessToken, token)
		}

		userID, err := getUserIdFromClaims(claims)
		if err != nil {
			return err
		}

		ctx = jwt.SetUserIdInContext(ctx, userID)

		next.ServeHTTP(w, r.WithContext(ctx))

		return nil
	}
}

var ErrUserNotMerchantEmployee = &apperr.Error{Code: "user_not_merchant_employee", Message: "user is not a member of this merchant's team"}

func (m *Manager) EmployeeAuthentication(next http.Handler) httputil.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		merchantId, err := uuid.Parse(chi.URLParam(r, "merchantId"))
		if err != nil {
			return validate.NewError("invalid merchant id")
		}

		userId := jwt.MustGetUserIDFromContext(r.Context())

		authInfo, err := m.userRepo.GetEmployeeByUser(ctx, merchantId, userId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return &apperr.APIError{
					Status: http.StatusUnauthorized,
					Err:    ErrUserNotMerchantEmployee,
				}
			}

			return err
		}

		ctx = actor.SetMerchantIdInContext(ctx, merchantId)
		ctx = actor.SetLocationIdInContext(ctx, authInfo.LocationId)
		ctx = actor.SetEmployeeIdInContext(ctx, authInfo.Id)
		ctx = actor.SetEmployeeRoleInContext(ctx, authInfo.Role)

		next.ServeHTTP(w, r.WithContext(ctx))

		return nil
	}
}

func getUserIdFromClaims(claims jwtlib.MapClaims) (uuid.UUID, error) {
	uuidStr, err := claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("jwt: error parsing claims: %w", err)
	}

	userID, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("jwt: error parsing claims: %w", err)
	}

	return userID, nil
}

func getRefreshVersionFromClaims(claims jwtlib.MapClaims) (int, bool) {
	val, ok := claims["refresh_version"]
	if !ok {
		return 0, false
	}

	switch refreshVersion := val.(type) {
	case float64:
		return int(refreshVersion), true

	case json.Number:
		val, _ := refreshVersion.Float64()

		return int(val), true
	}

	return 0, false
}

var ErrInvalidAccessToken = &apperr.Error{Code: "invalid_access_token", Message: "invalid access token"}
var ErrInvalidRefreshToken = &apperr.Error{Code: "invalid_refresh_token", Message: "invalid refresh token"}

// parse and validate jwt, returning the claims if valid
func verifyToken(tokenString string, tokenType jwt.JwtType) (jwtlib.MapClaims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, jwtlib.MapClaims{}, func(token *jwtlib.Token) (any, error) {
		switch tokenType {
		case jwt.AccessToken:
			return []byte(config.LoadEnvVars().JWT_ACCESS_SECRET), nil
		case jwt.RefreshToken:
			return []byte(config.LoadEnvVars().JWT_REFRESH_SECRET), nil
		default:
			return "", fmt.Errorf("jwt: unexpected token type: %v", tokenType)
		}
	})
	if err != nil {
		switch tokenType {
		case jwt.AccessToken:
			return nil, apperr.Wrap(ErrInvalidAccessToken, err)
		case jwt.RefreshToken:
			return nil, apperr.Wrap(ErrInvalidRefreshToken, err)
		default:
			return nil, err
		}
	}

	claims, ok := token.Claims.(jwtlib.MapClaims)
	if !ok || !token.Valid {
		switch tokenType {
		case jwt.AccessToken:
			return nil, ErrInvalidAccessToken
		case jwt.RefreshToken:
			return nil, ErrInvalidRefreshToken
		default:
			return nil, fmt.Errorf("jwt: invalid token")
		}
	}

	return claims, nil
}

var ErrTokenMissing = errors.New("jwt: token is missing")

// check if a token is sent with the request
func verifyRequest(r *http.Request, tokenType jwt.JwtType, findTokenFns ...func(r *http.Request, tokenType jwt.JwtType) string) (jwtlib.MapClaims, error) {
	var tokenString string

	for _, fn := range findTokenFns {
		tokenString = fn(r, tokenType)
		if tokenString != "" {
			break
		}
	}
	if tokenString == "" {
		return nil, ErrTokenMissing
	}

	return verifyToken(tokenString, tokenType)
}

func getTokenFromCookie(r *http.Request, tokenType jwt.JwtType) string {
	var cookieName string

	switch tokenType {
	case jwt.AccessToken:
		cookieName = jwt.AccessCookieName
	case jwt.RefreshToken:
		cookieName = jwt.RefreshCookieName
	default:
		assert.Never("Jwt token type can be either refresh or access", tokenType)
	}

	cookie, err := r.Cookie(cookieName)

	if err != nil {
		return ""
	}

	return cookie.Value
}
