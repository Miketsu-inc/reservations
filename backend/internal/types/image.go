package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type ImagePurpose struct {
	purpose string
}

func (p ImagePurpose) String() string {
	return p.purpose
}

var (
	ImagePurposeUserAvatar     = ImagePurpose{"user_avatar"}
	ImagePurposeEmployeeAvatar = ImagePurpose{"employee_avatar"}
	ImagePurposeMerchantLogo   = ImagePurpose{"merchant_logo"}
	ImagePurposeLocationImage  = ImagePurpose{"location_image"}
)

func NewImagePurpose(purposeStr string) (ImagePurpose, error) {
	switch strings.ToLower(purposeStr) {
	case "user_avatar":
		return ImagePurposeUserAvatar, nil
	case "employee_avatar":
		return ImagePurposeEmployeeAvatar, nil
	case "merchant_logo":
		return ImagePurposeMerchantLogo, nil
	case "location_image":
		return ImagePurposeLocationImage, nil
	default:
		return ImagePurpose{}, fmt.Errorf("invalid image purpose: %s", purposeStr)
	}
}

func (p ImagePurpose) IsSquare() bool {
	return p != ImagePurposeLocationImage
}

func (p ImagePurpose) Value() (driver.Value, error) {
	return p.purpose, nil
}

func (p *ImagePurpose) Scan(src any) error {
	purposeStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(purposeStr) == 0 {
		return nil
	}

	purpose, err := NewImagePurpose(purposeStr)
	if err != nil {
		return err
	}

	*p = purpose
	return nil
}

func (p ImagePurpose) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.purpose)
}

func (p *ImagePurpose) UnmarshalJSON(data []byte) error {
	var purposeStr string
	if err := json.Unmarshal(data, &purposeStr); err != nil {
		return err
	}

	purpose, err := NewImagePurpose(purposeStr)
	if err != nil {
		return err
	}

	*p = purpose
	return nil
}

type ImageStatus struct {
	status string
}

func (s ImageStatus) String() string {
	return s.status
}

var (
	// Upload url was issued, waiting for the client to upload the file
	ImageStatusPending = ImageStatus{"pending"}
	// Upload is completed, waiting for the processing job
	ImageStatusProcessing = ImageStatus{"processing"}
	ImageStatusReady      = ImageStatus{"ready"}
	ImageStatusFailed     = ImageStatus{"failed"}
)

func NewImageStatus(statusStr string) (ImageStatus, error) {
	switch strings.ToLower(statusStr) {
	case "pending":
		return ImageStatusPending, nil
	case "processing":
		return ImageStatusProcessing, nil
	case "ready":
		return ImageStatusReady, nil
	case "failed":
		return ImageStatusFailed, nil
	default:
		return ImageStatus{}, fmt.Errorf("invalid image status: %s", statusStr)
	}
}

func (s ImageStatus) Value() (driver.Value, error) {
	return s.status, nil
}

func (s *ImageStatus) Scan(src any) error {
	statusStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(statusStr) == 0 {
		return nil
	}

	status, err := NewImageStatus(statusStr)
	if err != nil {
		return err
	}

	*s = status
	return nil
}

func (s ImageStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.status)
}

func (s *ImageStatus) UnmarshalJSON(data []byte) error {
	var statusStr string
	if err := json.Unmarshal(data, &statusStr); err != nil {
		return err
	}

	status, err := NewImageStatus(statusStr)
	if err != nil {
		return err
	}

	*s = status
	return nil
}
