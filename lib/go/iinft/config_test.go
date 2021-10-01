package iinft_test

import (
	"testing"

	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/stretchr/testify/require"
)

func TestLoadFlowKitAccount(t *testing.T) {
	acct, err := iinft.LoadFlowKitAccount("01cf0e2f2f715450", "d5457a187e9642a8e49d4032b3b4f85c92da7202c79681d9302c6e444e7033a8")
	require.NoError(t, err)
	require.NotEmpty(t, acct)
}
