package media

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
)

func objectKey(imageId uuid.UUID, variant Variant) string {
	return fmt.Sprintf("images/%s/%s", imageId, variant)
}

func NewVariantURL(imageId uuid.UUID, variant Variant) string {
	return strings.TrimRight(config.LoadEnvVars().MEDIA_PUBLIC_URL, "/") + "/" + objectKey(imageId, variant)
}

func NewURL(imageId *uuid.UUID, variant Variant) *string {
	if imageId == nil {
		return nil
	}

	url := NewVariantURL(*imageId, variant)
	return &url
}
