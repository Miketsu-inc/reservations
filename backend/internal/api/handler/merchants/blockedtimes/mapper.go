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
		if in.BlockedDay == nil || in.FromDate != nil || in.ToDate != nil {
			return blockedtimeServ.NewInput{}, fmt.Errorf("an all-day blocked time requires blocked_day only")
		}

		blockedDay, err := time.Parse(time.DateOnly, *in.BlockedDay)
		if err != nil {
			return blockedtimeServ.NewInput{}, fmt.Errorf("invalid blocked_day: %s", err.Error())
		}

		input.BlockedDay = &blockedDay
		return input, nil
	}

	if in.FromDate == nil || in.ToDate == nil || in.BlockedDay != nil {
		return blockedtimeServ.NewInput{}, fmt.Errorf("a timed blocked time requires from_date and to_date only")
	}

	fromDate, err := time.Parse(time.RFC3339, *in.FromDate)
	if err != nil {
		return blockedtimeServ.NewInput{}, fmt.Errorf("invalid from date: %s", err.Error())
	}

	toDate, err := time.Parse(time.RFC3339, *in.ToDate)
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
		if in.BlockedDay == nil || in.FromDate != nil || in.ToDate != nil {
			return blockedtimeServ.UpdateInput{}, fmt.Errorf("an all-day blocked time requires blocked_day only")
		}

		blockedDay, err := time.Parse(time.DateOnly, *in.BlockedDay)
		if err != nil {
			return blockedtimeServ.UpdateInput{}, fmt.Errorf("invalid blocked_day: %s", err.Error())
		}

		input.BlockedDay = &blockedDay
		return input, nil
	}

	if in.FromDate == nil || in.ToDate == nil || in.BlockedDay != nil {
		return blockedtimeServ.UpdateInput{}, fmt.Errorf("a timed blocked time requires from_date and to_date only")
	}

	fromDate, err := time.Parse(time.RFC3339, *in.FromDate)
	if err != nil {
		return blockedtimeServ.UpdateInput{}, fmt.Errorf("invalid from date: %s", err.Error())
	}

	toDate, err := time.Parse(time.RFC3339, *in.ToDate)
	if err != nil {
		return blockedtimeServ.UpdateInput{}, fmt.Errorf("invalid to date: %s", err.Error())
	}

	input.FromDate = &fromDate
	input.ToDate = &toDate

	return input, nil
}
