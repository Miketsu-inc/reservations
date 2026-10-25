package workers

import (
	"context"
	"errors"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/service/media"
	"github.com/riverqueue/river"
)

func cancelIfUnprocessable(err error) error {
	if errors.Is(err, media.ErrUnprocessable) {
		return river.JobCancel(err)
	}

	return err
}

type ProcessImage struct {
	river.WorkerDefaults[args.ProcessImage]

	mediaService *media.Service
}

func NewProcessImage(mediaService *media.Service) *ProcessImage {
	return &ProcessImage{mediaService: mediaService}
}

func (w *ProcessImage) Timeout(*river.Job[args.ProcessImage]) time.Duration {
	return 2 * time.Minute
}

func (w *ProcessImage) Work(ctx context.Context, job *river.Job[args.ProcessImage]) error {
	return cancelIfUnprocessable(w.mediaService.ProcessImage(ctx, job.Args.ImageId, job.Args.Crop, job.Args.ContentType))
}

type DeleteImage struct {
	river.WorkerDefaults[args.DeleteImage]

	mediaService *media.Service
}

func NewDeleteImage(mediaService *media.Service) *DeleteImage {
	return &DeleteImage{mediaService: mediaService}
}

func (w *DeleteImage) Work(ctx context.Context, job *river.Job[args.DeleteImage]) error {
	return w.mediaService.DeleteImage(ctx, job.Args.ImageId)
}

type CleanupImages struct {
	river.WorkerDefaults[args.CleanupImages]

	mediaService *media.Service
}

func NewCleanupImages(mediaService *media.Service) *CleanupImages {
	return &CleanupImages{mediaService: mediaService}
}

func (w *CleanupImages) Work(ctx context.Context, job *river.Job[args.CleanupImages]) error {
	return w.mediaService.CleanupImages(ctx)
}
