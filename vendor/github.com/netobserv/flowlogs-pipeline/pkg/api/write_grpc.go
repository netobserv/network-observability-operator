package api

import "errors"

type WriteGRPC struct {
	TargetHost string `yaml:"targetHost,omitempty" json:"targetHost,omitempty" doc:"the host name or IP of the target Flow collector"`
	TargetPort int    `yaml:"targetPort,omitempty" json:"targetPort,omitempty" doc:"the port of the target Flow collector"`
}

func (w *WriteGRPC) Validate() error {
	if w == nil {
		return errors.New("you must provide a configuration")
	}
	if w.TargetHost == "" {
		return errors.New("targetHost can't be empty")
	}
	if w.TargetPort == 0 {
		return errors.New("targetPort can't be empty")
	}
	return nil
}
