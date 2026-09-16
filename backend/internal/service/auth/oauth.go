package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/lang"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/repository"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/oauthutil"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

type oauthProfile struct {
	ProviderId    string
	FirstName     string
	LastName      string
	Email         string
	EmailVerified bool
}

type oauthProvider struct {
	authProvider         types.AuthProviderType
	config               *oauth2.Config
	userInfoURL          string
	requireVerifiedEmail bool
	decodeProfile        func(io.Reader) (oauthProfile, error)
}

var googleProvider = oauthProvider{
	authProvider: types.AuthProviderTypeGoogle,
	config: &oauth2.Config{
		ClientID:     config.LoadEnvVars().GOOGLE_OAUTH_CLIENT_ID,
		ClientSecret: config.LoadEnvVars().GOOGLE_OAUTH_CLIENT_SECRET,
		RedirectURL:  "http://localhost:8080/api/v1/auth/oauth/google/callback",
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	},
	userInfoURL:          "https://openidconnect.googleapis.com/v1/userinfo",
	requireVerifiedEmail: true,
	decodeProfile:        decodeGoogleProfile,
}

func decodeGoogleProfile(r io.Reader) (oauthProfile, error) {
	var user struct {
		Id            string `json:"sub"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := json.NewDecoder(r).Decode(&user); err != nil {
		return oauthProfile{}, err
	}

	return oauthProfile{
		ProviderId:    user.Id,
		FirstName:     user.GivenName,
		LastName:      user.FamilyName,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
	}, nil
}

func (s *Service) GoogleLogin(ctx context.Context) (string, string, error) {
	return oauthLogin(googleProvider)
}

func (s *Service) GoogleCallback(ctx context.Context, code string) (jwt.TokenPair, error) {
	return s.oauthCallback(ctx, code, googleProvider)
}

var facebookProvider = oauthProvider{
	authProvider: types.AuthProviderTypeFacebook,
	config: &oauth2.Config{
		ClientID:     config.LoadEnvVars().FACEBOOK_OAUTH_CLIENT_ID,
		ClientSecret: config.LoadEnvVars().FACEBOOK_OAUTH_CLIENT_SECRET,
		RedirectURL:  "http://localhost:8080/api/v1/auth/oauth/facebook/callback",
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     facebook.Endpoint,
	},
	userInfoURL:   "https://graph.facebook.com/v24.0/me?fields=id,first_name,last_name,email",
	decodeProfile: decodeFacebookProfile,
}

func decodeFacebookProfile(r io.Reader) (oauthProfile, error) {
	var user struct {
		Id        string `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	}
	if err := json.NewDecoder(r).Decode(&user); err != nil {
		return oauthProfile{}, err
	}

	return oauthProfile{
		ProviderId: user.Id,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Email:      user.Email,
	}, nil
}

func (s *Service) FacebookLogin(ctx context.Context) (string, string, error) {
	return oauthLogin(facebookProvider)
}

func (s *Service) FacebookCallback(ctx context.Context, code string) (jwt.TokenPair, error) {
	return s.oauthCallback(ctx, code, facebookProvider)
}

func oauthLogin(provider oauthProvider) (string, string, error) {
	state, err := oauthutil.RandomString(32)
	if err != nil {
		return "", "", err
	}

	return provider.config.AuthCodeURL(state, oauth2.AccessTypeOffline), state, nil
}

func (s *Service) oauthCallback(ctx context.Context, code string, provider oauthProvider) (jwt.TokenPair, error) {
	profile, err := fetchOauthProfile(ctx, code, provider)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	userId, refreshVersion, err := s.findOrCreateOauthUser(ctx, provider.authProvider, profile)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	return newJwtTokens(userId, refreshVersion)
}

func fetchOauthProfile(ctx context.Context, code string, provider oauthProvider) (oauthProfile, error) {
	token, err := provider.config.Exchange(ctx, code)
	if err != nil {
		return oauthProfile{}, fmt.Errorf("error during %s oauth exchange: %w", provider.authProvider, err)
	}

	resp, err := provider.config.Client(ctx, token).Get(provider.userInfoURL)
	if err != nil {
		return oauthProfile{}, fmt.Errorf("error during request to %s user endpoint: %w", provider.authProvider, err)
	}
	// nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return oauthProfile{}, fmt.Errorf("%s user endpoint returned status %s", provider.authProvider, resp.Status)
	}

	profile, err := provider.decodeProfile(resp.Body)
	if err != nil {
		return oauthProfile{}, fmt.Errorf("error decoding %s user profile: %w", provider.authProvider, err)
	}

	profile.ProviderId = strings.TrimSpace(profile.ProviderId)
	profile.Email = strings.TrimSpace(profile.Email)

	if profile.ProviderId == "" {
		return oauthProfile{}, fmt.Errorf("%s user profile did not contain a provider id", provider.authProvider)
	}

	if profile.Email == "" || (provider.requireVerifiedEmail && !profile.EmailVerified) {
		return oauthProfile{}, ErrOauthEmailUnavailable
	}

	return profile, nil
}

func (s *Service) findOrCreateOauthUser(ctx context.Context, provider types.AuthProviderType, profile oauthProfile) (uuid.UUID, int, error) {
	userId, err := s.userRepo.FindOauthUser(ctx, provider, profile.ProviderId)
	if err == nil {
		refreshVersion, err := s.userRepo.GetUserJwtRefreshVersion(ctx, userId)
		if err != nil {
			return uuid.Nil, 0, err
		}

		return userId, refreshVersion, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, 0, err
	}

	unique, err := s.userRepo.IsEmailUnique(ctx, profile.Email)
	if err != nil {
		return uuid.Nil, 0, err
	}

	if !unique {
		// Email equality alone is not enough proof to link login methods. Linking must
		// happen in a separate flow authenticated as the existing user.
		return uuid.Nil, 0, ErrOauthAccountAlreadyExists
	}

	userId, err = uuid.NewV7()
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("error creating user id: %w", err)
	}

	err = s.userRepo.NewUser(ctx, domain.User{
		Id:                userId,
		FirstName:         profile.FirstName,
		LastName:          profile.LastName,
		Email:             profile.Email,
		PhoneNumber:       nil,
		PasswordHash:      nil,
		JwtRefreshVersion: 0,
		Language:          lang.LangFromContext(ctx).String(),
		AuthProvider:      &provider,
		ProviderId:        &profile.ProviderId,
	})
	if err != nil {
		if db.IsUniqueConstraintViolation(err, repository.UserEmailLowerUniqueConstraint) {
			return uuid.Nil, 0, ErrOauthAccountAlreadyExists
		}

		if db.IsUniqueConstraintViolation(err, repository.UserOauthIdentityUniqueConstraint) {
			return s.existingOauthUser(ctx, provider, profile.ProviderId)
		}

		return uuid.Nil, 0, err
	}

	return userId, 0, nil
}

func (s *Service) existingOauthUser(ctx context.Context, provider types.AuthProviderType, providerId string) (uuid.UUID, int, error) {
	userId, err := s.userRepo.FindOauthUser(ctx, provider, providerId)
	if err != nil {
		return uuid.Nil, 0, err
	}

	refreshVersion, err := s.userRepo.GetUserJwtRefreshVersion(ctx, userId)
	if err != nil {
		return uuid.Nil, 0, err
	}

	return userId, refreshVersion, err
}
