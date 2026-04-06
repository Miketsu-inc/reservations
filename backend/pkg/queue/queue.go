package queue

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
)

type Enqueuer interface {
	InsertTx(ctx context.Context, tx pgx.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
	Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
	JobCancel(ctx context.Context, jobID int64) (*rivertype.JobRow, error)
	JobCancelTx(ctx context.Context, tx pgx.Tx, jobID int64) (*rivertype.JobRow, error)
}

type RegisterWorkersFunc[T any] func(workers *river.Workers, deps T)

func NewClient[T any](dbConn *pgxpool.Pool, deps T, registerWorkersFunc RegisterWorkersFunc[T], periodicJobs []*river.PeriodicJob) (*river.Client[pgx.Tx], error) {
	riverWorkers := river.NewWorkers()

	registerWorkersFunc(riverWorkers, deps)

	return river.NewClient(riverpgxv5.New(dbConn), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
			"email":            {MaxWorkers: 100},
		},
		Workers:      riverWorkers,
		PeriodicJobs: periodicJobs,
	})
}
