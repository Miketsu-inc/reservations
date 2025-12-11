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
	"github.com/miketsu-inc/reservations/backend/cmd/types"
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

var userIDCtxKey = &contextKey{"UserID"}
var merchantIDCtxKey = &contextKey{"MerchnatID"}
var employeeIDCtxKey = &contextKey{"EmployeeID"}
var locationIDCtxKey = &contextKey{"LocationID"}
var employeeRoleCtxKey = &contextKey{"EmployeeRole"}

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

type EmployeeContext struct {
	Id         int
	Role       types.EmployeeRole
	LocationId int
	UserId     uuid.UUID
	MerchantId uuid.UUID
}

// Get employee details from the request's context. Panics if not present!
func MustGetEmployeeFromContext(ctx context.Context) EmployeeContext {
	employeeId, hasEmpId := ctx.Value(employeeIDCtxKey).(int)
	locationId, hasLocId := ctx.Value(locationIDCtxKey).(int)
	role, hasRole := ctx.Value(employeeRoleCtxKey).(types.EmployeeRole)
	userId, hasUsrId := ctx.Value(userIDCtxKey).(uuid.UUID)
	merchantId, hasMerchId := ctx.Value(merchantIDCtxKey).(uuid.UUID)

	assert.True(hasEmpId, "employee id not in context", ctx.Value(employeeIDCtxKey), employeeId, hasEmpId)
	assert.True(hasLocId, "location id not in context", ctx.Value(locationIDCtxKey), locationId, hasLocId)
	assert.True(hasRole, "employee role not in context", ctx.Value(employeeRoleCtxKey), role, hasRole)
	assert.True(hasUsrId, "user id not in context", ctx.Value(userIDCtxKey), userId, hasUsrId)
	assert.True(hasMerchId, "merchant id not in context", ctx.Value(merchantIDCtxKey), merchantId, hasMerchId)

	return EmployeeContext{
		Id:         employeeId,
		Role:       role,
		LocationId: locationId,
		UserId:     userId,
		MerchantId: merchantId,
	}
}

// Returns employee details from the request's context
func GetEmployeeFromContext(ctx context.Context) (EmployeeContext, bool) {
	employeeId, hasEmpId := ctx.Value(employeeIDCtxKey).(int)
	locationId, hasLocId := ctx.Value(locationIDCtxKey).(int)
	role, hasRole := ctx.Value(employeeRoleCtxKey).(types.EmployeeRole)
	userId, hasUsrId := ctx.Value(userIDCtxKey).(uuid.UUID)
	merchantId, hasMerchId := ctx.Value(merchantIDCtxKey).(uuid.UUID)

	if !hasEmpId || !hasLocId || !hasRole || !hasUsrId || !hasMerchId {
		return EmployeeContext{}, false
	}

	return EmployeeContext{
		Id:         employeeId,
		Role:       role,
		LocationId: locationId,
		UserId:     userId,
		MerchantId: merchantId,
	}, true
}

// Jwt authentication middleware. Uses refresh and access tokens
func JwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// try to verify request with access token
		claims, err := verifyRequest(r, AccessToken, getTokenFromCookie)
		if err != nil {
			// if access token could not be found in cookies it means it's either expired or did not exist
			// if it is found but invalid unauthorized status can be returned
			if !errors.Is(err, ErrJwtNotFound) {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("%v", err.Error()))
				return
			}

			// try to verify request with refresh token
			claims, err = verifyRequest(r, RefreshToken, getTokenFromCookie)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("%v", err.Error()))
				return
			}

			userID, err := getUserIdFromClaims(claims)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("could not parse jwt claims: %s", err.Error()))
				return
			}

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

			merchantId := getMerchantIdFromClaims(claims)
			employeeId := getEmployeeIdFromClaims(claims)
			locationId := getLocationIdFromClaims(claims)
			role := getEmployeeRoleFromCalims(claims)

			err = NewAccessToken(w, userID, merchantId, employeeId, locationId, role)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("could not create new access jwt"))
				return
			}
		}

		userID, err := getUserIdFromClaims(claims)
		if err != nil {
			httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("could not parse jwt claims: %s", err.Error()))
			return
		}

		ctx = context.WithValue(ctx, userIDCtxKey, userID)

		if merchantID := getMerchantIdFromClaims(claims); merchantID != nil {
			ctx = context.WithValue(ctx, merchantIDCtxKey, *merchantID)
		}

		if employeeID := getEmployeeIdFromClaims(claims); employeeID != nil {
			ctx = context.WithValue(ctx, employeeIDCtxKey, *employeeID)
		}

		if locationID := getLocationIdFromClaims(claims); locationID != nil {
			ctx = context.WithValue(ctx, locationIDCtxKey, *locationID)
		}

		if employeeRole := getEmployeeRoleFromCalims(claims); employeeRole != nil {
			ctx = context.WithValue(ctx, employeeRoleCtxKey, *employeeRole)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
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

// Adds the employee relevant fields to the claim if they are not nil
func addEmployeeClaims(claims jwtlib.MapClaims, merchantId *uuid.UUID, employeeId *int, locationId *int, role *types.EmployeeRole) {
	if merchantId == nil {
		return
	}

	assert.NotNil(employeeId, "jwt: employeeId can't be nil if merchantId is not nil", employeeId, merchantId)
	assert.NotNil(locationId, "jwt: locationId can't be nil if merchantId is not nil", locationId, merchantId)
	assert.NotNil(role, "jwt: role can't be nil if merchantId is not nil", role, merchantId)

	claims["merchant_id"] = merchantId.String()
	claims["employee_id"] = *employeeId
	claims["location_id"] = *locationId
	claims["employee_role"] = role.String()
}

// Create a new access token and set it in cookies
func NewAccessToken(w http.ResponseWriter, userID uuid.UUID, merchantId *uuid.UUID, employeeId *int, locationId *int, role *types.EmployeeRole) error {
	expMin := config.LoadEnvVars().JWT_ACCESS_EXP_MIN
	expMinDuration := time.Minute * time.Duration(expMin)

	claims := jwtlib.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(expMinDuration).Unix(),
	}

	addEmployeeClaims(claims, merchantId, employeeId, locationId, role)

	token, err := new([]byte(config.LoadEnvVars().JWT_ACCESS_SECRET), claims)
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
		Domain:   ".reservations.local",
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// Create a new refresh token and set it in cookies
func NewRefreshToken(w http.ResponseWriter, userID uuid.UUID, merchantId *uuid.UUID, employeeId *int, locationId *int, role *types.EmployeeRole, refreshVersion int) error {
	expMin := config.LoadEnvVars().JWT_REFRESH_EXP_MIN
	expMinDuration := time.Minute * time.Duration(expMin)

	claims := jwtlib.MapClaims{
		"sub":             userID,
		"exp":             time.Now().Add(expMinDuration).Unix(),
		"refresh_version": refreshVersion,
	}

	addEmployeeClaims(claims, merchantId, employeeId, locationId, role)

	token, err := new([]byte(config.LoadEnvVars().JWT_REFRESH_SECRET), claims)
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
		Domain:   ".reservations.local",
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
		Domain:   ".reservations.local",
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
		Domain:   ".reservations.local",
		SameSite: http.SameSiteLaxMode,
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
func verifyToken(tokenString string, tokenType JwtType) (jwtlib.MapClaims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, jwtlib.MapClaims{}, func(token *jwtlib.Token) (interface{}, error) {
		switch tokenType {
		case AccessToken:
			return []byte(config.LoadEnvVars().JWT_ACCESS_SECRET), nil
		case RefreshToken:
			return []byte(config.LoadEnvVars().JWT_REFRESH_SECRET), nil
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

var ErrJwtNotFound = errors.New("jwt token could not be found")

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
		return nil, ErrJwtNotFound
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
