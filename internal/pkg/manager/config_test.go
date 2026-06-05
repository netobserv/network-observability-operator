package manager

import (
	"testing"

	"github.com/netobserv/netobserv-operator/internal/controller/constants"
	"github.com/netobserv/netobserv-operator/internal/pkg/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pf4Image = "quay.io/netobserv/console-plugin:test-pf4"
	pf5Image = "quay.io/netobserv/console-plugin:test-pf5"
	pf6Image = "quay.io/netobserv/console-plugin:test"
)

func TestResolveWebConsoleImage_OpenShift(t *testing.T) {
	cfg := &Config{
		WebConsoleImage:    pf6Image,
		WebConsolePF4Image: pf4Image,
		WebConsolePF5Image: pf5Image,
		Vendor:             constants.VendorOpenShift,
	}
	type testCase struct {
		version  string
		expected string
	}
	for _, tc := range []testCase{
		{
			expected: pf4Image,
		},
		{
			version:  "4.1.0",
			expected: pf4Image,
		},
		{
			version:  "4.13.0",
			expected: pf4Image,
		},
		{
			version:  "4.14.9",
			expected: pf4Image,
		},
		{
			version:  "4.15.0",
			expected: pf5Image,
		},
		{
			version:  "4.18.3",
			expected: pf5Image,
		},
		{
			version:  "4.21.0",
			expected: pf5Image,
		},
		{
			version:  "4.22.0",
			expected: pf6Image,
		},
		{
			version:  "4.25.0-rc5",
			expected: pf6Image,
		},
	} {
		info := &cluster.Info{}
		info.Mock(tc.version, "")
		img, err := cfg.ResolveWebConsoleImage(info)
		require.NoError(t, err)

		assert.Equal(t, tc.expected, img, "Wrong web console image for OpenShift %v", tc.version)
	}
}

func TestResolveWebConsoleImage_NoVendor(t *testing.T) {
	cfg := &Config{
		WebConsoleImage:    pf6Image,
		WebConsolePF4Image: pf4Image,
		WebConsolePF5Image: pf5Image,
		Vendor:             "",
	}
	info := &cluster.Info{}
	info.Mock("", "")
	img, err := cfg.ResolveWebConsoleImage(info)
	require.NoError(t, err)
	assert.Equal(t, pf6Image, img, "should default to WebConsoleImage")
}

func TestResolveWebConsoleImage_OpenShiftNilInfo(t *testing.T) {
	cfg := &Config{
		Vendor: constants.VendorOpenShift,
	}
	_, err := cfg.ResolveWebConsoleImage(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster info is nil")
}

func TestResolveWebConsoleImage_NoVendorNilInfo(t *testing.T) {
	cfg := &Config{
		WebConsoleImage: pf6Image,
	}
	img, err := cfg.ResolveWebConsoleImage(nil)
	require.NoError(t, err)
	assert.Equal(t, pf6Image, img, "should default to WebConsoleImage")
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		EBPFAgentImage:        "agent:test",
		FlowlogsPipelineImage: "flp:test",
		WebConsoleImage:       "console:test",
		Namespace:             "netobserv",
	}
	assert.NoError(t, cfg.Validate())
}
