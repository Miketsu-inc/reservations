package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/miketsu-inc/reservations/backend/cmd/server"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

func main() {
	// pgx uses the local time for db queries
	// and I did not find a way to configure it to not do so
	time.Local = time.UTC

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	server := server.NewServer()

	err := server.ListenAndServe()
	assert.Nil(err, fmt.Sprintf("cannot start server: %s", err))
}
