package watchers

import (
	"encoding/base64"
	"encoding/json"
	"hash/fnv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Watchable interface {
	ProvidePlaceholder() client.Object
	GetDigest(client.Object, []string) (string, error)
	PrepareForCreate(client.Object, *metav1.ObjectMeta)
	PrepareForUpdate(client.Object, client.Object)
}

type SecretWatchable struct {
	Watchable
}

func (w *SecretWatchable) ProvidePlaceholder() client.Object {
	return &corev1.Secret{}
}

func (w *SecretWatchable) GetDigest(obj client.Object, keys []string) (string, error) {
	secret := obj.(*corev1.Secret)
	return getDigest(keys, func(k string) interface{} {
		return secret.Data[k]
	})
}

func (w *SecretWatchable) PrepareForCreate(obj client.Object, m *metav1.ObjectMeta) {
	fromSecret := obj.(*corev1.Secret)
	fromSecret.ObjectMeta = *m
}

func (w *SecretWatchable) PrepareForUpdate(from, to client.Object) {
	fromSecret := from.(*corev1.Secret)
	toSecret := to.(*corev1.Secret)
	toSecret.Data = fromSecret.Data
}

type ConfigWatchable struct {
	Watchable
}

func (w *ConfigWatchable) ProvidePlaceholder() client.Object {
	return &corev1.ConfigMap{}
}

func (w *ConfigWatchable) GetDigest(obj client.Object, keys []string) (string, error) {
	cm := obj.(*corev1.ConfigMap)
	return getDigest(keys, func(k string) interface{} {
		return cm.Data[k]
	})
}

func (w *ConfigWatchable) PrepareForCreate(obj client.Object, m *metav1.ObjectMeta) {
	fromSecret := obj.(*corev1.ConfigMap)
	fromSecret.ObjectMeta = *m
}

func (w *ConfigWatchable) PrepareForUpdate(from, to client.Object) {
	fromSecret := from.(*corev1.ConfigMap)
	toSecret := to.(*corev1.ConfigMap)
	toSecret.Data = fromSecret.Data
}

func getDigest(keys []string, chunker func(string) interface{}) (string, error) {
	// Inspired from https://github.com/openshift/library-go/blob/master/pkg/operator/resource/resourcehash/as_configmap.go
	hasher := fnv.New32()
	encoder := json.NewEncoder(hasher)
	for _, k := range keys {
		if err := encoder.Encode(chunker(k)); err != nil {
			return "", err
		}
	}
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil)), nil
}
