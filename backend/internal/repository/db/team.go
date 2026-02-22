package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

type teamRepository struct {
	db *pgxpool.Pool
}

func NewTeamRepository(db *pgxpool.Pool) domain.TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) GetEmployeesByMerchant(ctx context.Context, merchantId uuid.UUID) ([]domain.PublicEmployee, error) {
	query := `
	select e.id, e.role, coalesce(e.first_name, u.first_name) as first_name, coalesce(e.last_name, u.last_name) as last_name,
		coalesce(e.email, u.email) as email, coalesce(e.phone_number, u.phone_number) as phone_number, e.is_active
	from "Employee" e
	left join "User" u on u.id = e.user_id
	where merchant_id = $1`

	rows, _ := r.db.Query(ctx, query, merchantId)
	members, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.PublicEmployee])
	if err != nil {
		return []domain.PublicEmployee{}, err
	}

	return members, nil
}

func (r *teamRepository) GetEmployeeById(ctx context.Context, merchantId uuid.UUID, memberId int) (domain.PublicEmployee, error) {
	query := `
	select e.id, e.role, coalesce(e.first_name, u.first_name) as first_name, coalesce(e.last_name, u.last_name) as last_name,
		coalesce(e.email, u.email) as email, coalesce(e.phone_number, u.phone_number) as phone_number, e.is_active
	from "Employee" e
	left join "User" u on u.id = e.user_id
	where merchant_id = $1 and e.id = $2
	`

	rows, _ := r.db.Query(ctx, query, merchantId, memberId)
	member, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.PublicEmployee])
	if err != nil {
		return domain.PublicEmployee{}, err
	}

	return member, nil
}

func (r *teamRepository) NewEmployee(ctx context.Context, merchantId uuid.UUID, emp domain.PublicEmployee) error {
	query := `
	insert into "Employee" (merchant_id, role, first_name, last_name, email, phone_number, is_active)
	values ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query, merchantId, emp.Role, emp.FirstName, emp.LastName, emp.Email, emp.PhoneNumber, emp.IsActive)
	if err != nil {
		return err
	}

	return nil
}

func (r *teamRepository) UpdateEmployeeById(ctx context.Context, merchantId uuid.UUID, employee domain.PublicEmployee) error {
	query := `
	update "Employee"
	set role = $3, first_name = $4, last_name = $5, email = $6, phone_number = $7, is_active = $8
	where merchant_id = $1 and id = $2
	`

	_, err := r.db.Exec(ctx, query, merchantId, employee.Id, employee.Role, employee.FirstName, employee.LastName, employee.Email,
		employee.PhoneNumber, employee.IsActive)
	if err != nil {
		return err
	}

	return nil
}

func (r *teamRepository) DeleteEmployeeById(ctx context.Context, merchantId uuid.UUID, employeeId int) error {
	query := `
	delete from "Employee"
	where merchant_id = $1 and id = $2 and role not in ('owner')
	`

	_, err := r.db.Exec(ctx, query, merchantId, employeeId)
	if err != nil {
		return err
	}

	return nil
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
		return uuid.UUID{}, nil
	}

	return merchantId, nil
}

func (r *teamRepository) GetEmployeesForCalendarByMerchant(ctx context.Context, merchantId uuid.UUID) ([]domain.EmployeeForCalendar, error) {
	query := `
	select e.id, coalesce(e.first_name, u.first_name) as first_name, coalesce(e.last_name, u.last_name) as last_name
	from "Employee" e
	left join "User" u on u.id = e.user_id
	where merchant_id = $1`

	rows, _ := r.db.Query(ctx, query, merchantId)
	employees, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.EmployeeForCalendar])
	if err != nil {
		return []domain.EmployeeForCalendar{}, err
	}

	return employees, nil
}
