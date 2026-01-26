package api

import (
	"errors"
	"time"
)

type WriteIpfix struct {
	TargetHost      string   `yaml:"targetHost,omitempty" json:"targetHost,omitempty" doc:"IPFIX Collector host target IP"`
	TargetPort      int      `yaml:"targetPort,omitempty" json:"targetPort,omitempty" doc:"IPFIX Collector host target port"`
	Transport       string   `yaml:"transport,omitempty" json:"transport,omitempty" doc:"Transport protocol (tcp/udp) to be used for the IPFIX connection"`
	EnterpriseID    int      `yaml:"enterpriseId,omitempty" json:"enterpriseId,omitempty" doc:"Enterprise ID for exporting transformations"`
	TplSendInterval Duration `yaml:"tplSendInterval,omitempty" json:"tplSendInterval,omitempty" doc:"Interval for resending templates to the collector (default: 1m)"`
}

func (w *WriteIpfix) SetDefaults() {
	if w.Transport == "" {
		w.Transport = "tcp"
	}
	if w.TplSendInterval.Duration == 0 {
		w.TplSendInterval.Duration = time.Minute
	}
}

func (w *WriteIpfix) Validate() error {
	if w == nil {
		return errors.New("you must provide a configuration")
	}
	if w.TargetHost == "" {
		return errors.New("targetHost can't be empty")
	}
	if w.TargetPort == 0 {
		return errors.New("targetPort can't be empty")
	}
	if w.Transport != "tcp" && w.Transport != "udp" && w.Transport != "" {
		return errors.New("transport should be tcp/udp")
	}
	return nil
}
