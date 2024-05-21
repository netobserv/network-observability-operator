package helper

import (
	"testing"

	"github.com/netobserv/network-observability-operator/controllers/consoleplugin/config"
	"github.com/stretchr/testify/assert"
)

func TestLabelCardinality(t *testing.T) {
	r, err := CheckCardinality("SrcK8S_Name")
	assert.NoError(t, err)
	assert.Equal(t, config.CardinalityWarnCareful, r.GetOverall())
	assert.Equal(t, `Cardinality level 'careful': 1 labels (SrcK8S_Name); `, r.GetDetails())

	r, err = CheckCardinality("DstK8S_Name")
	assert.NoError(t, err)
	assert.Equal(t, config.CardinalityWarnCareful, r.GetOverall())
	assert.Equal(t, `Cardinality level 'careful': 1 labels (DstK8S_Name); `, r.GetDetails())

	r, err = CheckCardinality("TimeReceived")
	assert.NoError(t, err)
	assert.Equal(t, config.CardinalityWarnAvoid, r.GetOverall())
	assert.Equal(t, `Cardinality level 'avoid': 1 labels (TimeReceived); `, r.GetDetails())

	r, err = CheckCardinality("SrcK8S_OwnerName")
	assert.NoError(t, err)
	assert.Equal(t, config.CardinalityWarnFine, r.GetOverall())
	assert.Equal(t, `Cardinality level 'fine': 1 labels (SrcK8S_OwnerName); `, r.GetDetails())

	r, err = CheckCardinality("SrcK8S_Name", "DstK8S_Name", "SrcK8S_OwnerName")
	assert.NoError(t, err)
	assert.Equal(t, config.CardinalityWarnCareful, r.GetOverall())
	assert.Equal(t, `Cardinality level 'careful': 2 labels (SrcK8S_Name, DstK8S_Name); Cardinality level 'fine': 1 labels (SrcK8S_OwnerName); `, r.GetDetails())
}
