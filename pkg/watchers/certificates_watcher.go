package watchers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	flowlatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
)

type CertificatesWatcher struct {
	watched   map[string]watchedObject
	namespace string
}

func NewCertificatesWatcher() CertificatesWatcher {
	return CertificatesWatcher{watched: make(map[string]watchedObject)}
}

type watchedObject struct {
	nsName types.NamespacedName
	kind   string
}

func RegisterCertificatesWatcher(builder *builder.Builder) *CertificatesWatcher {
	watcher := NewCertificatesWatcher()
	builder.Watches(
		&source.Kind{Type: &corev1.Secret{}},
		handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
			if watcher.isWatched(flowlatest.CertRefTypeSecret, o) {
				// Trigger FlowCollector reconcile
				return []reconcile.Request{{NamespacedName: constants.FlowCollectorName}}
			}
			return []reconcile.Request{}
		}),
	)
	builder.Watches(
		&source.Kind{Type: &corev1.ConfigMap{}},
		handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
			if watcher.isWatched(flowlatest.CertRefTypeConfigMap, o) {
				// Trigger FlowCollector reconcile
				return []reconcile.Request{{NamespacedName: constants.FlowCollectorName}}
			}
			return []reconcile.Request{}
		}),
	)
	return &watcher
}

func (w *CertificatesWatcher) Reset(namespace string) {
	w.namespace = namespace
	w.watched = make(map[string]watchedObject)
}

func (w *CertificatesWatcher) SetWatchedCertificate(key string, ref *flowlatest.CertificateReference) {
	w.watched[key] = watchedObject{
		nsName: types.NamespacedName{
			Name:      ref.Name,
			Namespace: w.namespace,
		},
		kind: ref.Type,
	}
}

func (w *CertificatesWatcher) isWatched(kind string, o client.Object) bool {
	for _, watched := range w.watched {
		if watched.kind == kind && watched.nsName.Name == o.GetName() && watched.nsName.Namespace == o.GetNamespace() {
			return true
		}
	}
	return false
}

func (w *CertificatesWatcher) AnnotatePod(ctx context.Context, cl client.Client, pod *corev1.PodTemplateSpec, keyPrefixes ...string) error {
	for _, keyPrefix := range keyPrefixes {
		caName := constants.CertCAName(keyPrefix)
		if err := w.AnnotatePodSingleVolume(ctx, cl, pod, caName); err != nil {
			return err
		}
		userName := constants.CertUserName(keyPrefix)
		if err := w.AnnotatePodSingleVolume(ctx, cl, pod, userName); err != nil {
			return err
		}
	}
	return nil
}

func (w *CertificatesWatcher) AnnotatePodSingleVolume(ctx context.Context, cl client.Client, pod *corev1.PodTemplateSpec, key string) error {
	if watched, ok := w.watched[key]; ok {
		var sourceMeta *metav1.ObjectMeta
		if watched.kind == flowlatest.CertRefTypeConfigMap {
			var cm corev1.ConfigMap
			err := cl.Get(ctx, watched.nsName, &cm)
			if err != nil {
				return err
			}
			sourceMeta = &cm.ObjectMeta
		} else {
			var s corev1.Secret
			err := cl.Get(ctx, watched.nsName, &s)
			if err != nil {
				return err
			}
			sourceMeta = &s.ObjectMeta
		}
		pod.Annotations[constants.PodCertIDSuffix+key] = string(sourceMeta.GetUID()) + "/" + sourceMeta.GetResourceVersion()
	}
	return nil
}
