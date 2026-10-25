package args

import (
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/riverqueue/river"
)

const MediaQueue = "media"

type ProcessImage struct {
	ImageId     uuid.UUID         `json:"image_id"`
	Crop        *domain.ImageCrop `json:"crop"`
	ContentType string            `json:"content_type"`
}

func (ProcessImage) Kind() string { return "process_image" }

func (ProcessImage) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: MediaQueue,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}

type DeleteImage struct {
	ImageId uuid.UUID `json:"image_id"`
}

func (DeleteImage) Kind() string { return "delete_image" }

func (DeleteImage) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}

type CleanupImages struct{}

func (CleanupImages) Kind() string { return "cleanup_images" }

func (CleanupImages) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByPeriod: time.Hour,
		},
	}
}
