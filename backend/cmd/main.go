package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/internal/app"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

func main() {
	// pgx uses the local time for db queries
	// and I did not find a way to configure it to not do so
	time.Local = time.UTC

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	cfg := config.LoadEnvVars()
	cfg.Validate()

	application := app.New(cfg)
	defer application.Stop()

	err := application.Start()
	assert.Nil(err, fmt.Sprintf("cannot start server: %s", err))
}
