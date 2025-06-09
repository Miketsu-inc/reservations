package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	Id            int       `json:"ID"`
	MerchantId    uuid.UUID `json:"merchant_id"`
	CategoryId    *int      `json:"category_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Color         string    `json:"color"`
	TotalDuration int       `json:"total_duration"`
	Price         int       `json:"price"`
	PriceNote     *string   `json:"price_note"`
	Cost          int       `json:"cost"`
	IsActive      bool      `json:"is_active"`
	DeletedOn     *string   `json:"deleted_on"`
}

type ServicePhase struct {
	Id        int     `json:"ID"`
	ServiceId int     `json:"service_id"`
	Name      string  `json:"name"`
	Sequence  int     `json:"sequence"`
	Duration  int     `json:"duration"`
	PhaseType string  `json:"phase_type"`
	DeletedOn *string `json:"deleted_on"`
}

type ServiceCategory struct {
	Id   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func (s *service) NewService(ctx context.Context, serv Service, servPhases []ServicePhase) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	serviceQuery := `
	insert into "Service" (merchant_id, category_id, name, description, color, total_duration, price, price_note, cost, is_active)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	returning id
	`

	var serviceId int
	err = tx.QueryRow(ctx, serviceQuery, serv.MerchantId, serv.CategoryId, serv.Name, serv.Description, serv.Color,
		serv.TotalDuration, serv.Price, serv.PriceNote, serv.Cost, serv.IsActive).Scan(&serviceId)
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

	return tx.Commit(ctx)
}

func (s *service) GetServiceWithPhasesById(ctx context.Context, serviceID int, merchantId uuid.UUID) (PublicServiceWithPhases, error) {
	query := `
	select s.id, s.merchant_id, s.category_id, s.name, s.description, s.color, s.total_duration, s.price, s.price_note,
		s.cost, s.is_active, sp.id, sp.service_id, sp.name, sp.sequence, sp.duration, sp.phase_type
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

		err := rows.Scan(&ts.Id, &ts.MerchantId, &ts.CategoryId, &ts.Name, &ts.Description, &ts.Color, &ts.TotalDuration,
			&ts.Price, &ts.PriceNote, &ts.Cost, &ts.IsActive, &spId, &p.ServiceId, &p.Name, &p.Sequence, &p.Duration, &p.PhaseType)
		if err != nil {
			return PublicServiceWithPhases{}, err
		}

		if firstRow {
			pswp = PublicServiceWithPhases{
				Id:            ts.Id,
				MerchantId:    ts.MerchantId,
				CategoryId:    ts.CategoryId,
				Name:          ts.Name,
				Description:   ts.Description,
				Color:         ts.Color,
				TotalDuration: ts.TotalDuration,
				Price:         ts.Price,
				PriceNote:     ts.PriceNote,
				Cost:          ts.Cost,
				IsActive:      ts.IsActive,
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
	Id        int    `json:"id" db:"id"`
	ServiceId int    `json:"service_id" db:"service_id"`
	Name      string `json:"name" db:"name"`
	Sequence  int    `json:"sequence" db:"sequence"`
	Duration  int    `json:"duration" db:"duration"`
	PhaseType string `json:"phase_type" db:"phase_type"`
}

type PublicServiceWithPhases struct {
	Id            int                  `json:"id"`
	MerchantId    uuid.UUID            `json:"merchant_id"`
	CategoryId    *int                 `json:"category_id"`
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	Color         string               `json:"color"`
	TotalDuration int                  `json:"total_duration"`
	Price         int                  `json:"price"`
	PriceNote     *string              `json:"price_note"`
	Cost          int                  `json:"cost"`
	IsActive      bool                 `json:"is_active"`
	Phases        []PublicServicePhase `json:"phases"`
}

type ServicesGroupedByCategory struct {
	Id       *int                      `json:"id"`
	Name     *string                   `json:"name"`
	Services []PublicServiceWithPhases `json:"services"`
}

// TODO: full outer joins can be expensive, this should be reevaluated later for performance
func (s *service) GetServicesByMerchantId(ctx context.Context, merchantId uuid.UUID, filterForActive bool) ([]ServicesGroupedByCategory, error) {
	query := `
	with services as (
		select s.id, s.merchant_id, s.category_id, s.name, s.description, s.color, s.total_duration, s.price, s.price_note,
			s.cost, s.is_active,
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
		where s.merchant_id = $1 and ($2::bool = false or s.is_active = true) and s.deleted_on is null and sp.deleted_on is null
		group by s.id
	)
	select sc.id, sc.name,
	coalesce (
		jsonb_agg(
			jsonb_build_object(
				'id', s.id,
				'merchant_id', s.merchant_id,
				'name', s.name,
				'description', s.description,
				'color', s.color,
				'total_duration', s.total_duration,
				'price', s.price,
				'price_note', s.price_note,
				'cost', s.cost,
				'is_active', s.is_active,
				'phases', s.phases
			) order by s.name
		) filter (where s.id is not null),
	'[]'::jsonb) as services
	from "ServiceCategory" sc
	full outer join services s on s.category_id = sc.id
	where sc.merchant_id = $1 or s.merchant_id = $1
	group by sc.id, sc.name
	order by sc.name nulls first
	`

	rows, _ := s.db.Query(ctx, query, merchantId, filterForActive)
	servicesGroupByCategory, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (ServicesGroupedByCategory, error) {
		var sgby ServicesGroupedByCategory
		var services []byte

		err := row.Scan(&sgby.Id, &sgby.Name, &services)
		if err != nil {
			return ServicesGroupedByCategory{}, err
		}

		fmt.Println(sgby)
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

func (s *service) UpdateServicWithPhaseseById(ctx context.Context, pswp PublicServiceWithPhases) error {
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
	update "Service"
	set category_id = $3, name = $4, description = $5, color = $6, total_duration = $7, price = $8, price_note = $9,
		cost = $10, is_active = $11
	where ID = $1 and merchant_id = $2 and deleted_on is null
	`

	_, err = s.db.Exec(ctx, updateServiceQuery, pswp.Id, pswp.MerchantId, pswp.CategoryId, pswp.Name, pswp.Description, pswp.Color, pswp.TotalDuration,
		pswp.Price, pswp.PriceNote, pswp.Cost, pswp.IsActive)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *service) NewServiceCategory(ctx context.Context, merchantId uuid.UUID, sc ServiceCategory) error {
	query := `
	insert into "ServiceCategory" (merchant_id, name)
	values ($1, $2)
	`

	_, err := s.db.Exec(ctx, query, merchantId, sc.Name)
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
	Description   string                        `json:"description"`
	Color         string                        `json:"color"`
	TotalDuration int                           `json:"total_duration"`
	Price         int                           `json:"price"`
	PriceNote     *string                       `json:"price_note"`
	Cost          int                           `json:"cost"`
	IsActive      bool                          `json:"is_active"`
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
	select s.id, s.name, s.category_id, s.description, s.color, s.total_duration, s.price, s.price_note, s.cost, s.is_active,
		coalesce(phases.phases, '[]'::jsonb) as phases,
		coalesce(products.products, '[]'::jsonb) as products
	from "Service" s
	left join phases on s.id = phases.service_id
	left join products on s.id = products.service_id
	where s.id = $1 and s.merchant_id = $2 and s.deleted_on is null
	`

	var spd ServicePageData
	var phaseJson []byte
	var productJson []byte

	err := s.db.QueryRow(ctx, query, serviceId, merchantId).Scan(&spd.Id, &spd.Name, &spd.CategoryId, &spd.Description,
		&spd.Color, &spd.TotalDuration, &spd.Price, &spd.PriceNote, &spd.Cost, &spd.IsActive, &phaseJson, &productJson)
	if err != nil {
		return ServicePageData{}, err
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
}

type ServicePageFormOptions struct {
	Products   []MinimalProductInfo `json:"products"`
	Categories []ServiceCategory    `json:"categories"`
}

func (s *service) GetServicePageFormOptions(ctx context.Context, merchantId uuid.UUID) (ServicePageFormOptions, error) {
	query := `
	with product as (
		select id, name from "Product" where merchant_id = $1
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
