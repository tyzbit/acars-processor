package main

import (
	"errors"
	"fmt"
	// English only
)

// Filter
func (f FilterStep) Filter(m APMessage) (s string, filtered bool, errs error) {
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
			s = fmt.Sprintf("%s(%s)", filter.Name(), reason)
			break
		}
	}
	return s, filtered, errs
}
