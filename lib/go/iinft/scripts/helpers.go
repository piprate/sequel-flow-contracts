package scripts

import (
	_ "github.com/kevinburke/go-bindata"
	"github.com/onflow/cadence"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
)

func CreateSealDigitalArtTx(script string, client *gwtf.GoWithTheFlow, metadata *iinft.Metadata) gwtf.FlowTransactionBuilder {
	return client.Transaction(script).
		StringArgument(metadata.MetadataLink).
		StringArgument(metadata.Name).
		StringArgument(metadata.Artist).
		Argument(cadence.Address(metadata.ArtistAddress)).
		StringArgument(metadata.Description).
		StringArgument(metadata.Type).
		StringArgument(metadata.ContentLink).
		StringArgument(metadata.ContentPreviewLink).
		StringArgument(metadata.Mimetype).
		UInt64Argument(metadata.MaxEdition).
		StringArgument(metadata.Asset).
		StringArgument(metadata.Record).
		StringArgument(metadata.AssetHead)
}
