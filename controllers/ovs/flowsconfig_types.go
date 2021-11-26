package ovs

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"

	"github.com/netobserv/network-observability-operator/api/v1alpha1"
)

type flowsConfig struct {
	v1alpha1.FlowCollectorIPFIX `json:",inline" mapstructure:",squash"`
	SharedTarget                string `json:"sharedTarget,omitempty" mapstructure:"sharedTarget,omitempty"`
	NodePort                    int32  `json:"nodePort,omitempty" mapstructure:"nodePort,omitempty"`
}

func configFromMap(data map[string]string) (*flowsConfig, error) {
	config := flowsConfig{}
	err := mapstructure.WeakDecode(data, &config)
	return &config, err
}

func (fc *flowsConfig) asStringMap() map[string]string {
	vals := map[string]interface{}{}
	if err := mapstructure.WeakDecode(fc, &vals); err != nil {
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
