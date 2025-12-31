package handlers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/lang"
	"github.com/miketsu-inc/reservations/backend/cmd/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

type Auth struct {
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
func (a *Auth) newJwts(w http.ResponseWriter, ctx context.Context, userID uuid.UUID, merchantId *uuid.UUID, employeeId *int, locationId *int, role *types.EmployeeRole) error {
	refreshVersion, err := a.Postgresdb.GetUserJwtRefreshVersion(ctx, userID)
	if err != nil {
		return fmt.Errorf("unexpected error when getting refresh version: %s", err.Error())
	}

	err = jwt.NewRefreshToken(w, userID, merchantId, employeeId, locationId, role, refreshVersion)
	if err != nil {
		return fmt.Errorf("unexpected error when creating refresh jwt token: %s", err.Error())
	}

	err = jwt.NewAccessToken(w, userID, merchantId, employeeId, locationId, role)
	if err != nil {
		return fmt.Errorf("unexpected error when creating access jwt token: %s", err.Error())
	}

	return nil
}

func (a *Auth) UserLogin(w http.ResponseWriter, r *http.Request) {
	type loginData struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,ascii"`
	}
	var login loginData

	if err := validate.ParseStruct(r, &login); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID, password, err := a.Postgresdb.GetUserPasswordAndIDByUserEmail(r.Context(), login.Email)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("incorrect email or password %s", err.Error()))
		return
	}

	err = hashCompare(login.Password, *password)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, err)
		return
	}

	employeeAuthInfo, err := a.Postgresdb.GetEmployeesByUser(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("unexpected error when reading employees associated with user: %s", err.Error()))
		return
	}

	var merchantId *uuid.UUID
	var employeeId *int
	var locationId *int
	var role *types.EmployeeRole

	// TODO: later user should be able to select which merchant to log into
	if len(employeeAuthInfo) >= 1 {
		merchantId = &employeeAuthInfo[0].MerchantId
		employeeId = &employeeAuthInfo[0].Id
		locationId = &employeeAuthInfo[0].LocationId
		role = &employeeAuthInfo[0].Role
	}

	err = a.newJwts(w, r.Context(), userID, merchantId, employeeId, locationId, role)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, err)
		return
	}
}

func (a *Auth) UserSignup(w http.ResponseWriter, r *http.Request) {
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

	err := a.Postgresdb.IsEmailUnique(r.Context(), signup.Email)
	if err != nil {
		httputil.Error(w, http.StatusConflict, err)
		return
	}

	err = a.Postgresdb.IsPhoneNumberUnique(r.Context(), signup.PhoneNumber)
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

	err = a.Postgresdb.NewUser(r.Context(), database.User{
		Id:                userID,
		FirstName:         signup.FirstName,
		LastName:          signup.LastName,
		Email:             signup.Email,
		PhoneNumber:       &signup.PhoneNumber,
		PasswordHash:      &hashedPassword,
		JwtRefreshVersion: 0,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error when creating user: %s", err.Error()))
		return
	}

	err = a.newJwts(w, r.Context(), userID, nil, nil, nil, nil)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// TODO: create new jwts here... just don't know what to put as location
func (a *Auth) MerchantSignup(w http.ResponseWriter, r *http.Request) {
	type signUpData struct {
		Name         string `json:"name" validate:"required"`
		ContactEmail string `json:"contact_email" validate:"required,email"`
		Timezone     string `json:"timezone" validate:"required,timezone"`
	}
	var signup signUpData

	if err := validate.ParseStruct(r, &signup); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
	}

	urlName, err := validate.MerchantNameToUrlName(signup.Name)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("unexpected error during merchant url name conversion: %s", err.Error()))
		return
	}

	err = a.Postgresdb.IsMerchantUrlUnique(r.Context(), urlName)
	if err != nil {
		httputil.Error(w, http.StatusConflict, err)
		return
	}

	userID := jwt.MustGetUserIDFromContext(r.Context())

	merchantID, err := uuid.NewV7()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating merchant id: %s", err.Error()))
		return
	}

	language := lang.LangFromContext(r.Context())
	curr := currencyx.FindBest(language)

	err = a.Postgresdb.NewMerchant(r.Context(), userID, database.Merchant{
		Id:               merchantID,
		Name:             signup.Name,
		UrlName:          urlName,
		ContactEmail:     signup.ContactEmail,
		Introduction:     "",
		Announcement:     "",
		AboutUs:          "",
		ParkingInfo:      "",
		PaymentInfo:      "",
		Timezone:         signup.Timezone,
		CurrencyCode:     curr,
		SubscriptionTier: types.SubTierFree,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error creating a merchant: %s", err.Error()))
		return
	}

	businessHours := map[int][]database.TimeSlot{
		0: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
		1: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
		2: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
		3: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
		4: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
		5: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
		6: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
	}

	err = a.Postgresdb.UpdateBusinessHours(r.Context(), merchantID, businessHours)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating business hours for merchant: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// The jwt auth middleware should always run before this as that is what verifies the user.
func (a *Auth) UserIsAuthenticated(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	jwt.DeleteJwts(w)
}

func (a *Auth) LogoutAllDevices(w http.ResponseWriter, r *http.Request) {
	userID := jwt.MustGetUserIDFromContext(r.Context())

	err := a.Postgresdb.IncrementUserJwtRefreshVersion(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	jwt.DeleteJwts(w)
}

func setOauthStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth-state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   5 * 60,
		Expires:  time.Now().UTC().Add(time.Minute * 5),
		// needs to be true in production
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func generateSate() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func validateOauthState(r *http.Request) error {
	state := r.URL.Query().Get("state")
	if state == "" {
		return fmt.Errorf("missing state in callback")
	}

	stateCookie, err := r.Cookie("oauth-state")
	if err != nil {
		return fmt.Errorf("missing oauth-sate cookie")
	}

	if subtle.ConstantTimeCompare([]byte(state), []byte(stateCookie.Value)) != 1 {
		return fmt.Errorf("invalid oauth state")
	}

	return nil
}

var googleConf = &oauth2.Config{
	ClientID:     config.LoadEnvVars().GOOGLE_OAUTH_CLIENT_ID,
	ClientSecret: config.LoadEnvVars().GOOGLE_OAUTH_CLIENT_SECRET,
	RedirectURL:  "http://localhost:8080/api/v1/auth/callback/google",
	Scopes:       []string{"email", "profile"},
	Endpoint:     google.Endpoint,
}

func (a *Auth) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateSate()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	setOauthStateCookie(w, state)

	url := googleConf.AuthCodeURL(state, oauth2.AccessTypeOffline)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// TODO: if not unique the user already registered without oauth
// we should probably show a prompt to login with the original method
func (a *Auth) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	if err := validateOauthState(r); err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error during oauth state validation: %s", err.Error()))
		return
	}

	token, err := googleConf.Exchange(r.Context(), code)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error during google oauth exchange: %s", err.Error()))
		return
	}

	client := googleConf.Client(r.Context(), token)

	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error during request to google user endpoint: %s", err.Error()))
		return
	}
	// nolint:errcheck
	defer resp.Body.Close()

	type googleUser struct {
		Id            string `json:"sub"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Locale        string `json:"locale"`
	}
	var g googleUser

	err = json.NewDecoder(resp.Body).Decode(&g)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("%s", err.Error()))
		return
	}

	userId, err := a.Postgresdb.FindOauthUser(r.Context(), types.AuthProviderTypeFacebook, g.Id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error during finding oauth user: %s", err.Error()))
			return
		}

		userId, err = uuid.NewV7()
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating user id: %s", err.Error()))
			return
		}

		err = a.Postgresdb.NewUser(r.Context(), database.User{
			Id:                userId,
			FirstName:         g.Name,
			LastName:          "",
			Email:             g.Email,
			PhoneNumber:       nil,
			PasswordHash:      nil,
			JwtRefreshVersion: 0,
			PreferredLang:     nil,
			AuthProvider:      &types.AuthProviderTypeGoogle,
			ProviderId:        &g.Id,
		})
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error creating new oauth user: %s", err.Error()))
			return
		}
	}

	err = a.newJwts(w, r.Context(), userId, nil, nil, nil, nil)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error creating new jwts: %s", err.Error()))
		return
	}

	http.Redirect(w, r, "http://localhost:8080/", http.StatusPermanentRedirect)
}

var facebookConf = &oauth2.Config{
	ClientID:     config.LoadEnvVars().FACEBOOK_OAUTH_CLIENT_ID,
	ClientSecret: config.LoadEnvVars().FACEBOOK_OAUTH_CLIENT_SECRET,
	RedirectURL:  "http://localhost:8080/api/v1/auth/callback/facebook",
	Scopes:       []string{"email", "public_profile"},
	Endpoint:     facebook.Endpoint,
}

func (a *Auth) FacebookLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateSate()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	setOauthStateCookie(w, state)

	url := facebookConf.AuthCodeURL(state, oauth2.AccessTypeOffline)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// TODO: if not unique the user already registered without oauth
// we should probably show a prompt to login with the original method
func (a *Auth) FacebookCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	if err := validateOauthState(r); err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error during oauth state validation: %s", err.Error()))
		return
	}

	token, err := facebookConf.Exchange(r.Context(), code)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error during facebook oauth exchange: %s", err.Error()))
		return
	}

	client := facebookConf.Client(r.Context(), token)

	resp, err := client.Get("https://graph.facebook.com/v24.0/me?fields=id,name,email,picture")
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error during request to facebook user endpoint: %s", err.Error()))
		return
	}
	// nolint:errcheck
	defer resp.Body.Close()

	type facebookUser struct {
		Id      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	var fb facebookUser

	err = json.NewDecoder(resp.Body).Decode(&fb)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("%s", err.Error()))
		return
	}

	userId, err := a.Postgresdb.FindOauthUser(r.Context(), types.AuthProviderTypeFacebook, fb.Id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error during finding oauth user: %s", err.Error()))
			return
		}

		userId, err = uuid.NewV7()
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating user id: %s", err.Error()))
			return
		}

		err = a.Postgresdb.NewUser(r.Context(), database.User{
			Id:                userId,
			FirstName:         fb.Name,
			LastName:          "",
			Email:             fb.Email,
			PhoneNumber:       nil,
			PasswordHash:      nil,
			JwtRefreshVersion: 0,
			PreferredLang:     nil,
			AuthProvider:      &types.AuthProviderTypeFacebook,
			ProviderId:        &fb.Id,
		})
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error creating new oauth user: %s", err.Error()))
			return
		}
	}

	err = a.newJwts(w, r.Context(), userId, nil, nil, nil, nil)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error creating new jwts: %s", err.Error()))
		return
	}

	http.Redirect(w, r, "http://localhost:8080/", http.StatusPermanentRedirect)
}
