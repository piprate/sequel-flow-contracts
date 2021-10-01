package test

import (
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
)

// NewGoWithTheFlowEmbedded creates a new test go with the flow client based on embedded setup
func NewGoWithTheFlowEmbedded(network string, inMemory bool) (*gwtf.GoWithTheFlow, error) {
	return iinft.NewGoWithTheFlowError(&embeddedFileLoader{}, network, inMemory)
}
