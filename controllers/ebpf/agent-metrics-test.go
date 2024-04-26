package ebpf

import (
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"

	"github.com/stretchr/testify/assert" // Import the testify library for assertions
	"k8s.io/utils/ptr"
)

func TestPromService(t *testing.T) {
	// Create a new instance of your controller
	controller := &AgentController{}

	// Create a sample FlowCollectorEBPF object for testing
	target := &flowslatest.FlowCollectorEBPF{
		Metrics: flowslatest.EBPFMetrics{
			Server: flowslatest.MetricsServerConfig{
				Port: ptr.To(int32(8080)), // Sample port for testing
			},
		},
	}

	// Call the promService function
	service := controller.promService(target)

	// Assert that the returned service is not nil
	assert.NotNil(t, service)
	// Assert that the service name is as expected
	assert.Equal(t, constants.EBPFAgentMetricsSvcName, service.ObjectMeta.Name)
	// Add more assertions as needed for other properties of the service
}

func TestPromServiceMonitoring(t *testing.T) {
	// Create a new instance of your controller
	controller := &AgentController{}

	// Create a sample FlowCollectorEBPF object for testing
	target := &flowslatest.FlowCollectorEBPF{
		Metrics: flowslatest.EBPFMetrics{
			Server: flowslatest.MetricsServerConfig{
				Port: ptr.To(int32(8080)), // Sample port for testing
			},
		},
	}
	// Call the promServiceMonitoring function
	monitor := controller.promServiceMonitoring(target)

	// Assert that the returned monitor is not nil
	assert.NotNil(t, monitor)
	// Assert that the monitor name is as expected
	assert.Equal(t, constants.EBPFAgentMetricsSvcMonitoringName, monitor.ObjectMeta.Name)
	// Add more assertions as needed for other properties of the monitor
}
