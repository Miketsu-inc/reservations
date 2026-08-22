package workers

import (
	"context"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/riverqueue/river"
)

type HandleInvitationExpiration struct {
	river.WorkerDefaults[args.HandleInvitationExpiration]

	teamRepo domain.TeamRepository
}

func NewHandleInvitationExpiration(teamRepo domain.TeamRepository) *HandleInvitationExpiration {
	return &HandleInvitationExpiration{teamRepo: teamRepo}
}

func (w *HandleInvitationExpiration) Work(ctx context.Context, job *river.Job[args.HandleInvitationExpiration]) error {
	invitations, err := w.teamRepo.GetExpiredEmployeeInvitations(ctx, time.Now().UTC(), 100)
	if err != nil {
		return err
	}

	var invIds []int

	for _, inv := range invitations {
		if inv.IsExpired() && inv.Status == types.EmployeeInvitationStatusPending {
			invIds = append(invIds, inv.Id)
		}
	}

	return w.teamRepo.UpdateEmployeeInvitationsStatus(ctx, invIds, types.EmployeeInvitationStatusExpired)
}
