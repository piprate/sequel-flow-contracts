package scripts

import (
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/evergreen"
)

type (
	MintOnDemandParameters struct {
		Metadata *iinft.DigitalArtMetadata
		Profile  *evergreen.Profile
	}
)
