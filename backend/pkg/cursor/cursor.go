package cursor

import (
	"encoding/base64"
	"encoding/json"
)

func Encode[T any](value T) (string, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func Decode[T any](encoded string) (T, error) {
	var value T

	if encoded == "" {
		return value, nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return value, err
	}

	if err := json.Unmarshal(decoded, &value); err != nil {
		return value, err
	}

	return value, nil
}
