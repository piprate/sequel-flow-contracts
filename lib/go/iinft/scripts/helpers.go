package scripts

import (
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/evergreen"
)

type (
	// MintOnDemandParameters provides inputs for "digitalart_mint_on_demand_flow" and
	// "digitalart_mint_on_demand_fusd" transaction templates.
	// If Metadata is nil, the transactions won't include checks if the master is sealed
	// (and sealing it, if it's not).
	MintOnDemandParameters struct {
		Metadata *iinft.DigitalArtMetadata
		Profile  *evergreen.Profile
	}
)
