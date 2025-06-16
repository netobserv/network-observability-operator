package cardinality

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelCardinality(t *testing.T) {
	r, err := CheckCardinality("SrcK8S_Name")
	assert.NoError(t, err)
	assert.Equal(t, WarnCareful, r.GetOverall())
	assert.Equal(t, `Cardinality level 'careful': 1 labels (SrcK8S_Name); `, r.GetDetails())

	r, err = CheckCardinality("DstK8S_Name")
	assert.NoError(t, err)
	assert.Equal(t, WarnCareful, r.GetOverall())
	assert.Equal(t, `Cardinality level 'careful': 1 labels (DstK8S_Name); `, r.GetDetails())

	r, err = CheckCardinality("TimeReceived")
	assert.NoError(t, err)
	assert.Equal(t, WarnAvoid, r.GetOverall())
	assert.Equal(t, `Cardinality level 'avoid': 1 labels (TimeReceived); `, r.GetDetails())

	r, err = CheckCardinality("SrcK8S_OwnerName")
	assert.NoError(t, err)
	assert.Equal(t, WarnFine, r.GetOverall())
	assert.Equal(t, `Cardinality level 'fine': 1 labels (SrcK8S_OwnerName); `, r.GetDetails())

	r, err = CheckCardinality("SrcK8S_Name", "DstK8S_Name", "SrcK8S_OwnerName")
	assert.NoError(t, err)
	assert.Equal(t, WarnCareful, r.GetOverall())
	assert.Equal(t, `Cardinality level 'careful': 2 labels (SrcK8S_Name, DstK8S_Name); Cardinality level 'fine': 1 labels (SrcK8S_OwnerName); `, r.GetDetails())
}

func TestFieldsCardinalityWarns(t *testing.T) {
	allowed := []Warn{WarnAvoid, WarnCareful, WarnFine}
	mapCardinality, err := GetCardinalities()
	assert.Equal(t, err, nil)

	for name, card := range mapCardinality {
		assert.Containsf(t, allowed, card, "Field %s: cardinalityWarn '%s' is invalid", name, card)

		if strings.HasPrefix(name, "Src") {
			base := strings.TrimPrefix(name, "Src")
			dst, ok := mapCardinality["Dst"+base]
			assert.True(t, ok)
			assert.Equalf(t, card, dst, "Cardinality for %s and %s differs", name, "Dst"+base)
		}
	}
}
