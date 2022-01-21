package scripts

import (
	_ "github.com/kevinburke/go-bindata"
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
