package contracts

//go:generate go run github.com/kevinburke/go-bindata/go-bindata -prefix ../../../../contracts -o internal/assets/assets.go -pkg assets -nometadata -nomemcopy ../../../../contracts

import (
	"regexp"

	_ "github.com/kevinburke/go-bindata"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/contracts/internal/assets"
)

const (
	digitalArtFile = "DigitalArt.cdc"
)

var (
	nftAddressPlaceholder = regexp.MustCompile(`"[^"\s].*/NonFungibleToken.cdc"`)
)

func GenerateDigitalArtContract(nftAddr string) []byte {

	code := assets.MustAssetString(digitalArtFile)

	codeWithNFTAddr := nftAddressPlaceholder.ReplaceAllString(code, "0x"+nftAddr)

	return []byte(codeWithNFTAddr)
}
