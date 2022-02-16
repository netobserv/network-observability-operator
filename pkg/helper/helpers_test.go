package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractVersion(t *testing.T) {
	assert := assert.New(t)

	v := ExtractVersion("quay.io/netobserv/flowlogs-pipeline:v0.1.0")
	assert.Equal("v0.1.0", v)
}

func TestExtractUnknownVersion(t *testing.T) {
	assert := assert.New(t)

	v := ExtractVersion("flowlogs-pipeline")
	assert.Equal("unknown", v)
}
