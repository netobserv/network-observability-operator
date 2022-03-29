package ebpf

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/netobserv/network-observability-operator/api/v1alpha1"
)

type agentConfig struct {
	v1alpha1.FlowCollectorEBPF `json:",inline" mapstructure:",squash"`
}

func configFromMap(data map[string]string) (*agentConfig, error) {
	config := agentConfig{}
	err := mapstructure.WeakDecode(data, &config)
	return &config, err
}

func (ac *agentConfig) asStringMap() map[string]string {
	vals := map[string]interface{}{}
	if err := mapstructure.WeakDecode(ac, &vals); err != nil {
		panic(err)
	}
	stringVals := map[string]string{}
	for k, v := range vals {
		if reflect.ValueOf(v).IsZero() {
			continue
		}
		stringVals[k] = fmt.Sprint(v)
	}
	return stringVals
}
