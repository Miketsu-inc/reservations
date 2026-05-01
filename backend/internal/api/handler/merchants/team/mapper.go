package team

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	teamServ "github.com/miketsu-inc/reservations/backend/internal/service/team"
)

func mapToNewMemberInput(in newMemberReq) teamServ.NewMemberInput {
	return teamServ.NewMemberInput{
		Role:        in.Role,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		IsActive:    in.IsActive,
	}
}

func mapToUpdateMemberInput(in updateMemberReq) teamServ.UpdateMemberInput {
	return teamServ.UpdateMemberInput{
		Role:        in.Role,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		IsActive:    in.IsActive,
	}
}

func mapToGetMemberResp(in domain.PublicEmployee) getMemberResp {
	return getMemberResp{
		Id:          in.Id,
		Role:        in.Role,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		IsActive:    in.IsActive,
	}
}
