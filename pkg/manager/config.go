package manager

import (
	"errors"
)

// Config of the operator.
type Config struct {
	// EBPFAgentImage is the image of the eBPF agent that is managed by the operator
	EBPFAgentImage string
	// FlowlogsPipelineImage is the image of the Flowlogs-Pipeline that is managed by the operator
	FlowlogsPipelineImage string
	// ConsolePluginImage is the image of the Console Plugin that is managed by the operator
	ConsolePluginImage string
	// Release kind is either upstream or downstream
	DownstreamDeployment bool
}

func (cfg *Config) Validate() error {
	if cfg.EBPFAgentImage == "" {
		return errors.New("eBPF agent image argument can't be empty")
	}
	if cfg.FlowlogsPipelineImage == "" {
		return errors.New("flowlogs-pipeline image argument can't be empty")
	}
	if cfg.ConsolePluginImage == "" {
		return errors.New("console plugin image argument can't be empty")
	}
	return nil
}
