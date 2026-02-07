package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

type Service struct {
	Id              int               `json:"ID"`
	MerchantId      uuid.UUID         `json:"merchant_id"`
	CategoryId      *int              `json:"category_id"`
	BookingType     types.BookingType `json:"booking_type"`
	Name            string            `json:"name"`
	Description     *string           `json:"description"`
	Color           string            `json:"color"`
	TotalDuration   int               `json:"total_duration"`
	Price           *currencyx.Price  `json:"price"`
	Cost            *currencyx.Price  `json:"cost"`
	PriceType       types.PriceType   `json:"price_type"`
	IsActive        bool              `json:"is_active"`
	Sequence        int               `json:"sequence"`
	MinParticipants int               `json:"min_participants"`
	MaxParticipants int               `json:"max_participants"`
	ServiceSettings
	DeletedOn *time.Time `json:"deleted_on"`
}

type ServicePhase struct {
	Id        int                    `json:"ID"`
	ServiceId int                    `json:"service_id"`
	Name      string                 `json:"name"`
	Sequence  int                    `json:"sequence"`
	Duration  int                    `json:"duration"`
	PhaseType types.ServicePhaseType `json:"phase_type"`
	DeletedOn *time.Time             `json:"deleted_on"`
}

type ServiceCategory struct {
	Id         int       `json:"id" db:"id"`
	MerchantId uuid.UUID `json:"merchant_id"`
	LocationId int       `json:"location_id"`
	Name       string    `json:"name" db:"name"`
	Sequence   int       `json:"sequence"`
}

type ServiceSettings struct {
	CancelDeadline   *int `json:"cancel_deadline"`
	BookingWindowMin *int `json:"booking_window_min"`
	BookingWindowMax *int `json:"booking_window_max"`
	BufferTime       *int `json:"buffer_time"`
}

func (s *service) NewService(ctx context.Context, serv Service, servPhases []ServicePhase, ConnProducts []ConnectedProducts) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	serviceQuery := `
	insert into "Service" (merchant_id, category_id, booking_type, name, description, color, total_duration, price_per_person, cost_per_person,
		price_type, is_active, sequence, min_participants, max_participants, cancel_deadline, booking_window_min, booking_window_max, buffer_time)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, coalesce((
		select max(sequence) + 1 from "Service" where category_id is not distinct from $2 and merchant_id = $1 and deleted_on is null
		), 1), $12, $13, $14, $15, $16, $17)
	returning id
	`

	var serviceId int
	err = tx.QueryRow(ctx, serviceQuery, serv.MerchantId, serv.CategoryId, serv.BookingType, serv.Name, serv.Description, serv.Color,
		serv.TotalDuration, serv.Price, serv.Cost, serv.PriceType, serv.IsActive, serv.MinParticipants, serv.MaxParticipants,
		serv.CancelDeadline, serv.BookingWindowMin, serv.BookingWindowMax, serv.BufferTime).Scan(&serviceId)
	if err != nil {
		return err
	}

	servicePhaseQuery := `
	insert into "ServicePhase" (service_id, name, sequence, duration, phase_type)
	values ($1, $2, $3, $4, $5)
	`

	for _, phase := range servPhases {
		_, err := tx.Exec(ctx, servicePhaseQuery, serviceId, phase.Name, phase.Sequence, phase.Duration, phase.PhaseType)
		if err != nil {
			return err
		}
	}

	serviceProductQuery := `
	insert into "ServiceProduct" (service_id, product_id, amount_used)
	select $1, p.id, $3
	from "Product" p where p.id = $2 and p.merchant_id = $4`

	for _, cp := range ConnProducts {
		_, err := tx.Exec(ctx, serviceProductQuery, serviceId, cp.ProductId, cp.AmountUsed, serv.MerchantId)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *service) GetServiceWithPhasesById(ctx context.Context, serviceID int, merchantId uuid.UUID) (PublicServiceWithPhases, error) {
	query := `
	select s.id, s.merchant_id, s.booking_type, s.category_id, s.name, s.description, s.color, s.total_duration, s.price_per_person as price, s.cost_per_person as cost,
		s.price_type, s.min_participants, s.max_participants, s.is_active, sp.id, sp.service_id, sp.name, sp.sequence, sp.duration, sp.phase_type
	from "Service" s
	left join "ServicePhase" sp on s.id = sp.service_id
	where s.id = $1 and s.merchant_id = $2 and s.deleted_on is null and sp.deleted_on is null
	order by sp.sequence asc
	`

	rows, err := s.db.Query(ctx, query, serviceID, merchantId)
	if err != nil {
		return PublicServiceWithPhases{}, err
	}
	defer rows.Close()

	var pswp PublicServiceWithPhases

	firstRow := true
	for rows.Next() {
		var ts Service
		var p PublicServicePhase
		var spId *int

		err := rows.Scan(&ts.Id, &ts.MerchantId, &ts.BookingType, &ts.CategoryId, &ts.Name, &ts.Description, &ts.Color, &ts.TotalDuration,
			&ts.Price, &ts.Cost, &ts.PriceType, &ts.MinParticipants, &ts.MaxParticipants, &ts.IsActive, &spId, &p.ServiceId, &p.Name, &p.Sequence, &p.Duration, &p.PhaseType)
		if err != nil {
			return PublicServiceWithPhases{}, err
		}

		if firstRow {
			pswp = PublicServiceWithPhases{
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

type PublicServicePhase struct {
	Id        int                    `json:"id" db:"id"`
	ServiceId int                    `json:"service_id" db:"service_id"`
	Name      string                 `json:"name" db:"name"`
	Sequence  int                    `json:"sequence" db:"sequence"`
	Duration  int                    `json:"duration" db:"duration"`
	PhaseType types.ServicePhaseType `json:"phase_type" db:"phase_type"`
}

type PublicServiceWithPhases struct {
	Id              int                  `json:"id"`
	MerchantId      uuid.UUID            `json:"merchant_id"`
	BookingType     types.BookingType    `json:"booking_type"`
	CategoryId      *int                 `json:"category_id"`
	Name            string               `json:"name"`
	Description     *string              `json:"description"`
	Color           string               `json:"color"`
	TotalDuration   int                  `json:"total_duration"`
	Price           *currencyx.Price     `json:"price"`
	Cost            *currencyx.Price     `json:"cost"`
	PriceType       types.PriceType      `json:"price_type"`
	IsActive        bool                 `json:"is_active"`
	MinParticipants int                  `json:"min_participants"`
	MaxParticipants int                  `json:"max_participants"`
	Sequence        int                  `json:"sequence"`
	Phases          []PublicServicePhase `json:"phases"`
}

type ServicesGroupedByCategory struct {
	Id       *int                      `json:"id"`
	Name     *string                   `json:"name"`
	Sequence *int                      `json:"sequence"`
	Services []PublicServiceWithPhases `json:"services"`
}

// TODO: full outer joins can be expensive, this should be reevaluated later for performance
func (s *service) GetServicesByMerchantId(ctx context.Context, merchantId uuid.UUID) ([]ServicesGroupedByCategory, error) {
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

	rows, _ := s.db.Query(ctx, query, merchantId)
	servicesGroupByCategory, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ServicesGroupedByCategory, error) {
		var sgby ServicesGroupedByCategory
		var services []byte

		err := row.Scan(&sgby.Id, &sgby.Name, &sgby.Sequence, &services)
		if err != nil {
			return ServicesGroupedByCategory{}, err
		}

		if len(services) > 0 {
			err = json.Unmarshal(services, &sgby.Services)
			if err != nil {
				return ServicesGroupedByCategory{}, err
			}
		} else {
			sgby.Services = []PublicServiceWithPhases{}
		}

		return sgby, nil
	})
	if err != nil {
		return []ServicesGroupedByCategory{}, err
	}

	// if services array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(servicesGroupByCategory) == 0 {
		servicesGroupByCategory = []ServicesGroupedByCategory{}
	}

	return servicesGroupByCategory, nil
}

type MerchantPageService struct {
	Id            int                       `json:"id"`
	CategoryId    *int                      `json:"category_id"`
	Name          string                    `json:"name"`
	Description   *string                   `json:"description"`
	TotalDuration int                       `json:"total_duration"`
	Price         *currencyx.FormattedPrice `json:"price"`
	PriceType     types.PriceType           `json:"price_type"`
	Sequence      int                       `json:"sequence"`
}

type MerchantPageServicesGroupedByCategory struct {
	Id       *int                  `json:"id"`
	Name     *string               `json:"name"`
	Sequence *int                  `json:"sequence"`
	Services []MerchantPageService `json:"services"`
}

func (s *service) GetServicesForMerchantPage(ctx context.Context, merchantId uuid.UUID) ([]MerchantPageServicesGroupedByCategory, error) {
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

	rows, _ := s.db.Query(ctx, query, merchantId)
	servicesGroupByCategory, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (MerchantPageServicesGroupedByCategory, error) {
		var sgby MerchantPageServicesGroupedByCategory
		var services []byte

		err := row.Scan(&sgby.Id, &sgby.Name, &sgby.Sequence, &services)
		if err != nil {
			return MerchantPageServicesGroupedByCategory{}, err
		}

		if len(services) > 0 {
			err = json.Unmarshal(services, &sgby.Services)
			if err != nil {
				return MerchantPageServicesGroupedByCategory{}, err
			}
		} else {
			sgby.Services = []MerchantPageService{}
		}

		return sgby, nil
	})
	if err != nil {
		return []MerchantPageServicesGroupedByCategory{}, err
	}

	// if services array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(servicesGroupByCategory) == 0 {
		servicesGroupByCategory = []MerchantPageServicesGroupedByCategory{}
	}

	return servicesGroupByCategory, nil
}

func (s *service) DeleteServiceById(ctx context.Context, merchantId uuid.UUID, serviceId int) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	deleteServiceQuery := `
	update "Service"
	set deleted_on = $1
	where merchant_id = $2 and ID = $3
	`

	_, err = s.db.Exec(ctx, deleteServiceQuery, time.Now().UTC(), merchantId, serviceId)
	if err != nil {
		return err
	}

	deletePhasesQuery := `
	update "ServicePhase"
	set deleted_on = $2
	where id = $1
	`

	_, err = s.db.Exec(ctx, deletePhasesQuery, serviceId, time.Now().UTC())
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func servicePhasesEqual(a, b PublicServicePhase) bool {
	return a.Name == b.Name &&
		a.Sequence == b.Sequence &&
		a.Duration == b.Duration &&
		a.PhaseType == b.PhaseType
}

type ServiceWithPhasesAndSettings struct {
	Id            int                  `json:"id"`
	MerchantId    uuid.UUID            `json:"merchant_id"`
	CategoryId    *int                 `json:"category_id"`
	Name          string               `json:"name"`
	Description   *string              `json:"description"`
	Color         string               `json:"color"`
	TotalDuration int                  `json:"total_duration"`
	Price         *currencyx.Price     `json:"price"`
	Cost          *currencyx.Price     `json:"cost"`
	PriceType     types.PriceType      `json:"price_type"`
	IsActive      bool                 `json:"is_active"`
	Sequence      int                  `json:"sequence"`
	Settings      ServiceSettings      `json:"settings"`
	Phases        []PublicServicePhase `json:"phases"`
}

func (s *service) UpdateServiceWithPhaseseById(ctx context.Context, pswp ServiceWithPhasesAndSettings) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	phasesForServiceQuery := `
	select id, name, sequence, duration, phase_type
	from "ServicePhase"
	where service_id = $1 and deleted_on is null
	`

	existingPhasesMap := map[int]PublicServicePhase{}

	var psp PublicServicePhase
	rows, _ := tx.Query(ctx, phasesForServiceQuery, pswp.Id)
	_, err = pgx.ForEachRow(rows, []any{&psp.Id, &psp.Name, &psp.Sequence, &psp.Duration, &psp.PhaseType}, func() error {
		existingPhasesMap[psp.Id] = psp
		return nil
	})
	if err != nil {
		return err
	}

	updatedPhasesMap := map[int]PublicServicePhase{}
	newPhases := []PublicServicePhase{}
	for _, phase := range pswp.Phases {
		if phase.Id == 0 {
			newPhases = append(newPhases, phase)
		} else {
			updatedPhasesMap[phase.Id] = phase
		}
	}

	deletePhasesQuery := `
	update "ServicePhase"
	set deleted_on = $2
	where id = $1
	`

	for id := range existingPhasesMap {
		if _, exists := updatedPhasesMap[id]; !exists {
			_, err := tx.Exec(ctx, deletePhasesQuery, id, time.Now().UTC())
			if err != nil {
				return err
			}
		}
	}

	updatePhasesQuery := `
	update "ServicePhase"
	set name = $2, sequence = $3, duration = $4, phase_type = $5
	where id = $1
	`

	for id, phase := range updatedPhasesMap {
		existingPhase := existingPhasesMap[id]
		if !servicePhasesEqual(existingPhase, phase) {
			_, err := tx.Exec(ctx, updatePhasesQuery, id, phase.Name, phase.Sequence, phase.Duration, phase.PhaseType)
			if err != nil {
				return err
			}
		}
	}

	insertNewPhasesQuery := `
	insert into "ServicePhase" (service_id, name, sequence, duration, phase_type)
	values ($1, $2, $3, $4, $5)
	`

	for _, phase := range newPhases {
		_, err := tx.Exec(ctx, insertNewPhasesQuery, pswp.Id, phase.Name, phase.Sequence, phase.Duration, phase.PhaseType)
		if err != nil {
			return err
		}
	}

	updateServiceQuery := `
	with old as (
		select id, category_id from "Service"
		where id = $1 and merchant_id = $2 and deleted_on is null
	)
	update "Service"
	set category_id = $3, name = $4, description = $5, color = $6, total_duration = $7, price_per_person = $8, cost_per_person = $9,
		price_type = $10, is_active = $11, cancel_deadline = $12, booking_window_min = $13, booking_window_max = $14, buffer_time =$15,
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
	err = tx.QueryRow(ctx, updateServiceQuery, pswp.Id, pswp.MerchantId, pswp.CategoryId, pswp.Name, pswp.Description, pswp.Color, pswp.TotalDuration,
		pswp.Price, pswp.Cost, pswp.PriceType, pswp.IsActive, pswp.Settings.CancelDeadline, pswp.Settings.BookingWindowMin,
		pswp.Settings.BookingWindowMax, pswp.Settings.BufferTime).Scan(&oldCategoryId)
	if err != nil {
		return err
	}

	// the categoryId has changed, reordering services is needed
	if (oldCategoryId == nil && pswp.CategoryId != nil) || (oldCategoryId != nil && (pswp.CategoryId == nil || *oldCategoryId != *pswp.CategoryId)) {
		reorderOldCategoryQuery := `
		with reordered as (
			select id, row_number() over (order by sequence) as new_sequence
			from "Service"
			where category_id is not distinct from $1 and merchant_id = $2 and deleted_on is null and id != $3
		)
		update "Service" s
		set sequence = r.new_sequence
		from reordered r
		where s.id = r.id
		`

		_, err := tx.Exec(ctx, reorderOldCategoryQuery, oldCategoryId, pswp.MerchantId, pswp.Id)
		if err != nil {
			return err
		}

		reorderNewCategoryQuery := `
		with reordered as (
			select id, row_number() over (order by sequence) as new_sequence
			from "Service"
			where category_id is not distinct from $1 and merchant_id = $2 and deleted_on is null
		)
		update "Service" s
		set sequence = r.new_sequence
		from reordered r
		where s.id = r.id
		`

		_, err = tx.Exec(ctx, reorderNewCategoryQuery, pswp.CategoryId, pswp.MerchantId)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *service) NewServiceCategory(ctx context.Context, merchantId uuid.UUID, sc ServiceCategory) error {
	query := `
	insert into "ServiceCategory" (merchant_id, name, sequence)
	values ($1, $2, coalesce(
		(select max(sequence) + 1 from "ServiceCategory" where merchant_id = $1), 1)
	)
	`

	_, err := s.db.Exec(ctx, query, merchantId, sc.Name)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) ReorderServices(ctx context.Context, merchantId uuid.UUID, categoryId *int, serviceIds []int) error {
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

	_, err := s.db.Exec(ctx, query, params...)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateServiceCategoryById(ctx context.Context, merchantId uuid.UUID, sc ServiceCategory) error {
	query := `
	update "ServiceCategory"
	set name = $3
	where id = $1 and merchant_id = $2
	`

	_, err := s.db.Exec(ctx, query, sc.Id, merchantId, sc.Name)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteServiceCategoryById(ctx context.Context, merchantId uuid.UUID, categoryId int) error {
	query := `
	delete from "ServiceCategory"
	where id = $1 and merchant_id = $2
	`

	_, err := s.db.Exec(ctx, query, categoryId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

type MinimalProductInfoWithUsage struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Unit       string `json:"unit"`
	AmountUsed int    `json:"amount_used"`
}

type ServicePageData struct {
	Id            int                           `json:"id"`
	CategoryId    *int                          `json:"category_id"`
	Name          string                        `json:"name"`
	Description   *string                       `json:"description"`
	Color         string                        `json:"color"`
	TotalDuration int                           `json:"total_duration"`
	Price         *currencyx.Price              `json:"price"`
	Cost          *currencyx.Price              `json:"cost"`
	PriceType     types.PriceType               `json:"price_type"`
	IsActive      bool                          `json:"is_active"`
	Sequence      int                           `json:"sequence"`
	Settings      ServiceSettings               `json:"settings"`
	Phases        []PublicServicePhase          `json:"phases"`
	Products      []MinimalProductInfoWithUsage `json:"used_products"`
}

func (s *service) GetAllServicePageData(ctx context.Context, serviceId int, merchantId uuid.UUID) (ServicePageData, error) {
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

	var spd ServicePageData
	var settingsJson []byte
	var phaseJson []byte
	var productJson []byte

	err := s.db.QueryRow(ctx, query, serviceId, merchantId).Scan(&spd.Id, &spd.Name, &spd.CategoryId, &spd.Description,
		&spd.Color, &spd.TotalDuration, &spd.Price, &spd.Cost, &spd.PriceType, &spd.IsActive, &spd.Sequence, &settingsJson, &phaseJson, &productJson)
	if err != nil {
		return ServicePageData{}, err
	}

	if len(settingsJson) > 0 {
		if err := json.Unmarshal(settingsJson, &spd.Settings); err != nil {
			return ServicePageData{}, err
		}
	}

	if len(phaseJson) > 0 {
		err = json.Unmarshal(phaseJson, &spd.Phases)
		if err != nil {
			return ServicePageData{}, err
		}
	} else {
		spd.Phases = []PublicServicePhase{}
	}

	if len(productJson) > 0 {
		err = json.Unmarshal(productJson, &spd.Products)
		if err != nil {
			return ServicePageData{}, err
		}
	} else {
		spd.Products = []MinimalProductInfoWithUsage{}
	}

	return spd, nil
}

type MinimalProductInfo struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type ServicePageFormOptions struct {
	Products   []MinimalProductInfo `json:"products"`
	Categories []ServiceCategory    `json:"categories"`
}

func (s *service) GetServicePageFormOptions(ctx context.Context, merchantId uuid.UUID) (ServicePageFormOptions, error) {
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

	var spfo ServicePageFormOptions
	var products []byte
	var categories []byte

	err := s.db.QueryRow(ctx, query, merchantId).Scan(&products, &categories)
	if err != nil {
		return ServicePageFormOptions{}, err
	}

	if len(products) > 0 {
		err = json.Unmarshal(products, &spfo.Products)
		if err != nil {
			return ServicePageFormOptions{}, err
		}
	} else {
		spfo.Products = []MinimalProductInfo{}
	}

	if len(categories) > 0 {
		err = json.Unmarshal(categories, &spfo.Categories)
		if err != nil {
			return ServicePageFormOptions{}, err
		}
	} else {
		spfo.Categories = []ServiceCategory{}
	}

	return spfo, nil
}

func (s *service) DeactivateServiceById(ctx context.Context, merchantId uuid.UUID, serviceId int) error {
	query := `
	update "Service"
	set is_active = true
	where id = $1 and merchant_id = $2
	`

	_, err := s.db.Exec(ctx, query, serviceId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) ActivateServiceById(ctx context.Context, merchantId uuid.UUID, serviceId int) error {
	query := `
	update "Service"
	set is_active = false
	where id = $1 and merchant_id = $2
	`

	_, err := s.db.Exec(ctx, query, serviceId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) ReorderServiceCategories(ctx context.Context, merchantId uuid.UUID, categoryIds []int) error {
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

	_, err := s.db.Exec(ctx, query, params...)
	if err != nil {
		return err
	}

	return nil
}

type ConnectedProducts struct {
	ProductId  int `json:"product_id"`
	ServiceId  int `json:"service_id"`
	AmountUsed int `json:"amount_used"`
}

func (s *service) UpdateConnectedProducts(ctx context.Context, serviceId int, productData []ConnectedProducts) error {

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	// nolint:errcheck
	defer tx.Rollback(ctx)

	productConnectionsQuery := `
	select product_id, amount_used from "ServiceProduct"
	where service_id = $1
	`

	existingConnectionsMap := map[int]ConnectedProducts{}

	var conn ConnectedProducts
	rows, _ := tx.Query(ctx, productConnectionsQuery, serviceId)
	_, err = pgx.ForEachRow(rows, []any{&conn.ProductId, &conn.AmountUsed}, func() error {
		conn.ServiceId = serviceId
		existingConnectionsMap[conn.ProductId] = conn
		return nil
	})
	if err != nil {
		return err
	}

	updatedConnectionsMap := map[int]ConnectedProducts{}
	for _, connection := range productData {
		if _, exists := existingConnectionsMap[connection.ProductId]; exists {
			updatedConnectionsMap[connection.ProductId] = connection
		}
	}

	deleteConnectionsQuery := `
	delete from "ServiceProduct"
	where service_id = $1 and product_id = $2
	`

	for productId := range existingConnectionsMap {
		if _, exists := updatedConnectionsMap[productId]; !exists {
			_, err := tx.Exec(ctx, deleteConnectionsQuery, serviceId, productId)
			if err != nil {
				return err
			}
		}
	}

	upsertQuery := `
	insert into "ServiceProduct" (service_id, product_id, amount_used)
	values ($1, $2, $3)
	on conflict (service_id, product_id) do update
	set amount_used = excluded.amount_used
	`

	for _, conn := range productData {
		_, err := tx.Exec(ctx, upsertQuery, serviceId, conn.ProductId, conn.AmountUsed)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

type PublicServiceDetails struct {
	Id                int                       `json:"id"`
	Name              string                    `json:"name"`
	Description       *string                   `json:"description"`
	TotalDuration     int                       `json:"total_duration"`
	Price             *currencyx.FormattedPrice `json:"price"`
	PriceType         types.PriceType           `json:"price_type"`
	FormattedLocation string                    `json:"formatted_location"`
	GeoPoint          types.GeoPoint            `json:"geo_point"`
	Phases            []PublicServicePhase      `json:"phases"`
}

func (s *service) GetServiceDetailsForMerchantPage(ctx context.Context, merchantId uuid.UUID, serviceId int, locationId int) (PublicServiceDetails, error) {
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

	var data PublicServiceDetails
	var phaseJson []byte

	err := s.db.QueryRow(ctx, query, serviceId, merchantId, locationId).Scan(&data.Id, &data.Name, &data.Description, &data.TotalDuration,
		&data.Price, &data.PriceType, &data.FormattedLocation, &data.GeoPoint, &phaseJson)
	if err != nil {
		return PublicServiceDetails{}, err
	}

	if len(phaseJson) > 0 {
		err = json.Unmarshal(phaseJson, &data.Phases)
		if err != nil {
			return PublicServiceDetails{}, err
		}
	} else {
		data.Phases = []PublicServicePhase{}
	}

	return data, nil
}

type MinimalServiceInfo struct {
	Name              string                    `json:"name"`
	TotalDuration     int                       `json:"total_duration"`
	Price             *currencyx.FormattedPrice `json:"price"`
	PriceType         types.PriceType           `json:"price_type"`
	FormattedLocation string                    `json:"formatted_location"`
}

func (s *service) GetMinimalServiceInfo(ctx context.Context, merchantId uuid.UUID, serviceId, locationId int) (MinimalServiceInfo, error) {
	query := `
	select s.name, s.total_duration, s.price_per_person as price, s.price_type, l.formatted_location
	from "Service" s
	left join "Location" l on l.merchant_id = $1 and l.id = $3
	where s.merchant_id = $1 and s.id = $2 and s.deleted_on is null
	`
	var msi MinimalServiceInfo
	err := s.db.QueryRow(ctx, query, merchantId, serviceId, locationId).Scan(&msi.Name, &msi.TotalDuration, &msi.Price, &msi.PriceType, &msi.FormattedLocation)
	if err != nil {
		return MinimalServiceInfo{}, err
	}

	return msi, nil
}

type ServiceForCalendar struct {
	Id            int               `json:"id" db:"id"`
	Name          string            `json:"name" db:"name"`
	TotalDuration int               `json:"total_duration" db:"total_duration"`
	BookingType   types.BookingType `json:"booking_type" db:"booking_type"`
}

func (s *service) GetServicesForCalendarByMerchant(ctx context.Context, merchantId uuid.UUID) ([]ServiceForCalendar, error) {
	query := `
	select id, name, total_duration, booking_type
	from "Service"
	where merchant_id = $1 and deleted_on is null
	`

	rows, _ := s.db.Query(ctx, query, merchantId)
	services, err := pgx.CollectRows(rows, pgx.RowToStructByName[ServiceForCalendar])
	if err != nil {
		return []ServiceForCalendar{}, err
	}

	return services, nil
}

type GroupServiceWithSettings struct {
	Id              int              `json:"id"`
	MerchantId      uuid.UUID        `json:"merchant_id"`
	CategoryId      *int             `json:"category_id"`
	Name            string           `json:"name"`
	Description     *string          `json:"description"`
	Color           string           `json:"color"`
	Duration        int              `json:"duration"`
	Price           *currencyx.Price `json:"price"`
	Cost            *currencyx.Price `json:"cost"`
	PriceType       types.PriceType  `json:"price_type"`
	IsActive        bool             `json:"is_active"`
	Sequence        int              `json:"sequence"`
	MinParticipants int              `json:"min_participants"`
	MaxParticipants int              `json:"max_participants"`
	Settings        ServiceSettings  `json:"settings"`
}

func (s *service) UpdateGroupServiceById(ctx context.Context, serv GroupServiceWithSettings) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	// nolint:errcheck
	defer tx.Rollback(ctx)

	updatePhaseQuery := `
	update "ServicePhase"
	set duration = $2
	where service_id = $1
	`

	_, err = tx.Exec(ctx, updatePhaseQuery, serv.Id, serv.Duration)
	if err != nil {
		return err
	}

	updateServiceQuery := `
    with old as (
        select id, category_id from "Service"
        where id = $1 and merchant_id = $2 and deleted_on is null
    )
    update "Service"
    set category_id = $3, name = $4, description = $5, color = $6, total_duration = $7, 
        price_per_person = $8, cost_per_person = $9, price_type = $10, is_active = $11, 
        cancel_deadline = $12, booking_window_min = $13, booking_window_max = $14, buffer_time = $15,
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
	err = tx.QueryRow(ctx, updateServiceQuery,
		serv.Id, serv.MerchantId, serv.CategoryId, serv.Name, serv.Description, serv.Color, serv.Duration,
		serv.Price, serv.Cost, serv.PriceType, serv.IsActive, serv.Settings.CancelDeadline, serv.Settings.BookingWindowMin,
		serv.Settings.BookingWindowMax, serv.Settings.BufferTime,
		serv.MinParticipants, serv.MaxParticipants,
	).Scan(&oldCategoryId)

	if err != nil {
		return err
	}

	if (oldCategoryId == nil && serv.CategoryId != nil) || (oldCategoryId != nil && (serv.CategoryId == nil || *oldCategoryId != *serv.CategoryId)) {
		reorderOldCategoryQuery := `
        with reordered as (
            select id, row_number() over (order by sequence) as new_sequence
            from "Service"
            where category_id is not distinct from $1 and merchant_id = $2 and deleted_on is null and id != $3
        )
        update "Service" s
        set sequence = r.new_sequence
        from reordered r
        where s.id = r.id
        `
		_, err := tx.Exec(ctx, reorderOldCategoryQuery, oldCategoryId, serv.MerchantId, serv.Id)
		if err != nil {
			return err
		}

		reorderNewCategoryQuery := `
        with reordered as (
            select id, row_number() over (order by sequence) as new_sequence
            from "Service"
            where category_id is not distinct from $1 and merchant_id = $2 and deleted_on is null
        )
        update "Service" s
        set sequence = r.new_sequence
        from reordered r
        where s.id = r.id
        `
		_, err = tx.Exec(ctx, reorderNewCategoryQuery, serv.CategoryId, serv.MerchantId)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

type GroupServicePageData struct {
	Id              int                           `json:"id"`
	CategoryId      *int                          `json:"category_id"`
	Name            string                        `json:"name"`
	Description     *string                       `json:"description"`
	Color           string                        `json:"color"`
	Duration        int                           `json:"duration"`
	Price           *currencyx.Price              `json:"price"`
	Cost            *currencyx.Price              `json:"cost"`
	PriceType       types.PriceType               `json:"price_type"`
	IsActive        bool                          `json:"is_active"`
	Sequence        int                           `json:"sequence"`
	MinParicipants  int                           `json:"min_participants"`
	MaxParticipants int                           `json:"max_participants"`
	Settings        ServiceSettings               `json:"settings"`
	Products        []MinimalProductInfoWithUsage `json:"used_products"`
}

func (s *service) GetGroupServicePageData(ctx context.Context, merchantId uuid.UUID, serviceId int) (GroupServicePageData, error) {
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

	var gspd GroupServicePageData
	var settingsJson []byte
	var productJson []byte

	err := s.db.QueryRow(ctx, query, serviceId, merchantId).Scan(&gspd.Id, &gspd.Name, &gspd.CategoryId, &gspd.Description,
		&gspd.Color, &gspd.Duration, &gspd.Price, &gspd.Cost, &gspd.PriceType, &gspd.IsActive, &gspd.Sequence,
		&gspd.MinParicipants, &gspd.MaxParticipants, &settingsJson, &productJson)
	if err != nil {
		return GroupServicePageData{}, err
	}

	if len(settingsJson) > 0 {
		if err := json.Unmarshal(settingsJson, &gspd.Settings); err != nil {
			return GroupServicePageData{}, err
		}
	}

	if len(productJson) > 0 {
		err = json.Unmarshal(productJson, &gspd.Products)
		if err != nil {
			return GroupServicePageData{}, err
		}
	} else {
		gspd.Products = []MinimalProductInfoWithUsage{}
	}

	return gspd, nil
}
