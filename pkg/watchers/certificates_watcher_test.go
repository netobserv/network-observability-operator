package watchers

import (
	"context"
	"testing"

	"github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClientMock struct {
	mock.Mock
	client.Client
	lastCreated client.Object
	lastUpdated client.Object
}

func (o *ClientMock) Get(ctx context.Context, nsname types.NamespacedName, obj client.Object) error {
	args := o.Called(ctx, nsname, obj)
	return args.Error(0)
}

func (o *ClientMock) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	args := o.Called(ctx, obj, opts)
	o.lastCreated = obj
	return args.Error(0)
}

func (o *ClientMock) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	args := o.Called(ctx, obj, opts)
	o.lastUpdated = obj
	return args.Error(0)
}

func (o *ClientMock) mockConfigMap(name, ns string, cm *corev1.ConfigMap) {
	o.On("Get", mock.Anything, types.NamespacedName{Namespace: ns, Name: name}, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*corev1.ConfigMap)
		arg.SetUID(cm.GetUID())
		arg.SetResourceVersion(cm.GetResourceVersion())
		arg.Data = cm.Data
	}).Return(nil)
}

func (o *ClientMock) mockSecret(name, ns string, s *corev1.Secret) {
	o.On("Get", mock.Anything, types.NamespacedName{Namespace: ns, Name: name}, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*corev1.Secret)
		arg.SetUID(s.GetUID())
		arg.SetResourceVersion(s.GetResourceVersion())
		arg.Data = s.Data
	}).Return(nil)
}

func (o *ClientMock) mockNotFound(name, ns string) {
	o.On("Get", mock.Anything, types.NamespacedName{Namespace: ns, Name: name}, mock.Anything).Return(errors.NewNotFound(schema.GroupResource{}, name))
}

func newClientMock() (helper.ClientHelper, *ClientMock) {
	cm := ClientMock{}
	cm.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cm.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	ch := helper.ClientHelper{Client: &cm, SetControllerReference: func(o client.Object) error { return nil }}
	return ch, &cm
}

var lokiCA = corev1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		UID:             "abcd",
		ResourceVersion: "1234",
	},
	Data: map[string]string{"loki-ca": "--cert--"},
}
var kafkaCA = corev1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		UID:             "efg",
		ResourceVersion: "567",
	},
	Data: map[string]string{"kafka-ca": "--cert--"},
}
var kafkaUser = corev1.Secret{
	ObjectMeta: v1.ObjectMeta{
		UID:             "hij",
		ResourceVersion: "890",
	},
	Data: map[string][]byte{"kafka-user": []byte("--cert--")},
}

func TestWatchingCertificates(t *testing.T) {
	assert := assert.New(t)
	cl, clientMock := newClientMock()
	clientMock.mockConfigMap("loki-ca", "ns", &lokiCA)
	clientMock.mockConfigMap("kafka-ca", "ns", &kafkaCA)
	clientMock.mockSecret("kafka-user", "ns", &kafkaUser)

	builder := builder.Builder{}
	watcher := RegisterCertificatesWatcher(&builder)
	assert.NotNil(watcher)

	watcher.Reset("ns")
	watcher.SetWatchedCertificate("loki-certificate-ca", &v1alpha1.CertificateReference{
		Type:     v1alpha1.CertRefTypeConfigMap,
		Name:     "loki-ca",
		CertFile: "ca.crt",
	})

	// isWatched only true for loki-ca in namespace ns
	assert.True(watcher.isWatched(v1alpha1.CertRefTypeConfigMap, &corev1.ConfigMap{ObjectMeta: v1.ObjectMeta{Name: "loki-ca", Namespace: "ns"}}))
	assert.False(watcher.isWatched(v1alpha1.CertRefTypeConfigMap, &corev1.ConfigMap{ObjectMeta: v1.ObjectMeta{Name: "loki-ca-other", Namespace: "ns"}}))
	assert.False(watcher.isWatched(v1alpha1.CertRefTypeConfigMap, &corev1.ConfigMap{ObjectMeta: v1.ObjectMeta{Name: "loki-ca", Namespace: "other-ns"}}))
	assert.False(watcher.isWatched(v1alpha1.CertRefTypeSecret, &corev1.Secret{ObjectMeta: v1.ObjectMeta{Name: "loki-ca", Namespace: "ns"}}))

	pod := corev1.PodTemplateSpec{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{},
		},
	}
	err := watcher.PrepareForPod(context.Background(), cl, &pod, "ns", "loki-certificate", "kafka-certificate")
	assert.NoError(err)

	// Pod annotated with info from the loki-ca configmap
	assert.Equal(map[string]string{"flows.netobserv.io/cert-loki-certificate-ca": "abcd/1234"}, pod.Annotations)

	watcher.SetWatchedCertificate("kafka-certificate-ca", &v1alpha1.CertificateReference{
		Type:     v1alpha1.CertRefTypeConfigMap,
		Name:     "kafka-ca",
		CertFile: "ca.crt",
	})

	watcher.SetWatchedCertificate("kafka-certificate-user", &v1alpha1.CertificateReference{
		Type:     v1alpha1.CertRefTypeSecret,
		Name:     "kafka-user",
		CertFile: "user.crt",
		CertKey:  "user.key",
	})

	err = watcher.PrepareForPod(context.Background(), cl, &pod, "ns", "loki-certificate", "kafka-certificate")
	assert.NoError(err)

	// Pod annotated with info from the kafka-ca configmap and kafka-user secret
	assert.Equal(map[string]string{
		"flows.netobserv.io/cert-loki-certificate-ca":    "abcd/1234",
		"flows.netobserv.io/cert-kafka-certificate-ca":   "efg/567",
		"flows.netobserv.io/cert-kafka-certificate-user": "hij/890",
	}, pod.Annotations)

	kafkaUser.SetResourceVersion("xxx")
	err = watcher.PrepareForPod(context.Background(), cl, &pod, "ns", "loki-certificate", "kafka-certificate")
	assert.NoError(err)

	// kafka-user secret updated with the new annotation
	assert.Equal(map[string]string{
		"flows.netobserv.io/cert-loki-certificate-ca":    "abcd/1234",
		"flows.netobserv.io/cert-kafka-certificate-ca":   "efg/567",
		"flows.netobserv.io/cert-kafka-certificate-user": "hij/xxx",
	}, pod.Annotations)
}

func TestWatchingCertificatesCopyAcrossNamespaces(t *testing.T) {
	assert := assert.New(t)
	cl, clientMock := newClientMock()
	clientMock.mockConfigMap("loki-ca", "ns-source", &lokiCA)
	clientMock.mockNotFound("loki-ca", "ns-dest")

	builder := builder.Builder{}
	watcher := RegisterCertificatesWatcher(&builder)
	assert.NotNil(watcher)

	watcher.Reset("ns")
	watcher.SetWatchedCertificate("loki-certificate-ca", &v1alpha1.CertificateReference{
		Type:      v1alpha1.CertRefTypeConfigMap,
		Name:      "loki-ca",
		Namespace: "ns-source",
		CertFile:  "ca.crt",
	})

	pod := corev1.PodTemplateSpec{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{},
		},
	}
	err := watcher.PrepareForPod(context.Background(), cl, &pod, "ns-dest", "loki-certificate", "kafka-certificate")
	assert.NoError(err)

	clientMock.AssertNumberOfCalls(t, "Create", 1)
	clientMock.AssertNotCalled(t, "Update")
	created := clientMock.lastCreated.(*corev1.ConfigMap)
	assert.Equal(map[string]string{"loki-ca": "--cert--"}, created.Data)

	// Pod annotated with info from the loki-ca configmap
	assert.Equal(map[string]string{"flows.netobserv.io/cert-loki-certificate-ca": "abcd/1234"}, pod.Annotations)

	// Reset mock to test update
	cl, clientMock = newClientMock()
	clientMock.mockConfigMap("loki-ca", "ns-source", &lokiCA)
	clientMock.mockConfigMap("loki-ca", "ns-dest", created)

	// Updating source certificate
	lokiCA.SetResourceVersion("xxx")
	lokiCA.Data["loki-ca"] = "--other-cert--"
	err = watcher.PrepareForPod(context.Background(), cl, &pod, "ns-dest", "loki-certificate", "kafka-certificate")
	assert.NoError(err)

	// kafka-user secret updated with the new annotation
	assert.Equal(map[string]string{
		"flows.netobserv.io/cert-loki-certificate-ca": "abcd/xxx",
	}, pod.Annotations)

	clientMock.AssertNumberOfCalls(t, "Create", 0)
	clientMock.AssertNumberOfCalls(t, "Update", 1)
	assert.Equal(map[string]string{"loki-ca": "--other-cert--"}, clientMock.lastUpdated.(*corev1.ConfigMap).Data)
}
