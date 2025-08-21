package main

import (
	"strings"

	hue "codeberg.org/tyzbit/huenique"
	"github.com/mcuadros/go-defaults"
	log "github.com/sirupsen/logrus"
)

var (
	schemaLine = `# yaml-language-server: $schema=./schema.json
# This file (and schema.json) are automatically generated 
# from the code by running ./acars-processor -s

`
)

func GenerateDocs(exampleConfigPath string) (updated bool) {
	log.Info(Content("Generating schema"))
	// Generate an example config
	var defaultConfig Config
	// Since this is an array, we need to add an item so the defaults will
	// populate.
	defaultConfig.Steps = []ProcessingStep{{}}

	// We need to do this to set an example value since Options is a slice.
	defaultConfig.Steps[0].Send.Webhook.Headers = []WebHookReceiverHeaders{{
		Name:  "APIKey",
		Value: "1234abcdef",
	}}
	// We need to do this to set an example value since Options is a slice.
	defaultConfig.Steps[0].Annotate.Ollama.Options = []OllamaOptionsConfig{{
		Name:  "num_predict",
		Value: 512,
	}}
	// We need to do this to set an example value since Options is a slice.
	defaultConfig.Steps[0].Filter.Ollama.Options = []OllamaOptionsConfig{{
		Name:  "num_predict",
		Value: 512,
	}}
	defaultConfig.Steps[0].Send.Discord.EmbedColorGradientSteps = []hue.Color{
		{R: 0, G: 255, B: 0},
		{R: 255, G: 255, B: 0},
		{R: 255, G: 0, B: 0},
	}

	// You can select fields on ACARS JSON feeders
	ac := &defaultConfig.ACARSProcessorSettings.ACARSHub
	ac.ACARS.SelectedFields = ac.ACARS.GetDefaultFields()
	ac.VDLM2.SelectedFields = ac.VDLM2.GetDefaultFields()

	// You can also select them on Annotators
	a := &defaultConfig.Steps[0].Annotate.ADSB
	o := &defaultConfig.Steps[0].Annotate.Ollama
	t := &defaultConfig.Steps[0].Annotate.Tar1090
	a.SelectedFields = a.GetDefaultFields()
	o.SelectedFields = o.GetDefaultFields()
	t.SelectedFields = t.GetDefaultFields()

	defaults.SetDefaults(&defaultConfig)
	configYaml, err := MarshalYAMLWithComments(defaultConfig)
	// Ugly hack for pointers, this assumes everything is a *bool
	configYaml = []byte(strings.ReplaceAll(string(configYaml), "<nil>", "true"))
	if err != nil {
		// Squelch errors
		log.Fatal(Attention("Error marshaling YAML: %w", err))
	}

	// Add the schema line so editors can use it
	configYaml = append([]byte(schemaLine), configYaml...)
	if UpdateFile(exampleConfigPath, configYaml) {
		updated = true
		log.SetLevel(log.InfoLevel)
		log.Info(Success("Updated %s", exampleConfigPath))
		log.SetLevel(log.FatalLevel)
	}
	return updated
}
