package cleanup

import (
	"context"
	"testing"

	"github.com/netobserv/network-observability-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var oldDashboard = corev1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		Name:      "grafana-dashboard-netobserv",
		Namespace: "openshift-config-managed",
	},
	Data: map[string]string{},
}

func TestCleanPastReferences(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.ClientMock{}
	clientMock.MockConfigMap(&oldDashboard)
	assert.Equal(1, clientMock.Len())
	didRun = false

	err := CleanPastReferences(context.Background(), &clientMock, "netobserv")
	assert.NoError(err)
	clientMock.AssertCalled(t,
		"Get",
		mock.Anything,
		types.NamespacedName{Name: "grafana-dashboard-netobserv", Namespace: "openshift-config-managed"},
		mock.Anything,
		mock.Anything)
	clientMock.AssertCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
	assert.Equal(0, clientMock.Len())
}

func TestCleanPastReferences_Empty(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.ClientMock{}
	clientMock.MockNonExisting(types.NamespacedName{Name: "grafana-dashboard-netobserv", Namespace: "openshift-config-managed"})
	assert.Equal(0, clientMock.Len())
	didRun = false

	err := CleanPastReferences(context.Background(), &clientMock, "netobserv")
	assert.NoError(err)
	clientMock.AssertCalled(t,
		"Get",
		mock.Anything,
		types.NamespacedName{Name: "grafana-dashboard-netobserv", Namespace: "openshift-config-managed"},
		mock.Anything,
		mock.Anything)
	clientMock.AssertNotCalled(t, "Delete")
}
