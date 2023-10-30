package cleanup

import (
	"context"
	"testing"

	"github.com/netobserv/network-observability-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

var oldDashboard = corev1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		Name:      "grafana-dashboard-netobserv",
		Namespace: "openshift-config-managed",
		OwnerReferences: []v1.OwnerReference{{
			APIVersion: "flows.netobserv.io/v1beta2",
			Kind:       "FlowCollector",
			Name:       "cluster",
			Controller: ptr.To(true),
		}},
	},
	Data: map[string]string{},
}

func TestCleanPastReferences(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.NewClient()
	clientMock.MockConfigMap(&oldDashboard)
	assert.Equal(1, clientMock.Len())
	didRun = false

	err := CleanPastReferences(context.Background(), clientMock, "netobserv")
	assert.NoError(err)
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: "grafana-dashboard-netobserv", Namespace: "openshift-config-managed"})
	clientMock.AssertDeleteCalled(t)
	assert.Equal(0, clientMock.Len())
}

func TestCleanPastReferences_Empty(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.NewClient()
	clientMock.MockNonExisting(types.NamespacedName{Name: "grafana-dashboard-netobserv", Namespace: "openshift-config-managed"})
	assert.Equal(0, clientMock.Len())
	didRun = false

	err := CleanPastReferences(context.Background(), clientMock, "netobserv")
	assert.NoError(err)
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: "grafana-dashboard-netobserv", Namespace: "openshift-config-managed"})
	clientMock.AssertDeleteNotCalled(t)
}

func TestCleanPastReferences_NotManaged(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.NewClient()
	unmanaged := oldDashboard
	unmanaged.OwnerReferences = nil
	clientMock.MockConfigMap(&unmanaged)
	assert.Equal(1, clientMock.Len())
	didRun = false

	err := CleanPastReferences(context.Background(), clientMock, "netobserv")
	assert.NoError(err)
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: "grafana-dashboard-netobserv", Namespace: "openshift-config-managed"})
	clientMock.AssertDeleteNotCalled(t)
	assert.Equal(1, clientMock.Len())
}

func TestCleanPastReferences_DifferentOwner(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.NewClient()
	unmanaged := oldDashboard
	unmanaged.OwnerReferences = []v1.OwnerReference{{
		APIVersion: "something/v1beta2",
		Kind:       "SomethingElse",
		Name:       "SomethingElse",
	}}
	clientMock.MockConfigMap(&unmanaged)
	assert.Equal(1, clientMock.Len())
	didRun = false

	err := CleanPastReferences(context.Background(), clientMock, "netobserv")
	assert.NoError(err)
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: "grafana-dashboard-netobserv", Namespace: "openshift-config-managed"})
	clientMock.AssertDeleteNotCalled(t)
	assert.Equal(1, clientMock.Len())
}
