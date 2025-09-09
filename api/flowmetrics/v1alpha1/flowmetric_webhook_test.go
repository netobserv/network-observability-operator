package v1alpha1

import (
	"context"
	"fmt"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFlowMetric(t *testing.T) {
	tests := []struct {
		desc          string
		m             *FlowMetric
		expectedError string
	}{
		{
			desc: "Valid FlowMetric",
			m: &FlowMetric{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "test-namespace",
				},
				Spec: FlowMetricSpec{
					Filters: []MetricFilter{
						{
							Field: "DstK8S_Zone",
						},
						{
							Field: "NetworkEvents>Type",
							Value: "acl",
						},
					},
					Labels: []string{
						"DstK8S_Zone",
						"NetworkEvents>Name",
					},
					ValueField: "Bytes",
					Flatten:    []string{"NetworkEvents"},
					Remap:      map[string]Label{"NetworkEvents>Name": "name"},
					Buckets:    []string{"0.01", "0.5", "1", "10"},
				},
			},
			expectedError: "",
		},
		{
			desc: "Invalid FlowMetric Filter",
			m: &FlowMetric{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "test-namespace",
				},
				Spec: FlowMetricSpec{
					Filters: []MetricFilter{
						{
							Field: "test",
						},
					},
				},
			},
			expectedError: "invalid filter field",
		},
		{
			desc: "Invalid FlowMetric Label",
			m: &FlowMetric{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "test-namespace",
				},
				Spec: FlowMetricSpec{
					Labels: []string{
						"test",
					},
				},
			},
			expectedError: "invalid label name",
		},
		{
			desc: "Invalid valueField",
			m: &FlowMetric{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "test-namespace",
				},
				Spec: FlowMetricSpec{
					ValueField: "DstAddr",
				},
			},
			expectedError: "invalid value field",
		},
		{
			desc: "Invalid buckets",
			m: &FlowMetric{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "test-namespace",
				},
				Spec: FlowMetricSpec{
					Buckets: []string{"a", ""},
				},
			},
			expectedError: `spec.buckets: Invalid value: ["a",""]: cannot be parsed as a float: "a"`,
		},
	}

	for _, test := range tests {
		_, err := validateFlowMetric(context.TODO(), test.m)
		if err == nil {
			if test.expectedError != "" {
				t.Errorf("%s: ValidateFlowMetric failed, no error found while expected: \"%s\"", test.desc, test.expectedError)
			}
		} else {
			if len(test.expectedError) == 0 {
				t.Errorf("%s: ValidateFlowMetric failed, unexpected error: \"%s\"", test.desc, err)
			}
			if !strings.Contains(fmt.Sprint(err), test.expectedError) {
				t.Errorf("%s: ValidateFlowMetric failed, expected error: \"%s\" to contain: \"%s\"", test.desc, err, test.expectedError)
			}
		}
	}
}
