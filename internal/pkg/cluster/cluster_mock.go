package cluster

import (
	"github.com/coreos/go-semver/semver"
	flowslatest "github.com/netobserv/netobserv-operator/api/flowcollector/v1beta2"
)

type MockOption func(*Info)

func WithOpenShiftVersion(v string) MockOption {
	return func(c *Info) {
		if v != "" {
			c.apisMap[OCPSecurity] = true
			c.openShiftVersion = semver.New(v)
		}
	}
}

func WithCNI(cni flowslatest.NetworkType) MockOption {
	return func(c *Info) {
		c.cni = cni
	}
}

func WithAPIs(apis ...APIName) MockOption {
	return func(c *Info) {
		for _, api := range apis {
			c.apisMap[api] = true
		}
	}
}

func WithKubeIPs(ips ...string) MockOption {
	return func(c *Info) {
		c.apiServerIPs = ips
	}
}

func WithKubePorts(ports ...int32) MockOption {
	return func(c *Info) {
		c.apiServerPorts = ports
	}
}

func Mock(opts ...MockOption) *Info {
	c := &Info{}
	c.apisMap = make(map[APIName]bool)
	for _, opt := range opts {
		opt(c)
	}
	c.ready = true
	return c
}
