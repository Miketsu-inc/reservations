package keys

import (
	"fmt"

	"github.com/google/uuid"
)

// Number of image uploads started by the user in the current rate limit window
type ImageUploadRate struct {
	UserId uuid.UUID
}

func (k ImageUploadRate) String() string {
	return fmt.Sprintf("image_upload_rate:%s", k.UserId)
}
