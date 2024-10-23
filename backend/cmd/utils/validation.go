package utils

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

func StructValidation(s interface{}) map[string]string {
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
