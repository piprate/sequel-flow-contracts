package main

/*
   This utility creates all accounts mentioned in the deployment section of flow.json
   to local emulator instance. This is for development purposes only.
*/

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flowkit/v2/config"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/splash"
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

	client, err := splash.NewNetworkConnector(
		config.DefaultPaths(),
		splash.NewFileSystemLoader("."),
		"emulator",
		splash.NewZeroLogger())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create network connector")
		os.Exit(-1)
	}

	client.CreateAccounts(context.Background(), "emulator-account")

	se, err := iinft.NewTemplateEngine(client)
	if err != nil {
		os.Exit(-1)
	}

	adminAcct := client.Account(sequelAdminName)

	_, err = se.NewTransaction("account_fund_flow").
		Argument(cadence.NewAddress(adminAcct.Address)).
		UFix64Argument(amount).
		SignProposeAndPayAsService().
		RunE(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to execute account funding transaction")
		os.Exit(-1)
	}
}
