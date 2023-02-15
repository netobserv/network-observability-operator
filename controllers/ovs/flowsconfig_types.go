package ovs

import (
	"context"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
)

type flowsConfig struct {
	flowslatest.FlowCollectorIPFIX `json:",inline" mapstructure:",squash"`
	SharedTarget                   string `json:"sharedTarget,omitempty" mapstructure:"sharedTarget,omitempty"`
	NodePort                       int32  `json:"nodePort,omitempty" mapstructure:"nodePort,omitempty"`
}

func configFromMap(data map[string]string) (*flowsConfig, error) {
	config := flowsConfig{}
	err := mapstructure.WeakDecode(data, &config)
	return &config, err
}

func (fc *flowsConfig) asStringMap() (map[string]string, error) {
	vals := map[string]interface{}{}
	if err := mapstructure.WeakDecode(fc, &vals); err != nil {
		return nil, err
	}
	stringVals := map[string]string{}
	for k, v := range vals {
		if reflect.ValueOf(v).IsZero() {
			continue
		}
		stringVals[k] = fmt.Sprint(v)
	}
	return stringVals, nil
}

// getSampling returns the configured sampling, or 1 if ipfix.forceSampleAll is true
// Note that configured sampling has a minimum value of 2.
// See also https://bugzilla.redhat.com/show_bug.cgi?id=2103136 , https://bugzilla.redhat.com/show_bug.cgi?id=2104943
func getSampling(ctx context.Context, cfg *flowslatest.FlowCollectorIPFIX) int32 {
	rlog := log.FromContext(ctx)
	if cfg.ForceSampleAll {
		rlog.Info("Warning, sampling is set to 1. This may put cluster stability at risk.")
		return 1
	}
	return cfg.Sampling
}
