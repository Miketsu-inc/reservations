package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/oauthutil"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

var googleConf = &oauth2.Config{
	ClientID:     config.LoadEnvVars().GOOGLE_OAUTH_CLIENT_ID,
	ClientSecret: config.LoadEnvVars().GOOGLE_OAUTH_CLIENT_SECRET,
	RedirectURL:  "http://localhost:8080/api/v1/auth/callback/google",
	Scopes:       []string{"email", "profile"},
	Endpoint:     google.Endpoint,
}

func (s *Service) GoogleLogin(ctx context.Context) (string, string, error) {
	state, err := oauthutil.RandomString(32)
	if err != nil {
		return "", "", err
	}

	url := googleConf.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return url, state, nil
}

// TODO: if not unique the user already registered without oauth
// we should probably show a prompt to login with the original method
func (s *Service) GoogleCallback(ctx context.Context, code string) (jwt.TokenPair, error) {
	token, err := googleConf.Exchange(ctx, code)
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("error during google oauth exchange: %s", err.Error())
	}

	client := googleConf.Client(ctx, token)

	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("error during request to google user endpoint: %s", err.Error())
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
		return jwt.TokenPair{}, err
	}

	userId, err := s.userRepo.FindOauthUser(ctx, types.AuthProviderTypeGoogle, g.Id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return jwt.TokenPair{}, fmt.Errorf("error during finding oauth user: %s", err.Error())
		}

		userId, err = uuid.NewV7()
		if err != nil {
			return jwt.TokenPair{}, fmt.Errorf("unexpected error during creating user id: %s", err.Error())
		}

		err = s.userRepo.NewUser(ctx, domain.User{
			Id:                userId,
			FirstName:         g.GivenName,
			LastName:          g.FamilyName,
			Email:             g.Email,
			PhoneNumber:       nil,
			PasswordHash:      nil,
			JwtRefreshVersion: 0,
			PreferredLang:     nil,
			AuthProvider:      &types.AuthProviderTypeGoogle,
			ProviderId:        &g.Id,
		})
		if err != nil {
			return jwt.TokenPair{}, fmt.Errorf("error creating new oauth user: %s", err.Error())
		}
	}

	accessToken, err := jwt.NewAccessToken(userId)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	refreshToken, err := jwt.NewRefreshToken(userId, 0)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	return jwt.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

var facebookConf = &oauth2.Config{
	ClientID:     config.LoadEnvVars().FACEBOOK_OAUTH_CLIENT_ID,
	ClientSecret: config.LoadEnvVars().FACEBOOK_OAUTH_CLIENT_SECRET,
	RedirectURL:  "http://localhost:8080/api/v1/auth/callback/facebook",
	Scopes:       []string{"email", "public_profile"},
	Endpoint:     facebook.Endpoint,
}

func (s *Service) FacebookLogin(ctx context.Context) (string, string, error) {
	state, err := oauthutil.RandomString(32)
	if err != nil {
		return "", "", err
	}

	url := facebookConf.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return url, state, nil
}

// TODO: if not unique the user already registered without oauth
// we should probably show a prompt to login with the original method
func (s *Service) FacebookCallback(ctx context.Context, code string) (jwt.TokenPair, error) {
	token, err := facebookConf.Exchange(ctx, code)
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("error during facebook oauth exchange: %s", err.Error())
	}

	client := facebookConf.Client(ctx, token)

	resp, err := client.Get("https://graph.facebook.com/v24.0/me?fields=id,name,first_name,last_name,email,picture")
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("error during request to facebook user endpoint: %s", err.Error())
	}
	// nolint:errcheck
	defer resp.Body.Close()

	type facebookUser struct {
		Id        string `json:"id"`
		Name      string `json:"name"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Picture   struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	var fb facebookUser

	err = json.NewDecoder(resp.Body).Decode(&fb)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	userId, err := s.userRepo.FindOauthUser(ctx, types.AuthProviderTypeFacebook, fb.Id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return jwt.TokenPair{}, fmt.Errorf("error during finding oauth user: %s", err.Error())
		}

		userId, err = uuid.NewV7()
		if err != nil {
			return jwt.TokenPair{}, fmt.Errorf("unexpected error during creating user id: %s", err.Error())
		}

		err = s.userRepo.NewUser(ctx, domain.User{
			Id:                userId,
			FirstName:         fb.FirstName,
			LastName:          fb.LastName,
			Email:             fb.Email,
			PhoneNumber:       nil,
			PasswordHash:      nil,
			JwtRefreshVersion: 0,
			PreferredLang:     nil,
			AuthProvider:      &types.AuthProviderTypeFacebook,
			ProviderId:        &fb.Id,
		})
		if err != nil {
			return jwt.TokenPair{}, fmt.Errorf("error creating new oauth user: %s", err.Error())
		}
	}

	accessToken, err := jwt.NewAccessToken(userId)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	refreshToken, err := jwt.NewRefreshToken(userId, 0)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	return jwt.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
