package database

import (
	"context"
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
	Id         int       `json:"ID"`
	MerchantId uuid.UUID `json:"merchant_id"`
	Name       string    `json:"name"`
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

func (s *service) GetServiceById(ctx context.Context, serviceID int, merchantId uuid.UUID) (PublicService, error) {
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
		return PublicService{}, err
	}
	defer rows.Close()

	var ps PublicService

	firstRow := true
	for rows.Next() {
		var ts Service
		var p PublicServicePhase
		var spId *int

		err := rows.Scan(&ts.Id, &ts.MerchantId, &ts.CategoryId, &ts.Name, &ts.Description, &ts.Color, &ts.TotalDuration,
			&ts.Price, &ts.PriceNote, &ts.Cost, &ts.IsActive, &spId, &p.ServiceId, &p.Name, &p.Sequence, &p.Duration, &p.PhaseType)
		if err != nil {
			return PublicService{}, err
		}

		if firstRow {
			ps = PublicService{
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
			ps.Phases = append(ps.Phases, p)
		}
	}

	return ps, nil
}

type PublicServicePhase struct {
	Id        int    `json:"id" db:"id"`
	ServiceId int    `json:"service_id" db:"service_id"`
	Name      string `json:"name" db:"name"`
	Sequence  int    `json:"sequence" db:"sequence"`
	Duration  int    `json:"duration" db:"duration"`
	PhaseType string `json:"phase_type" db:"phase_type"`
}

type PublicService struct {
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

func (s *service) GetServicesByMerchantId(ctx context.Context, merchantId uuid.UUID) ([]PublicService, error) {
	query := `
	select s.id, s.merchant_id, s.category_id, s.name, s.description, s.color, s.total_duration, s.price, s.price_note,
		s.cost, s.is_active, sp.id, sp.service_id, sp.name, sp.sequence, sp.duration, sp.phase_type
	from "Service" s
	left join "ServicePhase" sp on s.id = sp.service_id
	where s.merchant_id = $1 and s.deleted_on is null and sp.deleted_on is null
	order by s.id, sp.sequence asc
	`

	rows, err := s.db.Query(ctx, query, merchantId)
	if err != nil {
		return []PublicService{}, err
	}
	defer rows.Close()

	serviceMap := map[int]*PublicService{}

	for rows.Next() {
		var ts Service
		var p PublicServicePhase

		var spId *int

		err := rows.Scan(&ts.Id, &ts.MerchantId, &ts.CategoryId, &ts.Name, &ts.Description, &ts.Color, &ts.TotalDuration,
			&ts.Price, &ts.PriceNote, &ts.Cost, &ts.IsActive, &spId, &p.ServiceId, &p.Name, &p.Sequence, &p.Duration, &p.PhaseType)
		if err != nil {
			return []PublicService{}, err
		}

		ps, exists := serviceMap[ts.Id]
		if !exists {
			ps = &PublicService{
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
				Phases:        []PublicServicePhase{},
			}
			serviceMap[ts.Id] = ps
		}

		if spId != nil {
			p.Id = *spId
			ps.Phases = append(ps.Phases, p)
		}
	}

	var result []PublicService

	// if services array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(serviceMap) == 0 {
		result = []PublicService{}
	}

	for _, s := range serviceMap {
		result = append(result, *s)
	}

	return result, nil
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

func (s *service) UpdateServiceById(ctx context.Context, ps PublicService) error {
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
	rows, _ := tx.Query(ctx, phasesForServiceQuery, ps.Id)
	_, err = pgx.ForEachRow(rows, []any{&psp.Id, &psp.Name, &psp.Sequence, &psp.Duration, &psp.PhaseType}, func() error {
		existingPhasesMap[psp.Id] = psp
		return nil
	})
	if err != nil {
		return err
	}

	updatedPhasesMap := map[int]PublicServicePhase{}
	newPhases := []PublicServicePhase{}
	for _, phase := range ps.Phases {
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
		_, err := tx.Exec(ctx, insertNewPhasesQuery, ps.Id, phase.Name, phase.Sequence, phase.Duration, phase.PhaseType)
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

	_, err = s.db.Exec(ctx, updateServiceQuery, ps.Id, ps.MerchantId, ps.CategoryId, ps.Name, ps.Description, ps.Color, ps.TotalDuration,
		ps.Price, ps.PriceNote, ps.Cost, ps.IsActive)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
