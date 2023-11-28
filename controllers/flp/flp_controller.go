package flp

import (
	"context"
	"fmt"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/loki"
	"github.com/netobserv/network-observability-operator/pkg/manager"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
	configv1 "github.com/openshift/api/config/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler reconciles the current flowlogs-pipeline state with the desired configuration
type Reconciler struct {
	client.Client
	mgr       *manager.Manager
	watcher   *watchers.Watcher
	status    status.Instance
	clusterID string
}

func Start(ctx context.Context, mgr *manager.Manager) error {
	log := log.FromContext(ctx)
	log.Info("Starting Flowlogs Pipeline parent controller")

	r := Reconciler{
		Client: mgr.Client,
		mgr:    mgr,
		status: mgr.Status.ForComponent(status.FLPParent),
	}
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&flowslatest.FlowCollector{}, reconcilers.IgnoreStatusChange).
		Named("flp").
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&ascv2.HorizontalPodAutoscaler{}).
		Owns(&corev1.Namespace{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{})

	ctrl, err := builder.Build(&r)
	if err != nil {
		return err
	}
	r.watcher = watchers.NewWatcher(ctrl)

	return nil
}

type subReconciler interface {
	context(ctx context.Context) context.Context
	cleanupNamespace(ctx context.Context)
	reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error
	getStatus() *status.Instance
}

// Reconcile is the controller entry point for reconciling current state with desired state.
// It manages the controller status at a high level. Business logic is delegated into `reconcile`.
func (r *Reconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	l := log.Log.WithName("flp") // clear context (too noisy)
	ctx = log.IntoContext(ctx, l)

	r.status.SetUnknown()
	defer r.status.Commit(ctx, r.Client)

	err := r.reconcile(ctx)
	if err != nil {
		l.Error(err, "FLP reconcile failure")
		// Set status failure unless it was already set
		if !r.status.HasFailure() {
			r.status.SetFailure("FLPError", err.Error())
		}
		return ctrl.Result{}, err
	}

	r.status.SetReady()
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcile(ctx context.Context) error {
	log := log.FromContext(ctx)

	clh, fc, err := helper.NewFlowCollectorClientHelper(ctx, r.Client)
	if err != nil {
		return fmt.Errorf("failed to get FlowCollector: %w", err)
	} else if fc == nil {
		return nil
	}

	ns := helper.GetNamespace(&fc.Spec)
	previousNamespace := r.status.GetDeployedNamespace(fc)
	loki := helper.NewLokiConfig(&fc.Spec.Loki, ns)
	cmn := r.newCommonInfo(clh, ns, previousNamespace, &loki)

	r.watcher.Reset(ns)

	// obtain default cluster ID - api is specific to openshift
	if r.mgr.IsOpenShift() && r.clusterID == "" {
		cversion := &configv1.ClusterVersion{}
		key := client.ObjectKey{Name: "version"}
		if err := r.Client.Get(ctx, key, cversion); err != nil {
			log.Error(err, "unable to obtain cluster ID")
		} else {
			r.clusterID = string(cversion.Spec.ClusterID)
		}
	}

	// Create sub-reconcilers
	// TODO: refactor to move these subReconciler allocations in `Start`. It will involve some decoupling work, as currently
	// `reconcilers.Common` is dependent on the FlowCollector object, which isn't known at start time.
	reconcilers := []subReconciler{
		newMonolithReconciler(cmn.NewInstance(r.mgr.Config.FlowlogsPipelineImage, r.mgr.Status.ForComponent(status.FLPMonolith))),
		newTransformerReconciler(cmn.NewInstance(r.mgr.Config.FlowlogsPipelineImage, r.mgr.Status.ForComponent(status.FLPTransformOnly))),
		newIngesterReconciler(cmn.NewInstance(r.mgr.Config.FlowlogsPipelineImage, r.mgr.Status.ForComponent(status.FLPIngestOnly))),
	}

	// Check namespace changed
	if ns != previousNamespace {
		if previousNamespace != "" {
			log.Info("FlowCollector namespace change detected: cleaning up previous namespace", "old", previousNamespace, "new", ns)
			for _, sr := range reconcilers {
				sr.cleanupNamespace(sr.context(ctx))
			}
		}
		// Update namespace in status
		if err := r.status.SetDeployedNamespace(ctx, r.Client, ns); err != nil {
			return r.status.Error("ChangeNamespaceError", err)
		}
	}

	for _, sr := range reconcilers {
		if err := sr.reconcile(sr.context(ctx), fc); err != nil {
			return sr.getStatus().Error("FLPReconcileError", err)
		}
	}

	return nil
}

func (r *Reconciler) newCommonInfo(clh *helper.Client, ns, prevNs string, loki *helper.LokiConfig) reconcilers.Common {
	return reconcilers.Common{
		Client:            *clh,
		Namespace:         ns,
		PreviousNamespace: prevNs,
		UseOpenShiftSCC:   r.mgr.IsOpenShift(),
		AvailableAPIs:     &r.mgr.AvailableAPIs,
		Watcher:           r.watcher,
		Loki:              loki,
		ClusterID:         r.clusterID,
	}
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
