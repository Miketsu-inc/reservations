package blockedtimes

import (
	"fmt"
	"time"

	blockedtimeServ "github.com/miketsu-inc/reservations/backend/internal/service/blockedtime"
)

func mapToNewInput(in newReq) (blockedtimeServ.NewInput, error) {
	fromDate, err := time.Parse(time.RFC3339, in.FromDate)
	if err != nil {
		return blockedtimeServ.NewInput{}, fmt.Errorf("invalid from date: %s", err.Error())
	}

	toDate, err := time.Parse(time.RFC3339, in.ToDate)
	if err != nil {
		return blockedtimeServ.NewInput{}, fmt.Errorf("invalid to date: %s", err.Error())
	}

	return blockedtimeServ.NewInput{
		Name:          in.Name,
		BlockedTypeId: in.BlockedTypeId,
		FromDate:      fromDate,
		ToDate:        toDate,
		AllDay:        in.AllDay,
	}, nil
}

func mapToUpdateInput(in updateReq) (blockedtimeServ.UpdateInput, error) {
	fromDate, err := time.Parse(time.RFC3339, in.FromDate)
	if err != nil {
		return blockedtimeServ.UpdateInput{}, fmt.Errorf("invalid from date: %s", err.Error())
	}

	toDate, err := time.Parse(time.RFC3339, in.ToDate)
	if err != nil {
		return blockedtimeServ.UpdateInput{}, fmt.Errorf("invalid to date: %s", err.Error())
	}

	return blockedtimeServ.UpdateInput{
		Id:            in.Id,
		Name:          in.Name,
		BlockedTypeId: in.BlockedTypeId,
		FromDate:      fromDate,
		ToDate:        toDate,
		AllDay:        in.AllDay,
	}, nil
}

// func mapToDeleteInput(in deleteReq) blockedtimeServ.DeleteInput {
// 	return blockedtimeServ.DeleteInput{
// 		EmployeeId: in.EmployeeId,
// 	}
// }
