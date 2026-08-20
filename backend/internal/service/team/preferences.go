package team

import (
	"context"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

func (s *Service) GetPreferences(ctx context.Context, employeeId int) (domain.EmployeePreferences, error) {
	actor := actor.MustGetFromContext(ctx)

	if actor.EmployeeId != employeeId {
		return domain.EmployeePreferences{}, ErrPreferencesForbidden
	}

	preferences, err := s.teamRepo.GetEmployeePreferences(ctx, employeeId)
	if err != nil {
		return domain.EmployeePreferences{}, err
	}

	return preferences, nil
}

type UpdatePreferencesInput struct {
	FirstDayOfWeek     string
	TimeFormat         string
	CalendarView       string
	CalendarViewMobile string
	StartHour          time.Time
	EndHour            time.Time
	TimeFrequency      time.Time
}

func (s *Service) UpdatePreferences(ctx context.Context, employeeId int, input UpdatePreferencesInput) error {
	actor := actor.MustGetFromContext(ctx)

	if actor.EmployeeId != employeeId {
		return ErrPreferencesForbidden
	}

	err := s.teamRepo.UpdateEmployeePreferences(ctx, employeeId, domain.EmployeePreferences{
		FirstDayOfWeek:     input.FirstDayOfWeek,
		TimeFormat:         input.TimeFormat,
		CalendarView:       input.CalendarView,
		CalendarViewMobile: input.CalendarViewMobile,
		StartHour:          input.StartHour,
		EndHour:            input.EndHour,
		TimeFrequency:      input.TimeFrequency,
	})
	if err != nil {
		return err
	}

	return nil
}
