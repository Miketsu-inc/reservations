package handlers

import (
	"fmt"
	"html"
	"net/http"
	"reflect"
	"unicode"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Postgresdb database.PostgreSQL
}

func NewAuthHandler() *Auth {
	return &Auth{}
}

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpData struct {
	Firstname string `json:"firstName"`
	Lastname  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 15)
	return string(bytes), err
}

func HashCompare(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Sanitize(s interface{}) (interface{}, error) {
	value := reflect.ValueOf(s)
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a string")
	}

	sanitizedData := reflect.New(value.Type()).Elem()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		sanitizedField := sanitizedData.Field(i)
		fieldName := value.Type().Field(i).Name

		if field.Kind() == reflect.String {
			strValue := field.String()

			if !isASCII(strValue) {
				return nil, fmt.Errorf("%s contains non-ASCII characters", fieldName)
			}

			escapedValue := html.EscapeString(strValue)
			sanitizedField.SetString(escapedValue)
		} else {
			sanitizedField.Set(field)
		}
	}
	return sanitizedData.Interface(), nil
}

var stored_email *string = new(string)
var stored_hash *string = new(string)

func (a *Auth) HandleLogin(w http.ResponseWriter, r *http.Request) {

	var data LoginData
	if err := utils.ParseJSON(r, &data); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	sanitizedInterface, err := Sanitize(data)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	sanitizedData, ok := sanitizedInterface.(LoginData)
	if !ok {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during sanitization"))
		return
	}

	if sanitizedData.Email == *stored_email && HashCompare(sanitizedData.Password, *stored_hash) {
		utils.WriteJSON(w, http.StatusOK, map[string]string{"Response": "User logged in successfully"})
	} else {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid email or password"))
	}

}

func (a *Auth) HandleSignup(w http.ResponseWriter, r *http.Request) {

	var data SignUpData
	if err := utils.ParseJSON(r, &data); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	sanitizedInterface, err := Sanitize(data)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	//change interface to SignupData
	sanitizedData, ok := sanitizedInterface.(SignUpData)
	if !ok {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("assertation failed"))
		return
	}

	if *stored_email == sanitizedData.Email {
		utils.WriteError(w, http.StatusConflict, fmt.Errorf("this email is already used"))
		return
	}
	hashedPassword, err := HashPassword(sanitizedData.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	*stored_email = sanitizedData.Email
	*stored_hash = hashedPassword

	utils.WriteJSON(w, http.StatusOK, map[string]string{"Response": "User created successfully"})
}
