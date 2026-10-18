package args

import (
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/riverqueue/river"
)

// Image processing is memory heavy, so it runs on it's own queue with limited workers
const MediaQueue = "media"

type ProcessImage struct {
	ImageId uuid.UUID         `json:"image_id"`
	Crop    *domain.ImageCrop `json:"crop"`
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

type ImportOauthAvatar struct {
	UserId uuid.UUID `json:"user_id"`
	Url    string    `json:"url"`
}

func (ImportOauthAvatar) Kind() string { return "import_oauth_avatar" }

func (ImportOauthAvatar) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: MediaQueue,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}
