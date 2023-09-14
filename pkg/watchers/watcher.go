package watchers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	rec "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

type Watcher struct {
	watched          map[string]interface{}
	defaultNamespace string
	secrets          SecretWatchable
	configs          ConfigWatchable
}

func NewWatcher() Watcher {
	return Watcher{
		watched: make(map[string]interface{}),
	}
}

func RegisterWatcher(builder *builder.Builder) *Watcher {
	w := NewWatcher()
	w.registerWatches(builder, &w.secrets, flowslatest.RefTypeSecret)
	w.registerWatches(builder, &w.configs, flowslatest.RefTypeConfigMap)
	return &w
}

func (w *Watcher) registerWatches(builder *builder.Builder, watchable Watchable, kind flowslatest.MountableType) {
	builder.Watches(
		&source.Kind{Type: watchable.ProvidePlaceholder()},
		handler.EnqueueRequestsFromMapFunc(func(o client.Object) []rec.Request {
			if w.isWatched(kind, o.GetName(), o.GetNamespace()) {
				// Trigger FlowCollector reconcile
				return []rec.Request{{NamespacedName: constants.FlowCollectorName}}
			}
			return []rec.Request{}
		}),
	)
}

func (w *Watcher) Reset(namespace string) {
	w.defaultNamespace = namespace
	w.watched = make(map[string]interface{})
}

func key(kind flowslatest.MountableType, name, namespace string) string {
	return string(kind) + "/" + namespace + "/" + name
}

func (w *Watcher) watch(kind flowslatest.MountableType, name, namespace string) {
	w.watched[key(kind, name, namespace)] = true
}

func (w *Watcher) isWatched(kind flowslatest.MountableType, name, namespace string) bool {
	if _, ok := w.watched[key(kind, name, namespace)]; ok {
		return true
	}
	return false
}

func (w *Watcher) ProcessMTLSCerts(ctx context.Context, cl helper.Client, tls *flowslatest.ClientTLS, targetNamespace string) (caDigest string, userDigest string, err error) {
	if tls.Enable && tls.CACert.Name != "" {
		caRef := w.refFromCert(&tls.CACert)
		caDigest, err = w.reconcile(ctx, cl, caRef, targetNamespace)
		if err != nil {
			return "", "", err
		}
	}
	if tls.Enable && tls.UserCert.Name != "" {
		userRef := w.refFromCert(&tls.UserCert)
		userDigest, err = w.reconcile(ctx, cl, userRef, targetNamespace)
		if err != nil {
			return "", "", err
		}
	}
	return caDigest, userDigest, nil
}

func (w *Watcher) ProcessCACert(ctx context.Context, cl helper.Client, tls *flowslatest.ClientTLS, targetNamespace string) (caDigest string, err error) {
	if tls.Enable && tls.CACert.Name != "" {
		caRef := w.refFromCert(&tls.CACert)
		caDigest, err = w.reconcile(ctx, cl, caRef, targetNamespace)
		if err != nil {
			return "", err
		}
	}
	return caDigest, nil
}

func (w *Watcher) ProcessCertRef(ctx context.Context, cl helper.Client, cert *flowslatest.CertificateReference, targetNamespace string) (certDigest string, err error) {
	if cert != nil {
		certRef := w.refFromCert(cert)
		certDigest, err = w.reconcile(ctx, cl, certRef, targetNamespace)
		if err != nil {
			return "", err
		}
	}

	return certDigest, nil
}

func (w *Watcher) ProcessFileReference(ctx context.Context, cl helper.Client, file flowslatest.FileReference, targetNamespace string) (fileDigest string, err error) {
	fileDigest, err = w.reconcile(ctx, cl, w.refFromFile(&file), targetNamespace)
	if err != nil {
		return "", err
	}
	return fileDigest, nil
}

func (w *Watcher) ProcessSASL(ctx context.Context, cl helper.Client, sasl *flowslatest.SASLConfig, targetNamespace string) (idDigest string, secretDigest string, err error) {
	idDigest, err = w.reconcile(ctx, cl, w.refFromFile(&sasl.ClientIDReference), targetNamespace)
	if err != nil {
		return "", "", err
	}
	secretDigest, err = w.reconcile(ctx, cl, w.refFromFile(&sasl.ClientSecretReference), targetNamespace)
	if err != nil {
		return "", "", err
	}
	return idDigest, secretDigest, nil
}

func (w *Watcher) reconcile(ctx context.Context, cl helper.Client, ref objectRef, destNamespace string) (string, error) {
	rlog := log.FromContext(ctx, "Name", ref.name, "Source namespace", ref.namespace, "Target namespace", destNamespace)
	report := helper.NewChangeReport("Watcher for " + string(ref.kind) + " " + ref.name)
	defer report.LogIfNeeded(ctx)

	w.watch(ref.kind, ref.name, ref.namespace)
	var watchable Watchable
	if ref.kind == flowslatest.RefTypeConfigMap {
		watchable = &w.configs
	} else {
		watchable = &w.secrets
	}

	obj := watchable.ProvidePlaceholder()
	err := cl.Get(ctx, types.NamespacedName{Name: ref.name, Namespace: ref.namespace}, obj)
	if err != nil {
		return "", err
	}
	digest, err := watchable.GetDigest(obj, ref.keys)
	if err != nil {
		return "", err
	}
	if ref.namespace != destNamespace {
		// copy to namespace
		target := watchable.ProvidePlaceholder()
		err := cl.Get(ctx, types.NamespacedName{Name: ref.name, Namespace: destNamespace}, target)
		if err != nil {
			if !errors.IsNotFound(err) {
				return "", err
			}
			rlog.Info(fmt.Sprintf("creating %s %s in namespace %s", ref.kind, ref.name, destNamespace))
			watchable.PrepareForCreate(obj, &metav1.ObjectMeta{
				Name:      ref.name,
				Namespace: destNamespace,
				Annotations: map[string]string{
					constants.NamespaceCopyAnnotation: ref.namespace + "/" + ref.name,
				},
			})
			if err := cl.CreateOwned(ctx, obj); err != nil {
				return "", err
			}
		} else {
			// Check for update
			targetDigest, err := watchable.GetDigest(target, ref.keys)
			if err != nil {
				return "", err
			}
			if report.Check("Digest changed", targetDigest != digest) {
				// Update existing
				rlog.Info(fmt.Sprintf("updating %s %s in namespace %s", ref.kind, ref.name, destNamespace))
				watchable.PrepareForUpdate(obj, target)
				if err := cl.UpdateOwned(ctx, target, target); err != nil {
					return "", err
				}
			}
		}
	}
	return digest, nil
}

func Annotation(key string) string {
	return constants.PodWatchedSuffix + key
}
