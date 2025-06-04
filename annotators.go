package main

import log "github.com/sirupsen/logrus"

func ConfigureAnnotators() {
	// ACARS-type messages
	if config.Annotators.ACARS.Enabled {
		log.Info("ACARS annotator enabled")
		enabledACARSAnnotators = append(enabledACARSAnnotators, ACARSHandlerAnnotator{})
	}
	if config.Annotators.ADSBExchange.APIKey != "" {
		log.Info("ADSB annotator enabled")
		if config.Annotators.ADSBExchange.APIKey == "" {
			log.Error("ADSB API key not set")
		}
		enabledACARSAnnotators = append(enabledACARSAnnotators, ADSBHandlerAnnotator{})
	}
	if config.Annotators.Tar1090.URL != "" {
		log.Info("TAR1090 VDLM2 annotator enabled")
		enabledACARSAnnotators = append(enabledACARSAnnotators, Tar1090Handler{})
	}
	if config.Annotators.Ollama.URL != "" && config.ACARSHub.ACARS.Host != "" {
		log.Info("Ollama ACARS annotator enabled")
		enabledACARSAnnotators = append(enabledACARSAnnotators, OllamaHandler{})
	}
	if len(enabledACARSAnnotators) == 0 {
		log.Warn("no acars annotators are enabled")
	}

	// VDLM2-type messages
	if config.Annotators.VDLM2.Enabled {
		log.Info("VDLM2 annotator enabled")
		enabledVDLM2Annotators = append(enabledVDLM2Annotators, VDLM2HandlerAnnotator{})
	}
	if config.Annotators.Tar1090.URL != "" {
		log.Info("TAR1090 VDLM2 annotator enabled")
		enabledVDLM2Annotators = append(enabledVDLM2Annotators, Tar1090Handler{})
	}
	if config.Annotators.Ollama.URL != "" && config.ACARSHub.VDLM2.Host != "" {
		log.Info("Ollama VDLM2 annotator enabled")
		enabledVDLM2Annotators = append(enabledVDLM2Annotators, OllamaHandler{})
	}
	if len(enabledVDLM2Annotators) == 0 {
		log.Info("no vdlm2 annotators are enabled")
	}
}
