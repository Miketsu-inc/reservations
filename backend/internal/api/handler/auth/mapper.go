package auth

import authServ "github.com/miketsu-inc/reservations/backend/internal/service/auth"

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
