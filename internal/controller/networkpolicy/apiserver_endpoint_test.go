package networkpolicy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPToCIDR(t *testing.T) {
	assert := assert.New(t)

	// Test IPv4 addresses
	assert.Equal("192.168.1.1/32", ipToCIDR("192.168.1.1"))
	assert.Equal("10.0.0.1/32", ipToCIDR("10.0.0.1"))
	assert.Equal("172.20.0.1/32", ipToCIDR("172.20.0.1"))

	// Test IPv6 addresses
	assert.Equal("2001:db8::1/128", ipToCIDR("2001:db8::1"))
	assert.Equal("fe80::1/128", ipToCIDR("fe80::1"))
	assert.Equal("::1/128", ipToCIDR("::1"))

	// Test invalid IP
	assert.Equal("", ipToCIDR("invalid"))
	assert.Equal("", ipToCIDR(""))
	assert.Equal("", ipToCIDR("256.256.256.256"))
}
