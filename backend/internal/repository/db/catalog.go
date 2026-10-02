package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type catalogRepository struct {
	db db.DBTX
}

func NewCatalogRepository(db db.DBTX) domain.CatalogRepository {
	return &catalogRepository{db: db}
}

func (r *catalogRepository) WithTx(tx db.DBTX) domain.CatalogRepository {
	return &catalogRepository{db: tx}
}

func (r *catalogRepository) NewService(ctx context.Context, serv domain.Service) (int, error) {
	query := `
	insert into "Service" (merchant_id, category_id, booking_type, name, description, color, total_duration, price_per_person,
		price_type, is_active, sequence, min_participants, max_participants, cancel_deadline, booking_window_min, booking_window_max, buffer_time, approval_policy)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, coalesce((
		select max(sequence) + 1 from "Service" where category_id is not distinct from $2 and merchant_id = $1
		), 1), $11, $12, $13, $14, $15, $16, $17)
	returning id
	`

	var serviceId int
	err := r.db.QueryRow(ctx, query, serv.MerchantId, serv.CategoryId, serv.BookingType, serv.Name, serv.Description, serv.Color,
		serv.TotalDuration, serv.Price, serv.PriceType, serv.IsActive, serv.MinParticipants, serv.MaxParticipants,
		serv.CancelDeadline, serv.BookingWindowMin, serv.BookingWindowMax, serv.BufferTime, serv.ApprovalPolicy).Scan(&serviceId)
	if err != nil {
		return 0, fmt.Errorf("NewService: %w", err)
	}

	return serviceId, nil
}

func (r *catalogRepository) UpdateService(ctx context.Context, s domain.Service) (*int, error) {
	query := `
	with old as (
		select id, category_id from "Service"
		where id = $1 and merchant_id = $2
	)
	update "Service"
	set category_id = $3, name = $4, description = $5, color = $6, total_duration = $7, price_per_person = $8,
		price_type = $9, is_active = $10, cancel_deadline = $11, booking_window_min = $12, booking_window_max = $13, buffer_time = $14,
		approval_policy = $15, min_participants = $16, max_participants = $17,
		sequence = case
			when old.category_id is distinct from $3 then (
				coalesce((
					select max(sequence) + 1 from "Service" where category_id is not distinct from $3 and merchant_id = $2
				), 1)
			)
			else sequence
		end
	from old
	where "Service".id = old.id
	returning old.category_id
	`

	var oldCategoryId *int
	err := r.db.QueryRow(ctx, query, s.Id, s.MerchantId, s.CategoryId, s.Name, s.Description, s.Color, s.TotalDuration,
		s.Price, s.PriceType, s.IsActive, s.CancelDeadline, s.BookingWindowMin, s.BookingWindowMax, s.BufferTime,
		s.ApprovalPolicy, s.MinParticipants, s.MaxParticipants).Scan(&oldCategoryId)
	if err != nil {
		return nil, fmt.Errorf("UpdateService: %w", err)
	}

	return oldCategoryId, nil
}

func (r *catalogRepository) DeleteService(ctx context.Context, merchantId uuid.UUID, serviceId int) error {
	query := `
	delete from "Service"
	where merchant_id = $1 and ID = $2
	`

	_, err := r.db.Exec(ctx, query, merchantId, serviceId)
	if err != nil {
		return fmt.Errorf("DeleteService: %w", err)
	}

	return nil
}

func (r *catalogRepository) DeactivateService(ctx context.Context, merchantId uuid.UUID, serviceId int) error {
	query := `
	update "Service"
	set is_active = false
	where id = $1 and merchant_id = $2
	`

	_, err := r.db.Exec(ctx, query, serviceId, merchantId)
	if err != nil {
		return fmt.Errorf("DeactivateService: %w", err)
	}

	return nil
}

func (r *catalogRepository) ActivateService(ctx context.Context, merchantId uuid.UUID, serviceId int) error {
	query := `
	update "Service"
	set is_active = true
	where id = $1 and merchant_id = $2
	`

	_, err := r.db.Exec(ctx, query, serviceId, merchantId)
	if err != nil {
		return fmt.Errorf("ActivateService: %w", err)
	}

	return nil
}

func (r *catalogRepository) ReorderServices(ctx context.Context, merchantId uuid.UUID, categoryId *int, serviceIds []int) error {
	query := `
	update "Service" s
	set sequence = x.seq
	from unnest($1::int[]) with ordinality as x(id, seq)
	where s.id = x.id and s.merchant_id = $2 and (
		($3::int is null and s.category_id is null) or s.category_id = $3
	)
	`

	_, err := r.db.Exec(ctx, query, serviceIds, merchantId, categoryId)
	if err != nil {
		return fmt.Errorf("ReorderServices: %w", err)
	}

	return nil
}

func (r *catalogRepository) ReorderServicesAfterUpdate(ctx context.Context, categoryId *int, merchantId uuid.UUID, exludeServiceId *int) error {
	query := `
	with reordered as (
		select id, row_number() over (order by sequence) as new_sequence
		from "Service"
		where category_id is not distinct from $1 and merchant_id = $2 and ($3::int is null or id != $3)
	)
	update "Service" s
	set sequence = r.new_sequence
	from reordered r
	where s.id = r.id
	`

	_, err := r.db.Exec(ctx, query, categoryId, merchantId, exludeServiceId)
	if err != nil {
		return fmt.Errorf("ReorderServicesAfterUpdate: %w", err)
	}

	return nil
}

// TODO: full outer joins can be expensive, this should be reevaluated later for performance
func (r *catalogRepository) GetServicesGroupedByCategory(ctx context.Context, merchantId uuid.UUID) ([]domain.ServicesGroupedByCategory, error) {
	query := `
	with services as (
		select s.id, s.merchant_id, s.category_id, s.booking_type, s.name, s.description, s.color, s.total_duration, s.price_per_person, s.price_type,
			s.is_active, s.sequence, s.min_participants, s.max_participants, s.cancel_deadline, s.booking_window_min, s.booking_window_max,
			s.buffer_time, s.approval_policy,
		coalesce (
			jsonb_agg(
				jsonb_build_object(
					'id', sp.id,
					'service_id', sp.service_id,
					'name', sp.name,
					'sequence', sp.sequence,
					'duration', sp.duration,
					'phase_type', sp.phase_type
				) order by sp.sequence
			) filter (where sp.id is not null),
		'[]'::jsonb) as phases
		from "Service" s
		left join "ServicePhase" sp on s.id = sp.service_id
		where s.merchant_id = $1
		group by s.id
	)
	select sc.id, sc.name, sc.sequence,
	coalesce (
		jsonb_agg(
			jsonb_build_object(
				'id', s.id,
				'merchant_id', s.merchant_id,
				'booking_type', s.booking_type,
				'name', s.name,
				'description', s.description,
				'color', s.color,
				'total_duration', s.total_duration,
				'price_per_person', s.price_per_person,
				'price_type', s.price_type,
				'is_active', s.is_active,
				'sequence', s.sequence,
				'min_participants', s.min_participants,
				'max_participants', s.max_participants,
				'cancel_deadline', s.cancel_deadline,
				'booking_window_min', s.booking_window_min,
				'booking_window_max', s.booking_window_max,
				'buffer_time', s.buffer_time,
				'approval_policy', s.approval_policy,
				'phases', s.phases
			) order by s.sequence
		) filter (where s.id is not null),
	'[]'::jsonb) as services
	from "ServiceCategory" sc
	full outer join services s on s.category_id = sc.id
	where sc.merchant_id = $1 or s.merchant_id = $1
	group by sc.id, sc.name
	order by sc.sequence, sc.name nulls last
	`

	rows, _ := r.db.Query(ctx, query, merchantId)
	servicesGroupByCategory, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ServicesGroupedByCategory, error) {
		var sgby domain.ServicesGroupedByCategory
		var services []byte

		err := row.Scan(&sgby.Id, &sgby.Name, &sgby.Sequence, &services)
		if err != nil {
			return domain.ServicesGroupedByCategory{}, err
		}

		if len(services) > 0 {
			err = json.Unmarshal(services, &sgby.Services)
			if err != nil {
				return domain.ServicesGroupedByCategory{}, err
			}
		} else {
			sgby.Services = []domain.Service{}
		}

		return sgby, nil
	})
	if err != nil {
		return []domain.ServicesGroupedByCategory{}, fmt.Errorf("GetServicesGroupedByCategory: %w", err)
	}

	// if services array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(servicesGroupByCategory) == 0 {
		servicesGroupByCategory = []domain.ServicesGroupedByCategory{}
	}

	return servicesGroupByCategory, nil
}

func (r *catalogRepository) GetServicesForCalendar(ctx context.Context, merchantId uuid.UUID) ([]domain.ServicesGroupedByCategoriesForCalendar, error) {
	query := `
	select sc.id, sc.name,
	coalesce (
		jsonb_agg(
			jsonb_build_object(
				'id', s.id,
				'name', s.name,
				'duration', s.total_duration,
				'price', s.price_per_person,
				'price_type', s.price_type,
				'color', s.color,
				'booking_type', s.booking_type,
				'max_participants', s.max_participants
			) order by s.sequence
		) filter (where s.id is not null),
	'[]'::jsonb) as services
	from "Service" s
	left join "ServiceCategory" sc on s.category_id = sc.id
	where s.merchant_id = $1 and s.is_active = true
	group by sc.id, sc.name
	order by sc.sequence, sc.name
	`

	rows, _ := r.db.Query(ctx, query, merchantId)
	servicesGroupByCategory, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ServicesGroupedByCategoriesForCalendar, error) {
		var sgby domain.ServicesGroupedByCategoriesForCalendar
		var services []byte

		err := row.Scan(&sgby.Id, &sgby.Name, &services)
		if err != nil {
			return domain.ServicesGroupedByCategoriesForCalendar{}, err
		}

		if len(services) > 0 {
			err = json.Unmarshal(services, &sgby.Services)
			if err != nil {
				return domain.ServicesGroupedByCategoriesForCalendar{}, err
			}
		} else {
			sgby.Services = []domain.CalendarService{}
		}

		return sgby, nil
	})
	if err != nil {
		return []domain.ServicesGroupedByCategoriesForCalendar{}, fmt.Errorf("GetServicesForCalendar: %w", err)
	}

	// if services array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(servicesGroupByCategory) == 0 {
		servicesGroupByCategory = []domain.ServicesGroupedByCategoriesForCalendar{}
	}

	return servicesGroupByCategory, nil
}

func (r *catalogRepository) GetServiceWithPhases(ctx context.Context, serviceID int, merchantId uuid.UUID) (domain.Service, error) {
	query := `
	select s.id, s.merchant_id, s.category_id, s.booking_type, s.name, s.description, s.color, s.total_duration, s.price_per_person, s.price_type,
		s.is_active, s.sequence, s.min_participants, s.max_participants, s.cancel_deadline, s.booking_window_min, s.booking_window_max,
		s.buffer_time, s.approval_policy,
	coalesce (
		jsonb_agg(
			jsonb_build_object(
				'id', sp.id,
				'service_id', sp.service_id,
				'name', sp.name,
				'sequence', sp.sequence,
				'duration', sp.duration,
				'phase_type', sp.phase_type
			) order by sp.sequence
		) filter (where sp.id is not null),
	'[]'::jsonb) as phases
	from "Service" s
	left join "ServicePhase" sp on s.id = sp.service_id
	where s.id = $1 and s.merchant_id = $2
	group by s.id
	`

	var s domain.Service
	var phasesJson []byte

	err := r.db.QueryRow(ctx, query, serviceID, merchantId).Scan(&s.Id, &s.MerchantId, &s.CategoryId, &s.BookingType, &s.Name, &s.Description, &s.Color, &s.TotalDuration,
		&s.Price, &s.PriceType, &s.IsActive, &s.Sequence, &s.MinParticipants, &s.MaxParticipants, &s.CancelDeadline, &s.BookingWindowMin,
		&s.BookingWindowMax, &s.BufferTime, &s.ApprovalPolicy, &phasesJson)
	if err != nil {
		return domain.Service{}, fmt.Errorf("GetServiceWithPhases: %w", err)
	}

	if len(phasesJson) > 0 {
		err = json.Unmarshal(phasesJson, &s.Phases)
		if err != nil {
			return domain.Service{}, fmt.Errorf("GetServiceWithPhases: %w", err)
		}
	} else {
		s.Phases = []domain.ServicePhase{}
	}

	return s, nil
}

func (r *catalogRepository) GetServicesForMerchantPage(ctx context.Context, merchantId uuid.UUID, employeeId *int) ([]domain.MerchantPageServicesGroupedByCategory, error) {
	query := `
	with emp_overrides as (
		select
			es.service_id,
			min(coalesce(es.total_duration, s.total_duration)) as min_duration,
			max(coalesce(es.total_duration, s.total_duration)) as max_duration,
			max(coalesce(es.max_participants, s.max_participants)) as max_participants,
			coalesce(
				jsonb_agg(
					jsonb_build_object(
						'price_per_person', es.price_per_person,
						'price_type', es.price_type
					)
				) filter (where es.employee_id is not null),
				'[]'::jsonb
			) as price_overrides
		from "EmployeeService" es
		join "Employee" e on e.id = es.employee_id and e.merchant_id = $1 and e.is_active = true
		join "Service" s on s.id = es.service_id
		where ($2::int is null or es.employee_id = $2)
		group by es.service_id
	)
	select sc.id, sc.name, sc.sequence,
	coalesce (
		jsonb_agg(
			jsonb_build_object(
				'id', s.id,
				'category_id', s.category_id,
				'name', s.name,
				'description', s.description,
				'total_duration', coalesce(eo.min_duration, s.total_duration),
				'min_duration', coalesce(eo.min_duration, s.total_duration),
				'max_duration', coalesce(eo.max_duration, s.total_duration),
				'price', s.price_per_person,
				'price_type', s.price_type,
				'max_participants', coalesce(eo.max_participants, s.max_participants),
				'booking_type', s.booking_type,
				'sequence', s.sequence,
				'price_overrides', coalesce(eo.price_overrides, '[]'::jsonb)
			) order by s.sequence
		) filter (where s.id is not null),
	'[]'::jsonb) as services
	from "Service" s
	left join "ServiceCategory" sc on s.category_id = sc.id
	left join emp_overrides eo on eo.service_id = s.id
	where s.merchant_id = $1 and s.is_active = true
		and ($2::int is null or eo.service_id is not null)
	group by sc.id, sc.name
	order by sc.sequence, sc.name
	`

	rows, _ := r.db.Query(ctx, query, merchantId, employeeId)
	servicesGroupByCategory, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.MerchantPageServicesGroupedByCategory, error) {
		var sgby domain.MerchantPageServicesGroupedByCategory
		var servicesJson []byte

		err := row.Scan(&sgby.Id, &sgby.Name, &sgby.Sequence, &servicesJson)
		if err != nil {
			return domain.MerchantPageServicesGroupedByCategory{}, err
		}

		if len(servicesJson) > 0 {
			err = json.Unmarshal(servicesJson, &sgby.Services)
			if err != nil {
				return domain.MerchantPageServicesGroupedByCategory{}, err
			}
		} else {
			sgby.Services = []domain.MerchantPageService{}
		}

		return sgby, nil
	})
	if err != nil {
		return []domain.MerchantPageServicesGroupedByCategory{}, fmt.Errorf("GetServicesForMerchantPage: %w", err)
	}

	// if services array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(servicesGroupByCategory) == 0 {
		servicesGroupByCategory = []domain.MerchantPageServicesGroupedByCategory{}
	}

	return servicesGroupByCategory, nil
}

func (r *catalogRepository) GetServiceDetailsForMerchantPage(ctx context.Context, merchantId uuid.UUID, serviceId int, locationId int) (domain.PublicServiceDetails, error) {
	query := `
	select s.id, s.name, s.description, s.total_duration, s.price_per_person as price, s.price_type, l.formatted_location, l.geo_point,
	coalesce(
		jsonb_agg(
			jsonb_build_object(
				'id', sp.id,
				'service_id', sp.service_id,
                'name', sp.name,
                'sequence', sp.sequence,
                'duration', sp.duration,
                'phase_type', sp.phase_type
			) order by sp.sequence
		) filter (where sp.id is not null),
		'[]'::jsonb
	) as phases
	from "Service" s
	left join "ServicePhase" sp on s.id = sp.service_id
	left join "Location" l on l.merchant_id = $2 and l.id = $3
	where s.id = $1 and s.merchant_id = $2
	group by s.id, l.formatted_location, l.geo_point`

	var data domain.PublicServiceDetails
	var phaseJson []byte

	err := r.db.QueryRow(ctx, query, serviceId, merchantId, locationId).Scan(&data.Id, &data.Name, &data.Description, &data.TotalDuration,
		&data.Price, &data.PriceType, &data.FormattedLocation, &data.GeoPoint, &phaseJson)
	if err != nil {
		return domain.PublicServiceDetails{}, fmt.Errorf("GetServiceDetailsForMerchantPage: %w", err)
	}

	if len(phaseJson) > 0 {
		err = json.Unmarshal(phaseJson, &data.Phases)
		if err != nil {
			return domain.PublicServiceDetails{}, fmt.Errorf("GetServiceDetailsForMerchantPage: %w", err)
		}
	} else {
		data.Phases = []domain.ServicePhase{}
	}

	return data, nil
}

func (r *catalogRepository) GetAllServicePageData(ctx context.Context, serviceId int, merchantId uuid.UUID) (domain.ServicePageData, error) {
	query := `
	with phases as (
		select sp.service_id,
			jsonb_agg(
				jsonb_build_object(
					'id', sp.id,
					'service_id', sp.service_id,
					'name', sp.name,
					'sequence', sp.sequence,
					'duration', sp.duration,
					'phase_type', sp.phase_type
				) order by sp.sequence
			) as phases
		from "ServicePhase" sp
		group by sp.service_id
	),
	products as (
		select sprod.service_id,
			jsonb_agg(
				jsonb_build_object(
					'id', p.id,
					'name', p.name,
					'unit', p.unit,
					'amount_used', sprod.amount_used
				)
			) as products
		from "ServiceProduct" sprod
		join "Product" p on sprod.product_id = p.id
		group by sprod.service_id
	)
	select s.id, s.name, s.booking_type, s.category_id, s.description, s.color, s.total_duration, s.price_per_person, s.price_type, s.is_active, s.sequence,
		min_participants, max_participants,
		jsonb_build_object(
		 	'cancel_deadline', s.cancel_deadline,
         	'booking_window_min', s.booking_window_min,
         	'booking_window_max', s.booking_window_max,
         	'buffer_time', s.buffer_time,
			'approval_policy', s.approval_policy
		) as settings,
		coalesce(phases.phases, '[]'::jsonb) as phases,
		coalesce((
			select array_agg(es.employee_id order by es.employee_id)
			from "EmployeeService" es
			join "Employee" e on e.id = es.employee_id and e.is_active
			where es.service_id = s.id
		), '{}'::int[]) as employee_ids,
		coalesce(products.products, '[]'::jsonb) as products
	from "Service" s
	left join phases on s.id = phases.service_id
	left join products on s.id = products.service_id
	where s.id = $1 and s.merchant_id = $2
	`

	var spd domain.ServicePageData
	var settingsJson []byte
	var phaseJson []byte
	var productJson []byte

	err := r.db.QueryRow(ctx, query, serviceId, merchantId).Scan(&spd.Id, &spd.Name, &spd.BookingType, &spd.CategoryId, &spd.Description,
		&spd.Color, &spd.TotalDuration, &spd.Price, &spd.PriceType, &spd.IsActive, &spd.Sequence, &spd.MinParicipants, &spd.MaxParticipants,
		&settingsJson, &phaseJson, &spd.EmployeeIds, &productJson)
	if err != nil {
		return domain.ServicePageData{}, fmt.Errorf("GetAllServicePageData: %w", err)
	}

	if len(settingsJson) > 0 {
		if err := json.Unmarshal(settingsJson, &spd.Settings); err != nil {
			return domain.ServicePageData{}, fmt.Errorf("GetAllServicePageData: %w", err)
		}
	}

	if len(phaseJson) > 0 {
		err = json.Unmarshal(phaseJson, &spd.Phases)
		if err != nil {
			return domain.ServicePageData{}, fmt.Errorf("GetAllServicePageData: %w", err)
		}
	} else {
		spd.Phases = []domain.ServicePhase{}
	}

	if len(productJson) > 0 {
		err = json.Unmarshal(productJson, &spd.Products)
		if err != nil {
			return domain.ServicePageData{}, fmt.Errorf("GetAllServicePageData: %w", err)
		}
	} else {
		spd.Products = []domain.MinimalProductInfoWithUsage{}
	}

	return spd, nil
}

func (r *catalogRepository) GetServicePageFormOptions(ctx context.Context, merchantId uuid.UUID) (domain.ServicePageFormOptions, error) {
	query := `
	with product as (
		select id, name, unit from "Product" where merchant_id = $1
	),
	category as (
		select id, name from "ServiceCategory" where merchant_id = $1
	)
	select
		coalesce((select jsonb_agg(p) from product p), '[]'::jsonb) as products,
		coalesce((select jsonb_agg(c) from category c), '[]'::jsonb) as categories
	`

	var spfo domain.ServicePageFormOptions
	var products []byte
	var categories []byte

	err := r.db.QueryRow(ctx, query, merchantId).Scan(&products, &categories)
	if err != nil {
		return domain.ServicePageFormOptions{}, fmt.Errorf("GetServicePageFormOptions: %w", err)
	}

	if len(products) > 0 {
		err = json.Unmarshal(products, &spfo.Products)
		if err != nil {
			return domain.ServicePageFormOptions{}, fmt.Errorf("GetServicePageFormOptions: %w", err)
		}
	} else {
		spfo.Products = []domain.MinimalProductInfo{}
	}

	if len(categories) > 0 {
		err = json.Unmarshal(categories, &spfo.Categories)
		if err != nil {
			return domain.ServicePageFormOptions{}, fmt.Errorf("GetServicePageFormOptions: %w", err)
		}
	} else {
		spfo.Categories = []domain.ServiceCategory{}
	}

	return spfo, nil
}

func (r *catalogRepository) GetMinimalServiceInfo(ctx context.Context, merchantId uuid.UUID, serviceId, locationId int, employeeId *int) (domain.MinimalServiceInfo, error) {
	query := `
	with emp_overrides as (
		select
			min(coalesce(es.total_duration, s.total_duration)) as min_duration,
			max(coalesce(es.total_duration, s.total_duration)) as max_duration,
			coalesce(
				jsonb_agg(
					jsonb_build_object(
						'price_per_person', es.price_per_person,
						'price_type', es.price_type
					)
				) filter (where es.employee_id is not null),
				'[]'::jsonb
			) as price_overrides
		from "EmployeeService" es
		join "Employee" e on e.id = es.employee_id and e.merchant_id = $1 and e.is_active = true
		join "Service" s on s.id = es.service_id
		where es.service_id = $2 and ($4::int is null or es.employee_id = $4)
	)
	select
        s.name,
        s.total_duration,
        coalesce(eo.min_duration, s.total_duration) as min_duration,
        coalesce(eo.max_duration, s.total_duration) as max_duration,
        s.price_per_person,
        s.price_type,
        l.formatted_location,
        coalesce(eo.price_overrides, '[]'::jsonb) as price_overrides
    from "Service" s
    inner join "Location" l on l.merchant_id = $1 and l.id = $3
    left join emp_overrides eo on true
    where s.merchant_id = $1 and s.id = $2 and s.is_active = true and ($4::int is null or eo.min_duration is not null)
	`

	var msi domain.MinimalServiceInfo
	var priceOverridesJson []byte

	err := r.db.QueryRow(ctx, query, merchantId, serviceId, locationId, employeeId).Scan(
		&msi.Name, &msi.TotalDuration, &msi.MinDuration, &msi.MaxDuration, &msi.Price, &msi.PriceType, &msi.FormattedLocation, &priceOverridesJson,
	)
	if err != nil {
		return domain.MinimalServiceInfo{}, fmt.Errorf("GetMinimalServiceInfo: %w", err)
	}

	if len(priceOverridesJson) > 0 {
		if err := json.Unmarshal(priceOverridesJson, &msi.PriceOverrides); err != nil {
			return domain.MinimalServiceInfo{}, fmt.Errorf("GetMinimalServiceInfo unmarshal price overrides: %w", err)
		}
	}

	return msi, nil
}

func (r *catalogRepository) GetServiceWithPhasesForEmployees(ctx context.Context, serviceId int, employeeIds []int) (map[int]domain.Service, error) {
	query := `
	select es.employee_id,
		s.id, s.merchant_id, s.category_id, s.booking_type, s.name, s.description, s.color,
		coalesce(es.total_duration, s.total_duration) as total_duration,
		coalesce(es.price_per_person, s.price_per_person) as price_per_person,
		coalesce(es.price_type, s.price_type) as price_type,
		s.is_active, s.sequence,
		coalesce(es.min_participants, s.min_participants) as min_participants,
		coalesce(es.max_participants, s.max_participants) as max_participants,
		s.cancel_deadline, s.booking_window_min, s.booking_window_max,
		coalesce(es.buffer_time, s.buffer_time) as buffer_time, s.approval_policy,
		coalesce(
			jsonb_agg(
				jsonb_build_object(
					'id', sp.id,
					'service_id', sp.service_id,
					'name', sp.name,
					'sequence', sp.sequence,
					'duration', coalesce(esp.duration, sp.duration),
					'phase_type', sp.phase_type
				) order by sp.sequence
			) filter (where sp.id is not null),
		'[]'::jsonb) as phases
	from "Service" s
	join "EmployeeService" es on es.service_id = s.id and es.employee_id = any($2::int[])
	left join "ServicePhase" sp on s.id = sp.service_id
	left join "EmployeeServicePhase" esp on esp.service_phase_id = sp.id and esp.employee_id = es.employee_id
	where s.id = $1
	group by s.id, es.employee_id, es.service_id`

	result := make(map[int]domain.Service, len(employeeIds))

	var empId int
	var svc domain.Service
	var phasesJSON []byte

	rows, _ := r.db.Query(ctx, query, serviceId, employeeIds)
	_, err := pgx.ForEachRow(rows, []any{
		&empId,
		&svc.Id, &svc.MerchantId, &svc.CategoryId, &svc.BookingType, &svc.Name, &svc.Description, &svc.Color,
		&svc.TotalDuration, &svc.Price, &svc.PriceType, &svc.IsActive, &svc.Sequence,
		&svc.MinParticipants, &svc.MaxParticipants,
		&svc.CancelDeadline, &svc.BookingWindowMin, &svc.BookingWindowMax,
		&svc.BufferTime, &svc.ApprovalPolicy,
		&phasesJSON,
	}, func() error {
		if len(phasesJSON) > 0 {
			if err := json.Unmarshal(phasesJSON, &svc.Phases); err != nil {
				return fmt.Errorf("unmarshal phases: %w", err)
			}
		} else {
			svc.Phases = []domain.ServicePhase{}
		}

		result[empId] = svc
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("GetServiceWithPhasesForEmployees: %w", err)
	}
	return result, nil
}

func (r *catalogRepository) GetEmployeeIdsForService(ctx context.Context, serviceId int) ([]int, error) {
	query := `
	select coalesce(
		array_agg(es.employee_id order by es.employee_id),
		'{}'::int[]
		)
	from "EmployeeService" es
	join "Employee" e on e.id = es.employee_id and e.is_active is true
	where es.service_id = $1
	`

	var employeeIds []int
	err := r.db.QueryRow(ctx, query, serviceId).Scan(&employeeIds)
	if err != nil {
		return []int{}, fmt.Errorf("GetEmployeeIdsForService: %w", err)
	}

	return employeeIds, nil
}

func (r *catalogRepository) GetEmployeeServiceSettings(ctx context.Context, merchantId uuid.UUID, serviceId int) ([]domain.EmployeeServiceSettings, error) {
	query := `
	select e.id,
		coalesce(e.first_name, u.first_name) as first_name,
		coalesce(e.last_name, u.last_name) as last_name,
		e.role,
		(es.employee_id is not null) as is_assigned,
		es.total_duration,
		es.price_per_person,
		es.price_type,
		es.min_participants,
		es.max_participants,
		es.buffer_time
	from "Employee" e
	left join "EmployeeService" es on es.employee_id = e.id and es.service_id = $1
	left join "User" u on u.id = e.user_id
	where e.merchant_id = $2 and e.is_active is true
	order by coalesce(e.first_name, u.first_name), coalesce(e.last_name, u.last_name), e.id
	`

	rows, _ := r.db.Query(ctx, query, serviceId, merchantId)
	settings, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.EmployeeServiceSettings])
	if err != nil {
		return nil, fmt.Errorf("GetEmployeeServiceSettings: %w", err)
	}

	return settings, nil
}

func (r *catalogRepository) GetEmployeeServicePhaseOverrides(ctx context.Context, merchantId uuid.UUID, serviceId int) ([]domain.EmployeeServicePhase, error) {
	query := `
	select esp.employee_id, esp.service_id, esp.service_phase_id, esp.duration
	from "EmployeeServicePhase" esp
	join "Employee" e on e.id = esp.employee_id
	where esp.service_id = $1 and e.merchant_id = $2 and e.is_active is true
		and esp.duration is not null
	order by esp.employee_id, esp.service_phase_id
	`
	rows, _ := r.db.Query(ctx, query, serviceId, merchantId)
	phaseOverrides, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.EmployeeServicePhase])
	if err != nil {
		return nil, fmt.Errorf("GetEmployeeServicePhaseOverrides: %w", err)
	}

	return phaseOverrides, nil
}

func (r *catalogRepository) NewServicePhases(ctx context.Context, serviceId int, phases []domain.ServicePhase) error {
	query := `
	insert into "ServicePhase" (service_id, name, sequence, duration, phase_type)
	select $1, unnest($2::text[]), unnest($3::int[]), unnest($4::int[]), unnest($5::text[])
	`

	names := make([]string, len(phases))
	sequences := make([]int, len(phases))
	durations := make([]int, len(phases))
	phaseTypes := make([]string, len(phases))

	for i, p := range phases {
		names[i] = p.Name
		sequences[i] = p.Sequence
		durations[i] = p.Duration
		phaseTypes[i] = p.PhaseType.String()
	}

	_, err := r.db.Exec(ctx, query, serviceId, names, sequences, durations, phaseTypes)
	if err != nil {
		return fmt.Errorf("NewServicePhases: %w", err)
	}

	return nil
}

func (r *catalogRepository) UpdateServicePhases(ctx context.Context, phases []domain.ServicePhase) error {
	query := `
	update "ServicePhase" sp
	set name = u.name, sequence = u.sequence, duration = u.duration, phase_type = u.phase_type
	from unnest($1::int[], $2::text[], $3::int[], $4::int[], $5::text[]) as u(id, name, sequence, duration, phase_type)
	where sp.id = u.id
	`

	ids := make([]int, len(phases))
	names := make([]string, len(phases))
	sequences := make([]int, len(phases))
	durations := make([]int, len(phases))
	phaseTypes := make([]string, len(phases))

	for i, p := range phases {
		ids[i] = p.Id
		names[i] = p.Name
		sequences[i] = p.Sequence
		durations[i] = p.Duration
		phaseTypes[i] = p.PhaseType.String()
	}

	_, err := r.db.Exec(ctx, query, ids, names, sequences, durations, phaseTypes)
	if err != nil {
		return fmt.Errorf("UpdateServicePhases: %w", err)
	}

	return nil
}

func (r *catalogRepository) UpdateServicePhaseDuration(ctx context.Context, serviceId int, duration int) error {
	query := `
	update "ServicePhase"
	set duration = $2
	where service_id = $1
	`

	_, err := r.db.Exec(ctx, query, serviceId, duration)
	if err != nil {
		return fmt.Errorf("UpdateServicePhaseDuration: %w", err)
	}

	return nil
}

func (r *catalogRepository) DeleteServicePhases(ctx context.Context, phaseIds []int) error {
	query := `
	delete from "ServicePhase"
	where id = any($1::int[])
	`

	_, err := r.db.Exec(ctx, query, phaseIds)
	if err != nil {
		return fmt.Errorf("DeleteServicePhases: %w", err)
	}

	return nil
}

func (r *catalogRepository) DeleteServicePhasesForService(ctx context.Context, serviceId int) error {
	query := `
	delete from "ServicePhase"
	where service_id = $1
	`

	_, err := r.db.Exec(ctx, query, serviceId)
	if err != nil {
		return fmt.Errorf("DeleteServicePhasesForService: %w", err)
	}

	return nil
}

func (r *catalogRepository) GetServicePhases(ctx context.Context, serviceId int) ([]domain.ServicePhase, error) {
	query := `
	select *
	from "ServicePhase"
	where service_id = $1
	order by sequence
	`

	rows, _ := r.db.Query(ctx, query, serviceId)
	phases, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.ServicePhase])
	if err != nil {
		return []domain.ServicePhase{}, fmt.Errorf("GetServicePhases: %w", err)
	}

	return phases, nil
}

func (r *catalogRepository) NewServiceCategory(ctx context.Context, merchantId uuid.UUID, sc domain.ServiceCategory) error {
	query := `
	insert into "ServiceCategory" (merchant_id, name, sequence)
	values ($1, $2, coalesce(
		(select max(sequence) + 1 from "ServiceCategory" where merchant_id = $1), 1)
	)
	`

	_, err := r.db.Exec(ctx, query, merchantId, sc.Name)
	if err != nil {
		return fmt.Errorf("NewServiceCategory: %w", err)
	}

	return nil
}

func (r *catalogRepository) UpdateServiceCategory(ctx context.Context, merchantId uuid.UUID, sc domain.ServiceCategory) error {
	query := `
	update "ServiceCategory"
	set name = $3
	where id = $1 and merchant_id = $2
	`

	_, err := r.db.Exec(ctx, query, sc.Id, merchantId, sc.Name)
	if err != nil {
		return fmt.Errorf("UpdateServiceCategory: %w", err)
	}

	return nil
}

func (r *catalogRepository) DeleteServiceCategory(ctx context.Context, merchantId uuid.UUID, categoryId int) error {
	query := `
	delete from "ServiceCategory"
	where id = $1 and merchant_id = $2
	`

	_, err := r.db.Exec(ctx, query, categoryId, merchantId)
	if err != nil {
		return fmt.Errorf("DeleteServiceCategory: %w", err)
	}

	return nil
}

func (r *catalogRepository) ReorderServiceCategories(ctx context.Context, merchantId uuid.UUID, categoryIds []int) error {
	query := `
	update "ServiceCategory" sc
	set sequence = x.seq
	from unnest($1::int[]) with ordinality as x(id, seq)
	where sc.id = x.id and sc.merchant_id = $2
	`

	_, err := r.db.Exec(ctx, query, categoryIds, merchantId)
	if err != nil {
		return fmt.Errorf("ReorderServiceCategories: %w", err)
	}

	return nil
}

func (r *catalogRepository) NewServiceProduct(ctx context.Context, merchantId uuid.UUID, connectedProducts []domain.ConnectedProducts) error {
	query := `
	insert into "ServiceProduct" (service_id, product_id, amount_used)
	select $1, p.id, u.amount_used
	from unnest($2::int[], $3::int[]) as u(product_id, amount_used)
	join "Product" p on p.id = u.product_id
	where p.merchant_id = $4
	`

	productIds := make([]int, len(connectedProducts))
	amountUseds := make([]int, len(connectedProducts))

	for i, cp := range connectedProducts {
		productIds[i] = cp.ProductId
		amountUseds[i] = cp.AmountUsed
	}

	_, err := r.db.Exec(ctx, query, connectedProducts[0].ServiceId, productIds, amountUseds, merchantId)
	if err != nil {
		return fmt.Errorf("NewServiceProduct: %w", err)
	}

	return nil
}

func (r *catalogRepository) UpdateServiceProducts(ctx context.Context, serviceId int, connectedProducts []domain.ConnectedProducts) error {
	query := `
	insert into "ServiceProduct" (service_id, product_id, amount_used)
	select $1, u.product_id, u.amount_used
	from unnest($2::int[], $3::int[]) as u(product_id, amount_used)
	on conflict (service_id, product_id) do update
	set amount_used = excluded.amount_used
	`

	productIds := make([]int, len(connectedProducts))
	amountUseds := make([]int, len(connectedProducts))

	for i, cp := range connectedProducts {
		productIds[i] = cp.ProductId
		amountUseds[i] = cp.AmountUsed
	}

	_, err := r.db.Exec(ctx, query, serviceId, productIds, amountUseds)
	if err != nil {
		return fmt.Errorf("UpdateServiceProducts: %w", err)
	}

	return nil
}

func (r *catalogRepository) DeleteServiceProducts(ctx context.Context, serviceId int, productIds []int) error {
	query := `
	delete from "ServiceProduct"
	where service_id = $1 and product_id = any($2::int[])
	`

	_, err := r.db.Exec(ctx, query, serviceId, productIds)
	if err != nil {
		return fmt.Errorf("DeleteServiceProducts: %w", err)
	}

	return nil
}

func (r *catalogRepository) GetServiceProducts(ctx context.Context, serviceId int) ([]domain.ConnectedProducts, error) {
	query := `
	select product_id, service_id, amount_used from "ServiceProduct"
	where service_id = $1
	`

	rows, _ := r.db.Query(ctx, query, serviceId)
	connectedProducts, err := pgx.CollectRows(rows, pgx.RowTo[domain.ConnectedProducts])
	if err != nil {
		return []domain.ConnectedProducts{}, fmt.Errorf("GetServiceProducts: %w", err)
	}

	return connectedProducts, nil
}

func (r *catalogRepository) BulkInsertEmployeeService(ctx context.Context, employeeServices []domain.EmployeeService) error {
	query := `
	insert into "EmployeeService" (employee_id, service_id, total_duration, price_per_person, price_type, min_participants, max_participants, buffer_time)
	select unnest($1::int[]), unnest($2::int[]), unnest($3::int[]), unnest($4::price[]), unnest($5::text[]), unnest($6::int[]), unnest($7::int[]), unnest($8::int[])
	`

	employeeServiceCount := len(employeeServices)

	employeeIds := make([]int, employeeServiceCount)
	serviceIds := make([]int, employeeServiceCount)
	totalDurations := make([]pgtype.Int4, employeeServiceCount)
	pricePerPersons := make([]*currencyx.Price, employeeServiceCount)
	priceTypes := make([]pgtype.Text, employeeServiceCount)
	minParticipants := make([]pgtype.Int4, employeeServiceCount)
	maxParticipants := make([]pgtype.Int4, employeeServiceCount)
	bufferTimes := make([]pgtype.Int4, employeeServiceCount)

	for i, e := range employeeServices {
		employeeIds[i] = e.EmployeeId
		serviceIds[i] = e.ServiceId
		if e.TotalDuration == nil {
			totalDurations[i] = pgtype.Int4{Valid: false}
		} else {
			totalDurations[i] = pgtype.Int4{Int32: int32(*e.TotalDuration), Valid: true}
		}
		pricePerPersons[i] = e.PricePerPerson
		if e.PriceType == nil {
			priceTypes[i] = pgtype.Text{Valid: false}
		} else {
			priceTypes[i] = pgtype.Text{String: e.PriceType.String(), Valid: true}
		}
		if e.MinParticipants == nil {
			minParticipants[i] = pgtype.Int4{Valid: false}
		} else {
			minParticipants[i] = pgtype.Int4{Int32: int32(*e.MinParticipants), Valid: true}
		}
		if e.MaxParticipants == nil {
			maxParticipants[i] = pgtype.Int4{Valid: false}
		} else {
			maxParticipants[i] = pgtype.Int4{Int32: int32(*e.MaxParticipants), Valid: true}
		}
		if e.BufferTime == nil {
			bufferTimes[i] = pgtype.Int4{Valid: false}
		} else {
			bufferTimes[i] = pgtype.Int4{Int32: int32(*e.BufferTime), Valid: true}
		}
	}

	_, err := r.db.Exec(ctx, query, employeeIds, serviceIds, totalDurations, pricePerPersons, priceTypes, minParticipants, maxParticipants, bufferTimes)
	if err != nil {
		return fmt.Errorf("BulkInsertEmployeeService: %w", err)
	}

	return nil
}

func (r *catalogRepository) BulkDeleteEmployeeService(ctx context.Context, serviceId int, employeeIds []int) error {
	query := `
	delete from "EmployeeService"
	where service_id = $1 and employee_id = any($2::int[])
	`

	_, err := r.db.Exec(ctx, query, serviceId, employeeIds)
	if err != nil {
		return fmt.Errorf("BulkDeleteEmployeeService: %w", err)
	}

	return nil
}

func (r *catalogRepository) BulkUpsertEmployeeServiceSettings(ctx context.Context, employeeServices []domain.EmployeeService) error {
	query := `
	insert into "EmployeeService" (employee_id, service_id, total_duration, price_per_person, price_type, min_participants, max_participants, buffer_time)
	select unnest($1::int[]), unnest($2::int[]), unnest($3::int[]), unnest($4::price[]),
		unnest($5::text[]), unnest($6::int[]), unnest($7::int[]), unnest($8::int[])
	on conflict (employee_id, service_id) do update
	set total_duration = excluded.total_duration,
		price_per_person = excluded.price_per_person,
		price_type = excluded.price_type,
		min_participants = excluded.min_participants,
		max_participants = excluded.max_participants,
		buffer_time = excluded.buffer_time
	`

	employeeIds := make([]int, len(employeeServices))
	serviceIds := make([]int, len(employeeServices))
	totalDurations := make([]pgtype.Int4, len(employeeServices))
	prices := make([]*currencyx.Price, len(employeeServices))
	priceTypes := make([]pgtype.Text, len(employeeServices))
	minParticipants := make([]pgtype.Int4, len(employeeServices))
	maxParticipants := make([]pgtype.Int4, len(employeeServices))
	bufferTimes := make([]pgtype.Int4, len(employeeServices))

	for i, employeeService := range employeeServices {
		employeeIds[i] = employeeService.EmployeeId
		serviceIds[i] = employeeService.ServiceId
		if employeeService.TotalDuration != nil {
			totalDurations[i] = pgtype.Int4{Int32: int32(*employeeService.TotalDuration), Valid: true}
		}
		prices[i] = employeeService.PricePerPerson
		if employeeService.PriceType != nil {
			priceTypes[i] = pgtype.Text{String: employeeService.PriceType.String(), Valid: true}
		}
		if employeeService.MinParticipants != nil {
			minParticipants[i] = pgtype.Int4{Int32: int32(*employeeService.MinParticipants), Valid: true}
		}
		if employeeService.MaxParticipants != nil {
			maxParticipants[i] = pgtype.Int4{Int32: int32(*employeeService.MaxParticipants), Valid: true}
		}
		if employeeService.BufferTime != nil {
			bufferTimes[i] = pgtype.Int4{Int32: int32(*employeeService.BufferTime), Valid: true}
		}
	}

	_, err := r.db.Exec(ctx, query, employeeIds, serviceIds, totalDurations, prices, priceTypes, minParticipants, maxParticipants, bufferTimes)
	if err != nil {
		return fmt.Errorf("BulkUpsertEmployeeServiceSettings: %w", err)
	}

	return nil
}

func (r *catalogRepository) BulkUpdateEmployeeServiceDurations(ctx context.Context, employeeServices []domain.EmployeeService) error {
	query := `
	update "EmployeeService" es
	set total_duration = updates.total_duration
	from unnest($1::int[], $2::int[], $3::int[])
		as updates(employee_id, service_id, total_duration)
	where es.employee_id = updates.employee_id and es.service_id = updates.service_id
	`

	employeeIds := make([]int, len(employeeServices))
	serviceIds := make([]int, len(employeeServices))
	totalDurations := make([]pgtype.Int4, len(employeeServices))
	for i, employeeService := range employeeServices {
		employeeIds[i] = employeeService.EmployeeId
		serviceIds[i] = employeeService.ServiceId
		if employeeService.TotalDuration != nil {
			totalDurations[i] = pgtype.Int4{Int32: int32(*employeeService.TotalDuration), Valid: true}
		}
	}

	_, err := r.db.Exec(ctx, query, employeeIds, serviceIds, totalDurations)
	if err != nil {
		return fmt.Errorf("BulkUpdateEmployeeServiceDurations: %w", err)
	}

	return nil
}

func (r *catalogRepository) BulkDeleteEmployeeServicePhases(ctx context.Context, serviceId int, employeeIds []int) error {
	query := `
	delete from "EmployeeServicePhase"
	where service_id = $1 and employee_id = any($2::int[])
	`

	_, err := r.db.Exec(ctx, query, serviceId, employeeIds)
	if err != nil {
		return fmt.Errorf("BulkDeleteEmployeeServicePhases: %w", err)
	}

	return nil
}

func (r *catalogRepository) BulkInsertEmployeeServicePhases(ctx context.Context, phases []domain.EmployeeServicePhase) error {
	query := `
	insert into "EmployeeServicePhase" (employee_id, service_id, service_phase_id, duration)
	select unnest($1::int[]), unnest($2::int[]), unnest($3::int[]), unnest($4::int[])
	`

	employeeIds := make([]int, len(phases))
	serviceIds := make([]int, len(phases))
	servicePhaseIds := make([]int, len(phases))
	durations := make([]int, len(phases))
	for i, phase := range phases {
		employeeIds[i] = phase.EmployeeId
		serviceIds[i] = phase.ServiceId
		servicePhaseIds[i] = phase.ServicePhaseId
		durations[i] = phase.Duration
	}

	_, err := r.db.Exec(ctx, query, employeeIds, serviceIds, servicePhaseIds, durations)
	if err != nil {
		return fmt.Errorf("BulkInsertEmployeeServicePhases: %w", err)
	}

	return nil
}
