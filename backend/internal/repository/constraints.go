package repository

import "github.com/miketsu-inc/reservations/backend/pkg/db"

const (
	UserEmailLowerUniqueConstraint    db.ConstraintName = "user_email_lower_unique"
	UserOauthIdentityUniqueConstraint db.ConstraintName = "user_oauth_identity_unique"
)
