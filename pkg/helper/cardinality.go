package helper

import (
	"fmt"
	"strings"

	"github.com/netobserv/network-observability-operator/controllers/consoleplugin/config"
)

func CheckCardinality(labels ...string) (*CardinalityReport, error) {
	frontendCfg, err := config.LoadStaticFrontendConfig()
	if err != nil {
		return nil, err
	}

	perLevel := make(map[config.CardinalityWarn][]string)
	for _, label := range labels {
		card := getSingleCardinality(&frontendCfg, label)
		perLevel[card] = append(perLevel[card], label)
	}
	return &CardinalityReport{perLevel: perLevel}, nil
}

func getSingleCardinality(cfg *config.FrontendConfig, label string) config.CardinalityWarn {
	for _, cfgLabel := range cfg.Fields {
		if label == cfgLabel.Name {
			return cfgLabel.CardinalityWarn
		}
	}
	return config.CardinalityWarnUnknown
}

type CardinalityReport struct {
	perLevel map[config.CardinalityWarn][]string
}

func (r *CardinalityReport) GetOverall() config.CardinalityWarn {
	for _, lvl := range []config.CardinalityWarn{config.CardinalityWarnAvoid, config.CardinalityWarnUnknown, config.CardinalityWarnCareful, config.CardinalityWarnFine} {
		labels := r.perLevel[lvl]
		if len(labels) > 0 {
			return lvl
		}
	}
	// No label
	return config.CardinalityWarnFine
}

func (r *CardinalityReport) GetDetails() string {
	sb := strings.Builder{}
	for _, lvl := range []config.CardinalityWarn{config.CardinalityWarnAvoid, config.CardinalityWarnUnknown, config.CardinalityWarnCareful, config.CardinalityWarnFine} {
		labels := r.perLevel[lvl]
		if len(labels) > 0 {
			sb.WriteString(fmt.Sprintf("Cardinality level '%s': %d labels (%s); ", lvl, len(labels), strings.Join(labels, ", ")))
		}
	}
	return sb.String()
}
