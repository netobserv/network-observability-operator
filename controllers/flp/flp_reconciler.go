package flp

import (
	"context"
	"fmt"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/loki"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
)

// Type alias
type flpSpec = flowslatest.FlowCollectorFLP

// Reconciler reconciles the current flowlogs-pipeline state with the desired configuration
type Reconciler struct {
	reconcilers []singleReconciler
}

const contextReconcilerName = "FLP kind"

type singleReconciler interface {
	context(ctx context.Context) context.Context
	cleanupNamespace(ctx context.Context)
	reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error
}

func NewReconciler(cmn *reconcilers.Common, image string) Reconciler {
	return Reconciler{
		reconcilers: []singleReconciler{
			newMonolithReconciler(cmn.NewInstance(image)),
			newTransformerReconciler(cmn.NewInstance(image)),
			newIngesterReconciler(cmn.NewInstance(image)),
		},
	}
}

// CleanupNamespace cleans up old namespace
func (r *Reconciler) CleanupNamespace(ctx context.Context) {
	for _, sr := range r.reconcilers {
		sr.cleanupNamespace(sr.context(ctx))
	}
}

func validateDesired(desired *flpSpec) error {
	if desired.Port == 4789 ||
		desired.Port == 6081 ||
		desired.Port == 500 ||
		desired.Port == 4500 {
		return fmt.Errorf("flowlogs-pipeline port value is not authorized")
	}
	return nil
}

func (r *Reconciler) Reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error {
	if err := validateDesired(&desired.Spec.Processor); err != nil {
		return err
	}
	for _, sr := range r.reconcilers {
		if err := sr.reconcile(sr.context(ctx), desired); err != nil {
			return err
		}
	}
	return nil
}

func annotateKafkaExporterCerts(ctx context.Context, info *reconcilers.Common, exp []*flowslatest.FlowCollectorExporter, annotations map[string]string) error {
	for i, exporter := range exp {
		if exporter.Type == flowslatest.KafkaExporter {
			if err := annotateKafkaCerts(ctx, info, &exporter.Kafka, fmt.Sprintf("kafka-export-%d", i), annotations); err != nil {
				return err
			}
		}
	}
	return nil
}

func annotateKafkaCerts(ctx context.Context, info *reconcilers.Common, spec *flowslatest.FlowCollectorKafka, prefix string, annotations map[string]string) error {
	caDigest, userDigest, err := info.Watcher.ProcessMTLSCerts(ctx, info.Client, &spec.TLS, info.Namespace)
	if err != nil {
		return err
	}
	if caDigest != "" {
		annotations[watchers.Annotation(prefix+"-ca")] = caDigest
	}
	if userDigest != "" {
		annotations[watchers.Annotation(prefix+"-user")] = userDigest
	}
	if helper.UseSASL(&spec.SASL) {
		saslDigest1, saslDigest2, err := info.Watcher.ProcessSASL(ctx, info.Client, &spec.SASL, info.Namespace)
		if err != nil {
			return err
		}
		if saslDigest1 != "" {
			annotations[watchers.Annotation(prefix+"-sd1")] = saslDigest1
		}
		if saslDigest2 != "" {
			annotations[watchers.Annotation(prefix+"-sd2")] = saslDigest2
		}
	}
	return nil
}

func reconcileMonitoringCerts(ctx context.Context, info *reconcilers.Common, tlsConfig *flowslatest.ServerTLS, ns string) error {
	if tlsConfig.Type == flowslatest.ServerTLSProvided && tlsConfig.Provided != nil {
		_, err := info.Watcher.ProcessCertRef(ctx, info.Client, tlsConfig.Provided, ns)
		if err != nil {
			return err
		}
	}
	if !tlsConfig.InsecureSkipVerify && tlsConfig.ProvidedCaFile != nil && tlsConfig.ProvidedCaFile.File != "" {
		_, err := info.Watcher.ProcessFileReference(ctx, info.Client, *tlsConfig.ProvidedCaFile, ns)
		if err != nil {
			return err
		}
	}

	return nil
}

func reconcileLokiRoles(ctx context.Context, r *reconcilers.Common, b *builder) error {
	roles := loki.ClusterRoles(b.desired.Loki.Mode)
	if len(roles) > 0 {
		for i := range roles {
			if err := r.ReconcileClusterRole(ctx, &roles[i]); err != nil {
				return err
			}
		}
		// Binding
		crb := loki.ClusterRoleBinding(b.name(), b.name(), b.info.Namespace)
		if err := r.ReconcileClusterRoleBinding(ctx, crb); err != nil {
			return err
		}
	}
	return nil
}
