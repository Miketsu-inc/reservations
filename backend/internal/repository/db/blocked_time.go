package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type blockedTimeRepository struct {
	db db.DBTX
}

func NewBlockedTimeRepository(db db.DBTX) domain.BlockedTimeRepository {
	return &blockedTimeRepository{db: db}
}

func (r *blockedTimeRepository) WithTx(tx db.DBTX) domain.BlockedTimeRepository {
	return &blockedTimeRepository{db: tx}
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

func (r *blockedTimeRepository) BulkInsertBlockedTime(ctx context.Context, bt []domain.BlockedTime) ([]int, error) {
	query := `
	insert into "BlockedTime" (merchant_id, employee_id, name, from_date, to_date, all_day, source)
	select $1, $2, unnest($3::text[]), unnest($4::timestamptz[]), unnest($5::timestamptz[]), unnest($6::boolean[]), $7
	returning id
	`

	merchantId := bt[0].MerchantId
	employeeId := bt[0].EmployeeId
	source := bt[0].Source

	names := make([]string, len(bt))
	fromDates := make([]time.Time, len(bt))
	toDates := make([]time.Time, len(bt))
	isAllDay := make([]bool, len(bt))

	for i, blockedTime := range bt {
		names[i] = blockedTime.Name
		fromDates[i] = blockedTime.FromDate
		toDates[i] = blockedTime.ToDate
		isAllDay[i] = blockedTime.AllDay
	}

	var btIds []int

	rows, _ := r.db.Query(ctx, query, merchantId, employeeId, names, fromDates, toDates, isAllDay, source)
	btIds, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return []int{}, err
	}

	return btIds, nil
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

func (r *blockedTimeRepository) BulkUpdateBlockedTime(ctx context.Context, bt []domain.BlockedTime) error {
	query := `
	update "BlockedTime" b
	set name = u.name, from_date = u.from_date, to_date = u.to_date, all_day = u.all_day
	from unnest($1::int[], $2::text[], $3::timestamptz[], $4::timestamptz[], $5::boolean[])
	as u(id, name, from_date, to_date, all_day)
	where b.id = u.id
	`

	ids := make([]int, len(bt))
	names := make([]string, len(bt))
	fromDates := make([]time.Time, len(bt))
	toDates := make([]time.Time, len(bt))
	isAllDay := make([]bool, len(bt))

	for i, blockedTime := range bt {
		ids[i] = blockedTime.Id
		names[i] = blockedTime.Name
		fromDates[i] = blockedTime.FromDate
		toDates[i] = blockedTime.ToDate
		isAllDay[i] = blockedTime.AllDay
	}

	_, err := r.db.Exec(ctx, query, ids, names, fromDates, toDates, isAllDay)
	if err != nil {
		return err
	}

	return nil
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

func (r *blockedTimeRepository) BulkDeleteBlockedTime(ctx context.Context, btIds []int) error {
	query := `
	delete from "BlockedTime"
	where id = any($1::int[])
	`

	_, err := r.db.Exec(ctx, query, btIds)
	if err != nil {
		return err
	}

	return nil
}

func (r *blockedTimeRepository) DeleteExternalCalendarBlockedTimes(ctx context.Context, extCalendarId int) error {
	query := `
	delete from "BlockedTime"
	where id in (
		select internal_id
		from "ExternalCalendarEvent"
		where external_calendar_id = $1 and internal_id is not null and internal_type = 'blocked_time'
	)
	`

	_, err := r.db.Exec(ctx, query, extCalendarId)
	if err != nil {
		return err
	}

	return nil
}

func (r *blockedTimeRepository) GetBlockedTime(ctx context.Context, blockedTimeId int) (domain.BlockedTime, error) {
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

func (r *blockedTimeRepository) GetBlockedTimesForCalendar(ctx context.Context, merchantId uuid.UUID, startTime, endTime string) ([]domain.BlockedTimeEvent, error) {
	query := `
	select b.id, b.employee_id, b.name, b.from_date, b.to_date, b.all_day, btt.icon, btt.id as blocked_type_id from "BlockedTime" b
	left join "BlockedTimeType" btt on btt.id = b.blocked_type_id
	where b.merchant_id = $1 and b.from_date <= $3 and b.to_date >= $2
	order by b.id
	`

	rows, _ := r.db.Query(ctx, query, merchantId, startTime, endTime)
	blockedTimes, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BlockedTimeEvent])
	if err != nil {
		return []domain.BlockedTimeEvent{}, err
	}

	return blockedTimes, nil
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
