package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/imgproxy"
)

const (
	// 10 MB
	MaxUploadSize = 10 << 20
	jpegQuality   = 82
)

const (
	contentTypeJPEG = "image/jpeg"
	contentTypePNG  = "image/png"
)

var allowedContentTypes = []string{contentTypeJPEG, contentTypePNG}

type Variant string

const (
	VariantSmall Variant = "sm"
	VariantLarge Variant = "lg"
)

// Returns the maximum size of the longest side of the small and large variant
func variantSizes(purpose types.ImagePurpose) (int, int) {
	if purpose.IsSquare() {
		return 256, 512
	}

	return 640, 1920
}

var errInvalidImage = errors.New("invalid image")

type processedVariant struct {
	variant     Variant
	contentType string
	data        []byte
	width       int
	height      int
}

func (s *Service) processImage(ctx context.Context, image domain.Image, crop *domain.ImageCrop) ([]processedVariant, error) {
	source := fmt.Sprintf("s3://%s/%s", s.uploads.Name(), uploadKey(image.Id))
	smallSize, largeSize := variantSizes(image.Purpose)
	format, contentType := outputFormat(image.Purpose)

	variants := make([]processedVariant, 0, 2)
	for _, v := range []struct {
		variant Variant
		size    int
	}{{VariantSmall, smallSize}, {VariantLarge, largeSize}} {
		data, err := s.imgproxy.Process(ctx, source, processingOptions(image.Purpose, crop, v.size), format)
		if err != nil {
			if errors.Is(err, imgproxy.ErrInvalidSource) {
				return nil, fmt.Errorf("%w: %w", errInvalidImage, err)
			}
			return nil, fmt.Errorf("error processing %s variant: %w", v.variant, err)
		}

		width, height, err := imageSize(data)
		if err != nil {
			return nil, fmt.Errorf("error reading %s variant size: %w", v.variant, err)
		}

		variants = append(variants, processedVariant{
			variant:     v.variant,
			contentType: contentType,
			data:        data,
			width:       width,
			height:      height,
		})
	}

	return variants, nil
}

// Logos often have a transparent background. The rest are photos, where JPEG is a lot smaller
func outputFormat(purpose types.ImagePurpose) (string, string) {
	if purpose == types.ImagePurposeMerchantLogo {
		return "png", contentTypePNG
	}

	return "jpg", contentTypeJPEG
}

// See https://docs.imgproxy.net/usage/processing
func processingOptions(purpose types.ImagePurpose, crop *domain.ImageCrop, size int) []string {
	// the orientation is applied first, the crop is relative to the oriented image just like on the client
	options := []string{"ar:1", "sm:1"}

	if crop != nil && crop.Width > 0 && crop.Height > 0 {
		options = append(options, fmt.Sprintf("c:%d:%d:nowe:%d:%d", crop.Width, crop.Height, max(crop.X, 0), max(crop.Y, 0)))
	}

	if purpose.IsSquare() {
		// crops the center to a square, then downscales it. Smaller images are not enlarged
		options = append(options, fmt.Sprintf("rs:fill-down:%d:%d", size, size))
	} else {
		options = append(options, fmt.Sprintf("rs:fit:%d:%d:0", size, size))
	}

	if purpose == types.ImagePurposeMerchantLogo {
		return options
	}

	// JPEG has no transparency, so transparent PNG pixels become white
	return append(options, "bg:FFFFFF", fmt.Sprintf("q:%d", jpegQuality))
}

func imageSize(data []byte) (int, int, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}
