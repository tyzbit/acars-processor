package annotator

import (
	log "github.com/sirupsen/logrus"
	. "github.com/tyzbit/acars-processor/config"
	. "github.com/tyzbit/acars-processor/decorate"
)

// ALL KEYS MUST BE UNIQUE AMONG ALL ANNOTATORS
type Annotation map[string]interface{}

type Annotators struct {
	ACARS []ACARSAnnotator
	VDLM2 []VDLM2Annotator
}

var EnabledAnnotators Annotators

func ConfigureAnnotators() {
	// ACARS-type messages
	if Config.Annotators.ACARS.Enabled {
		log.Info(Success("ACARS annotator enabled"))
		EnabledAnnotators.ACARS = append(EnabledAnnotators.ACARS, ACARSAnnotatorHandler{})
	}
	if Config.Annotators.ADSBExchange.Enabled {
		log.Info(Success("ADSB annotator enabled"))
		if Config.Annotators.ADSBExchange.APIKey == "" {
			log.Error(Attention("ADSB API key not set"))
		}
		EnabledAnnotators.ACARS = append(EnabledAnnotators.ACARS, ADSBAnnotatorHandler{})
	}
	if Config.Annotators.Tar1090.Enabled {
		log.Info(Success("TAR1090 ACARS annotator enabled"))
		EnabledAnnotators.ACARS = append(EnabledAnnotators.ACARS, Tar1090AnnotatorHandler{})
	}
	if Config.Annotators.Ollama.Enabled && Config.ACARSProcessorSettings.ACARSHub.ACARS.Host != "" {
		log.Info(Success("Ollama ACARS annotator enabled"))
		EnabledAnnotators.ACARS = append(EnabledAnnotators.ACARS, OllamaAnnotatorHandler{})
	}
	if len(EnabledAnnotators.ACARS) == 0 {
		log.Warn(Attention("no acars annotators are enabled"))
	}

	// VDLM2-type messages
	if Config.Annotators.VDLM2.Enabled {
		log.Info(Success("VDLM2 annotator enabled"))
		EnabledAnnotators.VDLM2 = append(EnabledAnnotators.VDLM2, VDLM2AnnotatorHandler{})
	}
	if Config.Annotators.Tar1090.Enabled {
		log.Info(Success("TAR1090 VDLM2 annotator enabled"))
		EnabledAnnotators.VDLM2 = append(EnabledAnnotators.VDLM2, Tar1090AnnotatorHandler{})
	}
	if Config.Annotators.Ollama.Enabled && Config.ACARSProcessorSettings.ACARSHub.VDLM2.Host != "" {
		log.Info(Success("Ollama VDLM2 annotator enabled"))
		EnabledAnnotators.VDLM2 = append(EnabledAnnotators.VDLM2, OllamaAnnotatorHandler{})
	}
	if len(EnabledAnnotators.VDLM2) == 0 {
		log.Info(Attention("no vdlm2 annotators are enabled"))
	}
}
