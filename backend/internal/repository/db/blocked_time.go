package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *blockedTimeRepository) BulkInsertBlockedTime(ctx context.Context, bt []domain.BlockedTime) ([]int, error) {
	query := `
	insert into "BlockedTime" (merchant_id, blocked_type_id, name, from_date, to_date, all_day, source)
	select $1, unnest($2::int[]), unnest($3::text[]), unnest($4::timestamptz[]), unnest($5::timestamptz[]), unnest($6::boolean[]), $7
	returning id
	`

	btCount := len(bt)

	merchantId := bt[0].MerchantId
	source := bt[0].Source

	blockedTimeTypeIds := make([]pgtype.Int4, btCount)
	names := make([]string, btCount)
	fromDates := make([]time.Time, btCount)
	toDates := make([]time.Time, btCount)
	isAllDay := make([]bool, btCount)

	for i, blockedTime := range bt {
		if blockedTime.BlockedTypeId == nil {
			blockedTimeTypeIds[i] = pgtype.Int4{Valid: false}
		} else {
			blockedTimeTypeIds[i] = pgtype.Int4{Int32: int32(*blockedTime.BlockedTypeId), Valid: true}
		}
		names[i] = blockedTime.Name
		fromDates[i] = blockedTime.FromDate
		toDates[i] = blockedTime.ToDate
		isAllDay[i] = blockedTime.AllDay
	}

	var btIds []int

	rows, _ := r.db.Query(ctx, query, merchantId, blockedTimeTypeIds, names, fromDates, toDates, isAllDay, source)
	btIds, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return []int{}, err
	}

	return btIds, nil
}

func (r *blockedTimeRepository) BulkInsertEmployeeBlockedTime(ctx context.Context, blockedTimeIds []int, employeeIds []int) error {
	query := `
	insert into "EmployeeBlockedTime" (blocked_time_id, employee_id)
	select unnest($1::int[]), unnest($2::int[])
	`

	_, err := r.db.Exec(ctx, query, blockedTimeIds, employeeIds)
	if err != nil {
		return err
	}

	return nil
}

func (r *blockedTimeRepository) UpdateBlockedTime(ctx context.Context, bt domain.BlockedTime) error {
	query := `
	update "BlockedTime"
	set blocked_type_id = $3, name = $4, from_date = $5, to_date = $6, all_day = $7
	where merchant_id = $1 and ID = $2`

	_, err := r.db.Exec(ctx, query, bt.MerchantId, bt.Id, bt.BlockedTypeId, bt.Name, bt.FromDate, bt.ToDate, bt.AllDay)
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

func (r *blockedTimeRepository) BulkDeleteEmployeeBlockedTime(ctx context.Context, blockedTimeIds []int, employeeIds []int) error {
	query := `
	delete from "EmployeeBlockedTime"
	where blocked_time_id = any($1::int[]) and employee_id = any($2::int[])
	`

	_, err := r.db.Exec(ctx, query, blockedTimeIds, employeeIds)
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

func (r *blockedTimeRepository) GetBlockedTimeForEmployee(ctx context.Context, blockedTimeId int, employeeId int) (domain.BlockedTime, error) {
	query := `
	select bt.*
	from "BlockedTime" bt
	where id = $1 and (
		not exists (
			select 1
			from "EmployeeBlockedTime" ebt
			where ebt.blocked_time_id = bt.id
		)
		or exists (
			select 1
			from "EmployeeBlockedTime" ebt
			where ebt.blocked_time_id = bt.id and ebt.employee_id = $2
		)
	)
	`

	rows, _ := r.db.Query(ctx, query, blockedTimeId, employeeId)
	blockedTime, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BlockedTime])
	if err != nil {
		return domain.BlockedTime{}, err
	}

	return blockedTime, nil
}

func (r *blockedTimeRepository) GetBlockedTimeEmployees(ctx context.Context, blockedTimeId int) (domain.BlockedTimeEmployees, error) {
	query := `
	select bt.*,
		coalesce(
			array_agg(ebt.employee_id order by ebt.employee_id) filter (where ebt.employee_id is not null),
			'{}'::int[]
		) as employee_ids
	from "BlockedTime" bt
	left join "EmployeeBlockedTime" ebt on ebt.blocked_time_id = bt.id
	where bt.id = $1
	group by bt.id
	`

	rows, _ := r.db.Query(ctx, query, blockedTimeId)
	blockedTime, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BlockedTimeEmployees])
	if err != nil {
		return domain.BlockedTimeEmployees{}, err
	}

	return blockedTime, nil
}

func (r *blockedTimeRepository) GetBlockedTimesForCalendar(ctx context.Context, merchantId uuid.UUID, startTime, endTime string) ([]domain.BlockedTimeEvent, error) {
	query := `
	select bt.id, bt.name, bt.from_date, bt.to_date, bt.all_day, btt.icon, btt.id as blocked_type_id,
		coalesce(
			array_agg(ebt.employee_id order by ebt.employee_id) filter (where ebt.employee_id is not null),
			'{}'::int[]
		) as employee_ids
	from "BlockedTime" bt
	left join "EmployeeBlockedTime" ebt on ebt.blocked_time_id = bt.id
	left join "BlockedTimeType" btt on btt.id = bt.blocked_type_id
	where bt.merchant_id = $1 and bt.from_date <= $3 and bt.to_date >= $2
	group by bt.id, btt.id, btt.icon
	order by bt.id
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
