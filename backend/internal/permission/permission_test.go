package permission

import (
	"testing"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestEmployeeScope(t *testing.T) {

	t.Run("With permission sees everyone", func(t *testing.T) {
		admin := actor.EmployeeContext{EmployeeId: 1, Role: types.EmployeeRoleAdmin}

		assert.Nil(t, EmployeeScope(admin, BookingsAll))
	})

	t.Run("Without permission sees only self", func(t *testing.T) {
		staff := actor.EmployeeContext{EmployeeId: 7, Role: types.EmployeeRoleStaff}

		scope := EmployeeScope(staff, BookingsAll)

		assert.NotNil(t, scope)
		assert.Equal(t, 7, *scope)
	})
}

func TestCanActFor(t *testing.T) {
	staff := actor.EmployeeContext{EmployeeId: 7, Role: types.EmployeeRoleStaff}
	admin := actor.EmployeeContext{EmployeeId: 1, Role: types.EmployeeRoleAdmin}

	assert.True(t, CanActFor(staff, BlockedTimesAll, 7))
	assert.False(t, CanActFor(staff, BlockedTimesAll, 7, 8))
	assert.True(t, CanActFor(admin, BlockedTimesAll, 7, 8))
}

func TestHas(t *testing.T) {
	owner := actor.EmployeeContext{Role: types.EmployeeRoleOwner}
	admin := actor.EmployeeContext{Role: types.EmployeeRoleAdmin}

	assert.True(t, Has(owner, MerchantDelete))
	assert.False(t, Has(admin, MerchantDelete))
	assert.False(t, Has(actor.EmployeeContext{}, SettingsEdit))
}
