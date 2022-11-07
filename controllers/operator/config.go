package operator

import "errors"

// Config of the operator.
type Config struct {
	// EBPFAgentImage is the image of the eBPF agent that is manged by the operator
	EBPFAgentImage string
}

func (cfg *Config) Validate() error {
	if cfg.EBPFAgentImage == "" {
		return errors.New("eBPF agent image argument can't be empty")
	}
	return nil
}
