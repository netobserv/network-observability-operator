package flp

import (
	"context"
	"fmt"

	flowslatest "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricslatest "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/flp/fmstatus"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
	configv1 "github.com/openshift/api/config/v1"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler reconciles the current flowlogs-pipeline state with the desired configuration
type Reconciler struct {
	client.Client
	mgr              *manager.Manager
	watcher          *watchers.Watcher
	status           status.Instance
	currentNamespace string
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
		Owns(&corev1.ServiceAccount{}).
		Watches(
			&metricslatest.FlowMetric{},
			handler.EnqueueRequestsFromMapFunc(func(_ context.Context, o client.Object) []reconcile.Request {
				if o.GetNamespace() == r.currentNamespace {
					return []reconcile.Request{{NamespacedName: constants.FlowCollectorName}}
				}
				return []reconcile.Request{}
			}),
		)

	ctrl, err := builder.Build(&r)
	if err != nil {
		return err
	}
	r.watcher = watchers.NewWatcher(ctrl)

	return nil
}

type subReconciler interface {
	context(context.Context) context.Context
	cleanupNamespace(context.Context)
	reconcile(context.Context, *flowslatest.FlowCollector, *metricslatest.FlowMetricList, []flowslatest.SubnetLabel) error
	getStatus() *status.Instance
}

// Reconcile is the controller entry point for reconciling current state with desired state.
// It manages the controller status at a high level. Business logic is delegated into `reconcile`.
func (r *Reconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	l := log.Log.WithName("flp") // clear context (too noisy)
	ctx = log.IntoContext(ctx, l)

	// Get flowcollector & create dedicated client
	clh, fc, err := helper.NewFlowCollectorClientHelper(ctx, r.Client)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get FlowCollector: %w", err)
	} else if fc == nil {
		// Delete case
		return ctrl.Result{}, nil
	}

	r.status.SetUnknown()
	defer r.status.Commit(ctx, r.Client)

	err = r.reconcile(ctx, clh, fc)
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

func (r *Reconciler) reconcile(ctx context.Context, clh *helper.Client, fc *flowslatest.FlowCollector) error {
	log := log.FromContext(ctx)

	ns := helper.GetNamespace(&fc.Spec)
	r.currentNamespace = ns
	previousNamespace := r.status.GetDeployedNamespace(fc)
	loki := helper.NewLokiConfig(&fc.Spec.Loki, ns)
	cmn := r.newCommonInfo(clh, ns, previousNamespace, &loki)

	r.watcher.Reset(ns)

	// obtain default cluster ID - api is specific to openshift
	if err := r.mgr.ClusterInfo.CheckClusterInfo(ctx, r.Client); err != nil {
		log.Error(err, "unable to obtain cluster ID")
	}

	// Auto-detect subnets
	var subnetLabels []flowslatest.SubnetLabel
	if r.mgr.ClusterInfo.IsOpenShift() && helper.AutoDetectOpenShiftNetworks(&fc.Spec.Processor) {
		var err error
		subnetLabels, err = r.getOpenShiftSubnets(ctx)
		if err != nil {
			log.Error(err, "error while reading subnet definitions")
		}
	}

	// List custom metrics
	fm := metricslatest.FlowMetricList{}
	if err := r.Client.List(ctx, &fm, &client.ListOptions{Namespace: ns}); err != nil {
		return r.status.Error("CantListFlowMetrics", err)
	}
	fmstatus.Reset()
	defer fmstatus.Sync(ctx, r.Client, &fm)

	// Create sub-reconcilers
	// TODO: refactor to move these subReconciler allocations in `Start`. It will involve some decoupling work, as currently
	// `reconcilers.Common` is dependent on the FlowCollector object, which isn't known at start time.
	reconcilers := []subReconciler{
		newMonolithReconciler(cmn.NewInstance([]string{r.mgr.Config.FlowlogsPipelineImage}, r.mgr.Status.ForComponent(status.FLPMonolith))),
		newTransformerReconciler(cmn.NewInstance([]string{r.mgr.Config.FlowlogsPipelineImage}, r.mgr.Status.ForComponent(status.FLPTransformOnly))),
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
		if err := sr.reconcile(sr.context(ctx), fc, &fm, subnetLabels); err != nil {
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
		ClusterInfo:       r.mgr.ClusterInfo,
		Watcher:           r.watcher,
		Loki:              loki,
		IsDownstream:      r.mgr.Config.DownstreamDeployment,
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

func (r *Reconciler) getOpenShiftSubnets(ctx context.Context) ([]flowslatest.SubnetLabel, error) {
	var subnets []flowslatest.SubnetLabel

	// Pods and Services subnets are found in CNO config
	if r.mgr.ClusterInfo.HasCNO() {
		network := &configv1.Network{}
		err := r.Get(ctx, types.NamespacedName{Name: "cluster"}, network)
		if err != nil {
			return nil, fmt.Errorf("can't get Network information: %w", err)
		}
		var podCIDRs []string
		for _, podsNet := range network.Spec.ClusterNetwork {
			podCIDRs = append(podCIDRs, podsNet.CIDR)
		}
		if len(podCIDRs) > 0 {
			subnets = append(subnets, flowslatest.SubnetLabel{
				Name:  "Pods",
				CIDRs: podCIDRs,
			})
		}
		if len(network.Spec.ServiceNetwork) > 0 {
			subnets = append(subnets, flowslatest.SubnetLabel{
				Name:  "Services",
				CIDRs: network.Spec.ServiceNetwork,
			})
		}
		if network.Spec.ExternalIP != nil && len(network.Spec.ExternalIP.AutoAssignCIDRs) > 0 {
			subnets = append(subnets, flowslatest.SubnetLabel{
				Name:  "ExternalIP",
				CIDRs: network.Spec.ExternalIP.AutoAssignCIDRs,
			})
		}
	}

	// Nodes subnet found in CM cluster-config-v1 (kube-system)
	cm := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Name: "cluster-config-v1", Namespace: "kube-system"}, cm); err != nil {
		return nil, fmt.Errorf(`can't read "cluster-config-v1" ConfigMap: %w`, err)
	}
	machines, err := readMachineNetworks(cm)
	if err != nil {
		return nil, err
	}

	if len(machines) > 0 {
		subnets = append(subnets, machines...)
	}

	return subnets, nil
}

func readMachineNetworks(cm *corev1.ConfigMap) ([]flowslatest.SubnetLabel, error) {
	var subnets []flowslatest.SubnetLabel

	type ClusterConfig struct {
		Networking struct {
			MachineNetwork []struct {
				CIDR string `yaml:"cidr"`
			} `yaml:"machineNetwork"`
		} `yaml:"networking"`
	}

	var rawConfig string
	var ok bool
	if rawConfig, ok = cm.Data["install-config"]; !ok {
		return nil, fmt.Errorf(`can't find key "install-config" in "cluster-config-v1" ConfigMap`)
	}
	var config ClusterConfig
	if err := yaml.Unmarshal([]byte(rawConfig), &config); err != nil {
		return nil, fmt.Errorf(`can't deserialize content of "cluster-config-v1" ConfigMap: %w`, err)
	}

	var cidrs []string
	for _, cidr := range config.Networking.MachineNetwork {
		cidrs = append(cidrs, cidr.CIDR)
	}
	if len(cidrs) > 0 {
		subnets = append(subnets, flowslatest.SubnetLabel{
			Name:  "Machines",
			CIDRs: cidrs,
		})
	}

	return subnets, nil
}
