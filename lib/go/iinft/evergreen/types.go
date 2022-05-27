package evergreen

import (
	"github.com/onflow/flow-go-sdk"
)

const (
	RoleArtist    = "Artist"
	RolePlatform  = "Platform"
	RoleCollector = "Collector"
	RoleOwner     = "Owner"
)

type (
	Role struct {
		ID                        string       `json:"id"`
		Description               string       `json:"description"`
		InitialSaleCommission     float64      `json:"initialSaleCommission,omitempty"`
		SecondaryMarketCommission float64      `json:"secondaryMarketCommission,omitempty"`
		Address                   flow.Address `json:"addr,omitempty"`
		ReceiverPath              string       `json:"receiverPath,omitempty"`
	}

	Profile struct {
		ID          string  `json:"id"`
		Description string  `json:"description"`
		Roles       []*Role `json:"roles"`
	}
)
