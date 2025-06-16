package test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type ClientMock struct {
	mock.Mock
	client.Client
	objs map[string]client.Object
}

func NewClient() *ClientMock {
	m := ClientMock{
		objs: map[string]client.Object{},
	}
	m.MockCreateUpdate()
	return &m
}

func key(obj client.Object) string {
	return obj.GetObjectKind().GroupVersionKind().Kind + "/" + obj.GetNamespace() + "/" + obj.GetName()
}

func (o *ClientMock) Len() int {
	return len(o.objs)
}

func (o *ClientMock) Get(ctx context.Context, nsname types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	args := o.Called(ctx, nsname, obj, opts)
	return args.Error(0)
}

func (o *ClientMock) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return apiutil.GVKForObject(obj, scheme.Scheme)
}

func (o *ClientMock) AssertGetCalledWith(t *testing.T, nsname types.NamespacedName) {
	o.AssertCalled(t, "Get", mock.Anything, nsname, mock.Anything, mock.Anything)
}

func (o *ClientMock) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	k := key(obj)
	if _, exists := o.objs[k]; exists {
		return errors.New("already exists")
	}
	o.objs[k] = obj
	args := o.Called(ctx, obj, opts)
	return args.Error(0)
}

func (o *ClientMock) AssertCreateCalled(t *testing.T) {
	o.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

func (o *ClientMock) AssertCreateNotCalled(t *testing.T) {
	o.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	o.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

func (o *ClientMock) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	k := key(obj)
	if _, exists := o.objs[k]; !exists {
		return errors.New("doesn't exist")
	}
	o.objs[k] = obj
	args := o.Called(ctx, obj, opts)
	return args.Error(0)
}

func (o *ClientMock) AssertUpdateCalled(t *testing.T) {
	o.AssertCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func (o *ClientMock) AssertUpdateNotCalled(t *testing.T) {
	o.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	o.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func (o *ClientMock) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	k := key(obj)
	if _, exists := o.objs[k]; !exists {
		return errors.New("doesn't exist")
	}
	delete(o.objs, k)
	args := o.Called(ctx, obj, opts)
	return args.Error(0)
}

func (o *ClientMock) AssertDeleteCalled(t *testing.T) {
	o.AssertCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
}

func (o *ClientMock) AssertDeleteNotCalled(t *testing.T) {
	o.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
	o.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
}

func (o *ClientMock) MockSecret(obj *v1.Secret) {
	o.objs[key(obj)] = obj
	o.On("Get", mock.Anything, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*v1.Secret)
		arg.SetName(obj.GetName())
		arg.SetNamespace(obj.GetNamespace())
		arg.SetOwnerReferences(obj.GetOwnerReferences())
		arg.Data = obj.Data
	}).Return(nil)
	o.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
}

func (o *ClientMock) MockConfigMap(obj *v1.ConfigMap) {
	o.objs[key(obj)] = obj
	o.On("Get", mock.Anything, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*v1.ConfigMap)
		arg.SetName(obj.GetName())
		arg.SetNamespace(obj.GetNamespace())
		arg.SetOwnerReferences(obj.GetOwnerReferences())
		arg.Data = obj.Data
	}).Return(nil)
	o.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
}

func (o *ClientMock) MockCRB(obj *rbacv1.ClusterRoleBinding) {
	o.objs[key(obj)] = obj
	o.On("Get", mock.Anything, types.NamespacedName{Name: obj.GetName()}, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*rbacv1.ClusterRoleBinding)
		arg.SetName(obj.GetName())
		arg.SetOwnerReferences(obj.GetOwnerReferences())
		arg.RoleRef = obj.RoleRef
		arg.Subjects = obj.Subjects
	}).Return(nil)
	o.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
}

func (o *ClientMock) UpdateObject(obj client.Object) {
	o.objs[key(obj)] = obj
}

func (o *ClientMock) MockNonExisting(nsn types.NamespacedName) {
	o.On("Get", mock.Anything, nsn, mock.Anything, mock.Anything).Return(kerr.NewNotFound(schema.GroupResource{}, ""))
}

func (o *ClientMock) MockCreateUpdate() {
	o.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	o.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
}
