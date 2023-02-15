package watchers

import (
	"context"
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClientMock struct {
	mock.Mock
	client.Client
}

func (o *ClientMock) Get(ctx context.Context, nsname types.NamespacedName, obj client.Object) error {
	args := o.Called(ctx, nsname, obj)
	return args.Error(0)
}

func (o *ClientMock) mockObject(name, ns string, meta *v1.ObjectMeta) {
	o.On("Get", mock.Anything, types.NamespacedName{Namespace: ns, Name: name}, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(client.Object)
		arg.SetUID(meta.GetUID())
		arg.SetResourceVersion(meta.GetResourceVersion())
	}).Return(nil)
}

var lokiCA = corev1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		UID:             "abcd",
		ResourceVersion: "1234",
	},
}
var kafkaCA = corev1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		UID:             "efg",
		ResourceVersion: "567",
	},
}
var kafkaUser = corev1.Secret{
	ObjectMeta: v1.ObjectMeta{
		UID:             "hij",
		ResourceVersion: "890",
	},
}

func TestWatchingCertificates(t *testing.T) {
	assert := assert.New(t)
	clientMock := ClientMock{}
	clientMock.mockObject("loki-ca", "ns", &lokiCA.ObjectMeta)
	clientMock.mockObject("kafka-ca", "ns", &kafkaCA.ObjectMeta)
	clientMock.mockObject("kafka-user", "ns", &kafkaUser.ObjectMeta)

	builder := builder.Builder{}
	watcher := RegisterCertificatesWatcher(&builder)
	assert.NotNil(watcher)

	watcher.Reset("ns")
	watcher.SetWatchedCertificate("loki-certificate-ca", &flowslatest.CertificateReference{
		Type:     flowslatest.CertRefTypeConfigMap,
		Name:     "loki-ca",
		CertFile: "ca.crt",
	})

	// isWatched only true for loki-ca in namespace ns
	assert.True(watcher.isWatched(flowslatest.CertRefTypeConfigMap, &corev1.ConfigMap{ObjectMeta: v1.ObjectMeta{Name: "loki-ca", Namespace: "ns"}}))
	assert.False(watcher.isWatched(flowslatest.CertRefTypeConfigMap, &corev1.ConfigMap{ObjectMeta: v1.ObjectMeta{Name: "loki-ca-other", Namespace: "ns"}}))
	assert.False(watcher.isWatched(flowslatest.CertRefTypeConfigMap, &corev1.ConfigMap{ObjectMeta: v1.ObjectMeta{Name: "loki-ca", Namespace: "other-ns"}}))
	assert.False(watcher.isWatched(flowslatest.CertRefTypeSecret, &corev1.Secret{ObjectMeta: v1.ObjectMeta{Name: "loki-ca", Namespace: "ns"}}))

	pod := corev1.PodTemplateSpec{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{},
		},
	}
	err := watcher.AnnotatePod(context.Background(), &clientMock, &pod, "loki-certificate", "kafka-certificate")
	assert.NoError(err)

	// Pod annotated with info from the loki-ca configmap
	assert.Equal(map[string]string{"flows.netobserv.io/cert-loki-certificate-ca": "abcd/1234"}, pod.Annotations)

	watcher.SetWatchedCertificate("kafka-certificate-ca", &flowslatest.CertificateReference{
		Type:     flowslatest.CertRefTypeConfigMap,
		Name:     "kafka-ca",
		CertFile: "ca.crt",
	})

	watcher.SetWatchedCertificate("kafka-certificate-user", &flowslatest.CertificateReference{
		Type:     flowslatest.CertRefTypeSecret,
		Name:     "kafka-user",
		CertFile: "user.crt",
		CertKey:  "user.key",
	})

	err = watcher.AnnotatePod(context.Background(), &clientMock, &pod, "loki-certificate", "kafka-certificate")
	assert.NoError(err)

	// Pod annotated with info from the kafka-ca configmap and kafka-user secret
	assert.Equal(map[string]string{
		"flows.netobserv.io/cert-loki-certificate-ca":    "abcd/1234",
		"flows.netobserv.io/cert-kafka-certificate-ca":   "efg/567",
		"flows.netobserv.io/cert-kafka-certificate-user": "hij/890",
	}, pod.Annotations)

	kafkaUser.SetResourceVersion("xxx")
	err = watcher.AnnotatePod(context.Background(), &clientMock, &pod, "loki-certificate", "kafka-certificate")
	assert.NoError(err)

	// kafka-user secret updated with the new annotation
	assert.Equal(map[string]string{
		"flows.netobserv.io/cert-loki-certificate-ca":    "abcd/1234",
		"flows.netobserv.io/cert-kafka-certificate-ca":   "efg/567",
		"flows.netobserv.io/cert-kafka-certificate-user": "hij/xxx",
	}, pod.Annotations)

}
