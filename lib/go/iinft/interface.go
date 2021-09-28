package iinft

import (
	"github.com/onflow/flow-go-sdk"
)

type (
	Edition struct {
		ID            int    `json:"id"`
		Asset         string `json:"asset"`
		EditionNumber int    `json:"editionNumber,omitempty"`
	}

	SequelFlow interface {
		SealDigitalArt(metadata *Metadata) error
		MintDigitalArtEdition(artID string, amount int, recipient flow.Address) ([]*Edition, error)
		GetDigitalArtMetadata(acct flow.Address, id uint64) (*Metadata, error)
	}
)
