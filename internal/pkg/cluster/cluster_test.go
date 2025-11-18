package cluster

import (
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
)

func TestIsOpenShiftVersionLessThan(t *testing.T) {
	info := Info{openShiftVersion: semver.New("4.14.9"), ready: true}
	b, _, err := info.IsOpenShiftVersionLessThan("4.15.0")
	assert.NoError(t, err)
	assert.True(t, b)

	info.openShiftVersion = semver.New("4.15.0")
	b, _, err = info.IsOpenShiftVersionLessThan("4.15.0")
	assert.NoError(t, err)
	assert.False(t, b)
}

func TestIsOpenShiftVersionAtLeast(t *testing.T) {
	info := Info{openShiftVersion: semver.New("4.14.9"), ready: true}
	b, _, err := info.IsOpenShiftVersionAtLeast("4.14.0")
	assert.NoError(t, err)
	assert.True(t, b)

	info.openShiftVersion = semver.New("4.14.0")
	b, _, err = info.IsOpenShiftVersionAtLeast("4.14.0")
	assert.NoError(t, err)
	assert.True(t, b)

	info.openShiftVersion = semver.New("4.12.0")
	b, _, err = info.IsOpenShiftVersionAtLeast("4.14.0")
	assert.NoError(t, err)
	assert.False(t, b)
}
