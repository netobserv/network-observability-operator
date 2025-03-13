package cleanup

import (
	"context"
	"testing"

	"github.com/netobserv/network-observability-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

var oldCRB = rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "netobserv-plugin",
		OwnerReferences: []metav1.OwnerReference{{
			APIVersion: "flows.netobserv.io/v1beta2",
			Kind:       "FlowCollector",
			Name:       "cluster",
			Controller: ptr.To(true),
		}},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "any",
	},
	Subjects: []rbacv1.Subject{{
		Kind:      "ServiceAccount",
		Name:      "any",
		Namespace: "any",
	}},
}

var oldCRB2 = rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "flowlogs-pipeline-transformer-role",
		OwnerReferences: []metav1.OwnerReference{{
			APIVersion: "flows.netobserv.io/v1beta2",
			Kind:       "FlowCollector",
			Name:       "cluster",
			Controller: ptr.To(true),
		}},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "any",
	},
	Subjects: []rbacv1.Subject{{
		Kind:      "ServiceAccount",
		Name:      "any",
		Namespace: "any",
	}},
}

func mockCRBs(m *test.ClientMock, crbs ...*rbacv1.ClusterRoleBinding) {
	for _, item := range cleanupList {
		if _, ok := item.placeholder.(*rbacv1.ClusterRoleBinding); ok {
			found := false
			for _, toMock := range crbs {
				if toMock.Name == item.ref.Name {
					m.MockCRB(toMock)
					found = true
					break
				}
			}
			if !found {
				m.MockNonExisting(types.NamespacedName{Name: item.ref.Name})
			}
		}
	}
}

func TestCleanPastReferences(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.NewClient()
	mockCRBs(clientMock, &oldCRB, &oldCRB2)
	assert.Equal(2, clientMock.Len())
	didRun = false

	err := CleanPastReferences(context.Background(), clientMock, "netobserv")
	assert.NoError(err)
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: "netobserv-plugin"})
	clientMock.AssertDeleteCalled(t)
	assert.Equal(0, clientMock.Len())
}

func TestCleanPastReferences_Empty(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.NewClient()
	mockCRBs(clientMock)
	assert.Equal(0, clientMock.Len())
	didRun = false

	err := CleanPastReferences(context.Background(), clientMock, "netobserv")
	assert.NoError(err)
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: "netobserv-plugin"})
	clientMock.AssertDeleteNotCalled(t)
}

func TestCleanPastReferences_NotManaged(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.NewClient()
	unmanaged := oldCRB
	unmanaged.OwnerReferences = nil
	mockCRBs(clientMock, &unmanaged)
	assert.Equal(1, clientMock.Len())
	didRun = false

	err := CleanPastReferences(context.Background(), clientMock, "netobserv")
	assert.NoError(err)
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: "netobserv-plugin"})
	clientMock.AssertDeleteNotCalled(t)
	assert.Equal(1, clientMock.Len())
}

func TestCleanPastReferences_DifferentOwner(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.NewClient()
	unmanaged := oldCRB
	unmanaged.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: "something/v1beta2",
		Kind:       "SomethingElse",
		Name:       "SomethingElse",
	}}
	mockCRBs(clientMock, &unmanaged)
	assert.Equal(1, clientMock.Len())
	didRun = false

	err := CleanPastReferences(context.Background(), clientMock, "netobserv")
	assert.NoError(err)
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: "netobserv-plugin"})
	clientMock.AssertDeleteNotCalled(t)
	assert.Equal(1, clientMock.Len())
}
