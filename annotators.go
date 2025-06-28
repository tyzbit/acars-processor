package main

import log "github.com/sirupsen/logrus"

func ConfigureAnnotators() {
	// ACARS-type messages
	if config.Annotators.ACARS.Enabled {
		log.Info(Success("ACARS annotator enabled"))
		enabledACARSAnnotators = append(enabledACARSAnnotators, ACARSAnnotatorHandler{})
	}
	if config.Annotators.ADSBExchange.Enabled {
		log.Info(Success("ADSB annotator enabled"))
		if config.Annotators.ADSBExchange.APIKey == "" {
			log.Error(Attention("ADSB API key not set"))
		}
		enabledACARSAnnotators = append(enabledACARSAnnotators, ADSBAnnotatorHandler{})
	}
	if config.Annotators.Tar1090.Enabled {
		log.Info(Success("TAR1090 VDLM2 annotator enabled"))
		enabledACARSAnnotators = append(enabledACARSAnnotators, Tar1090AnnotatorHandler{})
	}
	if config.Annotators.Ollama.Enabled && config.ACARSProcessorSettings.ACARSHub.ACARS.Host != "" {
		log.Info(Success("Ollama ACARS annotator enabled"))
		enabledACARSAnnotators = append(enabledACARSAnnotators, OllamaAnnotatorHandler{})
	}
	if len(enabledACARSAnnotators) == 0 {
		log.Warn(Attention("no acars annotators are enabled"))
	}

	// VDLM2-type messages
	if config.Annotators.VDLM2.Enabled {
		log.Info(Success("VDLM2 annotator enabled"))
		enabledVDLM2Annotators = append(enabledVDLM2Annotators, VDLM2AnnotatorHandler{})
	}
	if config.Annotators.Tar1090.Enabled {
		log.Info(Success("TAR1090 VDLM2 annotator enabled"))
		enabledVDLM2Annotators = append(enabledVDLM2Annotators, Tar1090AnnotatorHandler{})
	}
	if config.Annotators.Ollama.Enabled && config.ACARSProcessorSettings.ACARSHub.VDLM2.Host != "" {
		log.Info(Success("Ollama VDLM2 annotator enabled"))
		enabledVDLM2Annotators = append(enabledVDLM2Annotators, OllamaAnnotatorHandler{})
	}
	if len(enabledVDLM2Annotators) == 0 {
		log.Info(Attention("no vdlm2 annotators are enabled"))
	}
}
