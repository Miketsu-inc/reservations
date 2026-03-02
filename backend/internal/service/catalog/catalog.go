package catalog

import (
	"context"
	"fmt"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
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

type NewInput struct {
	Name         string
	Description  *string
	Color        string
	Price        *currencyx.Price
	Cost         *currencyx.Price
	PriceType    types.PriceType
	CategoryId   *int
	IsActive     bool
	Settings     ServiceSettingsInput
	Phases       []NewPhasesInput
	UsedProducts []ConnectedProductsInput
}

type ServiceSettingsInput struct {
	CancelDeadline   *int
	BookingWindowMin *int
	BookingWindowMax *int
	BufferTime       *int
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
	employee := jwt.MustGetEmployeeFromContext(ctx)

	if len(input.Phases) == 0 {
		return fmt.Errorf("service phases can not be empty")
	}

	var phases []domain.ServicePhase
	durationSum := 0
	for _, phase := range input.Phases {
		phases = append(phases, domain.ServicePhase{
			Id:        0,
			ServiceId: 0,
			Name:      phase.Name,
			Sequence:  phase.Sequence,
			Duration:  phase.Duration,
			PhaseType: phase.PhaseType,
		})
		durationSum += phase.Duration
	}

	var connectedProducts []domain.ConnectedProducts
	for _, product := range input.UsedProducts {
		connectedProducts = append(connectedProducts, domain.ConnectedProducts{
			ProductId:  product.ProductId,
			ServiceId:  0,
			AmountUsed: product.AmountUsed,
		})
	}

	curr, err := s.merchantRepo.GetMerchantCurrency(ctx, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error while getting merchant's currency: %s", err.Error())
	}

	if input.Price != nil {
		if input.Price.CurrencyCode() != curr {
			return fmt.Errorf("new service price's currency does not match merchant's currency")
		}
	}

	if input.Cost != nil {
		if input.Cost.CurrencyCode() != curr {
			return fmt.Errorf("new service cost's currency does not match merchant's currency")
		}
	}

	err = s.txManager.WithTransaction(ctx, func(tx db.DBTX) error {
		serviceId, err := s.catalogRepo.WithTx(tx).NewService(ctx, domain.Service{
			Id:              0,
			MerchantId:      employee.MerchantId,
			CategoryId:      input.CategoryId,
			BookingType:     types.BookingTypeAppointment,
			Name:            input.Name,
			Description:     input.Description,
			Color:           input.Color,
			TotalDuration:   durationSum,
			Price:           input.Price,
			Cost:            input.Cost,
			PriceType:       input.PriceType,
			IsActive:        input.IsActive,
			Sequence:        0,
			MinParticipants: 1,
			MaxParticipants: 1,
			ServiceSettings: domain.ServiceSettings{
				CancelDeadline:   input.Settings.CancelDeadline,
				BookingWindowMin: input.Settings.BookingWindowMin,
				BookingWindowMax: input.Settings.BookingWindowMax,
				BufferTime:       input.Settings.BufferTime,
			},
		})
		if err != nil {
			return err
		}

		err = s.catalogRepo.WithTx(tx).NewServicePhases(ctx, serviceId, phases)
		if err != nil {
			return err
		}

		err = s.catalogRepo.WithTx(tx).NewServiceProduct(ctx, employee.MerchantId, connectedProducts)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("unexpected error inserting service: %s", err.Error())
	}

	return nil
}

func servicePhasesEqual(a, b domain.PublicServicePhase) bool {
	return a.Name == b.Name &&
		a.Sequence == b.Sequence &&
		a.Duration == b.Duration &&
		a.PhaseType == b.PhaseType
}

type UpdateInput struct {
	Id          int
	Name        string
	Description *string
	Color       string
	Price       *currencyx.Price
	Cost        *currencyx.Price
	PriceType   types.PriceType
	CategoryId  *int
	IsActive    bool
	Settings    ServiceSettingsInput
	Phases      []PhasesInput
}

type PhasesInput struct {
	Id        int
	ServiceId int
	Name      string
	Sequence  int
	Duration  int
	PhaseType types.ServicePhaseType
}

func (s *Service) Update(ctx context.Context, serviceId int, input UpdateInput) error {
	if serviceId != input.Id {
		return fmt.Errorf("invalid service id provided")
	}

	if len(input.Phases) == 0 {
		return fmt.Errorf("service phases can not be empty")
	}

	employee := jwt.MustGetEmployeeFromContext(ctx)

	var phases []domain.PublicServicePhase
	durationSum := 0
	for _, phase := range input.Phases {
		phases = append(phases, domain.PublicServicePhase{
			Id:        phase.Id,
			ServiceId: phase.ServiceId,
			Name:      phase.Name,
			Sequence:  phase.Sequence,
			Duration:  phase.Duration,
			PhaseType: phase.PhaseType,
		})
		durationSum += phase.Duration
	}

	err := s.txManager.WithTransaction(ctx, func(tx db.DBTX) error {
		existingPhases, err := s.catalogRepo.WithTx(tx).GetServicePhases(ctx, serviceId)
		if err != nil {
			return err
		}

		existingPhasesMap := map[int]domain.PublicServicePhase{}
		for _, p := range existingPhases {
			existingPhasesMap[p.Id] = domain.PublicServicePhase{
				Id:        p.Id,
				ServiceId: p.ServiceId,
				Name:      p.Name,
				Sequence:  p.Sequence,
				Duration:  p.Duration,
				PhaseType: p.PhaseType,
			}
		}

		updatedPhasesMap := map[int]domain.PublicServicePhase{}
		newPhases := []domain.PublicServicePhase{}
		for _, p := range phases {
			if p.Id == 0 {
				newPhases = append(newPhases, p)
			} else {
				updatedPhasesMap[p.Id] = p
			}
		}

		var phaseIdsToDelete []int
		for id := range existingPhasesMap {
			if _, exists := updatedPhasesMap[id]; !exists {
				phaseIdsToDelete = append(phaseIdsToDelete, id)
			}
		}

		err = s.catalogRepo.WithTx(tx).DeleteServicePhases(ctx, phaseIdsToDelete)
		if err != nil {
			return err
		}

		var phasesToUpdate []domain.ServicePhase
		for id, phase := range updatedPhasesMap {
			existingPhase := existingPhasesMap[id]
			if !servicePhasesEqual(existingPhase, phase) {
				phasesToUpdate = append(phasesToUpdate, domain.ServicePhase{
					Id:        id,
					ServiceId: serviceId,
					Name:      phase.Name,
					Sequence:  phase.Sequence,
					Duration:  phase.Duration,
					PhaseType: phase.PhaseType,
				})
			}
		}

		err = s.catalogRepo.WithTx(tx).UpdateServicePhases(ctx, phasesToUpdate)
		if err != nil {
			return err
		}

		var phasesToInsert []domain.ServicePhase
		for _, p := range newPhases {
			phasesToInsert = append(phasesToInsert, domain.ServicePhase{
				ServiceId: serviceId,
				Name:      p.Name,
				Sequence:  p.Sequence,
				Duration:  p.Duration,
				PhaseType: p.PhaseType,
			})
		}

		err = s.catalogRepo.WithTx(tx).NewServicePhases(ctx, serviceId, phasesToInsert)
		if err != nil {
			return err
		}

		oldCategoryId, err := s.catalogRepo.WithTx(tx).UpdateService(ctx, domain.Service{
			Id:              serviceId,
			MerchantId:      employee.MerchantId,
			CategoryId:      input.CategoryId,
			Name:            input.Name,
			Description:     input.Description,
			Color:           input.Color,
			TotalDuration:   durationSum,
			Price:           input.Price,
			Cost:            input.Cost,
			PriceType:       input.PriceType,
			IsActive:        input.IsActive,
			MinParticipants: 1,
			MaxParticipants: 1,
			ServiceSettings: domain.ServiceSettings{
				CancelDeadline:   input.Settings.CancelDeadline,
				BookingWindowMin: input.Settings.BookingWindowMin,
				BookingWindowMax: input.Settings.BookingWindowMax,
				BufferTime:       input.Settings.BufferTime,
			},
		})
		if err != nil {
			return err
		}

		// the categoryId has changed, reordering services is needed
		if (oldCategoryId == nil && input.CategoryId != nil) || (oldCategoryId != nil && (input.CategoryId == nil || *oldCategoryId != *input.CategoryId)) {
			err = s.catalogRepo.WithTx(tx).ReorderServicesAfterUpdate(ctx, oldCategoryId, employee.MerchantId, &serviceId)
			if err != nil {
				return err
			}

			err = s.catalogRepo.WithTx(tx).ReorderServicesAfterUpdate(ctx, input.CategoryId, employee.MerchantId, nil)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error while updating service for merchant: %s", err.Error())
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, serviceId int) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.txManager.WithTransaction(ctx, func(tx db.DBTX) error {
		err := s.catalogRepo.WithTx(tx).DeleteService(ctx, employee.MerchantId, serviceId)
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
	employee := jwt.MustGetEmployeeFromContext(ctx)

	service, err := s.catalogRepo.GetAllServicePageData(ctx, serviceId, employee.MerchantId)
	if err != nil {
		return domain.ServicePageData{}, fmt.Errorf("error while retrieving service for merchant: %s", err.Error())
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

	err := s.txManager.WithTransaction(ctx, func(tx db.DBTX) error {
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
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.catalogRepo.DeactivateService(ctx, employee.MerchantId, serviceId)
	if err != nil {
		return fmt.Errorf("error while activating service: %s", err.Error())
	}

	return nil
}

func (s *Service) Deactivate(ctx context.Context, serviceId int) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.catalogRepo.ActivateService(ctx, employee.MerchantId, serviceId)
	if err != nil {
		return fmt.Errorf("error while deactivating service: %s", err.Error())
	}

	return nil
}

func (s *Service) GetAll(ctx context.Context) ([]domain.ServicesGroupedByCategory, error) {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	services, err := s.catalogRepo.GetServices(ctx, employee.MerchantId)
	if err != nil {
		return []domain.ServicesGroupedByCategory{}, fmt.Errorf("error while retrieving services for merchant: %s", err.Error())
	}

	return services, nil
}

type ReorderInput struct {
	CategoryId *int
	Services   []int
}

func (s *Service) Reorder(ctx context.Context, input ReorderInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.catalogRepo.ReorderServices(ctx, employee.MerchantId, input.CategoryId, input.Services)
	if err != nil {
		return fmt.Errorf("error while ordering services: %s", err.Error())
	}

	return nil
}

func (s *Service) GetFormOptions(ctx context.Context) (domain.ServicePageFormOptions, error) {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	formOptions, err := s.catalogRepo.GetServicePageFormOptions(ctx, employee.MerchantId)
	if err != nil {
		return domain.ServicePageFormOptions{}, fmt.Errorf("error while retrieving service form options for merchant: %s", err.Error())
	}

	return formOptions, nil
}

type NewGroupInput struct {
	Name            string
	Description     *string
	Color           string
	Price           *currencyx.Price
	Cost            *currencyx.Price
	PriceType       types.PriceType
	Duration        int
	CategoryId      *int
	MinParticipants *int
	MaxParticipants int
	IsActive        bool
	Settings        ServiceSettingsInput
	UsedProducts    []ConnectedProductsInput
}

func (s *Service) NewGroup(ctx context.Context, input NewGroupInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	var products []domain.ConnectedProducts
	for _, p := range input.UsedProducts {
		products = append(products, domain.ConnectedProducts{
			ProductId:  p.ProductId,
			ServiceId:  0,
			AmountUsed: p.AmountUsed,
		})
	}

	var phases = []domain.ServicePhase{{
		Id:        0,
		ServiceId: 0,
		Name:      "",
		Sequence:  1,
		Duration:  input.Duration,
		PhaseType: types.ServicePhaseTypeActive,
	}}

	curr, err := s.merchantRepo.GetMerchantCurrency(ctx, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error while getting merchant's currency: %s", err.Error())
	}

	if input.Price != nil {
		if input.Price.CurrencyCode() != curr {
			return fmt.Errorf("new service price's currency does not match merchant's currency")
		}
	}

	if input.Cost != nil {
		if input.Cost.CurrencyCode() != curr {
			return fmt.Errorf("new service cost's currency does not match merchant's currency")
		}
	}

	minParticipants := 1
	if input.MinParticipants != nil {
		minParticipants = *input.MinParticipants
	}

	err = s.txManager.WithTransaction(ctx, func(tx db.DBTX) error {
		serviceId, err := s.catalogRepo.WithTx(tx).NewService(ctx, domain.Service{
			Id:              0,
			MerchantId:      employee.MerchantId,
			CategoryId:      input.CategoryId,
			BookingType:     types.BookingTypeClass,
			Name:            input.Name,
			Description:     input.Description,
			Color:           input.Color,
			TotalDuration:   input.Duration,
			Price:           input.Price,
			Cost:            input.Cost,
			PriceType:       input.PriceType,
			IsActive:        input.IsActive,
			Sequence:        0,
			MinParticipants: minParticipants,
			MaxParticipants: input.MaxParticipants,
			ServiceSettings: domain.ServiceSettings{
				CancelDeadline:   input.Settings.CancelDeadline,
				BookingWindowMin: input.Settings.BookingWindowMin,
				BookingWindowMax: input.Settings.BookingWindowMax,
				BufferTime:       input.Settings.BufferTime,
			},
		})
		if err != nil {
			return err
		}

		err = s.catalogRepo.WithTx(tx).NewServicePhases(ctx, serviceId, phases)
		if err != nil {
			return err
		}

		err = s.catalogRepo.WithTx(tx).NewServiceProduct(ctx, employee.MerchantId, products)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("unexpected error inserting group service: %s", err.Error())
	}

	return nil
}

type UpdateGroupInput struct {
	Id              int
	Name            string
	Description     *string
	Color           string
	Price           *currencyx.Price
	Cost            *currencyx.Price
	PriceType       types.PriceType
	Duration        int
	CategoryId      *int
	MinParticipants *int
	MaxParticipants int
	IsActive        bool
	Settings        ServiceSettingsInput
}

func (s *Service) UpdateGroup(ctx context.Context, serviceId int, input UpdateGroupInput) error {
	if serviceId != input.Id {
		return fmt.Errorf("invalid service id")
	}

	employee := jwt.MustGetEmployeeFromContext(ctx)

	minParticipants := 1
	if input.MinParticipants != nil {
		minParticipants = *input.MinParticipants
	}

	return s.txManager.WithTransaction(ctx, func(tx db.DBTX) error {
		err := s.catalogRepo.WithTx(tx).UpdateServicePhaseDuration(ctx, serviceId, input.Duration)
		if err != nil {
			return err
		}

		oldCategoryId, err := s.catalogRepo.WithTx(tx).UpdateService(ctx, domain.Service{
			Id:              serviceId,
			MerchantId:      employee.MerchantId,
			CategoryId:      input.CategoryId,
			Name:            input.Name,
			Description:     input.Description,
			Color:           input.Color,
			TotalDuration:   input.Duration,
			Price:           input.Price,
			Cost:            input.Cost,
			PriceType:       input.PriceType,
			IsActive:        input.IsActive,
			MinParticipants: minParticipants,
			MaxParticipants: input.MaxParticipants,
			ServiceSettings: domain.ServiceSettings{
				CancelDeadline:   input.Settings.CancelDeadline,
				BookingWindowMin: input.Settings.BookingWindowMin,
				BookingWindowMax: input.Settings.BookingWindowMax,
				BufferTime:       input.Settings.BufferTime,
			},
		})
		if err != nil {
			return err
		}

		// the categoryId has changed, reordering services is needed
		if (oldCategoryId == nil && input.CategoryId != nil) || (oldCategoryId != nil && (input.CategoryId == nil || *oldCategoryId != *input.CategoryId)) {
			err = s.catalogRepo.WithTx(tx).ReorderServicesAfterUpdate(ctx, oldCategoryId, employee.MerchantId, &serviceId)
			if err != nil {
				return err
			}

			err = s.catalogRepo.WithTx(tx).ReorderServicesAfterUpdate(ctx, input.CategoryId, employee.MerchantId, nil)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Service) GetGroup(ctx context.Context, serviceId int) (domain.GroupServicePageData, error) {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	service, err := s.catalogRepo.GetGroupServicePageData(ctx, employee.MerchantId, serviceId)
	if err != nil {
		return domain.GroupServicePageData{}, fmt.Errorf("error while retrieving service for merchant: %s", err.Error())
	}

	return service, nil
}
