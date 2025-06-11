package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/ghodss/yaml"
	"github.com/invopop/jsonschema"
	"github.com/mcuadros/go-defaults"
	log "github.com/sirupsen/logrus"
)

var (
	configExample = "config_example.yaml"
	schemaLine    = `# yaml-language-server: $schema=https://raw.githubusercontent.com/tyzbit/acars-processor/refs/heads/main/schema.json
# This file (and schema.json) are automatically generated 
# from the code by running ./acars-processor -s

`
)

// These are called when jsonschema Reflects, so we don't need to call these.
func (ACARSAnnotatorConfig) JSONSchemaExtend(j *jsonschema.Schema) {
	s, ok := j.Properties.Get("SelectedFields")
	if !ok {
		log.Error(yo.Uhh("couldn't get selectedfields for acars annotator config type"))
		return
	}
	a := ACARSAnnotatorHandler{}
	fields := []string{}
	for field := range a.AnnotateACARSMessage(ACARSMessage{}) {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	s.Examples = append(s.Examples, fields)
	j.Properties.Set("SelectedFields", s)
}

func (VDLM2AnnotatorConfig) JSONSchemaExtend(j *jsonschema.Schema) {
	s, ok := j.Properties.Get("SelectedFields")
	if !ok {
		log.Error(yo.Uhh("couldn't get selectedfields for vdlm2 annotator config type"))
		return
	}
	a := VDLM2AnnotatorHandler{}
	fields := []string{}
	for field := range a.AnnotateVDLM2Message(VDLM2Message{}) {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	s.Examples = append(s.Examples, fields)
	j.Properties.Set("SelectedFields", s)
}

func (OllamaAnnotatorConfig) JSONSchemaExtend(j *jsonschema.Schema) {
	s, ok := j.Properties.Get("SelectedFields")
	if !ok {
		log.Error(yo.Uhh("couldn't get selectedfields for ollama annotator config type"))
		return
	}
	a := OllamaAnnotatorHandler{}
	// For Ollama, ACARS and VDLM2 fields are the same
	// This is not necessarily true for all annotators
	fields := []string{}
	for field := range a.AnnotateACARSMessage(ACARSMessage{}) {
		fields = append(fields, field)
	}
	for field := range a.AnnotateVDLM2Message(VDLM2Message{}) {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	s.Examples = append(s.Examples, fields)
	j.Properties.Set("SelectedFields", s)
}

func (Tar1090AnnotatorConfig) JSONSchemaExtend(j *jsonschema.Schema) {
	s, ok := j.Properties.Get("SelectedFields")
	if !ok {
		log.Error(yo.Uhh("couldn't get selectedfields for tar1090 annotator config type"))
		return
	}
	a := Tar1090AnnotatorHandler{}
	// For tar1090, ACARS and VDLM2 fields are the same
	// This is not necessarily true for all annotators
	fields := []string{}
	for field := range a.AnnotateACARSMessage(ACARSMessage{}) {
		fields = append(fields, field)
	}
	for field := range a.AnnotateVDLM2Message(VDLM2Message{}) {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	s.Examples = append(s.Examples, fields)
	j.Properties.Set("SelectedFields", s)
}

func (ADSBExchangeAnnotatorConfig) JSONSchemaExtend(j *jsonschema.Schema) {
	s, ok := j.Properties.Get("SelectedFields")
	if !ok {
		log.Error(yo.Uhh("couldn't get selectedfields for vdlm2 annotator config type"))
		return
	}
	a := ADSBAnnotatorHandler{}
	// For adsb, ACARS and VDLM2 fields are the same
	// This is not necessarily true for all annotators
	fields := []string{}
	for field := range a.AnnotateACARSMessage(ACARSMessage{}) {
		fields = append(fields, field)
	}
	for field := range a.AnnotateVDLM2Message(VDLM2Message{}) {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	s.Examples = append(s.Examples, fields)
	j.Properties.Set("SelectedFields", s)

}

func GenerateSchema(schemaPath string) {
	var configUpdated, schemaUpdated bool
	log.Info(yo.FYI("Generating schema").FRFR())
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

	// Squelch errors
	log.SetLevel(log.FatalLevel)
	// Get defaults for the example config for all fields with SelectFields
	defaultConfig.Annotators.ACARS.SelectedFields = ACARSAnnotatorHandler{}.DefaultFields()
	defaultConfig.Annotators.VDLM2.SelectedFields = VDLM2AnnotatorHandler{}.DefaultFields()
	defaultConfig.Annotators.ADSBExchange.SelectedFields = ADSBAnnotatorHandler{}.DefaultFields()
	defaultConfig.Annotators.Ollama.SelectedFields = OllamaAnnotatorHandler{}.DefaultFields()
	defaultConfig.Annotators.Tar1090.SelectedFields = Tar1090AnnotatorHandler{}.DefaultFields()

	// Set the values for defaultConfig to the defaults defined in the struct tags
	defaults.SetDefaults(&defaultConfig)
	configYaml, err := yaml.Marshal(defaultConfig)
	if err != nil {
		// Squelch errors
		log.Fatal(yo.Uhh("Error marshaling YAML:", err).FRFR())
	}
	// Add the schema line so editors can use it
	configYaml = append([]byte(schemaLine), configYaml...)
	if UpdateFile(configExample, configYaml) {
		configUpdated = true
		log.SetLevel(log.InfoLevel)
		log.Info(yo.Bet("Updated example config").FRFR())
		log.SetLevel(log.FatalLevel)
	}

	// First we generate the schema from the Config type with comments
	r := new(jsonschema.Reflector)
	err = r.AddGoComments("main", "./", jsonschema.WithFullComment())
	if err != nil {
		log.Fatal(yo.Uhh("unable to add comments to schema, %s", err).FRFR())
	}

	// Now we generate the schema and save it
	r.RequiredFromJSONSchemaTags = true
	schema := r.Reflect(&Config{})
	// Suppress further for clean output
	log.SetLevel(log.InfoLevel)
	json, _ := schema.MarshalJSON()
	if UpdateFile(fmt.Sprintf("./%s", schemaFilePath), json) {
		schemaUpdated = true
		log.Info(yo.Bet("Updated schema at %s", schemaFilePath).FRFR())
	}
	if configUpdated || schemaUpdated {
		log.Info(yo.Hmm("Files were updated, so exiting with status of 100").FRFR())
		os.Exit(100)
	}
	log.Info(yo.FYI("Schema and example config are up to date").FRFR())
}
