package invitations

import teamServ "github.com/miketsu-inc/reservations/backend/internal/service/team"

func mapToGetInvitation(in teamServ.GetInvitationResult) getInvitationResp {
	return getInvitationResp{
		MerchantName: in.MerchantName,
		InvitorName:  in.InvitorName,
		Role:         in.Role,
		Email:        in.Email,
		Status:       in.Status,
	}
}
