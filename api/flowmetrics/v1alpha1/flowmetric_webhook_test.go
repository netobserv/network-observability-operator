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
			desc: "Valid FlowMetric Filter",
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
					},
				},
			},
			expectedError: "",
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
			desc: "Valid FlowMetric Label",
			m: &FlowMetric{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "test-namespace",
				},
				Spec: FlowMetricSpec{
					Labels: []string{
						"DstK8S_Zone",
					},
				},
			},
			expectedError: "",
		},
		{
			desc: "Valid valueField",
			m: &FlowMetric{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "test-namespace",
				},
				Spec: FlowMetricSpec{
					ValueField: "Bytes",
				},
			},
			expectedError: "",
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
			desc: "Valid nested fields",
			m: &FlowMetric{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "test-namespace",
				},
				Spec: FlowMetricSpec{
					Labels:  []string{"NetworkEvents>Name"},
					Flatten: []string{"NetworkEvents"},
					Filters: []MetricFilter{
						{
							Field: "NetworkEvents>Type",
							Value: "acl",
						},
					},
					Remap: map[string]Label{"NetworkEvents>Name": "name"},
				},
			},
			expectedError: "",
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
