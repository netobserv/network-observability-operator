package watchers

import (
	"context"
	"testing"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

const baseNamespace = "base-ns"
const otherNamespace = "other-ns"

var lokiCA = corev1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		Name:      "loki-ca",
		Namespace: baseNamespace,
	},
	Data: map[string]string{
		"tls.crt": " -- LOKI CA --",
	},
}
var lokiTLS = flowslatest.ClientTLS{
	Enable: true,
	CACert: flowslatest.CertificateReference{
		Type:     flowslatest.RefTypeConfigMap,
		Name:     lokiCA.Name,
		CertFile: "tls.crt",
	},
}
var otherLokiCA = corev1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		Name:      "other-loki-ca",
		Namespace: otherNamespace,
	},
	Data: map[string]string{
		"tls.crt": " -- LOKI OTHER CA --",
	},
}
var otherLokiTLS = flowslatest.ClientTLS{
	Enable: true,
	CACert: flowslatest.CertificateReference{
		Type:      flowslatest.RefTypeConfigMap,
		Name:      otherLokiCA.Name,
		Namespace: otherLokiCA.Namespace,
		CertFile:  "tls.crt",
	},
}
var kafkaCA = corev1.ConfigMap{
	ObjectMeta: v1.ObjectMeta{
		Name:      "kafka-ca",
		Namespace: baseNamespace,
	},
	Data: map[string]string{
		"ca.crt": " -- KAFKA CA --",
	},
}
var kafkaUser = corev1.Secret{
	ObjectMeta: v1.ObjectMeta{
		Name:      "kafka-user",
		Namespace: baseNamespace,
	},
	Data: map[string][]byte{
		"user.crt": []byte(" -- KAFKA USER CERT --"),
		"user.key": []byte(" -- KAFKA USER KEY --"),
	},
}
var kafkaMTLS = flowslatest.ClientTLS{
	Enable: true,
	CACert: flowslatest.CertificateReference{
		Type:     flowslatest.RefTypeConfigMap,
		Name:     kafkaCA.Name,
		CertFile: "ca.crt",
	},
	UserCert: flowslatest.CertificateReference{
		Type:     flowslatest.RefTypeSecret,
		Name:     kafkaUser.Name,
		CertFile: "user.crt",
		CertKey:  "user.key",
	},
}
var kafkaSaslSecret = corev1.Secret{
	ObjectMeta: v1.ObjectMeta{
		Name:      "kafka-sasl",
		Namespace: baseNamespace,
	},
	Data: map[string][]byte{
		"id":    []byte("me"),
		"token": []byte("ssssaaaaassssslllll"),
	},
}
var kafkaSaslConfig = flowslatest.SASLConfig{
	ClientIDReference: flowslatest.FileReference{
		Type: flowslatest.RefTypeSecret,
		Name: kafkaSaslSecret.Name,
		File: "id",
	},
	ClientSecretReference: flowslatest.FileReference{
		Type: flowslatest.RefTypeSecret,
		Name: kafkaSaslSecret.Name,
		File: "token",
	},
}

func TestGenDigests(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.ClientMock{}
	clientMock.MockConfigMap(&lokiCA)
	clientMock.MockConfigMap(&kafkaCA)
	clientMock.MockSecret(&kafkaUser)
	clientMock.MockSecret(&kafkaSaslSecret)

	builder := builder.Builder{}
	watcher := RegisterWatcher(&builder)
	assert.NotNil(watcher)
	watcher.Reset(baseNamespace)
	cl := helper.UnmanagedClient(&clientMock)

	digLoki, err := watcher.ProcessCACert(context.Background(), cl, &lokiTLS, baseNamespace)
	assert.NoError(err)
	assert.Equal("XDamCg==", digLoki)

	// Same output expected from MTLS func
	dig1, dig2, err := watcher.ProcessMTLSCerts(context.Background(), cl, &lokiTLS, baseNamespace)
	assert.NoError(err)
	assert.Equal("XDamCg==", dig1)
	assert.Equal("", dig2)

	// Different output for kafka certs
	dig1, dig2, err = watcher.ProcessMTLSCerts(context.Background(), cl, &kafkaMTLS, baseNamespace)
	assert.NoError(err)
	assert.Equal("EPFv4Q==", dig1)
	assert.Equal("bNKS0Q==", dig2)

	// Different output for sasl via watcher.Process
	dig1, dig2, err = watcher.ProcessSASL(context.Background(), cl, &kafkaSaslConfig, baseNamespace)
	assert.NoError(err)
	assert.Equal("DTk0Pg==", dig1) // for client ID
	assert.Equal("ItNuCg==", dig2) // for client secret

	// Update object, verify the digest has changed
	caCopy := lokiCA
	caCopy.Data["tls.crt"] = " -- LOKI CA MODIFIED --"
	clientMock.UpdateObject(&caCopy)

	digUpdated, err := watcher.ProcessCACert(context.Background(), cl, &lokiTLS, baseNamespace)
	assert.NoError(err)
	assert.NotEqual(digLoki, digUpdated)
	assert.Equal("Hb65OQ==", digUpdated)

	// Update another key in object, verify the digest hasn't changed
	caCopy.Data["other"] = " -- OTHER --"
	clientMock.UpdateObject(&caCopy)

	digUpdated2, err := watcher.ProcessCACert(context.Background(), cl, &lokiTLS, baseNamespace)
	assert.NoError(err)
	assert.Equal(digUpdated2, digUpdated)
}

func TestNoCopy(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.ClientMock{}
	clientMock.MockConfigMap(&lokiCA)

	builder := builder.Builder{}
	watcher := RegisterWatcher(&builder)
	assert.NotNil(watcher)
	watcher.Reset(baseNamespace)
	cl := helper.UnmanagedClient(&clientMock)

	_, _, err := watcher.ProcessMTLSCerts(context.Background(), cl, &lokiTLS, baseNamespace)
	assert.NoError(err)
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: lokiCA.Name, Namespace: lokiCA.Namespace})
	clientMock.AssertCreateNotCalled(t)
	clientMock.AssertUpdateNotCalled(t)
}

func TestCopyCertificate(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.ClientMock{}
	clientMock.MockConfigMap(&otherLokiCA)
	clientMock.MockNonExisting(types.NamespacedName{Namespace: baseNamespace, Name: otherLokiCA.Name})
	clientMock.MockCreateUpdate()

	builder := builder.Builder{}
	watcher := RegisterWatcher(&builder)
	assert.NotNil(watcher)
	watcher.Reset(baseNamespace)
	cl := helper.UnmanagedClient(&clientMock)

	_, _, err := watcher.ProcessMTLSCerts(context.Background(), cl, &otherLokiTLS, baseNamespace)
	assert.NoError(err)
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: otherLokiCA.Name, Namespace: otherLokiCA.Namespace})
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: otherLokiCA.Name, Namespace: baseNamespace})
	clientMock.AssertCreateCalled(t)
	clientMock.AssertUpdateNotCalled(t)
}

func TestUpdateCertificate(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.ClientMock{}
	clientMock.MockConfigMap(&otherLokiCA)
	// Copy cert changing content => should be updated
	copied := otherLokiCA
	copied.Namespace = baseNamespace
	copied.Data = map[string]string{
		"tls.crt": " -- MODIFIED LOKI OTHER CA --",
	}

	clientMock.MockConfigMap(&copied)
	clientMock.MockCreateUpdate()

	builder := builder.Builder{}
	watcher := RegisterWatcher(&builder)
	assert.NotNil(watcher)
	watcher.Reset(baseNamespace)
	cl := helper.UnmanagedClient(&clientMock)

	_, _, err := watcher.ProcessMTLSCerts(context.Background(), cl, &otherLokiTLS, baseNamespace)
	assert.NoError(err)
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: otherLokiCA.Name, Namespace: otherLokiCA.Namespace})
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: otherLokiCA.Name, Namespace: baseNamespace})
	clientMock.AssertCreateNotCalled(t)
	clientMock.AssertUpdateCalled(t)
}

func TestNoUpdateCertificate(t *testing.T) {
	assert := assert.New(t)
	clientMock := test.ClientMock{}
	clientMock.MockConfigMap(&otherLokiCA)
	// Copy cert keeping same content => should not be updated
	copied := otherLokiCA
	copied.Namespace = baseNamespace
	copied.Data = map[string]string{
		"tls.crt": otherLokiCA.Data["tls.crt"],
	}
	clientMock.MockConfigMap(&copied)
	clientMock.MockCreateUpdate()

	builder := builder.Builder{}
	watcher := RegisterWatcher(&builder)
	assert.NotNil(watcher)
	watcher.Reset(baseNamespace)
	cl := helper.UnmanagedClient(&clientMock)

	_, _, err := watcher.ProcessMTLSCerts(context.Background(), cl, &otherLokiTLS, baseNamespace)
	assert.NoError(err)
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: otherLokiCA.Name, Namespace: otherLokiCA.Namespace})
	clientMock.AssertGetCalledWith(t, types.NamespacedName{Name: otherLokiCA.Name, Namespace: baseNamespace})
	clientMock.AssertCreateNotCalled(t)
	clientMock.AssertUpdateNotCalled(t)
}
