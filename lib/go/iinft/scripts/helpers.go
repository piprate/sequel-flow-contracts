package scripts

import (
	_ "github.com/kevinburke/go-bindata"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/evergreen"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
)

func CreateSealDigitalArtTx(se *Engine, client *gwtf.GoWithTheFlow, metadata *iinft.Metadata,
	profile *evergreen.Profile) gwtf.FlowTransactionBuilder {

	tx := client.Transaction(se.GetStandardScript("master_seal")).
		Argument(iinft.MetadataToCadence(metadata, flow.HexToAddress(se.WellKnownAddresses()["DigitalArt"]))).
		Argument(evergreen.ProfileToCadence(profile, flow.HexToAddress(se.WellKnownAddresses()["Evergreen"])))

	return tx
}

func CreateMintSingleDigitalArtTx(se *Engine, client *gwtf.GoWithTheFlow, metadata *iinft.Metadata,
	profile *evergreen.Profile, recipient flow.Address) gwtf.FlowTransactionBuilder {

	metadataCpy := *metadata
	metadataCpy.Edition = 1
	metadataCpy.MaxEdition = 1

	tx := client.Transaction(se.GetStandardScript("digitalart_mint_single")).
		Argument(iinft.MetadataToCadence(&metadataCpy, flow.HexToAddress(se.WellKnownAddresses()["DigitalArt"]))).
		Argument(evergreen.ProfileToCadence(profile, flow.HexToAddress(se.WellKnownAddresses()["Evergreen"]))).
		Argument(cadence.Address(recipient))

	return tx
}
