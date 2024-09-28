package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Postgresdb database.PostgreSQL
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func hashCompare(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		slog.Debug(err.Error())
	}
	return err == nil
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

func parseSanitizeConvert[T any](r *http.Request) (T, error) {
	var data T

	if err := utils.ParseJSON(r, &data); err != nil {
		return data, err
	}

	sanitizedInterface, err := sanitize(data)
	if err != nil {
		return data, err
	}

	sanitizedData, ok := sanitizedInterface.(T)
	if !ok {
		return data, errors.New("unexpected error during handling data")
	}

	return sanitizedData, nil
}

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *Auth) HandleLogin(w http.ResponseWriter, r *http.Request) {
	login, err := parseSanitizeConvert[LoginData](r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unexpected error during handling data"))
	}

	password, err := a.Postgresdb.GetUserPasswordByUserEmail(r.Context(), login.Email)
	if err != nil {
		slog.Error(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("incorrect email or password"))
	}

	if hashCompare(login.Password, password) {
		utils.WriteJSON(w, http.StatusOK, map[string]string{"Response": "User logged in successfully"})
	} else {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("incorrect email or password"))
	}
}

type SignUpData struct {
	Firstname string `json:"firstName"`
	Lastname  string `json:"lastName"`
	Email     string `json:"email"`
	Phonenum  string `json:"phoneNum"`
	Password  string `json:"password"`
}

func (a *Auth) HandleSignup(w http.ResponseWriter, r *http.Request) {
	signup, err := parseSanitizeConvert[SignUpData](r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unexpected error during handling data"))
		return
	}

	_, err = a.Postgresdb.GetUserPasswordByUserEmail(r.Context(), signup.Email)
	if !errors.Is(err, sql.ErrNoRows) {
		if err != nil {
			utils.WriteError(w, http.StatusConflict, fmt.Errorf("unexpected error during handling data"))
			return
		}

		utils.WriteError(w, http.StatusConflict, fmt.Errorf("the email %s is already used", signup.Email))
		return
	}

	hashedPassword, err := hashPassword(signup.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("the password is too long"))
		return
	}

	userId, err := uuid.NewV7()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating user id"))
		return
	}

	err = a.Postgresdb.NewUser(r.Context(), database.User{
		Id:           userId,
		FirstName:    signup.Firstname,
		LastName:     signup.Lastname,
		Email:        signup.Email,
		Password:     hashedPassword,
		Subscription: 0,
		Settings:     make(map[string]bool),
	})
	if err != nil {
		slog.Error(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unexpected error when creating user"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"Response": "User created successfully"})
}
