package catalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/bojanz/currency"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/service/team"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type Service struct {
	catalogRepo  domain.CatalogRepository
	merchantRepo domain.MerchantRepository
	teamService  *team.Service
	txManager    db.TransactionManager
}

func NewService(catalog domain.CatalogRepository, merchant domain.MerchantRepository, teamService *team.Service, txManager db.TransactionManager) *Service {
	return &Service{
		catalogRepo:  catalog,
		merchantRepo: merchant,
		teamService:  teamService,
		txManager:    txManager,
	}
}

func validateService(phaseCount int, bookingType types.BookingType, maxParticipants *int) error {
	if phaseCount == 0 {
		return ErrServicePhasesRequired
	}

	if bookingType == types.BookingTypeClass || bookingType == types.BookingTypeEvent {
		if phaseCount != 1 {
			return ErrGroupServiceRequiresSinglePhase
		}

		if maxParticipants == nil {
			return ErrGroupServiceRequiresMaxParticipants
		}
	}

	return nil
}

type NewInput struct {
	BookingType     types.BookingType
	Name            string
	Description     *string
	Color           string
	Price           *currencyx.Price
	PriceType       types.PriceType
	CategoryId      *int
	MinParticipants *int
	MaxParticipants *int
	IsActive        bool
	EmployeeIds     []int
	Settings        ServiceSettingsInput
	Phases          []NewPhasesInput
	UsedProducts    []ConnectedProductsInput
}

type ServiceSettingsInput struct {
	CancelDeadline   *int
	BookingWindowMin *int
	BookingWindowMax *int
	BufferTime       *int
	ApprovalPolicy   *types.ApprovalType
}

type NewPhasesInput struct {
	Name      string
	Sequence  int
	Duration  int
	PhaseType types.ServicePhaseType
}

type ConnectedProductsInput struct {
	ProductId  int
	AmountUsed int
}

func (s *Service) New(ctx context.Context, input NewInput) error {
	actor := actor.MustGetFromContext(ctx)

	if err := validateService(len(input.Phases), input.BookingType, input.MaxParticipants); err != nil {
		return err
	}

	if err := s.teamService.IsInActiveEmployees(ctx, actor.MerchantId, input.EmployeeIds); err != nil {
		return err
	}

	minParticipants := 1
	if input.MinParticipants != nil {
		minParticipants = *input.MinParticipants
	}

	maxParticipants := 1
	if input.MaxParticipants != nil {
		maxParticipants = *input.MaxParticipants
	}

	var phases []domain.ServicePhase

	totalDuration := 0
	for _, phase := range input.Phases {
		phases = append(phases, domain.ServicePhase{
			ServiceId: 0,
			Name:      phase.Name,
			Sequence:  phase.Sequence,
			Duration:  phase.Duration,
			PhaseType: phase.PhaseType,
		})

		totalDuration += phase.Duration
	}

	var connectedProducts []domain.ConnectedProducts
	for _, product := range input.UsedProducts {
		connectedProducts = append(connectedProducts, domain.ConnectedProducts{
			ProductId:  product.ProductId,
			ServiceId:  0,
			AmountUsed: product.AmountUsed,
		})
	}

	curr, err := s.merchantRepo.GetMerchantCurrency(ctx, actor.MerchantId)
	if err != nil {
		return err
	}

	if input.Price != nil {
		if input.Price.CurrencyCode() != curr {
			return ErrServicePriceCurrencyMismatch
		}
	}

	err = s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		serviceId, err := s.catalogRepo.WithTx(tx).NewService(ctx, domain.Service{
			Id:            0,
			MerchantId:    actor.MerchantId,
			CategoryId:    input.CategoryId,
			BookingType:   input.BookingType,
			Name:          input.Name,
			Description:   input.Description,
			Color:         input.Color,
			TotalDuration: totalDuration,
			Price:         input.Price,
			PriceType:     input.PriceType,
			IsActive:      input.IsActive,
			// sequence get's calculated in the query
			Sequence:         0,
			MinParticipants:  minParticipants,
			MaxParticipants:  maxParticipants,
			CancelDeadline:   input.Settings.CancelDeadline,
			BookingWindowMin: input.Settings.BookingWindowMin,
			BookingWindowMax: input.Settings.BookingWindowMax,
			BufferTime:       input.Settings.BufferTime,
			ApprovalPolicy:   input.Settings.ApprovalPolicy,
		})
		if err != nil {
			return err
		}

		err = s.catalogRepo.WithTx(tx).NewServicePhases(ctx, serviceId, phases)
		if err != nil {
			return err
		}

		employeeServices := make([]domain.EmployeeService, len(input.EmployeeIds))
		for i, e := range input.EmployeeIds {
			employeeServices[i] = domain.EmployeeService{
				EmployeeId: e,
				ServiceId:  serviceId,
			}
		}

		err = s.catalogRepo.WithTx(tx).BulkInsertEmployeeService(ctx, employeeServices)
		if err != nil {
			return err
		}

		if len(connectedProducts) != 0 {
			err = s.catalogRepo.WithTx(tx).NewServiceProduct(ctx, actor.MerchantId, connectedProducts)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("unexpected error inserting service: %w", err)
	}

	return nil
}

type phaseChanges struct {
	ToInsert []domain.ServicePhase
	ToUpdate []domain.ServicePhase
	ToDelete []int
}

func detectPhaseChanges(existingPhases []domain.ServicePhase, incomingPhases []domain.ServicePhase) phaseChanges {
	pc := phaseChanges{
		ToInsert: []domain.ServicePhase{},
		ToUpdate: []domain.ServicePhase{},
		ToDelete: []int{},
	}

	existingMap := map[int]domain.ServicePhase{}
	for _, p := range existingPhases {
		existingMap[p.Id] = domain.ServicePhase{
			Id:        p.Id,
			ServiceId: p.ServiceId,
			Name:      p.Name,
			Sequence:  p.Sequence,
			Duration:  p.Duration,
			PhaseType: p.PhaseType,
		}
	}

	incomingNotNewMap := map[int]domain.ServicePhase{}
	serviceId := existingPhases[0].ServiceId

	for _, p := range incomingPhases {
		if p.Id == -1 {
			pc.ToInsert = append(pc.ToInsert, domain.ServicePhase{
				ServiceId: serviceId,
				Name:      p.Name,
				Sequence:  p.Sequence,
				Duration:  p.Duration,
				PhaseType: p.PhaseType,
			})
		} else {
			existingPhase := existingMap[p.Id]

			if !existingPhase.IsEqual(p) {
				pc.ToUpdate = append(pc.ToUpdate, domain.ServicePhase{
					Id:        p.Id,
					ServiceId: existingPhase.ServiceId,
					Name:      p.Name,
					Sequence:  p.Sequence,
					Duration:  p.Duration,
					PhaseType: p.PhaseType,
				})
			}

			incomingNotNewMap[p.Id] = p
		}
	}

	for id := range existingMap {
		if _, exists := incomingNotNewMap[id]; !exists {
			pc.ToDelete = append(pc.ToDelete, id)
		}
	}

	return pc
}

func buildEmployeeTotalDurationUpdate(serviceId int, phases []domain.ServicePhase, phaseOverrides []domain.EmployeeServicePhase) []domain.EmployeeService {
	overridesByEmployee := make(map[int]map[int]int)
	employeeIds := make([]int, 0)

	for _, employeePhase := range phaseOverrides {
		if overridesByEmployee[employeePhase.EmployeeId] == nil {
			overridesByEmployee[employeePhase.EmployeeId] = make(map[int]int)
			employeeIds = append(employeeIds, employeePhase.EmployeeId)
		}

		overridesByEmployee[employeePhase.EmployeeId][employeePhase.ServicePhaseId] = employeePhase.Duration
	}

	updates := make([]domain.EmployeeService, 0, len(employeeIds))
	for _, employeeId := range employeeIds {
		updates = append(updates, domain.EmployeeService{
			EmployeeId:    employeeId,
			ServiceId:     serviceId,
			TotalDuration: calculateTotalDurationWithOverrides(phases, overridesByEmployee[employeeId]),
		})
	}

	return updates
}

func calculateTotalDurationWithOverrides(phases []domain.ServicePhase, overrides map[int]int) *int {
	total := 0
	hasOverride := false

	for _, phase := range phases {
		duration, isOverridden := overrides[phase.Id]
		if !isOverridden {
			duration = phase.Duration
		}

		hasOverride = hasOverride || isOverridden
		total += duration
	}

	if !hasOverride {
		return nil
	}

	return &total
}

type UpdateInput struct {
	Id              int
	BookingType     types.BookingType
	Name            string
	Description     *string
	Color           string
	Price           *currencyx.Price
	PriceType       types.PriceType
	CategoryId      *int
	MinParticipants *int
	MaxParticipants *int
	IsActive        bool
	EmployeeIds     []int
	Settings        ServiceSettingsInput
	Phases          []PhasesInput
}

type PhasesInput struct {
	Id        int
	Name      string
	Sequence  int
	Duration  int
	PhaseType types.ServicePhaseType
}

func (s *Service) Update(ctx context.Context, input UpdateInput) error {
	actor := actor.MustGetFromContext(ctx)

	if err := validateService(len(input.Phases), input.BookingType, input.MaxParticipants); err != nil {
		return err
	}

	if err := s.teamService.IsInActiveEmployees(ctx, actor.MerchantId, input.EmployeeIds); err != nil {
		return err
	}

	minParticipants := 1
	if input.MinParticipants != nil {
		minParticipants = *input.MinParticipants
	}

	maxParticipants := 1
	if input.MaxParticipants != nil {
		maxParticipants = *input.MaxParticipants
	}

	var phases []domain.ServicePhase
	totalDuration := 0

	for _, phase := range input.Phases {
		phases = append(phases, domain.ServicePhase{
			Id:        phase.Id,
			ServiceId: input.Id,
			Name:      phase.Name,
			Sequence:  phase.Sequence,
			Duration:  phase.Duration,
			PhaseType: phase.PhaseType,
		})

		totalDuration += phase.Duration
	}

	curr, err := s.merchantRepo.GetMerchantCurrency(ctx, actor.MerchantId)
	if err != nil {
		return err
	}

	if input.Price != nil {
		if input.Price.CurrencyCode() != curr {
			return ErrServicePriceCurrencyMismatch
		}
	}

	err = s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		existingPhases, err := s.catalogRepo.WithTx(tx).GetServicePhases(ctx, input.Id)
		if err != nil {
			return err
		}

		phaseChanges := detectPhaseChanges(existingPhases, phases)

		if len(phaseChanges.ToDelete) > 0 {
			err = s.catalogRepo.WithTx(tx).DeleteServicePhases(ctx, phaseChanges.ToDelete)
			if err != nil {
				return err
			}
		}

		if len(phaseChanges.ToUpdate) > 0 {
			err = s.catalogRepo.WithTx(tx).UpdateServicePhases(ctx, phaseChanges.ToUpdate)
			if err != nil {
				return err
			}
		}

		if len(phaseChanges.ToInsert) > 0 {
			err = s.catalogRepo.WithTx(tx).NewServicePhases(ctx, input.Id, phaseChanges.ToInsert)
			if err != nil {
				return err
			}
		}

		phaseOverrides, err := s.catalogRepo.WithTx(tx).GetEmployeeServicePhaseOverrides(ctx, actor.MerchantId, input.Id)
		if err != nil {
			return err
		}

		// when a non overriden phase changes, we have to recalculate the total duration on the EmployeeService
		totalDurationUpdates := buildEmployeeTotalDurationUpdate(input.Id, phases, phaseOverrides)
		if len(totalDurationUpdates) > 0 {
			if err := s.catalogRepo.WithTx(tx).BulkUpdateEmployeeServiceDurations(ctx, totalDurationUpdates); err != nil {
				return err
			}
		}

		oldCategoryId, err := s.catalogRepo.WithTx(tx).UpdateService(ctx, domain.Service{
			Id:            input.Id,
			MerchantId:    actor.MerchantId,
			CategoryId:    input.CategoryId,
			Name:          input.Name,
			Description:   input.Description,
			Color:         input.Color,
			TotalDuration: totalDuration,
			Price:         input.Price,
			PriceType:     input.PriceType,
			IsActive:      input.IsActive,
			// sequence get's calculated in the query
			Sequence:         0,
			MinParticipants:  minParticipants,
			MaxParticipants:  maxParticipants,
			CancelDeadline:   input.Settings.CancelDeadline,
			BookingWindowMin: input.Settings.BookingWindowMin,
			BookingWindowMax: input.Settings.BookingWindowMax,
			BufferTime:       input.Settings.BufferTime,
			ApprovalPolicy:   input.Settings.ApprovalPolicy,
		})
		if err != nil {
			return err
		}

		existingEmployees, err := s.catalogRepo.WithTx(tx).GetEmployeeIdsForService(ctx, input.Id)
		if err != nil {
			return err
		}

		employeeChanges, err := s.teamService.DetectEmployeeChanges(existingEmployees, input.EmployeeIds)
		if err != nil {
			return err
		}

		if len(employeeChanges.ToDelete) > 0 {
			err = s.catalogRepo.WithTx(tx).BulkDeleteEmployeeService(ctx, input.Id, employeeChanges.ToDelete)
			if err != nil {
				return err
			}
		}

		if len(employeeChanges.ToInsert) > 0 {
			employeeServices := make([]domain.EmployeeService, len(employeeChanges.ToInsert))
			for i, e := range employeeChanges.ToInsert {
				employeeServices[i] = domain.EmployeeService{
					EmployeeId: e,
					ServiceId:  input.Id,
				}
			}

			err = s.catalogRepo.WithTx(tx).BulkInsertEmployeeService(ctx, employeeServices)
			if err != nil {
				return err
			}
		}

		// the categoryId has changed, reordering services is needed
		if (oldCategoryId == nil && input.CategoryId != nil) || (oldCategoryId != nil && (input.CategoryId == nil || *oldCategoryId != *input.CategoryId)) {
			err = s.catalogRepo.WithTx(tx).ReorderServicesAfterUpdate(ctx, oldCategoryId, actor.MerchantId, &input.Id)
			if err != nil {
				return err
			}

			err = s.catalogRepo.WithTx(tx).ReorderServicesAfterUpdate(ctx, input.CategoryId, actor.MerchantId, nil)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error while updating service for merchant: %w", err)
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, serviceId int) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.catalogRepo.WithTx(tx).DeleteService(ctx, actor.MerchantId, serviceId)
		if err != nil {
			return err
		}

		err = s.catalogRepo.WithTx(tx).DeleteServicePhasesForService(ctx, serviceId)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error while deleting service for merchant: %w", err)
	}

	return nil
}

func (s *Service) Get(ctx context.Context, serviceId int) (domain.ServicePageData, error) {
	actor := actor.MustGetFromContext(ctx)

	service, err := s.catalogRepo.GetAllServicePageData(ctx, serviceId, actor.MerchantId)
	if err != nil {
		return domain.ServicePageData{}, err
	}

	return service, nil
}

type GetTeamMemberSettingsResult struct {
	ServiceId          int
	ServiceName        string
	CurrencyCode       string
	BookingType        types.BookingType
	TotalDuration      int
	PricePerPerson     *currencyx.Price
	PriceType          types.PriceType
	MinParticipants    int
	MaxParticipants    int
	BufferTime         *int
	Phases             []domain.ServicePhase
	TeamMemberSettings []domain.EmployeeServiceSettings
}

func (s *Service) GetTeamMemberSettings(ctx context.Context, serviceId int) (GetTeamMemberSettingsResult, error) {
	actor := actor.MustGetFromContext(ctx)

	service, err := s.catalogRepo.GetServiceWithPhases(ctx, serviceId, actor.MerchantId)
	if err != nil {
		return GetTeamMemberSettingsResult{}, err
	}

	var currencyCode string

	if service.Price != nil {
		currencyCode = service.Price.CurrencyCode()
	} else {
		curr, err := s.merchantRepo.GetMerchantCurrency(ctx, actor.MerchantId)
		if err != nil {
			return GetTeamMemberSettingsResult{}, err
		}

		currencyCode = curr
	}

	settings, err := s.catalogRepo.GetEmployeeServiceSettings(ctx, actor.MerchantId, serviceId)
	if err != nil {
		return GetTeamMemberSettingsResult{}, err
	}

	phaseOverrides, err := s.catalogRepo.GetEmployeeServicePhaseOverrides(ctx, actor.MerchantId, serviceId)
	if err != nil {
		return GetTeamMemberSettingsResult{}, err
	}

	settingIndexByEmployee := make(map[int]int, len(settings))
	for i := range settings {
		settingIndexByEmployee[settings[i].EmployeeId] = i
	}

	for _, phaseOverride := range phaseOverrides {
		if i, exists := settingIndexByEmployee[phaseOverride.EmployeeId]; exists {
			settings[i].PhaseOverrides = append(settings[i].PhaseOverrides, phaseOverride)
		}
	}

	return GetTeamMemberSettingsResult{
		ServiceId:          service.Id,
		ServiceName:        service.Name,
		CurrencyCode:       currencyCode,
		BookingType:        service.BookingType,
		TotalDuration:      service.TotalDuration,
		PricePerPerson:     service.Price,
		PriceType:          service.PriceType,
		MinParticipants:    service.MinParticipants,
		MaxParticipants:    service.MaxParticipants,
		BufferTime:         service.BufferTime,
		Phases:             service.Phases,
		TeamMemberSettings: settings,
	}, nil
}

type EmployeeServicePhaseInput struct {
	ServicePhaseId int
	Duration       int
}

type UpdateTeamMemberSettingsInput struct {
	EmployeeId      int
	IsAssigned      bool
	PricePerPerson  *currencyx.Price
	PriceType       *types.PriceType
	MinParticipants *int
	MaxParticipants *int
	BufferTime      *int
	PhaseOverrides  []EmployeeServicePhaseInput
}

func validateTeamMemberSettingsInputs(inputs []UpdateTeamMemberSettingsInput, currencyCode string) error {
	maxPrice, err := currency.NewAmount("1000000", currencyCode)
	if err != nil {
		return err
	}

	seenEmployeeIds := make(map[int]struct{}, len(inputs))

	for _, input := range inputs {
		if _, exists := seenEmployeeIds[input.EmployeeId]; exists {
			return domain.ErrDuplicateEmployee
		}

		seenEmployeeIds[input.EmployeeId] = struct{}{}

		if !input.IsAssigned {
			continue
		}

		if price := input.PricePerPerson; price != nil {
			if price.CurrencyCode() != currencyCode {
				return ErrEmployeeServicePriceCurrencyMismatch
			}

			priceComparison, err := price.Cmp(maxPrice)
			if err != nil {
				return err
			}

			if price.IsNegative() || priceComparison > 0 {
				return ErrEmployeeServicePriceOutOfRange
			}
		}

		for _, participants := range []*int{input.MinParticipants, input.MaxParticipants} {
			if participants != nil && *participants < 1 {
				return ErrEmployeeServiceParticipantsOutOfRange
			}
		}

		if input.BufferTime != nil && (*input.BufferTime < 0 || *input.BufferTime > 1440) {
			return ErrEmployeeServiceBufferTimeOutOfRange
		}
	}

	return nil
}

type teamMemberSettingsChanges struct {
	employeeServices           []domain.EmployeeService
	employeeIdsToDelete        []int
	employeeIdsToReplacePhases []int
	phaseOverrides             []domain.EmployeeServicePhase
}

func buildTeamMemberSettingsChanges(service domain.Service, settings []domain.EmployeeServiceSettings,
	inputs []UpdateTeamMemberSettingsInput) (teamMemberSettingsChanges, error) {

	assignmentByEmployeeId := make(map[int]bool, len(settings))
	assignedEmployeeCount := 0
	for _, employee := range settings {
		assignmentByEmployeeId[employee.EmployeeId] = employee.IsAssigned

		if employee.IsAssigned {
			assignedEmployeeCount++
		}
	}

	servicePhaseIds := make(map[int]struct{}, len(service.Phases))
	for _, phase := range service.Phases {
		servicePhaseIds[phase.Id] = struct{}{}
	}

	changes := teamMemberSettingsChanges{
		employeeServices:           make([]domain.EmployeeService, 0, len(inputs)),
		employeeIdsToDelete:        make([]int, 0, len(inputs)),
		employeeIdsToReplacePhases: make([]int, 0, len(inputs)),
	}

	for _, input := range inputs {
		// the queried employee service settings contains all active employees
		// therefore if not in it, it is inactive
		wasAssigned, isActiveEmployee := assignmentByEmployeeId[input.EmployeeId]
		if !isActiveEmployee {
			return teamMemberSettingsChanges{}, domain.ErrEmployeeNotActive
		}

		if input.IsAssigned && !wasAssigned {
			assignedEmployeeCount++
		} else if !input.IsAssigned && wasAssigned {
			assignedEmployeeCount--
		}

		if !input.IsAssigned {
			changes.employeeIdsToDelete = append(changes.employeeIdsToDelete, input.EmployeeId)
			continue
		}

		changes.employeeIdsToReplacePhases = append(changes.employeeIdsToReplacePhases, input.EmployeeId)

		phaseOverrides := make(map[int]int, len(input.PhaseOverrides))
		for _, phase := range input.PhaseOverrides {
			if _, duplicate := phaseOverrides[phase.ServicePhaseId]; duplicate {
				return teamMemberSettingsChanges{}, ErrDuplicateEmployeeServicePhase
			}

			if _, exists := servicePhaseIds[phase.ServicePhaseId]; !exists {
				return teamMemberSettingsChanges{}, ErrInvalidEmployeeServicePhase
			}

			if phase.Duration < 1 || phase.Duration > 1440 {
				return teamMemberSettingsChanges{}, ErrEmployeeServiceDurationOutOfRange
			}

			phaseOverrides[phase.ServicePhaseId] = phase.Duration

			changes.phaseOverrides = append(changes.phaseOverrides, domain.EmployeeServicePhase{
				EmployeeId:     input.EmployeeId,
				ServiceId:      service.Id,
				ServicePhaseId: phase.ServicePhaseId,
				Duration:       phase.Duration,
			})
		}

		minParticipants := input.MinParticipants
		maxParticipants := input.MaxParticipants
		if !service.IsGroupService() {
			minParticipants = nil
			maxParticipants = nil
		} else {
			effectiveMinParticipants := service.MinParticipants
			if minParticipants != nil {
				effectiveMinParticipants = *minParticipants
			}

			effectiveMaxParticipants := service.MaxParticipants
			if maxParticipants != nil {
				effectiveMaxParticipants = *maxParticipants
			}

			if effectiveMinParticipants > effectiveMaxParticipants {
				return teamMemberSettingsChanges{}, ErrEmployeeServiceParticipantRangeInvalid
			}
		}

		changes.employeeServices = append(changes.employeeServices, domain.EmployeeService{
			EmployeeId:      input.EmployeeId,
			ServiceId:       service.Id,
			TotalDuration:   calculateTotalDurationWithOverrides(service.Phases, phaseOverrides),
			PricePerPerson:  input.PricePerPerson,
			PriceType:       input.PriceType,
			MinParticipants: minParticipants,
			MaxParticipants: maxParticipants,
			BufferTime:      input.BufferTime,
		})
	}

	if assignedEmployeeCount == 0 {
		return teamMemberSettingsChanges{}, domain.ErrEmployeesRequired
	}

	return changes, nil
}

func (s *Service) UpdateTeamMemberSettings(ctx context.Context, serviceId int, inputs []UpdateTeamMemberSettingsInput) error {
	actor := actor.MustGetFromContext(ctx)

	currencyCode, err := s.merchantRepo.GetMerchantCurrency(ctx, actor.MerchantId)
	if err != nil {
		return err
	}

	if err := validateTeamMemberSettingsInputs(inputs, currencyCode); err != nil {
		return err
	}

	err = s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		service, err := s.catalogRepo.WithTx(tx).GetServiceWithPhases(ctx, serviceId, actor.MerchantId)
		if err != nil {
			return err
		}

		settings, err := s.catalogRepo.WithTx(tx).GetEmployeeServiceSettings(ctx, actor.MerchantId, serviceId)
		if err != nil {
			return err
		}

		changes, err := buildTeamMemberSettingsChanges(service, settings, inputs)
		if err != nil {
			return err
		}

		if len(changes.employeeIdsToDelete) > 0 {
			if err := s.catalogRepo.WithTx(tx).BulkDeleteEmployeeService(ctx, serviceId, changes.employeeIdsToDelete); err != nil {
				return err
			}
		}

		if len(changes.employeeServices) > 0 {
			if err := s.catalogRepo.WithTx(tx).BulkUpsertEmployeeServiceSettings(ctx, changes.employeeServices); err != nil {
				return err
			}
		}

		if len(changes.employeeIdsToReplacePhases) > 0 {
			if err := s.catalogRepo.WithTx(tx).BulkDeleteEmployeeServicePhases(ctx, serviceId, changes.employeeIdsToReplacePhases); err != nil {
				return err
			}
		}

		if len(changes.phaseOverrides) > 0 {
			if err := s.catalogRepo.WithTx(tx).BulkInsertEmployeeServicePhases(ctx, changes.phaseOverrides); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error while updating team member settings: %w", err)
	}

	return nil
}

type UpdateServiceProductInput struct {
	UsedProducts []ConnectedProductsInput
}

// TODO: this does not check wether the service and product belong to the merchant updating it
func (s *Service) UpdateServiceProduct(ctx context.Context, serviceId int, input UpdateServiceProductInput) error {
	var products []domain.ConnectedProducts
	for _, product := range input.UsedProducts {
		products = append(products, domain.ConnectedProducts{
			ProductId:  product.ProductId,
			ServiceId:  serviceId,
			AmountUsed: product.AmountUsed,
		})
	}

	err := s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		connProducts, err := s.catalogRepo.WithTx(tx).GetServiceProducts(ctx, serviceId)
		if err != nil {
			return err
		}

		existing := map[int]domain.ConnectedProducts{}
		for _, p := range connProducts {
			existing[p.ProductId] = p
		}

		updated := map[int]domain.ConnectedProducts{}
		for _, p := range products {
			updated[p.ProductId] = p
		}

		var productIds []int
		for productId := range existing {
			if _, exists := updated[productId]; !exists {
				productIds = append(productIds, productId)
			}
		}

		if len(productIds) > 0 {
			err = s.catalogRepo.WithTx(tx).DeleteServiceProducts(ctx, serviceId, productIds)
			if err != nil {
				return err
			}
		}

		err = s.catalogRepo.WithTx(tx).UpdateServiceProducts(ctx, serviceId, products)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error while updating products connected to service for merchant: %w", err)
	}

	return nil
}

func (s *Service) Activate(ctx context.Context, serviceId int) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.catalogRepo.ActivateService(ctx, actor.MerchantId, serviceId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Deactivate(ctx context.Context, serviceId int) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.catalogRepo.DeactivateService(ctx, actor.MerchantId, serviceId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetAll(ctx context.Context) ([]domain.ServicesGroupedByCategory, error) {
	actor := actor.MustGetFromContext(ctx)

	services, err := s.catalogRepo.GetServicesGroupedByCategory(ctx, actor.MerchantId)
	if err != nil {
		return []domain.ServicesGroupedByCategory{}, err
	}

	return services, nil
}

type ReorderInput struct {
	CategoryId *int
	Services   []int
}

func (s *Service) Reorder(ctx context.Context, input ReorderInput) error {
	actor := actor.MustGetFromContext(ctx)

	idSet := make(map[int]struct{}, len(input.Services))
	for _, id := range input.Services {
		if _, ok := idSet[id]; ok {
			return ErrDuplicateServiceId
		}

		idSet[id] = struct{}{}
	}

	err := s.catalogRepo.ReorderServices(ctx, actor.MerchantId, input.CategoryId, input.Services)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetFormOptions(ctx context.Context) (domain.ServicePageFormOptions, error) {
	actor := actor.MustGetFromContext(ctx)

	formOptions, err := s.catalogRepo.GetServicePageFormOptions(ctx, actor.MerchantId)
	if err != nil {
		return domain.ServicePageFormOptions{}, err
	}

	return formOptions, nil
}

func (s *Service) GetServicesGroupedByCategories(ctx context.Context, merchantName string) ([]domain.MerchantPageServicesGroupedByCategory, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return []domain.MerchantPageServicesGroupedByCategory{}, err
	}

	services, err := s.catalogRepo.GetServicesForMerchantPage(ctx, merchantId)
	if err != nil {
		return []domain.MerchantPageServicesGroupedByCategory{}, err
	}

	return services, nil
}

func (s *Service) GetServiceDetails(ctx context.Context, merchantName string, serviceId, locationId int) (domain.PublicServiceDetails, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return domain.PublicServiceDetails{}, err
	}

	serviceDetails, err := s.catalogRepo.GetServiceDetailsForMerchantPage(ctx, merchantId, serviceId, locationId)
	if err != nil {
		return domain.PublicServiceDetails{}, err
	}

	return serviceDetails, nil
}

func (s *Service) GetServicesForCalendar(ctx context.Context) ([]domain.ServicesGroupedByCategoriesForCalendar, error) {
	actor := actor.MustGetFromContext(ctx)

	services, err := s.catalogRepo.GetServicesForCalendar(ctx, actor.MerchantId)
	if err != nil {
		return []domain.ServicesGroupedByCategoriesForCalendar{}, err
	}

	return services, nil
}
