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

func TestIsSubset(t *testing.T) {
	assert.True(t, IsSubSet(
		map[string]string{"a": "b", "c": "d", "e": "f"},
		map[string]string{"a": "b", "c": "d", "e": "f"}))
	assert.True(t, IsSubSet(
		map[string]string{"a": "b", "c": "d", "e": "f"},
		map[string]string{"a": "b", "e": "f"}))
	assert.False(t, IsSubSet(
		map[string]string{"a": "b", "c": "d", "e": "f"},
		map[string]string{"a": "b", "e": "xxx"}))
	assert.False(t, IsSubSet(
		map[string]string{"a": "b", "c": "d", "e": "f"},
		map[string]string{"a": "b", "z": "d"}))
	assert.False(t, IsSubSet(
		map[string]string{"a": "b", "c": "d", "e": "f"},
		map[string]string{"a": "b", "c": "d", "e": "f", "g": "h"}))
}
