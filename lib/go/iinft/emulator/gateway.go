/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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
	"context"
	"fmt"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-emulator/convert/sdk"
	"google.golang.org/grpc/status"

	"github.com/onflow/flow-emulator/server/backend"
	"github.com/onflow/flow-go-sdk"
	flowGo "github.com/onflow/flow-go/model/flow"
	"github.com/sirupsen/logrus"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
)

/*
Gateway is a copy of github.com/onflow/flow-cli/pkg/flowkit/gateway/EmulatorGateway
with an added option to enable transaction fees
*/
type Gateway struct {
	emulator     *emulator.Blockchain
	backend      *backend.Backend
	ctx          context.Context
	logger       *logrus.Logger
	enableTxFees bool
}

func UnwrapStatusError(err error) error {
	return fmt.Errorf(status.Convert(err).Message())
}

func NewGateway(serviceAccount *flowkit.Account) *Gateway {
	return NewGatewayWithOpts(serviceAccount)
}

func NewGatewayWithOpts(serviceAccount *flowkit.Account, opts ...func(*Gateway)) *Gateway {

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	gateway := &Gateway{
		ctx:    context.Background(),
		logger: logger,
	}
	for _, opt := range opts {
		opt(gateway)
	}
	gateway.emulator = newEmulator(serviceAccount, gateway.enableTxFees)
	gateway.backend = backend.New(gateway.logger, gateway.emulator)
	gateway.backend.EnableAutoMine()

	return gateway
}

func WithLogger(logger *logrus.Logger) func(g *Gateway) {
	return func(g *Gateway) {
		g.logger = logger
	}
}

func WithContext(ctx context.Context) func(g *Gateway) {
	return func(g *Gateway) {
		g.ctx = ctx
	}
}

func WithTransactionFees() func(g *Gateway) {
	return func(g *Gateway) {
		g.enableTxFees = true
	}
}

func (g *Gateway) SetContext(ctx context.Context) {
	g.ctx = ctx
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
	account, err := g.backend.GetAccount(g.ctx, address)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return account, nil
}

func (g *Gateway) SendSignedTransaction(tx *flowkit.Transaction) (*flow.Transaction, error) {
	err := g.backend.SendTransaction(context.Background(), *tx.FlowTransaction())
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return tx.FlowTransaction(), nil
}

func (g *Gateway) GetTransactionResult(tx *flow.Transaction, waitSeal bool) (*flow.TransactionResult, error) {
	result, err := g.backend.GetTransactionResult(g.ctx, tx.ID())
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return result, nil
}

func (g *Gateway) GetTransaction(id flow.Identifier) (*flow.Transaction, error) {
	transaction, err := g.backend.GetTransaction(g.ctx, id)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return transaction, nil
}

func (g *Gateway) Ping() error {
	err := g.backend.Ping(g.ctx)
	if err != nil {
		return UnwrapStatusError(err)
	}
	return nil
}

func (g *Gateway) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {

	args, err := cadenceValuesToMessages(arguments)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}

	result, err := g.backend.ExecuteScriptAtLatestBlock(g.ctx, script, args)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}

	value, err := messageToCadenceValue(result)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}

	return value, nil
}

func (g *Gateway) GetLatestBlock() (*flow.Block, error) {
	block, err := g.backend.GetLatestBlock(g.ctx, true)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}

	return convertBlock(block), nil
}

func cadenceValuesToMessages(values []cadence.Value) ([][]byte, error) {
	msgs := make([][]byte, len(values))
	for i, val := range values {
		msg, err := jsoncdc.Encode(val)
		if err != nil {
			return nil, fmt.Errorf("convert: %w", err)
		}
		msgs[i] = msg
	}
	return msgs, nil
}

func messageToCadenceValue(m []byte) (cadence.Value, error) {
	v, err := jsoncdc.Decode(nil, m)
	if err != nil {
		return nil, fmt.Errorf("convert: %w", err)
	}

	return v, nil
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
) ([]flow.BlockEvents, error) {
	events := make([]flow.BlockEvents, 0)

	for height := startHeight; height <= endHeight; height++ {
		events = append(events, g.getBlockEvent(height, eventType))
	}

	return events, nil
}

func (g *Gateway) getBlockEvent(height uint64, eventType string) flow.BlockEvents {
	block, _ := g.backend.GetBlockByHeight(g.ctx, height)
	events, _ := g.backend.GetEventsForBlockIDs(g.ctx, eventType, []flow.Identifier{flow.Identifier(block.ID())})

	result := flow.BlockEvents{
		BlockID:        flow.Identifier(block.ID()),
		Height:         block.Header.Height,
		BlockTimestamp: block.Header.Timestamp,
		Events:         []flow.Event{},
	}

	for _, e := range events {
		if e.BlockID == block.ID() {
			result.Events, _ = sdk.FlowEventsToSDK(e.Events)
			return result
		}
	}

	return result
}

func (g *Gateway) GetCollection(id flow.Identifier) (*flow.Collection, error) {
	collection, err := g.backend.GetCollectionByID(g.ctx, id)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return collection, nil
}

func (g *Gateway) GetBlockByID(id flow.Identifier) (*flow.Block, error) {
	block, err := g.backend.GetBlockByID(g.ctx, id)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return convertBlock(block), nil
}

func (g *Gateway) GetBlockByHeight(height uint64) (*flow.Block, error) {
	block, err := g.backend.GetBlockByHeight(g.ctx, height)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return convertBlock(block), nil
}

func (g *Gateway) GetLatestProtocolStateSnapshot() ([]byte, error) {
	snapshot, err := g.backend.GetLatestProtocolStateSnapshot(g.ctx)
	if err != nil {
		return nil, UnwrapStatusError(err)
	}
	return snapshot, nil
}

// SecureConnection placeholder func to complete gateway interface implementation
func (g *Gateway) SecureConnection() bool {
	return false
}
