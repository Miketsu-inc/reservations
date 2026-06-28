package catalog

import (
	"testing"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestDetectPhaseChanges(t *testing.T) {

	t.Run("Basic", func(t *testing.T) {

		existingPhases := []domain.ServicePhase{
			{
				Id:        1,
				ServiceId: 9,
				Name:      "Wash",
				Sequence:  1,
				Duration:  15,
				PhaseType: types.ServicePhaseTypeActive,
			},
			{
				Id:        2,
				ServiceId: 9,
				Name:      "Wait",
				Sequence:  2,
				Duration:  5,
				PhaseType: types.ServicePhaseTypeWait,
			},
			{
				Id:        3,
				ServiceId: 9,
				Name:      "Hair",
				Sequence:  3,
				Duration:  30,
				PhaseType: types.ServicePhaseTypeActive,
			},
		}

		incomingPhases := []domain.ServicePhase{
			{
				Id:        1,
				ServiceId: 9,
				Name:      "Wash",
				Sequence:  1,
				Duration:  20,
				PhaseType: types.ServicePhaseTypeActive,
			},
			{
				Id:        -1,
				ServiceId: 9,
				Name:      "Wait but different",
				Sequence:  2,
				Duration:  10,
				PhaseType: types.ServicePhaseTypeWait,
			},
			{
				Id:        3,
				ServiceId: 9,
				Name:      "Hair",
				Sequence:  3,
				Duration:  30,
				PhaseType: types.ServicePhaseTypeActive,
			},
		}

		toInsert := incomingPhases[1]
		toInsert.Id = 0

		expectedToInsert := []domain.ServicePhase{toInsert}
		expectedToUpdate := []domain.ServicePhase{incomingPhases[0]}
		expectedToDelete := []int{existingPhases[1].Id}

		phaseChanges := detectPhaseChanges(existingPhases, incomingPhases)
		assert.Equal(t, expectedToInsert, phaseChanges.ToInsert, "phases to insert shall match")
		assert.Equal(t, expectedToUpdate, phaseChanges.ToUpdate, "phases to update shall match")
		assert.Equal(t, expectedToDelete, phaseChanges.ToDelete, "phases to delete shall match")
	})

	t.Run("Phase with 0 id", func(t *testing.T) {

		existingPhases := []domain.ServicePhase{
			{
				Id:        0,
				ServiceId: 9,
				Name:      "Wash",
				Sequence:  1,
				Duration:  15,
				PhaseType: types.ServicePhaseTypeActive,
			},
		}

		incomingPhases := []domain.ServicePhase{
			{
				Id:        0,
				ServiceId: 9,
				Name:      "Wash",
				Sequence:  1,
				Duration:  15,
				PhaseType: types.ServicePhaseTypeActive,
			},
			{
				Id:        -1,
				ServiceId: 9,
				Name:      "Wait",
				Sequence:  2,
				Duration:  15,
				PhaseType: types.ServicePhaseTypeWait,
			},
		}

		toInsert := incomingPhases[1]
		toInsert.Id = 0

		expectedToInsert := []domain.ServicePhase{toInsert}
		expectedToUpdate := []domain.ServicePhase{}
		expectedToDelete := []int{}

		phaseChanges := detectPhaseChanges(existingPhases, incomingPhases)
		assert.Equal(t, expectedToInsert, phaseChanges.ToInsert, "phases to insert shall match")
		assert.Equal(t, expectedToUpdate, phaseChanges.ToUpdate, "phases to update shall match")
		assert.Equal(t, expectedToDelete, phaseChanges.ToDelete, "phases to delete shall match")
	})
}
