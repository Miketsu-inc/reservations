package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolationCode = "23505"

type Constraint interface {
	Name() string
}

type ConstraintName string

func (c ConstraintName) Name() string {
	return string(c)
}

// IsUniqueConstraintViolation reports whether err was caused by the named
// PostgreSQL unique constraint or index.
func IsUniqueConstraintViolation(err error, constraint Constraint) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == uniqueViolationCode &&
		pgErr.ConstraintName == constraint.Name()
}
