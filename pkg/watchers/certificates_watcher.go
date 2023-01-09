package watchers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

type CertificatesWatcher struct {
	watched          map[string]watchedObject
	defaultNamespace string
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
			if watcher.isWatched(v1alpha1.CertRefTypeSecret, o) {
				// Trigger FlowCollector reconcile
				return []reconcile.Request{{NamespacedName: constants.FlowCollectorName}}
			}
			return []reconcile.Request{}
		}),
	)
	builder.Watches(
		&source.Kind{Type: &corev1.ConfigMap{}},
		handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
			if watcher.isWatched(v1alpha1.CertRefTypeConfigMap, o) {
				// Trigger FlowCollector reconcile
				return []reconcile.Request{{NamespacedName: constants.FlowCollectorName}}
			}
			return []reconcile.Request{}
		}),
	)
	return &watcher
}

func (w *CertificatesWatcher) Reset(namespace string) {
	w.defaultNamespace = namespace
	w.watched = make(map[string]watchedObject)
}

func (w *CertificatesWatcher) SetWatchedCertificate(key string, ref *v1alpha1.CertificateReference) {
	ns := ref.Namespace
	if ns == "" {
		ns = w.defaultNamespace
	}
	w.watched[key] = watchedObject{
		nsName: types.NamespacedName{
			Name:      ref.Name,
			Namespace: ns,
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

func (w *CertificatesWatcher) PrepareForPod(ctx context.Context, cl helper.ClientHelper, pod *corev1.PodTemplateSpec, namespace string, keyPrefixes ...string) error {
	for _, keyPrefix := range keyPrefixes {
		caName := constants.CertCAName(keyPrefix)
		if err := w.PrepareForPodSingleVolume(ctx, cl, pod, namespace, caName); err != nil {
			return err
		}
		userName := constants.CertUserName(keyPrefix)
		if err := w.PrepareForPodSingleVolume(ctx, cl, pod, namespace, userName); err != nil {
			return err
		}
	}
	return nil
}

func (w *CertificatesWatcher) PrepareForPodSingleVolume(ctx context.Context, cl helper.ClientHelper, pod *corev1.PodTemplateSpec, namespace string, key string) error {
	if watched, ok := w.watched[key]; ok {
		var idRev string
		var err error
		if watched.kind == v1alpha1.CertRefTypeConfigMap {
			idRev, err = reconcileConfigMap(ctx, cl, watched.nsName.Name, watched.nsName.Namespace, namespace)
		} else {
			idRev, err = reconcileSecret(ctx, cl, watched.nsName.Name, watched.nsName.Namespace, namespace)
		}
		if err != nil {
			return err
		}
		pod.Annotations[constants.PodCertIDSuffix+key] = idRev
	}
	return nil
}

func reconcileConfigMap(ctx context.Context, cl helper.ClientHelper, name, sourceNamespace, destNamespace string) (string, error) {
	rlog := log.FromContext(ctx, "Name", name, "Source namespace", sourceNamespace, "Target namespace", destNamespace)

	var cm corev1.ConfigMap
	err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: sourceNamespace}, &cm)
	if err != nil {
		return "", err
	}
	idRev := getIDRev(&cm.ObjectMeta)
	if sourceNamespace != destNamespace {
		// copy to namespace
		var cmTarget corev1.ConfigMap
		err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: destNamespace}, &cmTarget)
		if err != nil {
			if !errors.IsNotFound(err) {
				return "", err
			}
			rlog.Info(fmt.Sprintf("creating configmap %s in namespace %s", name, destNamespace))
			cm.ObjectMeta = metav1.ObjectMeta{
				Name:      name,
				Namespace: destNamespace,
			}
			if err := cl.CreateOwned(ctx, &cm); err != nil {
				return "", err
			}
		} else {
			// Update existing
			rlog.Info(fmt.Sprintf("updating configmap %s in namespace %s", name, destNamespace))
			cmTarget.Data = cm.Data
			if err := cl.UpdateOwned(ctx, &cmTarget, &cmTarget); err != nil {
				return "", err
			}
		}
	}
	return idRev, nil
}

func reconcileSecret(ctx context.Context, cl helper.ClientHelper, name, sourceNamespace, destNamespace string) (string, error) {
	rlog := log.FromContext(ctx, "Name", name, "Source namespace", sourceNamespace, "Target namespace", destNamespace)

	var s corev1.Secret
	err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: sourceNamespace}, &s)
	if err != nil {
		return "", err
	}
	idRev := getIDRev(&s.ObjectMeta)
	if sourceNamespace != destNamespace {
		// copy to namespace
		var sTarget corev1.Secret
		err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: destNamespace}, &sTarget)
		if err != nil {
			if !errors.IsNotFound(err) {
				return "", err
			}
			rlog.Info(fmt.Sprintf("creating secret %s in namespace %s", name, destNamespace))
			s.ObjectMeta = metav1.ObjectMeta{
				Name:      name,
				Namespace: destNamespace,
			}
			if err := cl.CreateOwned(ctx, &s); err != nil {
				return "", err
			}
		} else {
			// Update existing
			rlog.Info(fmt.Sprintf("updating secret %s in namespace %s", name, destNamespace))
			sTarget.Data = s.Data
			if err := cl.UpdateOwned(ctx, &sTarget, &sTarget); err != nil {
				return "", err
			}
		}
	}
	return idRev, nil
}

func getIDRev(m *metav1.ObjectMeta) string {
	return string(m.GetUID()) + "/" + m.GetResourceVersion()
}
