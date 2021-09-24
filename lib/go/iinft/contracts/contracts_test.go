package contracts_test

import (
	"testing"

	. "github.com/piprate/sequel-flow-contracts/lib/go/iinft/contracts"
	"github.com/stretchr/testify/assert"
)

var addrA = "0A"

func TestGenerateDigitalArtContract(t *testing.T) {
	contract := GenerateDigitalArtContract(addrA)
	assert.NotNil(t, contract)
	assert.Contains(t, string(contract), addrA)
}
