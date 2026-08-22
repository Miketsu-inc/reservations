package email

import (
	"context"

	"golang.org/x/text/language"
)

type EmployeeInvitationData struct {
	InviterName  string `json:"inviter_name"`
	MerchantName string `json:"merchant_name"`
	AcceptLink   string `json:"accept_link"`
}

// TODO: missing email
func (s *Service) EmployeeInvitation(ctx context.Context, lang language.Tag, to string, data EmployeeInvitationData) error {
	return nil
}
