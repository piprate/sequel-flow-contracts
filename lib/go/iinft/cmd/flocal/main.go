package main

/*
   This utility creates all accounts mentioned in the deployment section of flow.json
   to local emulator instance. This is for development purposes only.
*/

import (
	"flag"
	"os"
	"time"

	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})

	var sequelAdminName string
	var amount string

	flag.StringVar(&sequelAdminName, "admin", "sequel-admin", "Specify Sequel admin account name. Default is emulator-sequel-admin")
	flag.StringVar(&amount, "amount", "100.0", "Specify Flow amount to deposit. Default is 100.0")

	flag.Parse()

	client, err := iinft.NewGoWithTheFlowFS(".", "emulator", false, false)
	if err != nil {
		os.Exit(-1)
	}

	client.CreateAccounts("emulator-account")

	adminAcct := client.Account(sequelAdminName)
	if err = scripts.FundAccountWithFlowE(client, adminAcct.Address(), amount); err != nil {
		panic(err)
	}
}
