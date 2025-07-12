package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
	"golang.org/x/crypto/bcrypt"
)

type UserAuth struct {
	Postgresdb database.PostgreSQL
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func hashCompare(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return fmt.Errorf("incorrect email or password")

		} else if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return fmt.Errorf("password is too long")

		} else {
			// for debug purposes
			return err
		}
	}

	return nil
}

// Creates and sets both the resfresh and access jwt cookies
func (u *UserAuth) newJwts(w http.ResponseWriter, ctx context.Context, userID uuid.UUID) error {
	refreshVersion, err := u.Postgresdb.GetUserJwtRefreshVersion(ctx, userID)
	if err != nil {
		return fmt.Errorf("unexpected error when getting refresh version: %s", err.Error())
	}

	err = jwt.NewRefreshToken(w, userID, refreshVersion)
	if err != nil {
		return fmt.Errorf("unexpected error when creating refresh jwt token: %s", err.Error())
	}

	err = jwt.NewAccessToken(w, userID)
	if err != nil {
		return fmt.Errorf("unexpected error when creating access jwt token: %s", err.Error())
	}

	return nil
}

func (u *UserAuth) Login(w http.ResponseWriter, r *http.Request) {
	type loginData struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,ascii"`
	}
	var login loginData

	if err := validate.ParseStruct(r, &login); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID, password, err := u.Postgresdb.GetUserPasswordAndIDByUserEmail(r.Context(), login.Email)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("incorrect email or password %s", err.Error()))
		return
	}

	err = hashCompare(login.Password, password)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, err)
		return
	}

	err = u.newJwts(w, r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}
}

func (u *UserAuth) Signup(w http.ResponseWriter, r *http.Request) {
	type signUpData struct {
		FirstName   string `json:"first_name" validate:"required"`
		LastName    string `json:"last_name" validate:"required"`
		Email       string `json:"email" validate:"required,email"`
		PhoneNumber string `json:"phone_number" validate:"required,e164"`
		Password    string `json:"password" validate:"required,ascii"`
	}
	var signup signUpData

	if err := validate.ParseStruct(r, &signup); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := u.Postgresdb.IsEmailUnique(r.Context(), signup.Email)
	if err != nil {
		httputil.Error(w, http.StatusConflict, err)
		return
	}

	err = u.Postgresdb.IsPhoneNumberUnique(r.Context(), signup.PhoneNumber)
	if err != nil {
		httputil.Error(w, http.StatusConflict, err)
		return
	}

	hashedPassword, err := hashPassword(signup.Password)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("the password is too long"))
		return
	}

	userID, err := uuid.NewV7()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating user id: %s", err.Error()))
		return
	}

	err = u.Postgresdb.NewUser(r.Context(), database.User{
		Id:                userID,
		FirstName:         signup.FirstName,
		LastName:          signup.LastName,
		Email:             signup.Email,
		PhoneNumber:       signup.PhoneNumber,
		PasswordHash:      hashedPassword,
		JwtRefreshVersion: 0,
		Subscription:      0,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error when creating user: %s", err.Error()))
		return
	}

	err = u.newJwts(w, r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// The jwt auth middleware should always run before this as that is what verifies the user.
func (u *UserAuth) IsAuthenticated(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (u *UserAuth) Logout(w http.ResponseWriter, r *http.Request) {
	jwt.DeleteJwts(w)
}

func (u *UserAuth) LogoutAllDevices(w http.ResponseWriter, r *http.Request) {
	userID := jwt.UserIDFromContext(r.Context())

	err := u.Postgresdb.IncrementUserJwtRefreshVersion(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	jwt.DeleteJwts(w)
}
