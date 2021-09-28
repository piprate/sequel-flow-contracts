package service

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
)

type (
	SequelFlowService struct {
		client        *gwtf.GoWithTheFlow
		scriptsEngine *scripts.Engine
		sequelAccount *flowkit.Account
		gasLimit      uint64
	}
)

var _ iinft.SequelFlow = (*SequelFlowService)(nil)

func NewSequelFlowService(client *gwtf.GoWithTheFlow, sequelAccount *flowkit.Account) (*SequelFlowService, error) {
	eng, err := scripts.NewEngine(client, true)
	if err != nil {
		return nil, err
	}

	srv := &SequelFlowService{
		client:        client,
		scriptsEngine: eng,
		gasLimit:      9999,
		sequelAccount: sequelAccount,
	}

	return srv, nil
}

func (s *SequelFlowService) SealDigitalArt(metadata *iinft.Metadata) error {
	_, err := scripts.CreateSealDigitalArtTx(
		s.scriptsEngine.GetStandardScript("master_seal"), s.client, metadata).
		SignProposeAndPayAsAccount(s.sequelAccount).
		Gas(s.gasLimit).
		RunE()
	return err
}

func (s *SequelFlowService) MintDigitalArtEdition(artID string, amount int, recipient flow.Address) ([]*iinft.Edition, error) {
	var editions []*iinft.Edition

	events, err := s.client.Transaction(s.scriptsEngine.GetStandardScript("digitalart_mint")).
		SignProposeAndPayAsAccount(s.sequelAccount).
		Gas(s.gasLimit).
		StringArgument(artID).
		UInt64Argument(uint64(amount)).
		Argument(cadence.Address(recipient)).
		RunE()
	if err != nil {
		return nil, err
	}

	for _, e := range events {
		if e.Value.EventType.QualifiedIdentifier == "DigitalArt.Minted" {
			editions = append(editions, &iinft.Edition{
				ID:            int(e.Value.Fields[0].(cadence.UInt64)),
				Asset:         e.Value.Fields[1].ToGoValue().(string),
				EditionNumber: int(e.Value.Fields[2].(cadence.UInt64)),
			})
		}
	}

	return editions, nil
}

func (s *SequelFlowService) GetDigitalArtMetadata(acct flow.Address, id uint64) (*iinft.Metadata, error) {
	result, err := s.client.Services.Scripts.Execute(
		[]byte(s.scriptsEngine.GetStandardScript("digitalart_get_metadata")),
		[]cadence.Value{
			cadence.NewAddress(acct),
			cadence.NewUInt64(id),
		},
		"",
		s.client.Network)
	if err != nil {
		return nil, err
	}

	return iinft.ReadMetadata(result)
}

func (s *SequelFlowService) CreateDigitalArtCollection(recipient *flowkit.Account) error {
	_, err := s.client.Transaction(s.scriptsEngine.GetStandardScript("account_setup")).
		SignProposeAndPayAsAccount(recipient).
		Gas(s.gasLimit).
		RunE()
	return err
}
