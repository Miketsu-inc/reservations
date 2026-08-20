package team

import (
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	teamServ "github.com/miketsu-inc/reservations/backend/internal/service/team"
)

func mapToNewMemberInput(in newMemberReq) teamServ.NewMemberInput {
	return teamServ.NewMemberInput{
		Role:        in.Role,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		IsActive:    in.IsActive,
	}
}

func mapToUpdateMemberInput(in updateMemberReq) teamServ.UpdateMemberInput {
	return teamServ.UpdateMemberInput{
		Role:        in.Role,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		IsActive:    in.IsActive,
	}
}

func mapToGetMemberResp(in domain.PublicEmployee) getMemberResp {
	return getMemberResp{
		Id:          in.Id,
		Role:        in.Role,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		IsActive:    in.IsActive,
	}
}

func mapToGetPreferencesResp(in domain.EmployeePreferences) getPreferencesResp {
	return getPreferencesResp{
		FirstDayOfWeek:     in.FirstDayOfWeek,
		TimeFormat:         in.TimeFormat,
		CalendarView:       in.CalendarView,
		CalendarViewMobile: in.CalendarViewMobile,
		StartHour:          in.StartHour.Format("15:04"),
		EndHour:            in.EndHour.Format("15:04"),
		TimeFrequency:      in.TimeFrequency.Format("15:04"),
	}
}

func mapToUpdatePreferencesInput(in updatePreferencesReq) (teamServ.UpdatePreferencesInput, error) {
	startHour, err := time.Parse("15:04", in.StartHour)
	if err != nil {
		return teamServ.UpdatePreferencesInput{}, err
	}

	endHour, err := time.Parse("15:04", in.EndHour)
	if err != nil {
		return teamServ.UpdatePreferencesInput{}, err
	}

	timeFreq, err := time.Parse("15:04", in.TimeFrequency)
	if err != nil {
		return teamServ.UpdatePreferencesInput{}, err
	}

	return teamServ.UpdatePreferencesInput{
		FirstDayOfWeek:     in.FirstDayOfWeek,
		TimeFormat:         in.TimeFormat,
		CalendarView:       in.CalendarView,
		CalendarViewMobile: in.CalendarViewMobile,
		StartHour:          startHour,
		EndHour:            endHour,
		TimeFrequency:      timeFreq,
	}, nil
}
