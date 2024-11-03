package validate

import (
	"fmt"
	"html"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func Struct(s interface{}) map[string]string {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(s)
	if err != nil {

		var errorMsg = make(map[string]string)

		errors := err.(validator.ValidationErrors)
		for _, err := range errors {

			jsonTag, ok := getJsonTagForField(s, err.Field())
			assert.True(ok, fmt.Sprintf("%s field does not have a json tag", err.Field()), err.Error())

			errorMsg[jsonTag] = errorMessageForTag(err)
		}

		return errorMsg
	}

	return nil
}

func errorMessageForTag(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email"
	case "number":
		return "Must be a number"
	case "ascii":
		return "Cannot contain non-ASCII characters"
	case "min":
		return fmt.Sprintf("This field should be at least %s long", err.Param())
	case "max":
		return fmt.Sprintf("This field should be at most %s long", err.Param())
	}
	return err.Error()
}

func getJsonTagForField(s interface{}, fieldName string) (string, bool) {
	t := reflect.TypeOf(s)
	sf, ok := t.FieldByName(fieldName)
	if !ok {
		return "", false
	}
	return sf.Tag.Lookup("json")
}

func MerchantNameToUrlName(s string) (string, error) {
	result, err := replaceAccents(s)
	if err != nil {
		return "", err
	}

	result, err = replaceSpecialCharsWithHyphen(result)
	if err != nil {
		return "", err
	}

	result = reduceHyphens(result)
	if result == "" {
		return "", fmt.Errorf("urlName is empty after processing")
	}

	return result, nil
}

func replaceAccents(s string) (string, error) {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, s)
	if err != nil {
		return "", err
	}

	return result, nil
}

func replaceSpecialCharsWithHyphen(s string) (string, error) {
	specialChars := []rune{
		'`', '~', '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '_', '=', '+', ' ',
		'[', ']', '{', '}', ';', ':', '"', '\'', '\\', '|', ',', '<', '.', '>', '/', '?',
	}

	Fn := runes.Map(func(r rune) rune {
		if slices.Contains(specialChars, r) || unicode.IsControl(r) || r > unicode.MaxASCII {
			return '-'
		}

		return r
	})

	result, _, err := transform.String(Fn, s)
	if err != nil {
		return "", err
	}

	return result, nil
}

func reduceHyphens(s string) string {
	// replace multiple hyphens with a single hyphen
	re := regexp.MustCompile(`-+`)
	reduced := re.ReplaceAllLiteralString(s, "-")

	// remove leading and trailing hyphens
	return strings.Trim(reduced, "-")
}

func SanitizeStruct[T any](s any) (T, error) {
	var data T

	sanitizedInterface, err := sanitize(data)
	if err != nil {
		return data, err
	}

	sanitizedData, ok := sanitizedInterface.(T)
	if !ok {
		return data, fmt.Errorf("unexpected error during handling data")
	}

	return sanitizedData, nil
}

func sanitize(s interface{}) (interface{}, error) {
	value := reflect.ValueOf(s)
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a string")
	}

	sanitizedData := reflect.New(value.Type()).Elem()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		sanitizedField := sanitizedData.Field(i)

		if field.Kind() == reflect.String {
			strValue := field.String()
			escapedValue := html.EscapeString(strValue)
			sanitizedField.SetString(escapedValue)
		} else {
			sanitizedField.Set(field)
		}
	}
	return sanitizedData.Interface(), nil
}
