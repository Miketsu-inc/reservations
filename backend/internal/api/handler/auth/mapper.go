package auth

import (
	authServ "github.com/miketsu-inc/reservations/backend/internal/service/auth"
	teamServ "github.com/miketsu-inc/reservations/backend/internal/service/team"
)

func mapToLoginInput(in loginReq) authServ.LoginInput {
	return authServ.LoginInput{
		Email:    in.Email,
		Password: in.Password,
	}
}

func mapToUserSignupInput(in userSignupReq) authServ.UserSignupInput {
	return authServ.UserSignupInput{
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		Password:    in.Password,
	}
}

func mapToMerchantSignupInput(in merchantSignupReq) authServ.MerchantSignupInput {
	return authServ.MerchantSignupInput{
		Name:         in.Name,
		ContactEmail: in.ContactEmail,
		Timezone:     in.Timezone,
	}
}

func mapToMeResp(in teamServ.MeResult) meResp {
	memberships := make([]membershipsResp, len(in.Memberships))

	for i, e := range in.Memberships {
		memberships[i] = membershipsResp{
			MerchantId: e.MerchantId,
			LocationId: e.LocationId,
			EmployeeId: e.Id,
			Role:       e.Role,
		}
	}

	return meResp{
		UserId:      in.User.Id,
		FirstName:   in.User.FirstName,
		LastName:    in.User.LastName,
		Email:       in.User.Email,
		PhoneNumber: in.User.PhoneNumber,
		Memberships: memberships,
	}
}
