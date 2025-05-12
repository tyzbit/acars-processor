package main

import log "github.com/sirupsen/logrus"

func ConfigureAnnotators() {
	// ACARS-type messages
	if config.AnnotateACARS {
		log.Info("ACARS annotator enabled")
		enabledACARSAnnotators = append(enabledACARSAnnotators, ACARSHandlerAnnotator{})
	}
	if config.ADSBExchangeAPIKey != "" {
		log.Info("ADSB annotator enabled")
		if config.ADSBExchangeAPIKey == "" {
			log.Error("ADSB API key not set")
		}
		enabledACARSAnnotators = append(enabledACARSAnnotators, ADSBHandlerAnnotator{})
	}
	if config.TAR1090URL != "" {
		log.Info("TAR1090 VDLM2 annotator enabled")
		enabledACARSAnnotators = append(enabledACARSAnnotators, Tar1090Handler{})
	}
	if config.OllamaAnnotatorURL != "" && config.ACARSHubHost != "" {
		log.Info("Ollama ACARS annotator enabled")
		enabledACARSAnnotators = append(enabledACARSAnnotators, OllamaHandler{})
	}
	if len(enabledACARSAnnotators) == 0 {
		log.Warn("no acars annotators are enabled")
	}

	// VDLM2-type messages
	if config.AnnotateVDLM2 {
		log.Info("VDLM2 annotator enabled")
		enabledVDLM2Annotators = append(enabledVDLM2Annotators, VDLM2HandlerAnnotator{})
	}
	if config.TAR1090URL != "" {
		log.Info("TAR1090 VDLM2 annotator enabled")
		enabledVDLM2Annotators = append(enabledVDLM2Annotators, Tar1090Handler{})
	}
	if config.OllamaAnnotatorURL != "" && config.ACARSHubVDLM2Host != "" {
		log.Info("Ollama VDLM2 annotator enabled")
		enabledVDLM2Annotators = append(enabledVDLM2Annotators, OllamaHandler{})
	}
	if len(enabledVDLM2Annotators) == 0 {
		log.Info("no vdlm2 annotators are enabled")
	}
}
