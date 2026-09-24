package blockedtimes

import (
	"fmt"
	"time"

	blockedtimeServ "github.com/miketsu-inc/reservations/backend/internal/service/blockedtime"
)

func mapToNewInput(in newReq) (blockedtimeServ.NewInput, error) {
	input := blockedtimeServ.NewInput{
		Name:          in.Name,
		EmployeeIds:   in.EmployeeIds,
		BlockedTypeId: in.BlockedTypeId,
		IsAllDay:      in.IsAllDay,
	}

	if in.IsAllDay {
		blockedDay, err := time.Parse(time.DateOnly, in.BlockedDay)
		if err != nil {
			return blockedtimeServ.NewInput{}, fmt.Errorf("invalid date: %s", err.Error())
		}

		input.BlockedDay = &blockedDay
		return input, nil
	}

	fromDate, err := time.Parse(time.RFC3339, in.FromDate)
	if err != nil {
		return blockedtimeServ.NewInput{}, fmt.Errorf("invalid from date: %s", err.Error())
	}

	toDate, err := time.Parse(time.RFC3339, in.ToDate)
	if err != nil {
		return blockedtimeServ.NewInput{}, fmt.Errorf("invalid to date: %s", err.Error())
	}

	input.FromDate = &fromDate
	input.ToDate = &toDate

	return input, nil
}

func mapToUpdateInput(in updateReq) (blockedtimeServ.UpdateInput, error) {
	input := blockedtimeServ.UpdateInput{
		BlockedTimeId: in.Id,
		Name:          in.Name,
		BlockedTypeId: in.BlockedTypeId,
		EmployeeIds:   in.EmployeeIds,
		IsAllDay:      in.IsAllDay,
	}

	if in.IsAllDay {
		blockedDay, err := time.Parse(time.DateOnly, in.BlockedDay)
		if err != nil {
			return blockedtimeServ.UpdateInput{}, fmt.Errorf("invalid date: %s", err.Error())
		}

		input.BlockedDay = &blockedDay
		return input, nil
	}

	fromDate, err := time.Parse(time.RFC3339, in.FromDate)
	if err != nil {
		return blockedtimeServ.UpdateInput{}, fmt.Errorf("invalid from date: %s", err.Error())
	}

	toDate, err := time.Parse(time.RFC3339, in.ToDate)
	if err != nil {
		return blockedtimeServ.UpdateInput{}, fmt.Errorf("invalid to date: %s", err.Error())
	}

	input.FromDate = &fromDate
	input.ToDate = &toDate

	return input, nil
}
