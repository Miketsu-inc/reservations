package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type txManager struct {
	pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) TransactionManager {
	return &txManager{pool: pool}
}

func (m *txManager) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	err = fn(tx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
