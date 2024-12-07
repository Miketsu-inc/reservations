package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/miketsu-inc/reservations/backend/pkg/email"
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

	password, err := u.Postgresdb.GetUserPasswordByUserEmail(r.Context(), login.Email)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("incorrect email or password %s", err.Error()))
		return
	}

	err = hashCompare(login.Password, password)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, err)
		return
	}
}

func (u *UserAuth) Signup(w http.ResponseWriter, r *http.Request) {
	type signUpData struct {
		FirstName   string `json:"firstName" validate:"required"`
		LastName    string `json:"lastName" validate:"required"`
		Email       string `json:"email" validate:"required,email"`
		PhoneNumber string `json:"phoneNum" validate:"required,e164"`
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

	userId, err := uuid.NewV7()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating user id: %s", err.Error()))
		return
	}

	err = u.Postgresdb.NewUser(r.Context(), database.User{
		Id:             userId,
		FirstName:      signup.FirstName,
		LastName:       signup.LastName,
		Email:          signup.Email,
		PhoneNumber:    signup.PhoneNumber,
		PasswordHash:   hashedPassword,
		SubscriptionId: 0,
		// Settings:       make(map[string]bool),
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error when creating user: %s", err.Error()))
		return
	}

	token, err := jwt.New([]byte(os.Getenv("JWT_SECRET")), userId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error when creating jwt token: %s", err.Error()))
		return
	}

	exp_time, err := strconv.Atoi(os.Getenv("JWT_EXPIRATION_TIME"))
	assert.Nil(err, "JWT_EXPIRATION_TIME environment variable could not be found", err)

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
		MaxAge:   exp_time,
		Expires:  time.Now().UTC().Add(time.Hour * 24 * 30),
		Path:     "/",
		// needs to be true in production
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusCreated)
}

// The jwt auth middleware should always run before this as that is what verifies the user.
func (u *UserAuth) IsAuthenticated(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (u *UserAuth) Email(w http.ResponseWriter, r *http.Request) {
	type emailData struct {
		Email string `json:"email" validate:"required,email"`
	}
	var data emailData

	if err := validate.ParseStruct(r, &data); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := u.Postgresdb.IsEmailUnique(r.Context(), data.Email)
	if err != nil {
		httputil.Error(w, http.StatusConflict, err)
		return
	}

	if err := email.SendMail(data.Email, "http://localhost:5173/signup"); err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error sending the verification email: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (u *UserAuth) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	type verifData struct {
		Token string `json:"token" validate:"required"`
	}

	var data verifData

	if err := validate.ParseStruct(r, &data); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	if err := email.ValidateToken(data.Token); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
