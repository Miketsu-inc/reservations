package actor

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

type contextKey struct {
	name string
}

var employeeCtxKey = &contextKey{"Employee"}

type EmployeeContext struct {
	UserId       uuid.UUID
	MerchantId   uuid.UUID
	EmployeeId   int
	Role         types.EmployeeRole
	SubTier      types.SubTier
	Timezone     *time.Location
	CurrencyCode string
}

// Get employee details from the request's context. Panics if not present!
func MustGetFromContext(ctx context.Context) EmployeeContext {
	employee, ok := ctx.Value(employeeCtxKey).(EmployeeContext)
	assert.True(ok, "employee not in context", ctx.Value(employeeCtxKey), employee)

	return employee
}

// Returns employee details from the request's context
func GetFromContext(ctx context.Context) (EmployeeContext, bool) {
	employee, ok := ctx.Value(employeeCtxKey).(EmployeeContext)

	return employee, ok
}

func SetInContext(ctx context.Context, employee EmployeeContext) context.Context {
	return context.WithValue(ctx, employeeCtxKey, employee)
}
