package scripts

import (
	"bytes"
	"embed"
	"errors"
	"path"
	"text/template"

	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
)

//go:embed templates
var templateFS embed.FS
var goTemplates *template.Template

func init() {
	var err error
	goTemplates, err = template.New("").ParseFS(templateFS, "templates/transactions/*.cdc", "templates/scripts/*.cdc")
	if err != nil {
		panic(err)
	}
}

type (
	Engine struct {
		NFTAddress        string
		FUSDAddress       string
		DigitalArtAddress string

		client *gwtf.GoWithTheFlow

		preloadedTemplates map[string]string
	}
)

func NewEngine(client *gwtf.GoWithTheFlow, preload bool) (*Engine, error) {
	eng := &Engine{
		client:             client,
		preloadedTemplates: make(map[string]string),
	}

	if err := eng.loadContractAddresses(); err != nil {
		return nil, err
	}

	return eng, nil
}

func (e *Engine) loadContractAddresses() error {
	contracts, err := e.client.State.DeploymentContractsByNetwork(e.client.Network)
	if err != nil {
		return err
	}
	sourceTarget := make(map[string]string)
	for _, contract := range contracts {
		sourceTarget[path.Base(contract.Source)] = contract.Target.String()
	}

	var ok bool
	e.NFTAddress, ok = sourceTarget["NonFungibleToken.cdc"]
	if !ok {
		return errors.New("address not found for contract NonFungibleToken")
	}
	e.FUSDAddress, ok = sourceTarget["FUSD.cdc"]
	if !ok {
		return errors.New("address not found for contract FUSD")
	}
	e.DigitalArtAddress, ok = sourceTarget["DigitalArt.cdc"]
	if !ok {
		return errors.New("address not found for contract NonFungibleToken")
	}

	return nil
}

func (e *Engine) GetStandardScript(scriptID string) string {
	s, found := e.preloadedTemplates[scriptID]
	if !found {
		buf := &bytes.Buffer{}
		if err := goTemplates.ExecuteTemplate(buf, scriptID, map[string]interface{}{
			"NFTAddress":   e.NFTAddress,
			"TokenAddress": e.DigitalArtAddress,
		}); err != nil {
			panic(err)
		}

		s = string(buf.Bytes())
		e.preloadedTemplates[scriptID] = s

	}

	return s
}

func (e *Engine) NewTransaction(scriptID string) gwtf.FlowTransactionBuilder {
	return e.client.Transaction(e.GetStandardScript(scriptID))
}

func (e *Engine) NewScript(scriptID string) gwtf.FlowScriptBuilder {
	return e.client.Script(e.GetStandardScript(scriptID))
}

func (e *Engine) NewInlineScript(script string) gwtf.FlowScriptBuilder {
	return e.client.Script(script)
}
