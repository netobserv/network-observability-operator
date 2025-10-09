package consoleplugin

import (
	"context"
	"reflect"

	osv1 "github.com/openshift/api/console/v1"
	operatorsv1 "github.com/openshift/api/operator/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
	"github.com/netobserv/network-observability-operator/internal/controller/constants"
	"github.com/netobserv/network-observability-operator/internal/controller/reconcilers"
	"github.com/netobserv/network-observability-operator/internal/pkg/helper"
	"github.com/netobserv/network-observability-operator/internal/pkg/resources"
)

// Type alias
type pluginSpec = flowslatest.FlowCollectorConsolePlugin

// CPReconciler reconciles the current console plugin state with the desired configuration
type CPReconciler struct {
	*reconcilers.Instance
	deployment     *appsv1.Deployment
	service        *corev1.Service
	metricsService *corev1.Service
	hpa            *ascv2.HorizontalPodAutoscaler
	serviceAccount *corev1.ServiceAccount
	configMap      *corev1.ConfigMap
	serviceMonitor *monitoringv1.ServiceMonitor
}

func NewReconciler(cmn *reconcilers.Instance) CPReconciler {
	rec := CPReconciler{
		Instance:       cmn,
		deployment:     cmn.Managed.NewDeployment(constants.PluginName),
		service:        cmn.Managed.NewService(constants.PluginName),
		metricsService: cmn.Managed.NewService(metricsSvcName),
		hpa:            cmn.Managed.NewHPA(constants.PluginName),
		serviceAccount: cmn.Managed.NewServiceAccount(constants.PluginName),
		configMap:      cmn.Managed.NewConfigMap(configMapName),
	}
	if cmn.ClusterInfo.HasSvcMonitor() {
		rec.serviceMonitor = cmn.Managed.NewServiceMonitor(constants.PluginName)
	}
	return rec
}

// Reconcile is the reconciler entry point to reconcile the current plugin state with the desired configuration
func (r *CPReconciler) Reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error {
	l := log.FromContext(ctx).WithName("console-plugin")
	ctx = log.IntoContext(ctx, l)

	// Retrieve current owned objects
	err := r.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	if r.ClusterInfo.HasConsolePlugin() {
		if err = r.checkAutoPatch(ctx, desired, constants.PluginName); err != nil {
			return err
		}
	}

	if desired.Spec.UseConsolePlugin() && (r.ClusterInfo.HasConsolePlugin() || desired.Spec.UseTestConsolePlugin()) {
		// Create object builder
		builder := newBuilder(r.Instance, &desired.Spec, constants.PluginName)

		if err := r.reconcilePermissions(ctx, &builder, constants.PluginName); err != nil {
			return err
		}

		if r.ClusterInfo.HasConsolePlugin() {
			if err = r.reconcilePlugin(ctx, &builder, &desired.Spec, constants.PluginName, "NetObserv plugin"); err != nil {
				return err
			}
		}

		cmDigest, err := r.reconcileConfigMap(ctx, &builder)
		if err != nil {
			return err
		}

		if err = r.reconcileDeployment(ctx, &builder, &desired.Spec, constants.PluginName, cmDigest); err != nil {
			return err
		}

		if err = r.reconcileServices(ctx, &builder, constants.PluginName); err != nil {
			return err
		}

		if err = r.reconcileHPA(ctx, &builder, &desired.Spec); err != nil {
			return err
		}

		if desired.Spec.UseLoki() {
			// Watch for Loki certificates if necessary; we'll ignore in that case the returned digest, as we don't need to restart pods on cert rotation
			// because certificate is always reloaded from file
			if _, err = r.Watcher.ProcessCACert(ctx, r.Client, &r.Loki.TLS, r.Namespace); err != nil {
				return err
			}
			if _, _, err = r.Watcher.ProcessMTLSCerts(ctx, r.Client, &r.Loki.StatusTLS, r.Namespace); err != nil {
				return err
			}
		}
	} else {
		// delete any existing owned object
		r.Managed.TryDeleteAll(ctx)
	}

	return nil
}

func (r *CPReconciler) checkAutoPatch(ctx context.Context, desired *flowslatest.FlowCollector, name string) error {
	console := operatorsv1.Console{}
	advancedConfig := helper.GetAdvancedPluginConfig(desired.Spec.ConsolePlugin.Advanced)
	reg := desired.Spec.UseConsolePlugin() && *advancedConfig.Register
	if err := r.Client.Get(ctx, types.NamespacedName{Name: "cluster"}, &console); err != nil {
		// Console operator CR not found => warn but continue execution
		if reg {
			log.FromContext(ctx).Error(err, "Could not get the Console Operator resource for plugin registration. Please register manually.")
		}
		return nil
	}
	registered := helper.ContainsString(console.Spec.Plugins, name)
	if reg && !registered {
		console.Spec.Plugins = append(console.Spec.Plugins, name)
		return r.Client.Update(ctx, &console)
	}
	return nil
}

func (r *CPReconciler) reconcilePermissions(ctx context.Context, builder *builder, name string) error {
	if !r.Managed.Exists(r.serviceAccount) {
		return r.CreateOwned(ctx, builder.serviceAccount(name))
	} // update not needed for now

	binding := resources.GetClusterRoleBinding(
		r.Namespace,
		constants.PluginShortName,
		name,
		name,
		constants.ConsoleTokenReviewRole,
	)
	return r.ReconcileClusterRoleBinding(ctx, binding)
}

func (r *CPReconciler) reconcilePlugin(ctx context.Context, builder *builder, desired *flowslatest.FlowCollectorSpec, name, displayName string) error {
	// Console plugin is cluster-scope (it's not deployed in our namespace) however it must still be updated if our namespace changes
	oldPlg := osv1.ConsolePlugin{}
	pluginExists := true
	err := r.Get(ctx, types.NamespacedName{Name: name}, &oldPlg)
	if err != nil {
		if errors.IsNotFound(err) {
			pluginExists = false
		} else {
			return err
		}
	}

	// Check if objects need update
	consolePlugin := builder.consolePlugin(name, displayName)
	if !pluginExists {
		if err := r.CreateOwned(ctx, consolePlugin); err != nil {
			return err
		}
	} else if pluginNeedsUpdate(&oldPlg, &desired.ConsolePlugin) {
		if err := r.UpdateIfOwned(ctx, &oldPlg, consolePlugin); err != nil {
			return err
		}
	}
	return nil
}

func (r *CPReconciler) reconcileConfigMap(ctx context.Context, builder *builder) (string, error) {
	newCM, configDigest, err := builder.configMap(ctx)
	if err != nil {
		return "", err
	}
	if !r.Managed.Exists(r.configMap) {
		if err := r.CreateOwned(ctx, newCM); err != nil {
			return "", err
		}
	} else if !reflect.DeepEqual(newCM.Data, r.configMap.Data) {
		if err := r.UpdateIfOwned(ctx, r.configMap, newCM); err != nil {
			return "", err
		}
	}
	return configDigest, nil
}

func (r *CPReconciler) reconcileDeployment(ctx context.Context, builder *builder, desired *flowslatest.FlowCollectorSpec, name string, cmDigest string) error {
	report := helper.NewChangeReport("Console deployment")
	defer report.LogIfNeeded(ctx)

	return reconcilers.ReconcileDeployment(
		ctx,
		r.Instance,
		r.deployment,
		builder.deployment(name, cmDigest),
		name,
		!desired.ConsolePlugin.IsUnmanagedConsolePluginReplicas(),
		helper.PtrInt32(desired.ConsolePlugin.Replicas),
		&report,
	)
}

func (r *CPReconciler) reconcileServices(ctx context.Context, builder *builder, name string) error {
	report := helper.NewChangeReport("Console services")
	defer report.LogIfNeeded(ctx)

	if err := r.ReconcileService(ctx, r.service, builder.mainService(name), &report); err != nil {
		return err
	}
	if r.metricsService != nil {
		if err := r.ReconcileService(ctx, r.metricsService, builder.metricsService(), &report); err != nil {
			return err
		}
	}
	if r.serviceMonitor != nil && r.ClusterInfo.HasSvcMonitor() {
		serviceMonitor := builder.serviceMonitor()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.serviceMonitor, serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil {
			return err
		}
	}
	return nil
}

func (r *CPReconciler) reconcileHPA(ctx context.Context, builder *builder, desired *flowslatest.FlowCollectorSpec) error {
	report := helper.NewChangeReport("Console autoscaler")
	defer report.LogIfNeeded(ctx)

	return reconcilers.ReconcileHPA(
		ctx,
		r.Instance,
		r.hpa,
		builder.autoScaler(),
		&desired.ConsolePlugin.Autoscaler,
		&report,
	)
}

func pluginNeedsUpdate(plg *osv1.ConsolePlugin, desired *pluginSpec) bool {
	advancedConfig := helper.GetAdvancedPluginConfig(desired.Advanced)
	return plg.Spec.Backend.Service.Port != *advancedConfig.Port
}
