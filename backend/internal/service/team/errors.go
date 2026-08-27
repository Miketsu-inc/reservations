package team

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrPreferencesForbidden: http.StatusForbidden,

	ErrInvitationNotFound:        http.StatusNotFound,
	ErrInvitationForbidden:       http.StatusForbidden,
	ErrInvitationExpired:         http.StatusGone,
	ErrInvitationRevoked:         http.StatusConflict,
	ErrInvitationAlreadyAccepted: http.StatusConflict,
	ErrInvitationAlreadyDeclined: http.StatusConflict,
	ErrInvitationEmailMismatch:   http.StatusForbidden,

	ErrEmployeeNotFound:                            http.StatusNotFound,
	ErrUserEmployeeCannotBeEdited:                  http.StatusForbidden,
	ErrMultipleOwnersNotAllowed:                    http.StatusConflict,
	ErrOwnerRoleChangeRestrictedToMerchantSettings: http.StatusConflict,
}

var ErrPreferencesForbidden = &apperr.Error{Code: "preferences_forbidden", Message: "you do not have permission to access these preferences"}

var ErrInvitationNotFound = &apperr.Error{Code: "invitation_not_found", Message: "invitation not found for this id"}
var ErrInvitationForbidden = &apperr.Error{Code: "invitation_forbidden", Message: "you do not have permission to access this invitation"}
var ErrInvitationExpired = &apperr.Error{Code: "invitation_expired", Message: "the invitation has expired"}
var ErrInvitationRevoked = &apperr.Error{Code: "invitation_revoked", Message: "the invitation has been revoked"}
var ErrInvitationAlreadyAccepted = &apperr.Error{Code: "invitation_already_accepted", Message: "the invitation has already been accepted"}
var ErrInvitationAlreadyDeclined = &apperr.Error{Code: "invitation_already_declined", Message: "the invitation has already been declined"}
var ErrInvitationEmailMismatch = &apperr.Error{Code: "invitation_email_mismatch", Message: "the invitation email does not match the user email"}

var ErrEmployeeNotFound = &apperr.Error{Code: "employee_not_found", Message: "employee not found for this id"}
var ErrUserEmployeeCannotBeEdited = &apperr.Error{Code: "user_employee_cannot_be_edited", Message: "employee created from user account cannot be edited"}
var ErrMultipleOwnersNotAllowed = &apperr.Error{Code: "multiple_owners_not_allowed", Message: "there can be only one owner"}
var ErrOwnerRoleChangeRestrictedToMerchantSettings = &apperr.Error{Code: "owner_role_change_restricted_to_merchant_settings", Message: "owner role can only be changed in merchant settings"}
