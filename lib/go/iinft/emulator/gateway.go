/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package emulator

import (
	"fmt"
	"time"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"

	emulator "github.com/onflow/flow-emulator"
	flowGo "github.com/onflow/flow-go/model/flow"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/client/convert"
)

/*
	Gateway is a copy of github.com/onflow/flow-cli/pkg/flowkit/gateway/EmulatorGateway
	with an added option to enable transaction fees
*/
type Gateway struct {
	emulator *emulator.Blockchain
}

func NewGateway(serviceAccount *flowkit.Account, enableTxFees bool) *Gateway {
	return &Gateway{
		emulator: newEmulator(serviceAccount, enableTxFees),
	}
}

func newEmulator(serviceAccount *flowkit.Account, enableTxFees bool) *emulator.Blockchain {
	var opts []emulator.Option
	if serviceAccount != nil && serviceAccount.Key().Type() == config.KeyTypeHex {
		privKey, _ := serviceAccount.Key().PrivateKey()

		opts = append(opts, emulator.WithServicePublicKey(
			(*privKey).PublicKey(),
			serviceAccount.Key().SigAlgo(),
			serviceAccount.Key().HashAlgo(),
		))
	}

	if enableTxFees {
		opts = append(opts, emulator.WithTransactionFeesEnabled(enableTxFees))
	}

	b, err := emulator.NewBlockchain(opts...)
	if err != nil {
		panic(err)
	}

	return b
}

func (g *Gateway) GetAccount(address flow.Address) (*flow.Account, error) {
	return g.emulator.GetAccount(address)
}

func (g *Gateway) SendSignedTransaction(tx *flowkit.Transaction) (*flow.Transaction, error) {
	t := tx.FlowTransaction()
	err := g.emulator.AddTransaction(*t)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	_, err = g.emulator.ExecuteNextTransaction()
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	_, err = g.emulator.CommitBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	return t, nil
}

func (g *Gateway) GetTransactionResult(tx *flow.Transaction, waitSeal bool) (*flow.TransactionResult, error) {
	result, err := g.emulator.GetTransactionResult(tx.ID())
	if err != nil {
		return nil, err
	}

	if result.Status != flow.TransactionStatusSealed && waitSeal {
		time.Sleep(time.Second)
		return g.GetTransactionResult(tx, waitSeal)
	}

	return result, nil
}

func (g *Gateway) GetTransaction(id flow.Identifier) (*flow.Transaction, error) {
	return g.emulator.GetTransaction(id)
}

func (g *Gateway) Ping() error {
	return nil
}

func (g *Gateway) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {
	args, err := convert.CadenceValuesToMessages(arguments)
	if err != nil {
		return nil, err
	}

	result, err := g.emulator.ExecuteScript(script, args)
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return result.Value, nil
}

func (g *Gateway) GetLatestBlock() (*flow.Block, error) {
	block, err := g.emulator.GetLatestBlock()
	if err != nil {
		return nil, err
	}

	return convertBlock(block), nil
}

func convertBlock(block *flowGo.Block) *flow.Block {
	return &flow.Block{
		BlockHeader: flow.BlockHeader{
			ID:        flow.Identifier(block.Header.ID()),
			ParentID:  flow.Identifier(block.Header.ParentID),
			Height:    block.Header.Height,
			Timestamp: block.Header.Timestamp,
		},
		BlockPayload: flow.BlockPayload{
			CollectionGuarantees: nil,
			Seals:                nil,
		},
	}
}

func (g *Gateway) GetEvents(
	eventType string,
	startHeight uint64,
	endHeight uint64,
) ([]client.BlockEvents, error) {
	events := make([]client.BlockEvents, 0)

	for height := startHeight; height <= endHeight; height++ {
		events = append(events, g.getBlockEvent(height, eventType))
	}

	return events, nil
}

func (g *Gateway) getBlockEvent(height uint64, eventType string) client.BlockEvents {
	events, _ := g.emulator.GetEventsByHeight(height, eventType)
	block, _ := g.emulator.GetBlockByHeight(height)

	flowEvents := make([]flow.Event, 0)

	for _, e := range events {
		flowEvents = append(flowEvents, flow.Event{
			Type:             e.Type,
			TransactionID:    e.TransactionID,
			TransactionIndex: e.TransactionIndex,
			EventIndex:       e.EventIndex,
			Value:            e.Value,
		})
	}

	return client.BlockEvents{
		BlockID:        flow.Identifier(block.Header.ID()),
		Height:         block.Header.Height,
		BlockTimestamp: block.Header.Timestamp,
		Events:         flowEvents,
	}
}

func (g *Gateway) GetCollection(id flow.Identifier) (*flow.Collection, error) {
	return g.emulator.GetCollection(id)
}

func (g *Gateway) GetBlockByID(id flow.Identifier) (*flow.Block, error) {
	block, err := g.emulator.GetBlockByID(id)
	return convertBlock(block), err
}

func (g *Gateway) GetBlockByHeight(height uint64) (*flow.Block, error) {
	block, err := g.emulator.GetBlockByHeight(height)
	return convertBlock(block), err
}

// GetLatestProtocolStateSnapshot placeholder func to complete gateway interface implementation
func (g *Gateway) GetLatestProtocolStateSnapshot() ([]byte, error) {
	return []byte{}, nil
}

// SecureConnection placeholder func to complete gateway interface implementation
func (g *Gateway) SecureConnection() bool {
	return false
}
