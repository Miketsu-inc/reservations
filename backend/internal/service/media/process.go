package media

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"math"
	"net/http"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"golang.org/x/image/draw"
)

const (
	// 10 MB
	MaxUploadSize = 10 << 20
	// Images above this are rejected before decoding to protect against decompression bombs
	maxImagePixels = 50_000_000
	jpegQuality    = 82
)

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

// The uploaded file is not an image that can be processed, retrying will not help
var errInvalidImage = errors.New("invalid image")

type processedVariant struct {
	variant Variant
	data    []byte
	width   int
	height  int
}

// Decode, orient, crop, resize and re-encode the image into it's variants. Re-encoding strips every metadata (e.g. GPS location)
func processImage(data []byte, purpose types.ImagePurpose, crop *domain.ImageCrop) ([]processedVariant, error) {
	contentType, err := types.NewImageContentType(http.DetectContentType(data))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidImage, err)
	}

	imgConfig, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w: decode config: %w", errInvalidImage, err)
	}

	if imgConfig.Width <= 0 || imgConfig.Height <= 0 || imgConfig.Width*imgConfig.Height > maxImagePixels {
		return nil, fmt.Errorf("%w: unsupported dimensions %dx%d", errInvalidImage, imgConfig.Width, imgConfig.Height)
	}

	orientation := 1
	if contentType == types.ImageContentTypeJPEG {
		orientation = jpegOrientation(data)
	}

	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w: decode: %w", errInvalidImage, err)
	}

	bounds := src.Bounds()
	rawWidth, rawHeight := bounds.Dx(), bounds.Dy()

	displayWidth, displayHeight := rawWidth, rawHeight
	if isTransposed(orientation) {
		displayWidth, displayHeight = rawHeight, rawWidth
	}

	displayRect := sourceRect(displayWidth, displayHeight, purpose, crop)
	// The crop is relative to the displayed image, so it is mapped back to the stored pixels. Orientation
	// is only applied after resizing which is a lot cheaper than rotating the full sized image
	rawRect := toRawRect(displayRect, orientation, rawWidth, rawHeight).Add(bounds.Min)

	smallSize, largeSize := variantSizes(purpose)

	largeWidth, largeHeight := fitWithin(displayRect.Dx(), displayRect.Dy(), largeSize)

	scaledWidth, scaledHeight := largeWidth, largeHeight
	if isTransposed(orientation) {
		scaledWidth, scaledHeight = largeHeight, largeWidth
	}

	scaled := image.NewRGBA(image.Rect(0, 0, scaledWidth, scaledHeight))
	// JPEG has no transparency, so transparent PNG pixels are composited onto white
	draw.Draw(scaled, scaled.Bounds(), image.White, image.Point{}, draw.Src)
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), src, rawRect, draw.Over, nil)

	large := applyOrientation(scaled, orientation)

	smallWidth, smallHeight := fitWithin(largeWidth, largeHeight, smallSize)

	small := large
	if smallWidth != largeWidth || smallHeight != largeHeight {
		small = image.NewRGBA(image.Rect(0, 0, smallWidth, smallHeight))
		draw.CatmullRom.Scale(small, small.Bounds(), large, large.Bounds(), draw.Src, nil)
	}

	variants := make([]processedVariant, 0, 2)
	for _, v := range []struct {
		variant Variant
		img     *image.RGBA
	}{{VariantSmall, small}, {VariantLarge, large}} {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, v.img, &jpeg.Options{Quality: jpegQuality}); err != nil {
			return nil, fmt.Errorf("encode %s variant: %w", v.variant, err)
		}

		variants = append(variants, processedVariant{
			variant: v.variant,
			data:    buf.Bytes(),
			width:   v.img.Bounds().Dx(),
			height:  v.img.Bounds().Dy(),
		})
	}

	return variants, nil
}

// Returns the part of the displayed image that should be kept. Invalid crops fall back to the default
func sourceRect(width int, height int, purpose types.ImagePurpose, crop *domain.ImageCrop) image.Rectangle {
	full := image.Rect(0, 0, width, height)

	rect := full
	if crop != nil {
		cropped := image.Rect(crop.X, crop.Y, crop.X+crop.Width, crop.Y+crop.Height).Intersect(full)
		if !cropped.Empty() {
			rect = cropped
		}
	}

	if purpose.IsSquare() {
		side := min(rect.Dx(), rect.Dy())
		x := rect.Min.X + (rect.Dx()-side)/2
		y := rect.Min.Y + (rect.Dy()-side)/2
		rect = image.Rect(x, y, x+side, y+side)
	}

	return rect
}

// Scales the dimensions so the longest side is at most maxSize, never upscales
func fitWithin(width int, height int, maxSize int) (int, int) {
	if width <= maxSize && height <= maxSize {
		return width, height
	}

	scale := float64(maxSize) / float64(max(width, height))

	return max(1, int(math.Round(float64(width)*scale))), max(1, int(math.Round(float64(height)*scale)))
}

// Orientations 5-8 swap the width and height of the image
func isTransposed(orientation int) bool {
	return orientation >= 5 && orientation <= 8
}

// Maps a rectangle of the displayed image to the stored (raw) image's coordinates
func toRawRect(rect image.Rectangle, orientation int, rawWidth int, rawHeight int) image.Rectangle {
	x0, y0 := toRawPoint(rect.Min.X, rect.Min.Y, orientation, rawWidth, rawHeight)
	x1, y1 := toRawPoint(rect.Max.X, rect.Max.Y, orientation, rawWidth, rawHeight)

	// image.Rect normalizes the corners
	return image.Rect(x0, y0, x1, y1)
}

// Maps a point (pixel edge) of the displayed image to the stored image's coordinates
func toRawPoint(x int, y int, orientation int, w int, h int) (int, int) {
	switch orientation {
	case 2:
		return w - x, y
	case 3:
		return w - x, h - y
	case 4:
		return x, h - y
	case 5:
		return y, x
	case 6:
		return y, h - x
	case 7:
		return w - y, h - x
	case 8:
		return w - y, x
	default:
		return x, y
	}
}

// Rotates and flips the image according to the EXIF orientation, so it is stored the way it should be displayed
func applyOrientation(src *image.RGBA, orientation int) *image.RGBA {
	if orientation < 2 || orientation > 8 {
		return src
	}

	w, h := src.Bounds().Dx(), src.Bounds().Dy()

	dstWidth, dstHeight := w, h
	if isTransposed(orientation) {
		dstWidth, dstHeight = h, w
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))

	for dy := range dstHeight {
		for dx := range dstWidth {
			var rx, ry int

			switch orientation {
			case 2:
				rx, ry = w-1-dx, dy
			case 3:
				rx, ry = w-1-dx, h-1-dy
			case 4:
				rx, ry = dx, h-1-dy
			case 5:
				rx, ry = dy, dx
			case 6:
				rx, ry = dy, h-1-dx
			case 7:
				rx, ry = w-1-dy, h-1-dx
			case 8:
				rx, ry = w-1-dy, dx
			}

			srcOffset := ry*src.Stride + rx*4
			dstOffset := dy*dst.Stride + dx*4
			copy(dst.Pix[dstOffset:dstOffset+4], src.Pix[srcOffset:srcOffset+4])
		}
	}

	return dst
}
