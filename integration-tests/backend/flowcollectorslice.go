package e2etests

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// FlowcollectorSlice struct to handle FlowcollectorSlice resources
type FlowcollectorSlice struct {
	Name         string
	Namespace    string
	Sampling     string
	SubnetLabels string
	Template     string
}

// create flowcollector CRD for a given manifest file
func (flowSlice FlowcollectorSlice) CreateFlowcollectorSlice() {
	parameters := []string{"--ignore-unknown-parameters=true", "-f", flowSlice.Template, "-p"}

	flowCollector := reflect.ValueOf(&flowSlice).Elem()

	for i := 0; i < flowCollector.NumField(); i++ {
		if flowCollector.Field(i).Interface() != "" {
			if flowCollector.Type().Field(i).Name != "Template" {
				parameters = append(parameters, fmt.Sprintf("%s=%s", flowCollector.Type().Field(i).Name, flowCollector.Field(i).Interface()))
			}
		}
	}

	err := applyNsResourceFromTemplateByAdmin(flowSlice.Namespace, parameters...)
	if err != nil {
		e2e.Failf("Failed to create FlowCollectorSlice: %v", err)
	}
}

// DeleteFlowcollectorSlice deletes FlowCollectorSlice CRD from a cluster
func (flowSlice *FlowcollectorSlice) DeleteFlowcollectorSlice() error {
	return deleteDynamicResource("flowcollectorslice", flowSlice.Name, flowSlice.Namespace)
}

// WaitForFlowcollectorSliceReady waits for FlowCollectorSlice to be ready by checking status conditions
func (flowSlice *FlowcollectorSlice) WaitForFlowcollectorSliceReady() {
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 600*time.Second, false, func(context.Context) (done bool, err error) {
		obj, getErr := getDynamicResource("flowcollectorslice", flowSlice.Name, flowSlice.Namespace)
		if getErr != nil {
			e2e.Logf("Error getting Ready condition: %v", getErr)
			return false, nil
		}
		condStatus, _, _ := getConditionStatus(obj, "Ready")
		return condStatus == "True", nil
	})
	assertWaitPollNoErr(err, fmt.Sprintf("FlowCollectorSlice %s/%s did not become Ready", flowSlice.Namespace, flowSlice.Name))
}
