package evergreen_test

import (
	"testing"

	"github.com/onflow/flow-go-sdk"
	. "github.com/piprate/sequel-flow-contracts/lib/go/iinft/evergreen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	evergreenAddress = flow.HexToAddress("0x01cf0e2f2f715450")
	artist           = flow.HexToAddress("0xf3fcd2c1a78f5eee")
)

func TestRoleToCadence_RoleFromCadence(t *testing.T) {
	sourceRole := &Role{
		ID:                        RoleArtist,
		Description:               "Test Role",
		InitialSaleCommission:     0.8,
		SecondaryMarketCommission: 0.05,
		Address:                   artist,
	}

	val, err := RoleToCadence(sourceRole, evergreenAddress)
	require.NoError(t, err)
	require.NotEmpty(t, val)
	role, err := RoleFromCadence(val)
	require.NoError(t, err)

	assert.Equal(t, sourceRole.ID, role.ID)
	assert.Equal(t, sourceRole.Description, role.Description)
	assert.Equal(t, sourceRole.InitialSaleCommission, role.InitialSaleCommission)
	assert.Equal(t, sourceRole.SecondaryMarketCommission, role.SecondaryMarketCommission)
	assert.Equal(t, sourceRole.Address, role.Address)
	assert.Empty(t, role.ReceiverPath)

	// now, with a non-empty receiverPath

	sourceRole = &Role{
		ID:                        RoleArtist,
		Description:               "Test Role",
		InitialSaleCommission:     0.8,
		SecondaryMarketCommission: 0.05,
		Address:                   artist,
		ReceiverPath:              "/public/Test",
	}

	val, err = RoleToCadence(sourceRole, evergreenAddress)
	require.NoError(t, err)
	require.NotEmpty(t, val)

	role, err = RoleFromCadence(val)
	require.NoError(t, err)

	assert.Equal(t, sourceRole.ID, role.ID)
	assert.Equal(t, sourceRole.Description, role.Description)
	assert.Equal(t, sourceRole.InitialSaleCommission, role.InitialSaleCommission)
	assert.Equal(t, sourceRole.SecondaryMarketCommission, role.SecondaryMarketCommission)
	assert.Equal(t, sourceRole.Address, role.Address)
	assert.Equal(t, sourceRole.ReceiverPath, role.ReceiverPath)
}

func TestProfileToCadence_ProfileFromCadence(t *testing.T) {
	sourceProfile := &Profile{
		ID:          "did:sequel:evergreen1",
		Description: "Test Profile",
		Roles: []*Role{
			{
				ID:                        RoleArtist,
				Description:               "Test Role",
				InitialSaleCommission:     0.8,
				SecondaryMarketCommission: 0.05,
				Address:                   artist,
			},
		},
	}

	val, err := ProfileToCadence(sourceProfile, evergreenAddress)
	require.NoError(t, err)
	require.NotEmpty(t, val)

	profile, err := ProfileFromCadence(val)
	require.NoError(t, err)

	assert.Equal(t, profile.ID, profile.ID)
	assert.Equal(t, sourceProfile.Description, profile.Description)
	assert.Equal(t, 1, len(profile.Roles))
	assert.Equal(t, sourceProfile.Roles[0].ID, profile.Roles[0].ID)
}
