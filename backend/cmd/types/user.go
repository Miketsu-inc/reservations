package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type AuthProviderType struct {
	aptype string
}

func (a AuthProviderType) String() string {
	return a.aptype
}

var (
	AuthProviderTypeGoogle   = AuthProviderType{"google"}
	AuthProviderTypeFacebook = AuthProviderType{"facebook"}
)

func NewAuthProviderType(typeStr string) (AuthProviderType, error) {
	switch strings.ToLower(typeStr) {
	case "google":
		return AuthProviderTypeGoogle, nil
	case "facebook":
		return AuthProviderTypeFacebook, nil
	default:
		return AuthProviderType{}, fmt.Errorf("invalid AuthProviderType: %s", typeStr)
	}
}

func (a AuthProviderType) Value() (driver.Value, error) {
	return a.aptype, nil
}

func (a *AuthProviderType) Scan(src any) error {
	typeStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(typeStr) == 0 {
		return nil
	}

	aptype, err := NewAuthProviderType(typeStr)
	if err != nil {
		return err
	}

	*a = aptype
	return nil
}

func (a AuthProviderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.aptype)
}

func (a *AuthProviderType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	aptype, err := NewAuthProviderType(s)
	if err != nil {
		return err
	}

	*a = aptype
	return nil
}
