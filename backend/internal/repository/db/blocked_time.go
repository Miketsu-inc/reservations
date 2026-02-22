package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

type blockedTimeRepository struct {
	db *pgxpool.Pool
}

func NewBlockedTimeRepository(db *pgxpool.Pool) domain.BlockedTimeRepository {
	return &blockedTimeRepository{db: db}
}

func (r *blockedTimeRepository) NewBlockedTime(ctx context.Context, merchantId uuid.UUID, employeeIds []int, name string, fromDate, toDate time.Time, allDay bool, blockedTypeId *int) ([]int, error) {
	query := `
	insert into "BlockedTime" (merchant_id, employee_id, blocked_type_id, name, from_date, to_date, all_day) values ($1, $2, $3, $4, $5, $6, $7)
	returning id`

	var ids []int

	for _, empId := range employeeIds {
		var id int

		err := r.db.QueryRow(ctx, query, merchantId, empId, blockedTypeId, name, fromDate, toDate, allDay).Scan(&id)
		if err != nil {
			return []int{}, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (r *blockedTimeRepository) DeleteBlockedTime(ctx context.Context, blockedTimeId int, merchantId uuid.UUID, employeeId int) error {
	query := `
	delete from "BlockedTime"
	where merchant_id = $1 and employee_id = $2 and ID = $3`

	_, err := r.db.Exec(ctx, query, merchantId, employeeId, blockedTimeId)
	if err != nil {
		return err
	}

	return nil
}

func (r *blockedTimeRepository) UpdateBlockedTime(ctx context.Context, bt domain.BlockedTime) error {
	query := `
	update "BlockedTime"
	set blocked_type_id = $4, name = $5, from_date = $6, to_date = $7, all_day = $8
	where merchant_id = $1 and employee_id = $2 and ID = $3`

	_, err := r.db.Exec(ctx, query, bt.MerchantId, bt.EmployeeId, bt.Id, bt.BlockedTypeId, bt.Name, bt.FromDate, bt.ToDate, bt.AllDay)
	if err != nil {
		return err
	}

	return nil
}

func (r *blockedTimeRepository) GetBlockedTimes(ctx context.Context, merchantId uuid.UUID, start, end time.Time) ([]domain.BlockedTimes, error) {
	query := `
	select from_date, to_date, all_day from "BlockedTime"
	where merchant_id = $1 and to_date > $2 and from_date < $3
	order by from_date`

	rows, _ := r.db.Query(ctx, query, merchantId, start, end)
	blockedTimes, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BlockedTimes])
	if err != nil {
		return nil, err
	}

	return blockedTimes, nil

}

func (r *blockedTimeRepository) GetBlockedTimeById(ctx context.Context, blockedTimeId int) (domain.BlockedTime, error) {
	query := `
	select *
	from "BlockedTime"
	where id = $1
	`

	rows, _ := r.db.Query(ctx, query, blockedTimeId)
	blockedTime, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BlockedTime])
	if err != nil {
		return domain.BlockedTime{}, err
	}

	return blockedTime, nil
}

func (r *blockedTimeRepository) GetAllBlockedTimeTypes(ctx context.Context, merchantId uuid.UUID) ([]domain.BlockedTimeType, error) {
	query := `
	select ID, name, duration, icon from "BlockedTimeType"
	where merchant_id = $1
	order by id asc
	`

	rows, _ := r.db.Query(ctx, query, merchantId)
	types, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BlockedTimeType])
	if err != nil {
		return []domain.BlockedTimeType{}, err
	}

	return types, nil
}

func (r *blockedTimeRepository) NewBlockedTimeType(ctx context.Context, merchantId uuid.UUID, btt domain.BlockedTimeType) error {
	query := `
	insert into "BlockedTimeType" (merchant_id, name, duration, icon)
	values ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query, merchantId, btt.Name, btt.Duration, btt.Icon)
	if err != nil {
		return err
	}

	return nil
}

func (r *blockedTimeRepository) UpdateBlockedTimeType(ctx context.Context, merchantId uuid.UUID, btt domain.BlockedTimeType) error {
	query := `
	update "BlockedTimeType"
	set name = $3, duration = $4, icon = $5
	where merchant_id = $1 and id = $2
	`
	_, err := r.db.Exec(ctx, query, merchantId, btt.Id, btt.Name, btt.Duration, btt.Icon)
	if err != nil {
		return err
	}

	return nil
}

func (r *blockedTimeRepository) DeleteBlockedTimeType(ctx context.Context, merchantId uuid.UUID, typeId int) error {
	query := `
	delete from "BlockedTimeType" where merchant_id = $1 and id = $2`

	_, err := r.db.Exec(ctx, query, merchantId, typeId)
	if err != nil {
		return err
	}

	return nil
}
