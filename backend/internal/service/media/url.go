package media

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
)

// Object key of a processed image variant in the media bucket. Every upload gets a new image id,
// so the content behind a key never changes and can be cached forever
func objectKey(imageId uuid.UUID, variant Variant) string {
	return fmt.Sprintf("images/%s/%s.jpg", imageId, variant)
}

// Object key of the original upload in the uploads bucket
func uploadKey(imageId uuid.UUID) string {
	return imageId.String()
}

// Public URL of an image variant
func VariantURL(imageId uuid.UUID, variant Variant) string {
	return strings.TrimRight(config.LoadEnvVars().MEDIA_PUBLIC_URL, "/") + "/" + objectKey(imageId, variant)
}

// Public URL of an image variant, nil if there is no image
func URL(imageId *uuid.UUID, variant Variant) *string {
	if imageId == nil {
		return nil
	}

	url := VariantURL(*imageId, variant)
	return &url
}
