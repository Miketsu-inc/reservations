package team

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/lang"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/oauthutil"
)

func canUpdateInvitation(inv domain.EmployeeInvitation) error {
	if inv.Status == types.EmployeeInvitationStatusAccepted {
		return ErrInvitationAlreadyAccepted
	}

	if inv.Status == types.EmployeeInvitationStatusDeclined {
		return ErrInvitationAlreadyDeclined
	}

	return nil
}

func canRespondToInvitation(inv domain.EmployeeInvitation) error {
	if err := canUpdateInvitation(inv); err != nil {
		return err
	}

	if inv.Status == types.EmployeeInvitationStatusRevoked {
		return ErrInvitationRevoked
	}

	if inv.IsExpired() || inv.Status == types.EmployeeInvitationStatusExpired {
		return ErrInvitationExpired
	}

	return nil
}

func (s *Service) InviteMember(ctx context.Context, email string, role types.EmployeeInvitationRole) error {
	actor := actor.MustGetFromContext(ctx)

	token, err := oauthutil.RandomString(32)
	if err != nil {
		return err
	}

	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))

	invitedAt := time.Now().UTC()
	expiresAt := invitedAt.AddDate(0, 0, 7)

	var invitationId int

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		existing, err := s.teamRepo.WithTx(tx).GetEmployeeInvitationByEmail(ctx, actor.MerchantId, email)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}

			invitationId, err = s.teamRepo.WithTx(tx).NewEmployeeInvitation(ctx, domain.EmployeeInvitation{
				Status:     types.EmployeeInvitationStatusPending,
				MerchantId: actor.MerchantId,
				Email:      email,
				Role:       role,
				Token:      tokenHash,
				InvitedBy:  &actor.EmployeeId,
				InvitedAt:  invitedAt,
				ExpiresAt:  expiresAt,
			})
			if err != nil {
				return err
			}
		} else {
			invitationId = existing.Id

			if err = canUpdateInvitation(existing); err != nil {
				return err
			}

			err = s.teamRepo.WithTx(tx).UpdateEmployeeInvitation(ctx, domain.EmployeeInvitation{
				Id:        invitationId,
				Status:    types.EmployeeInvitationStatusPending,
				Role:      role,
				Token:     tokenHash,
				InvitedBy: &actor.EmployeeId,
				InvitedAt: invitedAt,
				ExpiresAt: expiresAt,
			})
			if err != nil {
				return err
			}
		}

		lang := lang.LangFromContext(ctx)

		_, err = s.enqueuer.InsertTx(ctx, tx, args.EmployeeInvitationEmail{
			// assume that the invitor and invitee speaks the same language
			// as this is probably better than defaulting to english
			Language:     lang,
			InvitationId: invitationId,
		}, nil)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) ResendInvitation(ctx context.Context, invitationId int) error {
	actor := actor.MustGetFromContext(ctx)

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		invitation, err := s.teamRepo.WithTx(tx).GetEmployeeInvitation(ctx, invitationId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvitationNotFound
			}

			return err
		}

		if invitation.MerchantId != actor.MerchantId {
			return ErrInvitationForbidden
		}

		if err = canUpdateInvitation(invitation); err != nil {
			return err
		}

		token, err := oauthutil.RandomString(32)
		if err != nil {
			return err
		}

		now := time.Now().UTC()

		err = s.teamRepo.WithTx(tx).UpdateEmployeeInvitation(ctx, domain.EmployeeInvitation{
			Id:        invitation.Id,
			Status:    types.EmployeeInvitationStatusPending,
			Role:      invitation.Role,
			Token:     fmt.Sprintf("%x", sha256.Sum256([]byte(token))),
			InvitedBy: &actor.EmployeeId,
			InvitedAt: now,
			ExpiresAt: now.AddDate(0, 0, 7),
		})
		if err != nil {
			return err
		}

		lang := lang.LangFromContext(ctx)

		_, err = s.enqueuer.InsertTx(ctx, tx, args.EmployeeInvitationEmail{
			// assume that the invitor and invitee speaks the same language
			// as this is probably better than defaulting to english
			Language:     lang,
			InvitationId: invitation.Id,
		}, nil)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) RevokeInvitation(ctx context.Context, invitationId int) error {
	actor := actor.MustGetFromContext(ctx)

	invitation, err := s.teamRepo.GetEmployeeInvitation(ctx, invitationId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvitationNotFound
		}

		return err
	}

	if invitation.MerchantId != actor.MerchantId {
		return ErrInvitationForbidden
	}

	if err = canUpdateInvitation(invitation); err != nil {
		return err
	}

	err = s.teamRepo.RevokeEmployeeInvitation(ctx, invitation.Id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetInvitations(ctx context.Context) ([]domain.EmployeeInvitation, error) {
	actor := actor.MustGetFromContext(ctx)

	invitations, err := s.teamRepo.GetEmployeeInvitations(ctx, actor.MerchantId)
	if err != nil {
		return []domain.EmployeeInvitation{}, err
	}

	return invitations, nil
}

type GetInvitationResult struct {
	MerchantName string
	InvitorName  *string
	Role         types.EmployeeInvitationRole
	Email        string
	Status       types.EmployeeInvitationStatus
}

func (s *Service) GetInvitation(ctx context.Context, token string) (GetInvitationResult, error) {
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))

	invitation, err := s.teamRepo.GetEmployeeInvitationByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetInvitationResult{}, ErrInvitationNotFound
		}

		return GetInvitationResult{}, err
	}

	merchant, err := s.merchantRepo.GetMerchant(ctx, invitation.MerchantId)
	if err != nil {
		return GetInvitationResult{}, err
	}

	var invitorName *string

	if invitation.InvitedBy != nil {
		employee, err := s.teamRepo.GetEmployee(ctx, invitation.MerchantId, *invitation.InvitedBy)
		if err != nil {
			return GetInvitationResult{}, err
		}

		if employee.FirstName != nil && employee.LastName != nil {
			name := fmt.Sprintf("%s %s", *employee.FirstName, *employee.LastName)
			invitorName = &name
		}
	}

	result := GetInvitationResult{
		MerchantName: merchant.Name,
		InvitorName:  invitorName,
		Role:         invitation.Role,
		Email:        invitation.Email,
		Status:       invitation.Status,
	}

	return result, nil
}

func (s *Service) validateInvitation(ctx context.Context, userEmail string, token string) (*domain.EmployeeInvitation, error) {
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))

	invitation, err := s.teamRepo.GetEmployeeInvitationByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvitationNotFound
		}

		return nil, err
	}

	if invitation.Email != userEmail {
		return nil, ErrInvitationEmailMismatch
	}

	if err = canRespondToInvitation(invitation); err != nil {
		return nil, err
	}

	return &invitation, nil
}

func (s *Service) AcceptInvitation(ctx context.Context, token string) (string, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	user, err := s.userRepo.GetUser(ctx, userId)
	if err != nil {
		return "", err
	}

	invitation, err := s.validateInvitation(ctx, user.Email, token)
	if err != nil {
		return "", err
	}

	err = s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err = s.teamRepo.AcceptEmployeeInvitation(ctx, invitation.Id)
		if err != nil {
			return err
		}

		_, err = s.teamRepo.NewEmployee(ctx, domain.Employee{
			UserId:     &user.Id,
			MerchantId: invitation.MerchantId,
			Role:       invitation.Role.ToEmployeeRole(),
			FirstName:  &user.FirstName,
			LastName:   &user.LastName,
			IsActive:   true,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/dashboard", config.LoadEnvVars().JABULANI_URL), nil
}

func (s *Service) DeclineInvitation(ctx context.Context, token string) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	user, err := s.userRepo.GetUser(ctx, userId)
	if err != nil {
		return err
	}

	invitation, err := s.validateInvitation(ctx, user.Email, token)
	if err != nil {
		return err
	}

	err = s.teamRepo.DeclineEmployeeInvitation(ctx, invitation.Id)
	if err != nil {
		return err
	}

	return nil
}
