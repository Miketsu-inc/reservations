package jwt

import (
	"context"
	"fmt"
	"net/http"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/internal/types"
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
var merchantIDCtxKey = &contextKey{"MerchantID"}
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

func SetMerchantIdInContext(ctx context.Context, merchantId uuid.UUID) context.Context {
	return context.WithValue(ctx, merchantIDCtxKey, merchantId)
}

func SetUsedIdInContext(ctx context.Context, userId uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDCtxKey, userId)
}

func SetEmployeeIdInContext(ctx context.Context, employeeId int) context.Context {
	return context.WithValue(ctx, employeeIDCtxKey, employeeId)
}

func SetLocationIdInContext(ctx context.Context, locationId int) context.Context {
	return context.WithValue(ctx, locationIDCtxKey, locationId)
}

func SetEmployeeRoleInContext(ctx context.Context, employeeRole types.EmployeeRole) context.Context {
	return context.WithValue(ctx, employeeRoleCtxKey, employeeRole)
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
func NewAccessToken(userID uuid.UUID, merchantId *uuid.UUID, employeeId *int, locationId *int, role *types.EmployeeRole) (string, error) {
	expMin := config.LoadEnvVars().JWT_ACCESS_EXP_MIN
	expMinDuration := time.Minute * time.Duration(expMin)

	claims := jwtlib.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(expMinDuration).Unix(),
	}

	addEmployeeClaims(claims, merchantId, employeeId, locationId, role)

	token, err := new([]byte(config.LoadEnvVars().JWT_ACCESS_SECRET), claims)
	if err != nil {
		return "", fmt.Errorf("unexpected error when creating jwt token: %s", err.Error())
	}

	return token, nil
}

// Create a new refresh token
func NewRefreshToken(userID uuid.UUID, merchantId *uuid.UUID, employeeId *int, locationId *int, role *types.EmployeeRole, refreshVersion int) (string, error) {
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
