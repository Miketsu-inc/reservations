package catalog

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type Service struct {
	catalogRepo  domain.CatalogRepository
	merchantRepo domain.MerchantRepository
	txManager    db.TransactionManager
}

func NewService(catalog domain.CatalogRepository, merchant domain.MerchantRepository, txManager db.TransactionManager) *Service {
	return &Service{
		catalogRepo:  catalog,
		merchantRepo: merchant,
		txManager:    txManager,
	}
}

func validateService(phaseCount int, bookingType types.BookingType, maxParticipants *int) error {
	if phaseCount == 0 {
		return fmt.Errorf("service phases can not be empty")
	}

	if bookingType == types.BookingTypeClass || bookingType == types.BookingTypeEvent {
		if phaseCount != 1 {
			return fmt.Errorf("group service shall have one phase")
		}

		if maxParticipants == nil {
			return fmt.Errorf("service must have max participants")
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
			return fmt.Errorf("new service price's currency does not match merchant's currency")
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

		if len(connectedProducts) != 0 {
			err = s.catalogRepo.WithTx(tx).NewServiceProduct(ctx, actor.MerchantId, connectedProducts)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("unexpected error inserting service: %s", err.Error())
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
			return fmt.Errorf("service price's currency does not match merchant's currency")
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
		return fmt.Errorf("error while deleting service for merchant: %s", err.Error())
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

type UpdateServiceProductInput struct {
	ServiceId    int
	UsedProducts []ConnectedProductsInput
}

// TODO: this does not check wether the service and product belong to the merchant updating it
func (s *Service) UpdateServiceProduct(ctx context.Context, serviceId int, input UpdateServiceProductInput) error {
	if serviceId != input.ServiceId {
		return fmt.Errorf("invalid service id")
	}

	if len(input.UsedProducts) == 0 {
		return nil
	}

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
		return fmt.Errorf("error while updating products connected to service for merchant: %s", err.Error())
	}

	return nil
}

// TODO: one query instead of separate activate and deactivate queries
func (s *Service) Activate(ctx context.Context, serviceId int) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.catalogRepo.DeactivateService(ctx, actor.MerchantId, serviceId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Deactivate(ctx context.Context, serviceId int) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.catalogRepo.ActivateService(ctx, actor.MerchantId, serviceId)
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
			return fmt.Errorf("duplicate service id: %d", id)
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
