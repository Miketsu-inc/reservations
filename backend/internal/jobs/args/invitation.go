package args

import (
	"time"

	"github.com/riverqueue/river"
)

type HandleInvitationExpiration struct{}

func (HandleInvitationExpiration) Kind() string { return "handle_invitation_expiration" }

func (HandleInvitationExpiration) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByPeriod: time.Hour * 24,
		},
	}
}
