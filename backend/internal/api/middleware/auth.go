package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
)

// Jwt authentication middleware. Uses refresh and access tokens
func (m *Manager) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// try to verify request with access token
		claims, err := verifyRequest(r, jwt.AccessToken, getTokenFromCookie)
		if err != nil {
			// if access token could not be found in cookies it means it's either expired or did not exist
			// if it is found but invalid unauthorized status can be returned
			if !errors.Is(err, ErrJwtNotFound) {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("%v", err.Error()))
				return
			}

			// try to verify request with refresh token
			claims, err = verifyRequest(r, jwt.RefreshToken, getTokenFromCookie)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("%v", err.Error()))
				return
			}

			userID, err := getUserIdFromClaims(claims)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("could not parse jwt claims: %s", err.Error()))
				return
			}

			dbRefreshVersion, err := m.userRepo.GetUserJwtRefreshVersion(ctx, userID)
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
				jwt.DeleteJwts(w)
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("refresh token version does not match"))
				return
			}

			merchantId := getMerchantIdFromClaims(claims)
			employeeId := getEmployeeIdFromClaims(claims)
			locationId := getLocationIdFromClaims(claims)
			role := getEmployeeRoleFromCalims(claims)

			token, err := jwt.NewAccessToken(userID, merchantId, employeeId, locationId, role)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("could not create new access jwt"))
				return
			}

			jwt.SetJwtCookie(w, jwt.AccessToken, token)
		}

		userID, err := getUserIdFromClaims(claims)
		if err != nil {
			httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("could not parse jwt claims: %s", err.Error()))
			return
		}

		ctx = jwt.SetUsedIdInContext(ctx, userID)

		if merchantID := getMerchantIdFromClaims(claims); merchantID != nil {
			ctx = jwt.SetMerchantIdInContext(ctx, *merchantID)
		}

		if employeeID := getEmployeeIdFromClaims(claims); employeeID != nil {
			ctx = jwt.SetEmployeeIdInContext(ctx, *employeeID)
		}

		if locationID := getLocationIdFromClaims(claims); locationID != nil {
			ctx = jwt.SetLocationIdInContext(ctx, *locationID)
		}

		if employeeRole := getEmployeeRoleFromCalims(claims); employeeRole != nil {
			ctx = jwt.SetEmployeeRoleInContext(ctx, *employeeRole)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserIdFromClaims(claims jwtlib.MapClaims) (uuid.UUID, error) {
	uuidStr, err := claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
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

func getMerchantIdFromClaims(claims jwtlib.MapClaims) *uuid.UUID {
	val, ok := claims["merchant_id"]
	if !ok {
		return nil
	}

	merchantIdStr, ok := val.(string)
	if !ok {
		return nil
	}

	merchantId, err := uuid.Parse(merchantIdStr)
	if err != nil {
		return nil
	}

	return &merchantId
}

func getEmployeeIdFromClaims(claims jwtlib.MapClaims) *int {
	val, ok := claims["employee_id"]
	if !ok {
		return nil
	}

	switch id := val.(type) {
	case float64:
		employeeId := int(id)
		return &employeeId
	case json.Number:
		num, _ := id.Float64()
		employeeId := int(num)
		return &employeeId
	}

	return nil
}

func getLocationIdFromClaims(claims jwtlib.MapClaims) *int {
	val, ok := claims["location_id"]
	if !ok {
		return nil
	}

	switch id := val.(type) {
	case float64:
		locationId := int(id)
		return &locationId
	case json.Number:
		num, _ := id.Float64()
		locationId := int(num)
		return &locationId
	}

	return nil
}

func getEmployeeRoleFromCalims(claims jwtlib.MapClaims) *types.EmployeeRole {
	val, ok := claims["employee_role"]
	if !ok {
		return nil
	}

	roleStr, ok := val.(string)
	if !ok {
		return nil
	}

	role, err := types.NewEmployeeRole(roleStr)
	if err != nil {
		return nil
	}

	return &role
}

// parse and validate jwt, returning the claims if valid
func verifyToken(tokenString string, tokenType jwt.JwtType) (jwtlib.MapClaims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, jwtlib.MapClaims{}, func(token *jwtlib.Token) (any, error) {
		switch tokenType {
		case jwt.AccessToken:
			return []byte(config.LoadEnvVars().JWT_ACCESS_SECRET), nil
		case jwt.RefreshToken:
			return []byte(config.LoadEnvVars().JWT_REFRESH_SECRET), nil
		default:
			assert.Never("jwt token type can be either refresh or access", tokenType)
			return "", fmt.Errorf("jwt token type can be either refresh or access")
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

var ErrJwtNotFound = errors.New("jwt token could not be found")

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
		return nil, ErrJwtNotFound
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
