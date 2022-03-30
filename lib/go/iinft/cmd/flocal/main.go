package main

/*
	This utility creates all accounts mentioned in the deployment section of flow.json
    to local emulator instance. This is for development purposes only.
*/

import (
	"os"
	"time"

	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})

	client, err := iinft.NewGoWithTheFlowFS(".", "emulator", false)
	if err != nil {
		os.Exit(-1)
	}

	client.CreateAccounts("emulator-account")
}
