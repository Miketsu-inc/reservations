package actor

import (
	"context"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

type contextKey struct {
	name string
}

var merchantIDCtxKey = &contextKey{"MerchantID"}
var employeeIDCtxKey = &contextKey{"EmployeeID"}
var employeeRoleCtxKey = &contextKey{"EmployeeRole"}

type EmployeeContext struct {
	UserId     uuid.UUID
	MerchantId uuid.UUID
	EmployeeId int
	Role       types.EmployeeRole
}

// Get employee details from the request's context. Panics if not present!
func MustGetFromContext(ctx context.Context) EmployeeContext {
	employeeId, hasEmpId := ctx.Value(employeeIDCtxKey).(int)
	role, hasRole := ctx.Value(employeeRoleCtxKey).(types.EmployeeRole)
	merchantId, hasMerchId := ctx.Value(merchantIDCtxKey).(uuid.UUID)
	userId := jwt.MustGetUserIDFromContext(ctx)

	assert.True(hasEmpId, "employee id not in context", ctx.Value(employeeIDCtxKey), employeeId, hasEmpId)
	assert.True(hasRole, "employee role not in context", ctx.Value(employeeRoleCtxKey), role, hasRole)
	assert.True(hasMerchId, "merchant id not in context", ctx.Value(merchantIDCtxKey), merchantId, hasMerchId)

	return EmployeeContext{
		UserId:     userId,
		MerchantId: merchantId,
		EmployeeId: employeeId,
		Role:       role,
	}
}

// Returns employee details from the request's context
func GetFromContext(ctx context.Context) (EmployeeContext, bool) {
	employeeId, hasEmpId := ctx.Value(employeeIDCtxKey).(int)
	role, hasRole := ctx.Value(employeeRoleCtxKey).(types.EmployeeRole)
	merchantId, hasMerchId := ctx.Value(merchantIDCtxKey).(uuid.UUID)
	userId, hasUsrId := jwt.GetUserIDFromContext(ctx)

	if !hasEmpId || !hasRole || !hasUsrId || !hasMerchId {
		return EmployeeContext{}, false
	}

	return EmployeeContext{
		UserId:     userId,
		MerchantId: merchantId,
		EmployeeId: employeeId,
		Role:       role,
	}, true
}

func SetMerchantIdInContext(ctx context.Context, merchantId uuid.UUID) context.Context {
	return context.WithValue(ctx, merchantIDCtxKey, merchantId)
}

func SetEmployeeIdInContext(ctx context.Context, employeeId int) context.Context {
	return context.WithValue(ctx, employeeIDCtxKey, employeeId)
}

func SetEmployeeRoleInContext(ctx context.Context, employeeRole types.EmployeeRole) context.Context {
	return context.WithValue(ctx, employeeRoleCtxKey, employeeRole)
}
