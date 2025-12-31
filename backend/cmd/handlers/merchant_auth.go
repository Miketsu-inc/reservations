package handlers

import (
	"github.com/miketsu-inc/reservations/backend/cmd/database"
)

type MerchantAuth struct {
	Postgresdb database.PostgreSQL
}
