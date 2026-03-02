package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
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
	insert into "Service" (merchant_id, category_id, booking_type, name, description, color, total_duration, price_per_person, cost_per_person,
		price_type, is_active, sequence, min_participants, max_participants, cancel_deadline, booking_window_min, booking_window_max, buffer_time)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, coalesce((
		select max(sequence) + 1 from "Service" where category_id is not distinct from $2 and merchant_id = $1 and deleted_on is null
		), 1), $12, $13, $14, $15, $16, $17)
	returning id
	`

	var serviceId int
	err := r.db.QueryRow(ctx, query, serv.MerchantId, serv.CategoryId, serv.BookingType, serv.Name, serv.Description, serv.Color,
		serv.TotalDuration, serv.Price, serv.Cost, serv.PriceType, serv.IsActive, serv.MinParticipants, serv.MaxParticipants,
		serv.CancelDeadline, serv.BookingWindowMin, serv.BookingWindowMax, serv.BufferTime).Scan(&serviceId)
	if err != nil {
		return 0, err
	}

	return serviceId, nil
}

func (r *catalogRepository) UpdateService(ctx context.Context, s domain.Service) (*int, error) {
	query := `
	with old as (
		select id, category_id from "Service"
		where id = $1 and merchant_id = $2 and deleted_on is null
	)
	update "Service"
	set category_id = $3, name = $4, description = $5, color = $6, total_duration = $7, price_per_person = $8, cost_per_person = $9,
		price_type = $10, is_active = $11, cancel_deadline = $12, booking_window_min = $13, booking_window_max = $14, buffer_time = $15,
		min_participants = $16, max_participants = $17,
		sequence = case
			when old.category_id is distinct from $3 then (
				coalesce((
					select max(sequence) + 1 from "Service" where category_id is not distinct from $3 and merchant_id = $2 and deleted_on is null
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
		s.Price, s.Cost, s.PriceType, s.IsActive, s.CancelDeadline, s.BookingWindowMin, s.BookingWindowMax, s.BufferTime,
		s.MinParticipants, s.MaxParticipants).Scan(&oldCategoryId)
	if err != nil {
		return nil, err
	}

	return oldCategoryId, nil
}

func (r *catalogRepository) DeleteService(ctx context.Context, merchantId uuid.UUID, serviceId int) error {
	query := `
	update "Service"
	set deleted_on = $1
	where merchant_id = $2 and ID = $3
	`

	_, err := r.db.Exec(ctx, query, time.Now().UTC(), merchantId, serviceId)
	if err != nil {
		return err
	}

	return nil
}

func (r *catalogRepository) DeactivateService(ctx context.Context, merchantId uuid.UUID, serviceId int) error {
	query := `
	update "Service"
	set is_active = true
	where id = $1 and merchant_id = $2
	`

	_, err := r.db.Exec(ctx, query, serviceId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

func (r *catalogRepository) ActivateService(ctx context.Context, merchantId uuid.UUID, serviceId int) error {
	query := `
	update "Service"
	set is_active = false
	where id = $1 and merchant_id = $2
	`

	_, err := r.db.Exec(ctx, query, serviceId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

// TODO: this implementation looks crazy
func (r *catalogRepository) ReorderServices(ctx context.Context, merchantId uuid.UUID, categoryId *int, serviceIds []int) error {
	var cases []string
	var inParams []string
	params := make([]any, len(serviceIds))

	for i, id := range serviceIds {
		params[i] = id
		paramIndex := i + 1
		cases = append(cases, fmt.Sprintf("when $%d then %d", paramIndex, i+1))
		inParams = append(inParams, fmt.Sprintf("$%d", paramIndex))
	}

	categoryIdParamIndex := len(serviceIds) + 1
	categoryCondition := ""
	if categoryId == nil {
		categoryCondition = "category_id is null"
	} else {
		categoryCondition = fmt.Sprintf("category_id = $%d", categoryIdParamIndex)
		params = append(params, *categoryId)
	}

	merchantIdParamIndex := len(params) + 1
	params = append(params, merchantId)

	query := fmt.Sprintf(`
	update "Service"
	set sequence = case id
	%s
	end
	where id in (%s) and %s and merchant_id = $%d`,
		strings.Join(cases, "\n"), strings.Join(inParams, ", "), categoryCondition, merchantIdParamIndex)

	_, err := r.db.Exec(ctx, query, params...)
	if err != nil {
		return err
	}

	return nil
}

func (r *catalogRepository) ReorderServicesAfterUpdate(ctx context.Context, categoryId *int, merchantId uuid.UUID, exludeServiceId *int) error {
	query := `
	with reordered as (
		select id, row_number() over (order by sequence) as new_sequence
		from "Service"
		where category_id is not distinct from $1 and merchant_id = $2 and deleted_on is null and ($3::int is null or id != $3)
	)
	update "Service" s
	set sequence = r.new_sequence
	from reordered r
	where s.id = r.id
	`

	_, err := r.db.Exec(ctx, query, categoryId, merchantId, exludeServiceId)
	if err != nil {
		return err
	}

	return nil
}

// TODO: full outer joins can be expensive, this should be reevaluated later for performance
func (r *catalogRepository) GetServices(ctx context.Context, merchantId uuid.UUID) ([]domain.ServicesGroupedByCategory, error) {
	query := `
	with services as (
		select s.id, s.merchant_id, s.category_id, s.booking_type, s.name, s.description, s.color, s.total_duration, s.price_per_person, s.cost_per_person,
			s.price_type, s.is_active, s.sequence,
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
		left join "ServicePhase" sp on s.id = sp.service_id and sp.deleted_on is null
		where s.merchant_id = $1 and s.deleted_on is null
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
				'price', s.price_per_person,
				'cost', s.cost_per_person,
				'price_type', s.price_type,
				'is_active', s.is_active,
				'sequence', s.sequence,
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
			sgby.Services = []domain.PublicServiceWithPhases{}
		}

		return sgby, nil
	})
	if err != nil {
		return []domain.ServicesGroupedByCategory{}, err
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
	where s.merchant_id = $1 and s.is_active = true and s.deleted_on is null
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
		return []domain.ServicesGroupedByCategoriesForCalendar{}, err
	}

	// if services array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(servicesGroupByCategory) == 0 {
		servicesGroupByCategory = []domain.ServicesGroupedByCategoriesForCalendar{}
	}

	return servicesGroupByCategory, nil
}

func (r *catalogRepository) GetServiceWithPhases(ctx context.Context, serviceID int, merchantId uuid.UUID) (domain.PublicServiceWithPhases, error) {
	query := `
	select s.id, s.merchant_id, s.booking_type, s.category_id, s.name, s.description, s.color, s.total_duration, s.price_per_person as price, s.cost_per_person as cost,
		s.price_type, s.min_participants, s.max_participants, s.is_active, sp.id, sp.service_id, sp.name, sp.sequence, sp.duration, sp.phase_type
	from "Service" s
	left join "ServicePhase" sp on s.id = sp.service_id
	where s.id = $1 and s.merchant_id = $2 and s.deleted_on is null and sp.deleted_on is null
	order by sp.sequence asc
	`

	rows, err := r.db.Query(ctx, query, serviceID, merchantId)
	if err != nil {
		return domain.PublicServiceWithPhases{}, err
	}
	defer rows.Close()

	var pswp domain.PublicServiceWithPhases

	firstRow := true
	for rows.Next() {
		var ts domain.Service
		var p domain.PublicServicePhase
		var spId *int

		err := rows.Scan(&ts.Id, &ts.MerchantId, &ts.BookingType, &ts.CategoryId, &ts.Name, &ts.Description, &ts.Color, &ts.TotalDuration,
			&ts.Price, &ts.Cost, &ts.PriceType, &ts.MinParticipants, &ts.MaxParticipants, &ts.IsActive, &spId, &p.ServiceId, &p.Name, &p.Sequence, &p.Duration, &p.PhaseType)
		if err != nil {
			return domain.PublicServiceWithPhases{}, err
		}

		if firstRow {
			pswp = domain.PublicServiceWithPhases{
				Id:              ts.Id,
				MerchantId:      ts.MerchantId,
				BookingType:     ts.BookingType,
				CategoryId:      ts.CategoryId,
				Name:            ts.Name,
				Description:     ts.Description,
				Color:           ts.Color,
				TotalDuration:   ts.TotalDuration,
				Price:           ts.Price,
				Cost:            ts.Cost,
				PriceType:       ts.PriceType,
				IsActive:        ts.IsActive,
				MinParticipants: ts.MinParticipants,
				MaxParticipants: ts.MaxParticipants,
			}
			firstRow = false
		}

		if spId != nil {
			p.Id = *spId
			pswp.Phases = append(pswp.Phases, p)
		}
	}

	return pswp, nil
}

func (r *catalogRepository) GetServicesForMerchantPage(ctx context.Context, merchantId uuid.UUID) ([]domain.MerchantPageServicesGroupedByCategory, error) {
	query := `
	select sc.id, sc.name, sc.sequence,
	coalesce (
		jsonb_agg(
			jsonb_build_object(
				'id', s.id,
				'name', s.name,
				'description', s.description,
				'total_duration', s.total_duration,
				'price', s.price_per_person,
				'price_type', s.price_type,
				'sequence', s.sequence
			) order by s.sequence
		) filter (where s.id is not null),
	'[]'::jsonb) as services
	from "Service" s
	left join "ServiceCategory" sc on s.category_id = sc.id
	where s.merchant_id = $1 and s.is_active = true and s.deleted_on is null and  s.booking_type = 'appointment'
	group by sc.id, sc.name
	order by sc.sequence, sc.name
	`

	rows, _ := r.db.Query(ctx, query, merchantId)
	servicesGroupByCategory, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.MerchantPageServicesGroupedByCategory, error) {
		var sgby domain.MerchantPageServicesGroupedByCategory
		var services []byte

		err := row.Scan(&sgby.Id, &sgby.Name, &sgby.Sequence, &services)
		if err != nil {
			return domain.MerchantPageServicesGroupedByCategory{}, err
		}

		if len(services) > 0 {
			err = json.Unmarshal(services, &sgby.Services)
			if err != nil {
				return domain.MerchantPageServicesGroupedByCategory{}, err
			}
		} else {
			sgby.Services = []domain.MerchantPageService{}
		}

		return sgby, nil
	})
	if err != nil {
		return []domain.MerchantPageServicesGroupedByCategory{}, err
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
	left join "ServicePhase" sp on s.id = sp.service_id and sp.deleted_on is null
	left join "Location" l on l.merchant_id = $2 and l.id = $3
	where s.id = $1 and s.merchant_id = $2 and s.deleted_on is null
	group by s.id, l.formatted_location, l.geo_point`

	var data domain.PublicServiceDetails
	var phaseJson []byte

	err := r.db.QueryRow(ctx, query, serviceId, merchantId, locationId).Scan(&data.Id, &data.Name, &data.Description, &data.TotalDuration,
		&data.Price, &data.PriceType, &data.FormattedLocation, &data.GeoPoint, &phaseJson)
	if err != nil {
		return domain.PublicServiceDetails{}, err
	}

	if len(phaseJson) > 0 {
		err = json.Unmarshal(phaseJson, &data.Phases)
		if err != nil {
			return domain.PublicServiceDetails{}, err
		}
	} else {
		data.Phases = []domain.PublicServicePhase{}
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
				)
			) as phases
		from "ServicePhase" sp
		where sp.deleted_on is null
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
		where p.deleted_on is null
		group by sprod.service_id
	)
	select s.id, s.name, s.category_id, s.description, s.color, s.total_duration, s.price_per_person as price, s.cost_per_person as cost, s.price_type, s.is_active, s.sequence,
		jsonb_build_object(
		 	'cancel_deadline', s.cancel_deadline,
         	'booking_window_min', s.booking_window_min,
         	'booking_window_max', s.booking_window_max,
         	'buffer_time', s.buffer_time
		) as settings,
		coalesce(phases.phases, '[]'::jsonb) as phases,
		coalesce(products.products, '[]'::jsonb) as products
	from "Service" s
	left join phases on s.id = phases.service_id
	left join products on s.id = products.service_id
	where s.id = $1 and s.merchant_id = $2 and s.deleted_on is null
	`

	var spd domain.ServicePageData
	var settingsJson []byte
	var phaseJson []byte
	var productJson []byte

	err := r.db.QueryRow(ctx, query, serviceId, merchantId).Scan(&spd.Id, &spd.Name, &spd.CategoryId, &spd.Description,
		&spd.Color, &spd.TotalDuration, &spd.Price, &spd.Cost, &spd.PriceType, &spd.IsActive, &spd.Sequence, &settingsJson, &phaseJson, &productJson)
	if err != nil {
		return domain.ServicePageData{}, err
	}

	if len(settingsJson) > 0 {
		if err := json.Unmarshal(settingsJson, &spd.Settings); err != nil {
			return domain.ServicePageData{}, err
		}
	}

	if len(phaseJson) > 0 {
		err = json.Unmarshal(phaseJson, &spd.Phases)
		if err != nil {
			return domain.ServicePageData{}, err
		}
	} else {
		spd.Phases = []domain.PublicServicePhase{}
	}

	if len(productJson) > 0 {
		err = json.Unmarshal(productJson, &spd.Products)
		if err != nil {
			return domain.ServicePageData{}, err
		}
	} else {
		spd.Products = []domain.MinimalProductInfoWithUsage{}
	}

	return spd, nil
}

func (r *catalogRepository) GetGroupServicePageData(ctx context.Context, merchantId uuid.UUID, serviceId int) (domain.GroupServicePageData, error) {
	query := `
		with products as (
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
		where p.deleted_on is null
		group by sprod.service_id
	)
	select s.id, s.name, s.category_id, s.description, s.color, s.total_duration as duration, s.price_per_person as price, s.cost_per_person as cost, s.price_type, s.is_active, s.sequence, s.min_participants, s.max_participants,
		jsonb_build_object(
		 	'cancel_deadline', s.cancel_deadline,
         	'booking_window_min', s.booking_window_min,
         	'booking_window_max', s.booking_window_max,
         	'buffer_time', s.buffer_time
		) as settings,
		coalesce(products.products, '[]'::jsonb) as products
	from "Service" s
	left join products on s.id = products.service_id
	where s.id = $1 and s.merchant_id = $2 and s.deleted_on is null `

	var gspd domain.GroupServicePageData
	var settingsJson []byte
	var productJson []byte

	err := r.db.QueryRow(ctx, query, serviceId, merchantId).Scan(&gspd.Id, &gspd.Name, &gspd.CategoryId, &gspd.Description,
		&gspd.Color, &gspd.Duration, &gspd.Price, &gspd.Cost, &gspd.PriceType, &gspd.IsActive, &gspd.Sequence,
		&gspd.MinParicipants, &gspd.MaxParticipants, &settingsJson, &productJson)
	if err != nil {
		return domain.GroupServicePageData{}, err
	}

	if len(settingsJson) > 0 {
		if err := json.Unmarshal(settingsJson, &gspd.Settings); err != nil {
			return domain.GroupServicePageData{}, err
		}
	}

	if len(productJson) > 0 {
		err = json.Unmarshal(productJson, &gspd.Products)
		if err != nil {
			return domain.GroupServicePageData{}, err
		}
	} else {
		gspd.Products = []domain.MinimalProductInfoWithUsage{}
	}

	return gspd, nil
}

func (r *catalogRepository) GetServicePageFormOptions(ctx context.Context, merchantId uuid.UUID) (domain.ServicePageFormOptions, error) {
	query := `
	with product as (
		select id, name, unit from "Product" where merchant_id = $1 and deleted_on is null
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
		return domain.ServicePageFormOptions{}, err
	}

	if len(products) > 0 {
		err = json.Unmarshal(products, &spfo.Products)
		if err != nil {
			return domain.ServicePageFormOptions{}, err
		}
	} else {
		spfo.Products = []domain.MinimalProductInfo{}
	}

	if len(categories) > 0 {
		err = json.Unmarshal(categories, &spfo.Categories)
		if err != nil {
			return domain.ServicePageFormOptions{}, err
		}
	} else {
		spfo.Categories = []domain.ServiceCategory{}
	}

	return spfo, nil
}

func (r *catalogRepository) GetMinimalServiceInfo(ctx context.Context, merchantId uuid.UUID, serviceId, locationId int) (domain.MinimalServiceInfo, error) {
	query := `
	select s.name, s.total_duration, s.price_per_person as price, s.price_type, l.formatted_location
	from "Service" s
	left join "Location" l on l.merchant_id = $1 and l.id = $3
	where s.merchant_id = $1 and s.id = $2 and s.deleted_on is null
	`
	var msi domain.MinimalServiceInfo
	err := r.db.QueryRow(ctx, query, merchantId, serviceId, locationId).Scan(&msi.Name, &msi.TotalDuration, &msi.Price, &msi.PriceType, &msi.FormattedLocation)
	if err != nil {
		return domain.MinimalServiceInfo{}, err
	}

	return msi, nil
}

func (r *catalogRepository) NewServicePhases(ctx context.Context, serviceId int, phases []domain.ServicePhase) error {
	query := `
	insert into "ServicePhase" (service_id, name, sequence, duration, phase_type)
	select $1, unnest($2::text[]), unnest($3::int[]), unnest($4::int[]), unnest($5::service_phase_type[])
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
		return err
	}

	return nil
}

func (r *catalogRepository) UpdateServicePhases(ctx context.Context, phases []domain.ServicePhase) error {
	query := `
	update "ServicePhase" sp
	set name = u.name, sequence = u.sequence, duration = u.duration, phase_type = u.phase_type
	from unnest($1::int[], $2::text[], $3::int[], $4::int[], $5::service_phase_type[]) as u(id, name, sequence, duration, phase_type)
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
		return err
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
		return err
	}

	return nil
}

func (r *catalogRepository) DeleteServicePhases(ctx context.Context, phaseIds []int) error {
	query := `
	update "ServicePhase"
	set deleted_on = $2
	where id = any($1::int[])
	`

	_, err := r.db.Exec(ctx, query, phaseIds, time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

func (r *catalogRepository) DeleteServicePhasesForService(ctx context.Context, serviceId int) error {
	query := `
	update "ServicePhase"
	set deleted_on = $2
	where service_id = $1
	`

	_, err := r.db.Exec(ctx, query, serviceId, time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

func (r *catalogRepository) GetServicePhases(ctx context.Context, serviceId int) ([]domain.ServicePhase, error) {
	query := `
	select *
	from "ServicePhase"
	where service_id = $1 and deleted_on is null
	`

	rows, _ := r.db.Query(ctx, query, serviceId)
	phases, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.ServicePhase])
	if err != nil {
		return []domain.ServicePhase{}, err
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
		return err
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
		return err
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
		return err
	}

	return nil
}

// TODO: this implementation looks crazy
func (r *catalogRepository) ReorderServiceCategories(ctx context.Context, merchantId uuid.UUID, categoryIds []int) error {
	var cases []string
	var inParams []string
	params := make([]any, len(categoryIds))

	for i, id := range categoryIds {
		params[i] = id
		paramIndex := i + 1
		cases = append(cases, fmt.Sprintf("when $%d then %d", paramIndex, i+1))
		inParams = append(inParams, fmt.Sprintf("$%d", paramIndex))
	}

	merchantIdParamIndex := len(categoryIds) + 1
	params = append(params, merchantId)

	query := fmt.Sprintf(`
	update "ServiceCategory"
	set sequence = case id
	%s
	end
	where id in (%s) and merchant_id = $%d`, strings.Join(cases, "\n"), strings.Join(inParams, ", "), merchantIdParamIndex)

	_, err := r.db.Exec(ctx, query, params...)
	if err != nil {
		return err
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
		return err
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
		return err
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
		return err
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
		return []domain.ConnectedProducts{}, err
	}

	return connectedProducts, nil
}
