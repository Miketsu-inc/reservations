package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type teamRepository struct {
	db db.DBTX
}

func NewTeamRepository(db db.DBTX) domain.TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) WithTx(tx db.DBTX) domain.TeamRepository {
	return &teamRepository{db: tx}
}

func (r *teamRepository) NewEmployee(ctx context.Context, emp domain.Employee) (int, error) {
	query := `
	insert into "Employee" (user_id, merchant_id, role, first_name, last_name, email, phone_number, is_active)
	values ($1, $2, $3, $4, $5, $6, $7, $8)
	returning id
	`

	var employeeId int
	err := r.db.QueryRow(ctx, query, emp.UserId, emp.MerchantId, emp.Role, emp.FirstName, emp.LastName, emp.Email, emp.PhoneNumber,
		emp.IsActive).Scan(&employeeId)
	if err != nil {
		return 0, fmt.Errorf("NewEmployee: %w", err)
	}

	return employeeId, nil
}

func (r *teamRepository) UpdateEmployee(ctx context.Context, employee domain.Employee) error {
	query := `
	update "Employee"
	set role = $3, first_name = $4, last_name = $5, email = $6, phone_number = $7, is_active = $8
	where merchant_id = $1 and id = $2
	`

	_, err := r.db.Exec(ctx, query, employee.MerchantId, employee.Id, employee.Role, employee.FirstName, employee.LastName, employee.Email,
		employee.PhoneNumber, employee.IsActive)
	if err != nil {
		return fmt.Errorf("UpdateEmployee: %w", err)
	}

	return nil
}

func (r *teamRepository) DeleteEmployee(ctx context.Context, merchantId uuid.UUID, employeeId int) error {
	query := `
	delete from "Employee"
	where merchant_id = $1 and id = $2 and role not in ('owner')
	`

	_, err := r.db.Exec(ctx, query, merchantId, employeeId)
	if err != nil {
		return fmt.Errorf("DeleteEmployee: %w", err)
	}

	return nil
}

func (r *teamRepository) GetEmployee(ctx context.Context, merchantId uuid.UUID, memberId int) (domain.Employee, error) {
	query := `
	select e.id, e.user_id, e.merchant_id, e.role, coalesce(e.first_name, u.first_name) as first_name, coalesce(e.last_name, u.last_name) as last_name,
		coalesce(e.email, u.email) as email, coalesce(e.phone_number, u.phone_number) as phone_number, e.is_active
	from "Employee" e
	left join "User" u on u.id = e.user_id
	where merchant_id = $1 and e.id = $2
	`

	rows, _ := r.db.Query(ctx, query, merchantId, memberId)
	member, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.Employee])
	if err != nil {
		return domain.Employee{}, fmt.Errorf("GetEmployee: %w", err)
	}

	return member, nil
}

func (r *teamRepository) GetEmployees(ctx context.Context, merchantId uuid.UUID) ([]domain.Employee, error) {
	query := `
	select e.id, e.user_id, e.merchant_id, e.role, coalesce(e.first_name, u.first_name) as first_name, coalesce(e.last_name, u.last_name) as last_name,
		coalesce(e.email, u.email) as email, coalesce(e.phone_number, u.phone_number) as phone_number, e.is_active
	from "Employee" e
	left join "User" u on u.id = e.user_id
	where merchant_id = $1`

	rows, _ := r.db.Query(ctx, query, merchantId)
	members, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Employee])
	if err != nil {
		return []domain.Employee{}, fmt.Errorf("GetEmployees: %w", err)
	}

	return members, nil
}

func (r *teamRepository) GetActiveEmployees(ctx context.Context, merchantId uuid.UUID) ([]domain.Employee, error) {
	query := `
	select e.id, e.user_id, e.merchant_id, e.role, coalesce(e.first_name, u.first_name) as first_name, coalesce(e.last_name, u.last_name) as last_name,
		coalesce(e.email, u.email) as email, coalesce(e.phone_number, u.phone_number) as phone_number, e.is_active
	from "Employee" e
	left join "User" u on u.id = e.user_id
	where merchant_id = $1 and e.is_active is true`

	rows, _ := r.db.Query(ctx, query, merchantId)
	members, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Employee])
	if err != nil {
		return []domain.Employee{}, fmt.Errorf("GetActiveEmployees: %w", err)
	}

	return members, nil
}

func (r *teamRepository) GetMerchantIdByEmployee(ctx context.Context, employeeId int) (uuid.UUID, error) {
	query := `
	select merchant_id
	from "Employee"
	where id = $1
	`

	var merchantId uuid.UUID

	err := r.db.QueryRow(ctx, query, employeeId).Scan(&merchantId)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("GetMerchantIdByEmployee: %w", err)
	}

	return merchantId, nil
}

func (r *teamRepository) NewEmployeeInvitation(ctx context.Context, inv domain.EmployeeInvitation) (int, error) {
	query := `
	insert into "EmployeeInvitation" (merchant_id, status, email, role, token, invited_by, invited_at, expires_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8)
	returning id
	`

	var invitationId int
	err := r.db.QueryRow(ctx, query, inv.MerchantId, inv.Status, inv.Email, inv.Role, inv.Token, inv.InvitedBy, inv.InvitedAt, inv.ExpiresAt).Scan(&invitationId)
	if err != nil {
		return 0, fmt.Errorf("NewEmployeeInvitation: %w", err)
	}

	return invitationId, nil
}

func (r *teamRepository) UpdateEmployeeInvitation(ctx context.Context, inv domain.EmployeeInvitation) error {
	query := `
	update "EmployeeInvitation"
	set status = $2, role = $3, token = $4, invited_by = $5, invited_at = $6, expires_at = $7
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, inv.Id, inv.Status, inv.Role, inv.Token, inv.InvitedBy, inv.InvitedAt, inv.ExpiresAt)
	if err != nil {
		return fmt.Errorf("UpdateEmployeeInvitation: %w", err)
	}

	return nil
}

func (r *teamRepository) UpdateEmployeeInvitationsStatus(ctx context.Context, invitationIds []int, status types.EmployeeInvitationStatus) error {
	query := `
	update "EmployeeInvitation"
	set status = $2
	where id = any($1::int[])
	`

	_, err := r.db.Exec(ctx, query, invitationIds, status)
	if err != nil {
		return fmt.Errorf("UpdateEmployeeInvitationsStatus: %w", err)
	}

	return nil
}

func (r *teamRepository) AcceptEmployeeInvitation(ctx context.Context, invitationId int) error {
	query := `
	update "EmployeeInvitation"
	set status = 'accepted', accepted_at = $2
	where id = $1 and status in ('pending')
	`

	_, err := r.db.Exec(ctx, query, invitationId, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("AcceptEmployeeInvitation: %w", err)
	}

	return nil
}

func (r *teamRepository) DeclineEmployeeInvitation(ctx context.Context, invitationId int) error {
	query := `
	update "EmployeeInvitation"
	set status = 'declined', declined_at = $2
	where id = $1 and status in ('pending')
	`

	_, err := r.db.Exec(ctx, query, invitationId, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("DeclineEmployeeInvitation: %w", err)
	}

	return nil
}

func (r *teamRepository) RevokeEmployeeInvitation(ctx context.Context, invitationId int) error {
	query := `
	update "EmployeeInvitation"
	set status = 'revoked', revoked_at = $2
	where id = $1 and status in ('pending')
	`

	_, err := r.db.Exec(ctx, query, invitationId, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("RevokeEmployeeInvitation: %w", err)
	}

	return nil
}

func (r *teamRepository) GetEmployeeInvitation(ctx context.Context, invitationId int) (domain.EmployeeInvitation, error) {
	query := `
	select *
	from "EmployeeInvitation"
	where id = $1
	`

	rows, _ := r.db.Query(ctx, query, invitationId)
	invitation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.EmployeeInvitation])
	if err != nil {
		return domain.EmployeeInvitation{}, fmt.Errorf("GetEmployeeInvitation: %w", err)
	}

	return invitation, nil
}

func (r *teamRepository) GetEmployeeInvitationByEmail(ctx context.Context, merchantId uuid.UUID, email string) (domain.EmployeeInvitation, error) {
	query := `
	select *
	from "EmployeeInvitation"
	where merchant_id = $1 and email = $2
	`

	rows, _ := r.db.Query(ctx, query, merchantId, email)
	invitation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.EmployeeInvitation])
	if err != nil {
		return domain.EmployeeInvitation{}, fmt.Errorf("GetEmployeeInvitationByEmail: %w", err)
	}

	return invitation, nil
}

func (r *teamRepository) GetEmployeeInvitationByToken(ctx context.Context, token string) (domain.EmployeeInvitation, error) {
	query := `
	select *
	from "EmployeeInvitation"
	where token = $1
	`

	rows, _ := r.db.Query(ctx, query, token)
	invitation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.EmployeeInvitation])
	if err != nil {
		return domain.EmployeeInvitation{}, fmt.Errorf("GetEmployeeInvitationByToken: %w", err)
	}

	return invitation, nil
}

func (r *teamRepository) GetEmployeeInvitations(ctx context.Context, merchantId uuid.UUID) ([]domain.EmployeeInvitation, error) {
	query := `
	select *
	from "EmployeeInvitation"
	where merchant_id = $1
	`

	rows, _ := r.db.Query(ctx, query, merchantId)
	invitations, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.EmployeeInvitation])
	if err != nil {
		return []domain.EmployeeInvitation{}, fmt.Errorf("GetEmployeeInvitations: %w", err)
	}

	return invitations, nil
}

func (r *teamRepository) GetExpiredEmployeeInvitations(ctx context.Context, expiry time.Time, limit int) ([]domain.EmployeeInvitation, error) {
	query := `
	select *
	from "EmployeeInvitation"
	where expires_at <= $1 and status = 'pending'
	limit $2
	`

	rows, _ := r.db.Query(ctx, query, expiry, limit)
	invitations, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.EmployeeInvitation])
	if err != nil {
		return []domain.EmployeeInvitation{}, fmt.Errorf("GetExpiredEmployeeInvitations: %w", err)
	}

	return invitations, nil
}

func (r *teamRepository) NewEmployeePreferences(ctx context.Context, employeeId int) error {
	query := `
	insert into "EmployeePreferences" (employee_id) values ($1)
	`

	_, err := r.db.Exec(ctx, query, employeeId)
	if err != nil {
		return fmt.Errorf("NewEmployeePreferences: %w", err)
	}

	return err
}

func (r *teamRepository) UpdateEmployeePreferences(ctx context.Context, employeeId int, p domain.EmployeePreferences) error {
	query := `
	update "EmployeePreferences"
	set first_day_of_week = $2, time_format = $3, calendar_view = $4, calendar_view_mobile = $5, start_hour = $6, end_hour = $7, time_frequency = $8
	where employee_id = $1
	`

	_, err := r.db.Exec(ctx, query, employeeId, p.FirstDayOfWeek, p.TimeFormat, p.CalendarView, p.CalendarViewMobile, p.StartHour, p.EndHour, p.TimeFrequency)
	if err != nil {
		return fmt.Errorf("UpdateEmployeePreferences: %w", err)
	}

	return nil
}

func (r *teamRepository) GetEmployeePreferences(ctx context.Context, employeeId int) (domain.EmployeePreferences, error) {
	query := `
	select first_day_of_week, time_format, calendar_view, calendar_view_mobile, start_hour, end_hour, time_frequency
	from "EmployeePreferences"
	where employee_id = $1
	`

	var p domain.EmployeePreferences
	err := r.db.QueryRow(ctx, query, employeeId).Scan(&p.FirstDayOfWeek, &p.TimeFormat, &p.CalendarView, &p.CalendarViewMobile, &p.StartHour, &p.EndHour, &p.TimeFrequency)
	if err != nil {
		return domain.EmployeePreferences{}, fmt.Errorf("GetEmployeePreferences: %w", err)
	}

	return p, nil
}
