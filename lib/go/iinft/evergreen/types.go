package evergreen

import "github.com/onflow/flow-go-sdk"

const (
	RoleArtist     = "Artist"
	RolePlatform   = "Platform"
	RoleGreenCause = "Green"
	RoleCollector  = "Collector"
)

type (
	Role struct {
		Role                      string       `json:"role"`
		InitialSaleCommission     float64      `json:"initialSaleCommission,omitempty"`
		SecondaryMarketCommission float64      `json:"secondaryMarketCommission,omitempty"`
		Address                   flow.Address `json:"addr,omitempty"`
	}

	Profile struct {
		ID    uint32  `json:"id"`
		Roles []*Role `json:"roles"`
	}
)
