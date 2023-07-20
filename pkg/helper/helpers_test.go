package helper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestRemoveAllStrings(t *testing.T) {
	assert := assert.New(t)

	s := RemoveAllStrings([]string{"one", "two", "three", "four", "three"}, "three")
	assert.Equal([]string{"one", "two", "four"}, s)

	s = RemoveAllStrings(s, "five")
	assert.Equal([]string{"one", "two", "four"}, s)
}

func TestKeySorted(t *testing.T) {
	set := map[string]string{
		"b": "1",
		"c": "2",
		"a": "3",
		"d": "4",
	}
	assert.Equal(t,
		[][2]string{{"a", "3"}, {"b", "1"}, {"c", "2"}, {"d", "4"}},
		KeySorted(set))
}

func TestMaxLabelLengt_Cut(t *testing.T) {
	assert.Equal(t, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcde",
		MaxLabelLength("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcde_CUT_HERE"))
}

func TestMaxLabelLengt_NoCut(t *testing.T) {
	assert.Equal(t, "0123456789", MaxLabelLength("0123456789"))
	assert.Equal(t, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcde",
		MaxLabelLength("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcde"))
}

func TestUnstructuredDuration(t *testing.T) {

	t.Run("nil input", func(t *testing.T) {
		var d *metav1.Duration
		got := UnstructuredDuration(d)
		want := ""

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("valid input", func(t *testing.T) {
		d := &metav1.Duration{Duration: time.Minute}
		want := "1m0s"
		got := UnstructuredDuration(d)

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
