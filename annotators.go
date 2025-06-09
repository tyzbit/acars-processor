package main

import (
	log "github.com/sirupsen/logrus"
)

func ConfigureAnnotators() {
	// ACARS-type messages
	if config.Annotators.ACARS.Enabled {
		yo().Bet("hi").FRFR()
		log.Info(yo().Bet("ACARS annotator enabled").FRFR())
		enabledACARSAnnotators = append(enabledACARSAnnotators, ACARSHandlerAnnotator{})
	}
	if config.Annotators.ADSBExchange.Enabled {
		log.Info(yo().Bet("ADSB annotator enabled").FRFR())
		if config.Annotators.ADSBExchange.APIKey == "" {
			log.Error(yo().Uhh("ADSB API key not set"))
		}
		enabledACARSAnnotators = append(enabledACARSAnnotators, ADSBHandlerAnnotator{})
	}
	if config.Annotators.Tar1090.Enabled {
		log.Info(yo().Bet("TAR1090 VDLM2 annotator enabled").FRFR())
		enabledACARSAnnotators = append(enabledACARSAnnotators, Tar1090Handler{})
	}
	if config.Annotators.Ollama.Enabled && config.ACARSHub.ACARS.Host != "" {
		log.Info(yo().Bet("Ollama ACARS annotator enabled").FRFR())
		enabledACARSAnnotators = append(enabledACARSAnnotators, OllamaHandler{})
	}
	if len(enabledACARSAnnotators) == 0 {
		log.Warn(yo().Uhh("no acars annotators are enabled").FRFR())
	}

	// VDLM2-type messages
	if config.Annotators.VDLM2.Enabled {
		log.Info(yo().Bet("VDLM2 annotator enabled").FRFR())
		enabledVDLM2Annotators = append(enabledVDLM2Annotators, VDLM2HandlerAnnotator{})
	}
	if config.Annotators.Tar1090.Enabled {
		log.Info(yo().Bet("TAR1090 VDLM2 annotator enabled").FRFR())
		enabledVDLM2Annotators = append(enabledVDLM2Annotators, Tar1090Handler{})
	}
	if config.Annotators.Ollama.Enabled && config.ACARSHub.VDLM2.Host != "" {
		log.Info(yo().Bet("Ollama VDLM2 annotator enabled").FRFR())
		enabledVDLM2Annotators = append(enabledVDLM2Annotators, OllamaHandler{})
	}
	if len(enabledVDLM2Annotators) == 0 {
		log.Info(yo().Uhh("no vdlm2 annotators are enabled").FRFR())
	}
}
