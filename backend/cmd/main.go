package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/miketsu-inc/reservations/backend/cmd/server"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	server := server.NewServer()

	err := server.ListenAndServe()
	assert.Nil(err, fmt.Sprintf("cannot start server: %s", err))
}
