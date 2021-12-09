package scripts

import (
	"fmt"

	_ "github.com/kevinburke/go-bindata"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
)

func CreateSealDigitalArtTx(script string, client *gwtf.GoWithTheFlow, metadata *iinft.Metadata) gwtf.FlowTransactionBuilder {
	tx := client.Transaction(script).
		StringArgument(metadata.MetadataLink).
		StringArgument(metadata.Name).
		StringArgument(metadata.Artist).
		StringArgument(metadata.Description).
		StringArgument(metadata.Type).
		StringArgument(metadata.ContentLink).
		StringArgument(metadata.ContentPreviewLink).
		StringArgument(metadata.Mimetype).
		UInt64Argument(metadata.MaxEdition).
		StringArgument(metadata.Asset).
		StringArgument(metadata.Record).
		StringArgument(metadata.AssetHead).
		UInt32Argument(metadata.ParticipationProfile.ID)

	artistRole, ok := metadata.ParticipationProfile.Roles[iinft.ParticipationRoleArtist]
	if ok {
		tx = tx.Argument(cadence.NewOptional(cadence.Address(artistRole.Address))).
			UFix64Argument(fmt.Sprintf("%.4f", artistRole.InitialSaleCommission)).
			UFix64Argument(fmt.Sprintf("%.4f", artistRole.SecondaryMarketCommission))
	} else {
		tx = tx.Argument(cadence.NewOptional(nil)).
			UFix64Argument("0.0").
			UFix64Argument("0.0")
	}

	platformRole, ok := metadata.ParticipationProfile.Roles[iinft.ParticipationRolePlatform]
	if ok {
		tx = tx.Argument(cadence.NewOptional(cadence.Address(platformRole.Address))).
			UFix64Argument(fmt.Sprintf("%.4f", platformRole.InitialSaleCommission)).
			UFix64Argument(fmt.Sprintf("%.4f", platformRole.SecondaryMarketCommission))
	} else {
		tx = tx.Argument(cadence.NewOptional(nil)).
			UFix64Argument("0.0").
			UFix64Argument("0.0")
	}

	return tx
}

func CreateMintSingleDigitalArtTx(script string, client *gwtf.GoWithTheFlow, metadata *iinft.Metadata, recipient flow.Address) gwtf.FlowTransactionBuilder {
	tx := client.Transaction(script).
		StringArgument(metadata.MetadataLink).
		StringArgument(metadata.Name).
		StringArgument(metadata.Artist).
		StringArgument(metadata.Description).
		StringArgument(metadata.Type).
		StringArgument(metadata.ContentLink).
		StringArgument(metadata.ContentPreviewLink).
		StringArgument(metadata.Mimetype).
		StringArgument(metadata.Asset).
		StringArgument(metadata.Record).
		StringArgument(metadata.AssetHead).
		UInt32Argument(metadata.ParticipationProfile.ID)

	artistRole, ok := metadata.ParticipationProfile.Roles[iinft.ParticipationRoleArtist]
	if ok {
		tx = tx.Argument(cadence.NewOptional(cadence.Address(artistRole.Address))).
			UFix64Argument(fmt.Sprintf("%.4f", artistRole.InitialSaleCommission)).
			UFix64Argument(fmt.Sprintf("%.4f", artistRole.SecondaryMarketCommission))
	} else {
		tx = tx.Argument(cadence.NewOptional(nil)).
			UFix64Argument("0.0").
			UFix64Argument("0.0")
	}

	platformRole, ok := metadata.ParticipationProfile.Roles[iinft.ParticipationRolePlatform]
	if ok {
		tx = tx.Argument(cadence.NewOptional(cadence.Address(platformRole.Address))).
			UFix64Argument(fmt.Sprintf("%.4f", platformRole.InitialSaleCommission)).
			UFix64Argument(fmt.Sprintf("%.4f", platformRole.SecondaryMarketCommission))
	} else {
		tx = tx.Argument(cadence.NewOptional(nil)).
			UFix64Argument("0.0").
			UFix64Argument("0.0")
	}

	tx = tx.Argument(cadence.Address(recipient))

	return tx
}
