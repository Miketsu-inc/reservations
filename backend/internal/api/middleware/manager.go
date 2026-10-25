package middleware

import "github.com/miketsu-inc/reservations/backend/internal/domain"

type Manager struct {
	userRepo domain.UserRepository
}

// Creates a new middleware manager that holds the middlewares repository dependencies
func NewManager(user domain.UserRepository) *Manager {
	return &Manager{
		userRepo: user,
	}
}
