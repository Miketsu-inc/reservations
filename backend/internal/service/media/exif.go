package media

import "encoding/binary"

const (
	exifOrientationTag = 0x0112
	exifTypeShort      = 3
)

// Returns the EXIF orientation (1-8) of a JPEG image or 1 if it is missing or invalid.
// Go's image decoders ignore the orientation so photos taken in portrait mode would appear rotated.
func jpegOrientation(data []byte) int {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return 1
	}

	pos := 2
	for pos+4 <= len(data) {
		if data[pos] != 0xFF {
			return 1
		}

		marker := data[pos+1]
		// fill bytes
		if marker == 0xFF {
			pos++
			continue
		}

		// start of scan or end of image, metadata segments can not follow
		if marker == 0xDA || marker == 0xD9 {
			return 1
		}

		// standalone markers without a length
		if (marker >= 0xD0 && marker <= 0xD7) || marker == 0x01 {
			pos += 2
			continue
		}

		length := int(binary.BigEndian.Uint16(data[pos+2 : pos+4]))
		if length < 2 || pos+2+length > len(data) {
			return 1
		}

		segment := data[pos+4 : pos+2+length]
		if marker == 0xE1 && len(segment) >= 6 && string(segment[:6]) == "Exif\x00\x00" {
			return tiffOrientation(segment[6:])
		}

		pos += 2 + length
	}

	return 1
}

func tiffOrientation(tiff []byte) int {
	if len(tiff) < 8 {
		return 1
	}

	var order binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 1
	}

	if order.Uint16(tiff[2:4]) != 42 {
		return 1
	}

	ifdOffset := int(order.Uint32(tiff[4:8]))
	if ifdOffset < 8 || ifdOffset+2 > len(tiff) {
		return 1
	}

	entryCount := int(order.Uint16(tiff[ifdOffset : ifdOffset+2]))
	for i := range entryCount {
		entry := ifdOffset + 2 + i*12
		if entry+12 > len(tiff) {
			return 1
		}

		if order.Uint16(tiff[entry:entry+2]) != exifOrientationTag {
			continue
		}

		if order.Uint16(tiff[entry+2:entry+4]) != exifTypeShort {
			return 1
		}

		orientation := int(order.Uint16(tiff[entry+8 : entry+10]))
		if orientation < 1 || orientation > 8 {
			return 1
		}

		return orientation
	}

	return 1
}
