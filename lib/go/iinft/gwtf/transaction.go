package gwtf

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/araddon/dateparse"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/transactions"
)

// TransactionFromFile will start a flow transaction builder
func (f *GoWithTheFlow) TransactionFromFile(filename string) FlowTransactionBuilder {
	return FlowTransactionBuilder{
		GoWithTheFlow:  f,
		FileName:       filename,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*accounts.Account{},
		GasLimit:       9999,
	}
}

// Transaction will start a flow transaction builder using the inline transaction
func (f *GoWithTheFlow) Transaction(content string) FlowTransactionBuilder {
	return FlowTransactionBuilder{
		GoWithTheFlow:  f,
		FileName:       "inline",
		Content:        content,
		Arguments:      []cadence.Value{},
		PayloadSigners: []*accounts.Account{},
		GasLimit:       9999,
	}
}

// Gas sets the gas limit for this transaction
func (tb FlowTransactionBuilder) Gas(limit uint64) FlowTransactionBuilder {
	tb.GasLimit = limit
	return tb
}

// ProposeAs sets the proposer
func (tb FlowTransactionBuilder) ProposeAs(proposer string) FlowTransactionBuilder {
	tb.Proposer = tb.GoWithTheFlow.Account(proposer)
	return tb
}

// PayAs sets the payer
func (tb FlowTransactionBuilder) PayAs(payer string) FlowTransactionBuilder {
	tb.Payer = tb.GoWithTheFlow.Account(payer)
	return tb
}

// SignAndProposeAs set the proposer and envelope signer
func (tb FlowTransactionBuilder) SignAndProposeAs(signer string) FlowTransactionBuilder {
	tb.Proposer = tb.GoWithTheFlow.Account(signer)
	tb.MainSigner = tb.Proposer
	return tb
}

// SignProposeAndPayAs set the payer, proposer and envelope signer
func (tb FlowTransactionBuilder) SignProposeAndPayAs(signer string) FlowTransactionBuilder {
	tb.Proposer = tb.GoWithTheFlow.Account(signer)
	tb.Payer = tb.Proposer
	tb.MainSigner = tb.Proposer
	return tb
}

// SignProposeAndPayAsService set the payer, proposer and envelope signer
func (tb FlowTransactionBuilder) SignProposeAndPayAsService() FlowTransactionBuilder {
	key := fmt.Sprintf("%s-account", tb.GoWithTheFlow.Services.Network().Name)
	account, err := tb.GoWithTheFlow.State.Accounts().ByName(key)
	if err != nil {
		log.Fatal(err)
	}
	tb.Proposer = account
	tb.Payer = tb.Proposer
	tb.MainSigner = tb.Proposer
	return tb
}

// RawAccountArgument add an account from a string as an argument
func (tb FlowTransactionBuilder) RawAccountArgument(key string) FlowTransactionBuilder {
	account := flow.HexToAddress(key)
	accountArg := cadence.BytesToAddress(account.Bytes())
	return tb.Argument(accountArg)
}

// AccountArgument add an account as an argument
func (tb FlowTransactionBuilder) AccountArgument(key string) FlowTransactionBuilder {
	f := tb.GoWithTheFlow

	account := f.Account(key)
	return tb.Argument(cadence.BytesToAddress(account.Address.Bytes()))
}

// StringArgument add a String Argument to the transaction
func (tb FlowTransactionBuilder) StringArgument(value string) FlowTransactionBuilder {
	return tb.Argument(cadence.String(value))
}

// BooleanArgument add a Boolean Argument to the transaction
func (tb FlowTransactionBuilder) BooleanArgument(value bool) FlowTransactionBuilder {
	return tb.Argument(cadence.NewBool(value))
}

// BytesArgument add a Bytes Argument to the transaction
func (tb FlowTransactionBuilder) BytesArgument(value []byte) FlowTransactionBuilder {
	return tb.Argument(cadence.NewBytes(value))
}

// IntArgument add an Int Argument to the transaction
func (tb FlowTransactionBuilder) IntArgument(value int) FlowTransactionBuilder {
	return tb.Argument(cadence.NewInt(value))
}

// Int8Argument add an Int8 Argument to the transaction
func (tb FlowTransactionBuilder) Int8Argument(value int8) FlowTransactionBuilder {
	return tb.Argument(cadence.NewInt8(value))
}

// Int16Argument add an Int16 Argument to the transaction
func (tb FlowTransactionBuilder) Int16Argument(value int16) FlowTransactionBuilder {
	return tb.Argument(cadence.NewInt16(value))
}

// Int32Argument add an Int32 Argument to the transaction
func (tb FlowTransactionBuilder) Int32Argument(value int32) FlowTransactionBuilder {
	return tb.Argument(cadence.NewInt32(value))
}

// Int64Argument add an Int64 Argument to the transaction
func (tb FlowTransactionBuilder) Int64Argument(value int64) FlowTransactionBuilder {
	return tb.Argument(cadence.NewInt64(value))
}

// Int128Argument add an Int128 Argument to the transaction
func (tb FlowTransactionBuilder) Int128Argument(value int) FlowTransactionBuilder {
	return tb.Argument(cadence.NewInt128(value))
}

// Int256Argument add an Int256 Argument to the transaction
func (tb FlowTransactionBuilder) Int256Argument(value int) FlowTransactionBuilder {
	return tb.Argument(cadence.NewInt256(value))
}

// UIntArgument add an UInt Argument to the transaction
func (tb FlowTransactionBuilder) UIntArgument(value uint) FlowTransactionBuilder {
	return tb.Argument(cadence.NewUInt(value))
}

// UInt8Argument add an UInt8 Argument to the transaction
func (tb FlowTransactionBuilder) UInt8Argument(value uint8) FlowTransactionBuilder {
	return tb.Argument(cadence.NewUInt8(value))
}

// UInt16Argument add an UInt16 Argument to the transaction
func (tb FlowTransactionBuilder) UInt16Argument(value uint16) FlowTransactionBuilder {
	return tb.Argument(cadence.NewUInt16(value))
}

// UInt32Argument add an UInt32 Argument to the transaction
func (tb FlowTransactionBuilder) UInt32Argument(value uint32) FlowTransactionBuilder {
	return tb.Argument(cadence.NewUInt32(value))
}

// UInt64Argument add an UInt64 Argument to the transaction
func (tb FlowTransactionBuilder) UInt64Argument(value uint64) FlowTransactionBuilder {
	return tb.Argument(cadence.NewUInt64(value))
}

// UInt128Argument add an UInt128 Argument to the transaction
func (tb FlowTransactionBuilder) UInt128Argument(value uint) FlowTransactionBuilder {
	return tb.Argument(cadence.NewUInt128(value))
}

// UInt256Argument add an UInt256 Argument to the transaction
func (tb FlowTransactionBuilder) UInt256Argument(value uint) FlowTransactionBuilder {
	return tb.Argument(cadence.NewUInt256(value))
}

// Word8Argument add a Word8 Argument to the transaction
func (tb FlowTransactionBuilder) Word8Argument(value uint8) FlowTransactionBuilder {
	return tb.Argument(cadence.NewWord8(value))
}

// Word16Argument add a Word16 Argument to the transaction
func (tb FlowTransactionBuilder) Word16Argument(value uint16) FlowTransactionBuilder {
	return tb.Argument(cadence.NewWord16(value))
}

// Word32Argument add a Word32 Argument to the transaction
func (tb FlowTransactionBuilder) Word32Argument(value uint32) FlowTransactionBuilder {
	return tb.Argument(cadence.NewWord32(value))
}

// Word64Argument add a Word64 Argument to the transaction
func (tb FlowTransactionBuilder) Word64Argument(value uint64) FlowTransactionBuilder {
	return tb.Argument(cadence.NewWord64(value))
}

// Fix64Argument add a Fix64 Argument to the transaction
func (tb FlowTransactionBuilder) Fix64Argument(value string) FlowTransactionBuilder {
	amount, err := cadence.NewFix64(value)
	if err != nil {
		panic(err)
	}
	return tb.Argument(amount)
}

// DateStringAsUnixTimestamp sends a dateString parsed in the timezone as a unix timezone ufix
func (tb FlowTransactionBuilder) DateStringAsUnixTimestamp(dateString string, timezone string) FlowTransactionBuilder {
	return tb.UFix64Argument(parseTime(dateString, timezone))
}

func parseTime(timeString, location string) string {
	loc, err := time.LoadLocation(location)
	if err != nil {
		panic(err)
	}

	time.Local = loc
	t, err := dateparse.ParseLocal(timeString)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%d.0", t.Unix())
}

// UFix64Argument add a UFix64 Argument to the transaction
func (tb FlowTransactionBuilder) UFix64Argument(value string) FlowTransactionBuilder {
	amount, err := cadence.NewUFix64(value)
	if err != nil {
		panic(err)
	}
	return tb.Argument(amount)
}

// Argument add an argument to the transaction
func (tb FlowTransactionBuilder) Argument(value cadence.Value) FlowTransactionBuilder {
	tb.Arguments = append(tb.Arguments, value)
	return tb
}

// StringArrayArgument add a string array argument to the transaction
func (tb FlowTransactionBuilder) StringArrayArgument(value ...string) FlowTransactionBuilder {
	array := make([]cadence.Value, len(value))
	for i, val := range value {
		v, err := cadence.NewString(val)
		if err != nil {
			panic(err)
		}
		array[i] = v
	}
	tb.Arguments = append(tb.Arguments, cadence.NewArray(array))
	return tb
}

// PayloadSigner set a signer for the payload
func (tb FlowTransactionBuilder) PayloadSigner(value string) FlowTransactionBuilder {
	signer := tb.GoWithTheFlow.Account(value)
	tb.PayloadSigners = append(tb.PayloadSigners, signer)
	return tb
}

// RunPrintEventsFull will run a transaction and print all events
func (tb FlowTransactionBuilder) RunPrintEventsFull(ctx context.Context) {
	PrintEvents(tb.Run(ctx), map[string][]string{})
}

// RunPrintEvents will run a transaction and print all events ignoring some fields
func (tb FlowTransactionBuilder) RunPrintEvents(ctx context.Context, ignoreFields map[string][]string) {
	PrintEvents(tb.Run(ctx), ignoreFields)
}

// Run runs the transaction
func (tb FlowTransactionBuilder) Run(ctx context.Context) []flow.Event {
	events, err := tb.RunE(ctx)
	if err != nil {
		tb.GoWithTheFlow.Logger.Error(fmt.Sprintf("Error executing script: %s output %v", tb.FileName, err))
		os.Exit(1)
	}
	return events
}

// RunE runs returns error
func (tb FlowTransactionBuilder) RunE(ctx context.Context) ([]flow.Event, error) {

	if tb.Proposer == nil {
		return nil, errors.New("you need to set the proposer")
	}
	if tb.Payer == nil {
		return nil, errors.New("you need to set the payer")
	}

	codeFileName := fmt.Sprintf("./transactions/%s.cdc", tb.FileName)
	code, err := tb.getContractCode(codeFileName)
	if err != nil {
		return nil, err
	}

	signerNames := map[string]bool{}
	authorizerAccounts := tb.PayloadSigners
	if tb.MainSigner != nil {
		authorizerAccounts = append(authorizerAccounts, tb.MainSigner)
	}

	authorizers := make([]flow.Address, len(authorizerAccounts))
	var signers []*accounts.Account
	for i, signer := range authorizerAccounts {
		authorizers[i] = signer.Address
		if signer.Name != tb.Payer.Name {
			signers = append(signers, signer)
		}
		signerNames[signer.Name] = true
	}

	if _, found := signerNames[tb.Proposer.Name]; !found && tb.Proposer.Name != tb.Payer.Name {
		return nil, errors.New("proposer doesn't match any authorizers or the payer")
	}

	// we append the Payer at the end here so that it signs last
	signers = append(signers, tb.Payer)

	tx, err := tb.GoWithTheFlow.Services.BuildTransaction(
		ctx,
		transactions.AddressesRoles{
			Proposer:    tb.Proposer.Address,
			Authorizers: authorizers,
			Payer:       tb.Payer.Address,
		},
		tb.Proposer.Key.Index(),
		flowkit.Script{
			Code:     code,
			Args:     tb.Arguments,
			Location: codeFileName,
		},
		tb.GasLimit,
	)
	if err != nil {
		return nil, err
	}

	for _, signer := range signers {
		err = tx.SetSigner(signer)
		if err != nil {
			return nil, err
		}

		tx, err = tx.Sign()
		if err != nil {
			return nil, err
		}
	}

	_, res, err := tb.GoWithTheFlow.Services.SendSignedTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	if res.Error != nil {
		return nil, res.Error
	}

	tb.GoWithTheFlow.Logger.Debug(fmt.Sprintf("Transaction %s successfully applied", tx.FlowTransaction().ID()))
	return res.Events, nil
}

func (tb FlowTransactionBuilder) getContractCode(codeFileName string) ([]byte, error) {
	code := []byte(tb.Content)
	var err error
	if tb.Content == "" {
		code, err = tb.GoWithTheFlow.State.ReaderWriter().ReadFile(codeFileName)
		if err != nil {
			return nil, fmt.Errorf("could not read transaction file from path=%s", codeFileName)
		}
	}
	return code, nil
}

// FlowTransactionBuilder used to create a builder pattern for a transaction
type FlowTransactionBuilder struct {
	GoWithTheFlow  *GoWithTheFlow
	FileName       string
	Content        string
	Arguments      []cadence.Value
	Proposer       *accounts.Account
	Payer          *accounts.Account
	MainSigner     *accounts.Account
	PayloadSigners []*accounts.Account
	GasLimit       uint64
}
