package middleware

import "github.com/miketsu-inc/reservations/backend/internal/domain"

type Manager struct {
	merchantRepo domain.MerchantRepository
	userRepo     domain.UserRepository
}

// Creates a new middleware manager that holds the middlewares repository dependencies
func NewManager(merchant domain.MerchantRepository, user domain.UserRepository) *Manager {
	return &Manager{
		merchantRepo: merchant,
		userRepo:     user,
	}
}
