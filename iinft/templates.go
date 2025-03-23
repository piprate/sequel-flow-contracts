package iinft

import (
	"embed"

	"github.com/piprate/splash"
)

//go:embed templates
var templateFS embed.FS

var (
	requiredWellKnownContracts = []string{
		"Burner", "FlowToken", "FungibleToken", "FungibleTokenMetadataViews",
		"FungibleTokenSwitchboard", "MetadataViews", "NFTStorefront",
		"NonFungibleToken", "NFTCatalog", "NFTRetrieval",
		"Art", "Content",
		"Evergreen", "DigitalArt", "SequelMarketplace",
	}
)

func NewTemplateEngine(client *splash.Connector) (*splash.TemplateEngine, error) {
	return splash.NewTemplateEngine(client, templateFS, []string{}, requiredWellKnownContracts, "templates/transactions/*.cdc", "templates/scripts/*.cdc", "templates/scripts/**/*.cdc")
}
