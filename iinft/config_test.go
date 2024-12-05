package iinft_test

import (
	"testing"

	. "github.com/piprate/sequel-flow-contracts/iinft"
	"github.com/stretchr/testify/require"
)

func TestLoadFlowKitAccount(t *testing.T) {
	acct, err := LoadFlowKitAccount("f669cb8d41ce0c74", "80025f0d1f2fd1ba0e18f447681fdd6a68a62ea86c2c2fefa811df086d40db3c")
	require.NoError(t, err)
	require.NotEmpty(t, acct)
}
