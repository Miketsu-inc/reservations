package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
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
	http.SetCookie(w, &http.Cookie{
		Name:     jwt.JwtRefreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Now().UTC(),
		// needs to be true in production
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     jwt.JwtAccessCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Now().UTC(),
		// needs to be true in production
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

// Creates and sets both the resfresh and access jwt cookies
func (u *UserAuth) newJwts(w http.ResponseWriter, ctx context.Context, userID uuid.UUID) error {
	refreshCookie, err := u.newJwtCookie(ctx, userID, jwt.RefreshToken)
	if err != nil {
		return fmt.Errorf("unexpected error when creating refresh jwt token: %s", err.Error())
	}
	http.SetCookie(w, refreshCookie)

	accessCookie, err := u.newJwtCookie(ctx, userID, jwt.AccessToken)
	if err != nil {
		return fmt.Errorf("unexpected error when creating access jwt token: %s", err.Error())
	}
	http.SetCookie(w, accessCookie)

	return nil
}

// Creates a new token and returns the cookie or an error
func (u *UserAuth) newJwtCookie(ctx context.Context, userID uuid.UUID, tokenType jwt.JwtType) (*http.Cookie, error) {
	var secret string
	var expMin int
	var cookieName string

	cfg := config.LoadEnvVars()

	switch tokenType {
	case jwt.RefreshToken:
		secret = cfg.JWT_REFRESH_SECRET
		expMin = cfg.JWT_REFRESH_EXP_MIN

		cookieName = jwt.JwtRefreshCookieName
	case jwt.AccessToken:
		secret = cfg.JWT_ACCESS_SECRET
		expMin = cfg.JWT_ACCESS_EXP_MIN

		cookieName = jwt.JwtAccessCookieName
	default:
		assert.Never("Jwt token type can be either refresh or access", tokenType)
	}

	expMinDuration := time.Minute * time.Duration(expMin)

	var claims jwtlib.MapClaims

	switch tokenType {
	case jwt.RefreshToken:
		refreshVersion, err := u.Postgresdb.IncrementUserJwtRefreshVersion(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("unexpected error when incrementing refresh version: %s", err.Error())
		}

		claims = jwtlib.MapClaims{
			"sub":             userID,
			"exp":             time.Now().Add(expMinDuration).Unix(),
			"refresh_version": refreshVersion,
		}
	case jwt.AccessToken:
		claims = jwtlib.MapClaims{
			"sub": userID,
			"exp": time.Now().Add(expMinDuration).Unix(),
		}
	}

	token, err := jwt.New([]byte(secret), claims)
	if err != nil {
		return nil, fmt.Errorf("unexpected error when creating jwt token: %s", err.Error())
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    token,
		HttpOnly: true,
		MaxAge:   expMin * 60,
		Expires:  time.Now().UTC().Add(expMinDuration),
		Path:     "/",
		// needs to be true in production
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	return cookie, nil
}
