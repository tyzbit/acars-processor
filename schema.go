package main

import (
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/invopop/jsonschema"
	"github.com/mcuadros/go-defaults"
	log "github.com/sirupsen/logrus"
)

var (
	configExample = "config_example.yaml"
	schemaLine    = `# yaml-language-server: $schema=https://raw.githubusercontent.com/tyzbit/acars-processor/refs/heads/main/schema.json

`
)

func GenerateSchema(schemaPath string) {
	updated := false
	// Generate an example config
	var defaultConfig Config
	// We need to do this to set an example value since Options is a slice.
	defaultConfig.Receivers.Webhook.Headers = []WebHookReceiverConfigHeaders{{
		Name:  "APIKey",
		Value: "1234abcdef",
	}}
	// We need to do this to set an example value since Options is a slice.
	defaultConfig.Annotators.Ollama.Options = []OllamaOptionsConfig{{
		Name:  "num_predict",
		Value: 512,
	}}
	// We need to do this to set an example value since Options is a slice.
	defaultConfig.Filters.Ollama.Options = []OllamaOptionsConfig{{
		Name:  "num_predict",
		Value: 512,
	}}
	// Set the values for defaultConfig to the defaults defined in the struct tags
	defaults.SetDefaults(&defaultConfig)
	configYaml, err := yaml.Marshal(defaultConfig)
	if err != nil {
		log.Fatal("Error marshaling YAML:", err)
	}
	// Add the schema line so editors can use it
	configYaml = append([]byte(schemaLine), configYaml...)
	if UpdateFile(configExample, configYaml) {
		updated = true
		log.Info("Updated example config")
	}

	// First we generate the schema from the Config type with comments
	r := new(jsonschema.Reflector)
	if err := r.AddGoComments("github.com/tyzbit/acars-processor",
		"./config.go", jsonschema.WithFullComment()); err != nil {
		log.Fatalf("unable to add comments to schema, %s", err)
	}

	// Now we generate the schema and save it
	r.RequiredFromJSONSchemaTags = true
	schema := r.Reflect(&Config{})
	json, _ := schema.MarshalJSON()
	if UpdateFile(fmt.Sprintf("./%s", schemaFilePath), json) {
		updated = true
		log.Infof("Updated schema at %s", schemaFilePath)
	}
	if updated {
		log.Info("Files were updated, so exiting with status of 100")
		os.Exit(100)
	}
}
