package main

import (
	"errors"
	"fmt"
	"slices"

	log "github.com/sirupsen/logrus"
	// English only
)

// Filter
func (f FilterStep) Filter(m APMessage) (name string, filtered bool, errs error) {
	filters := []Filterer{
		f.Builtin,
		f.Ollama,
		f.OpenAI,
	}
	for _, filter := range filters {
		if !filter.Configured() {
			continue
		}
		filterResult, reason, err := filter.Filter(m)
		if err != nil {
			errs = errors.Join(errs, err)
		}
		filtered = filterResult || filtered
		if filtered {
			name = fmt.Sprintf("%s(%s)", filter.Name(), reason)
			break
		}
		if reason != "" {
			reason = fmt.Sprintf("(%s)", reason)
		}
		log.Debug(Aside("message ending in \""),
			Note(Last20Characters(GetAPMessageCommonFieldAsString(m, "MessageText"))),
			Aside("\" was not filtered by %s%s", filter.Name(), reason))
	}
	// Only keep SelectedFields
	if len(f.SelectedFields) > 0 {
		for messageField := range m {
			if !slices.Contains(f.SelectedFields, messageField) {
				delete(m, messageField)
			}
		}
	}
	return name, filtered, errs
}
