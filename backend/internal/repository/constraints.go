package repository

import "github.com/miketsu-inc/reservations/backend/pkg/db"

const (
	UserEmailLowerUniqueConstraint           db.ConstraintName = "user_email_lower_unique"
	UserOauthIdentityUniqueConstraint        db.ConstraintName = "user_oauth_identity_unique"
	UniqueBookingParticipantConstraint       db.ConstraintName = "unique_booking_participant"
	UniqueBookingSeriesParticipantConstraint db.ConstraintName = "unique_booking_series_participant"
)
