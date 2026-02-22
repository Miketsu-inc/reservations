package blockedtimetypes

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	blockedtimeServ "github.com/miketsu-inc/reservations/backend/internal/service/blockedtime"
)

func mapToNewTypeInput(in newReq) blockedtimeServ.NewTypeInput {
	return blockedtimeServ.NewTypeInput{
		Name:     in.Name,
		Duration: in.Duration,
		Icon:     in.Icon,
	}
}

func mapToUpdateTypeInput(in updateReq) blockedtimeServ.UpdateTypeInput {
	return blockedtimeServ.UpdateTypeInput{
		Id:       in.Id,
		Name:     in.Name,
		Duration: in.Duration,
		Icon:     in.Icon,
	}
}

func mapToGetTypesResp(in []domain.BlockedTimeType) []getTypesResp {
	out := make([]getTypesResp, len(in))

	for i, btt := range in {
		out[i] = getTypesResp{
			Id:       btt.Id,
			Name:     btt.Name,
			Duration: btt.Duration,
			Icon:     btt.Icon,
		}
	}

	return out
}
