package cardinality

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

// The cardinalities relates to how the field is suitable for usage as a metric label wrt cardinality; it may have 3 values: fine, careful, avoid
type Warn string

const (
	WarnAvoid   Warn = "avoid"
	WarnCareful Warn = "careful"
	WarnFine    Warn = "fine"
	WarnUnknown Warn = "unknown"
)

//go:embed cardinality.json
var rawCardinality []byte
var cardinality map[string]Warn

func GetCardinalities() (map[string]Warn, error) {
	if cardinality == nil {
		cfg := make(map[string]Warn)
		err := json.Unmarshal(rawCardinality, &cfg)
		if err != nil {
			return cfg, err
		}
		cardinality = cfg
	}
	return cardinality, nil
}

func CheckCardinality(labels ...string) (*Report, error) {
	perLevel := make(map[Warn][]string)
	for _, label := range labels {
		card, err := getCardinality(label)
		if err != nil {
			return nil, err
		}
		perLevel[card] = append(perLevel[card], label)
	}
	return &Report{perLevel: perLevel}, nil
}

func getCardinality(label string) (Warn, error) {
	cardinalities, err := GetCardinalities()
	if err != nil {
		return WarnUnknown, err
	}

	v, ok := cardinalities[label]
	if ok {
		return v, nil
	}
	return WarnUnknown, nil
}

type Report struct {
	perLevel map[Warn][]string
}

func (r *Report) GetOverall() Warn {
	for _, lvl := range []Warn{WarnAvoid, WarnUnknown, WarnCareful, WarnFine} {
		labels := r.perLevel[lvl]
		if len(labels) > 0 {
			return lvl
		}
	}
	// No label
	return WarnFine
}

func (r *Report) GetDetails() string {
	sb := strings.Builder{}
	for _, lvl := range []Warn{WarnAvoid, WarnUnknown, WarnCareful, WarnFine} {
		labels := r.perLevel[lvl]
		if len(labels) > 0 {
			sb.WriteString(fmt.Sprintf("Cardinality level '%s': %d labels (%s); ", lvl, len(labels), strings.Join(labels, ", ")))
		}
	}
	return sb.String()
}
