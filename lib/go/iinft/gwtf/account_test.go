package gwtf

import (
	"context"
	"testing"

	"github.com/onflow/flowkit/v2/output"
	"github.com/stretchr/testify/assert"
)

func TestErrorsInAccountCreation(t *testing.T) {

	t.Run("Should give error on wrong saAccount name", func(t *testing.T) {
		g := NewGoWithTheFlow([]string{"../../../../flow.json"}, "emulator", true, output.NoneLog)
		_, err := g.CreateAccountsE(context.Background(), "foobar")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not find account with name foobar")
	})

	t.Run("Should give erro on wrong account name", func(t *testing.T) {
		_, err := NewGoWithTheFlowError([]string{"fixtures/invalid_account_in_deployment.json"}, "emulator", true, output.NoneLog)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deployment contains nonexisting account emulator-firs")
	})

	t.Run("Should fail when creating local accounts in the wrong order", func(t *testing.T) {
		g := NewGoWithTheFlow([]string{"fixtures/wrong_account_order_emulator.json"}, "emulator", true, output.NoneLog)
		_, err := g.CreateAccountsE(context.Background(), "emulator-first")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not find account with address 179b6b1cb6755e3")
	})
}
