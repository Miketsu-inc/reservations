package keys

import (
	"fmt"
)

type PasswordReset struct {
	Token string
}

func (k PasswordReset) String() string {
	return fmt.Sprintf("password_reset:%s", k.Token)
}
