package consoleplugin

import (
	"context"
	"reflect"

	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	operatorsv1 "github.com/openshift/api/operator/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

// Type alias
type pluginSpec = flowslatest.FlowCollectorConsolePlugin

// CPReconciler reconciles the current console plugin state with the desired configuration
type CPReconciler struct {
	*reconcilers.Instance
	owned ownedObjects
}

type ownedObjects struct {
	deployment     *appsv1.Deployment
	service        *corev1.Service
	metricsService *corev1.Service
	hpa            *ascv2.HorizontalPodAutoscaler
	serviceAccount *corev1.ServiceAccount
	configMap      *corev1.ConfigMap
	serviceMonitor *monitoringv1.ServiceMonitor
}

func NewReconciler(common *reconcilers.Common, imageName string) CPReconciler {
	owned := ownedObjects{
		deployment:     &appsv1.Deployment{},
		service:        &corev1.Service{},
		metricsService: &corev1.Service{},
		hpa:            &ascv2.HorizontalPodAutoscaler{},
		serviceAccount: &corev1.ServiceAccount{},
		configMap:      &corev1.ConfigMap{},
		serviceMonitor: &monitoringv1.ServiceMonitor{},
	}
	cmnInstance := common.NewInstance(imageName)
	cmnInstance.Managed.AddManagedObject(constants.PluginName, owned.deployment)
	cmnInstance.Managed.AddManagedObject(constants.PluginName, owned.service)
	cmnInstance.Managed.AddManagedObject(metricsSvcName, owned.metricsService)
	cmnInstance.Managed.AddManagedObject(constants.PluginName, owned.hpa)
	cmnInstance.Managed.AddManagedObject(constants.PluginName, owned.serviceAccount)
	cmnInstance.Managed.AddManagedObject(configMapName, owned.configMap)
	if common.AvailableAPIs.HasSvcMonitor() {
		cmnInstance.Managed.AddManagedObject(constants.PluginName, owned.serviceMonitor)
	}

	return CPReconciler{Instance: cmnInstance, owned: owned}
}

// CleanupNamespace cleans up old namespace
func (r *CPReconciler) CleanupNamespace(ctx context.Context) {
	r.Managed.CleanupPreviousNamespace(ctx)
}

// Reconcile is the reconciler entry point to reconcile the current plugin state with the desired configuration
func (r *CPReconciler) Reconcile(ctx context.Context, desired *flowslatest.FlowCollector) error {
	ns := r.Managed.Namespace
	// Retrieve current owned objects
	err := r.Managed.FetchAll(ctx)
	if err != nil {
		return err
	}

	if err = r.checkAutoPatch(ctx, desired); err != nil {
		return err
	}

	if helper.UseConsolePlugin(&desired.Spec) {
		// Create object builder
		builder := newBuilder(ns, r.Instance.Image, &desired.Spec)

		if err := r.reconcilePermissions(ctx, &builder); err != nil {
			return err
		}

		if err = r.reconcilePlugin(ctx, &builder, &desired.Spec); err != nil {
			return err
		}

		cmDigest, err := r.reconcileConfigMap(ctx, &builder)
		if err != nil {
			return err
		}

		if err = r.reconcileDeployment(ctx, &builder, &desired.Spec, cmDigest); err != nil {
			return err
		}

		if err = r.reconcileServices(ctx, &builder); err != nil {
			return err
		}

		if err = r.reconcileHPA(ctx, &builder, &desired.Spec); err != nil {
			return err
		}

		// Watch for Loki certificates if necessary; we'll ignore in that case the returned digest, as we don't need to restart pods on cert rotation
		// because certificate is always reloaded from file
		clientTLS := helper.LokiTLS(&desired.Spec.Loki)
		if _, err = r.Watcher.ProcessCACert(ctx, r.Client, clientTLS, r.Namespace); err != nil {
			return err
		}
		statusTLS := helper.LokiStatusTLS(&desired.Spec.Loki)
		if _, _, err = r.Watcher.ProcessMTLSCerts(ctx, r.Client, statusTLS, r.Namespace); err != nil {
			return err
		}
	} else {
		// delete any existing owned object
		r.Managed.TryDeleteAll(ctx)
	}

	return nil
}

func (r *CPReconciler) checkAutoPatch(ctx context.Context, desired *flowslatest.FlowCollector) error {
	console := operatorsv1.Console{}
	reg := helper.UseConsolePlugin(&desired.Spec) && helper.PtrBool(desired.Spec.ConsolePlugin.Register)
	if err := r.Client.Get(ctx, types.NamespacedName{Name: "cluster"}, &console); err != nil {
		// Console operator CR not found => warn but continue execution
		if reg {
			log.FromContext(ctx).Error(err, "Could not get the Console Operator resource for plugin registration. Please register manually.")
		}
		return nil
	}
	registered := helper.ContainsString(console.Spec.Plugins, constants.PluginName)
	if reg && !registered {
		console.Spec.Plugins = append(console.Spec.Plugins, constants.PluginName)
		return r.Client.Update(ctx, &console)
	} else if !reg && registered {
		console.Spec.Plugins = helper.RemoveAllStrings(console.Spec.Plugins, constants.PluginName)
		return r.Client.Update(ctx, &console)
	}
	return nil
}

func (r *CPReconciler) reconcilePermissions(ctx context.Context, builder *builder) error {
	if !r.Managed.Exists(r.owned.serviceAccount) {
		return r.CreateOwned(ctx, builder.serviceAccount())
	} // update not needed for now

	cr := buildClusterRole()
	if err := r.ReconcileClusterRole(ctx, cr); err != nil {
		return err
	}

	desired := builder.clusterRoleBinding()
	return r.ReconcileClusterRoleBinding(ctx, desired)
}

func (r *CPReconciler) reconcilePlugin(ctx context.Context, builder *builder, desired *flowslatest.FlowCollectorSpec) error {
	// Console plugin is cluster-scope (it's not deployed in our namespace) however it must still be updated if our namespace changes
	oldPlg := osv1alpha1.ConsolePlugin{}
	pluginExists := true
	err := r.Get(ctx, types.NamespacedName{Name: constants.PluginName}, &oldPlg)
	if err != nil {
		if errors.IsNotFound(err) {
			pluginExists = false
		} else {
			return err
		}
	}

	// Check if objects need update
	consolePlugin := builder.consolePlugin()
	if !pluginExists {
		if err := r.CreateOwned(ctx, consolePlugin); err != nil {
			return err
		}
	} else if pluginNeedsUpdate(&oldPlg, &desired.ConsolePlugin) {
		if err := r.UpdateOwned(ctx, &oldPlg, consolePlugin); err != nil {
			return err
		}
	}
	return nil
}

func (r *CPReconciler) reconcileConfigMap(ctx context.Context, builder *builder) (string, error) {
	newCM, configDigest := builder.configMap()
	if !r.Managed.Exists(r.owned.configMap) {
		if err := r.CreateOwned(ctx, newCM); err != nil {
			return "", err
		}
	} else if !reflect.DeepEqual(newCM.Data, r.owned.configMap.Data) {
		if err := r.UpdateOwned(ctx, r.owned.configMap, newCM); err != nil {
			return "", err
		}
	}
	return configDigest, nil
}

func (r *CPReconciler) reconcileDeployment(ctx context.Context, builder *builder, desired *flowslatest.FlowCollectorSpec, cmDigest string) error {
	report := helper.NewChangeReport("Console deployment")
	defer report.LogIfNeeded(ctx)

	newDepl := builder.deployment(cmDigest)
	if !r.Managed.Exists(r.owned.deployment) {
		if err := r.CreateOwned(ctx, newDepl); err != nil {
			return err
		}
	} else if helper.DeploymentChanged(r.owned.deployment, newDepl, constants.PluginName, helper.HPADisabled(&desired.ConsolePlugin.Autoscaler), helper.PtrInt32(desired.ConsolePlugin.Replicas), &report) {
		if err := r.UpdateOwned(ctx, r.owned.deployment, newDepl); err != nil {
			return err
		}
	} else {
		r.CheckDeploymentInProgress(r.owned.deployment)
	}
	return nil
}

func (r *CPReconciler) reconcileServices(ctx context.Context, builder *builder) error {
	report := helper.NewChangeReport("Console services")
	defer report.LogIfNeeded(ctx)

	if err := r.ReconcileService(ctx, r.owned.service, builder.mainService(), &report); err != nil {
		return err
	}
	if err := r.ReconcileService(ctx, r.owned.metricsService, builder.metricsService(), &report); err != nil {
		return err
	}
	if r.AvailableAPIs.HasSvcMonitor() {
		serviceMonitor := builder.serviceMonitor()
		if err := reconcilers.GenericReconcile(ctx, r.Managed, &r.Client, r.owned.serviceMonitor, serviceMonitor, &report, helper.ServiceMonitorChanged); err != nil {
			return err
		}
	}
	return nil
}

func (r *CPReconciler) reconcileHPA(ctx context.Context, builder *builder, desired *flowslatest.FlowCollectorSpec) error {
	report := helper.NewChangeReport("Console autoscaler")
	defer report.LogIfNeeded(ctx)

	// Delete or Create / Update Autoscaler according to HPA option
	if helper.HPADisabled(&desired.ConsolePlugin.Autoscaler) {
		r.Managed.TryDelete(ctx, r.owned.hpa)
	} else {
		newASC := builder.autoScaler()
		if !r.Managed.Exists(r.owned.hpa) {
			if err := r.CreateOwned(ctx, newASC); err != nil {
				return err
			}
		} else if helper.AutoScalerChanged(r.owned.hpa, desired.ConsolePlugin.Autoscaler, &report) {
			if err := r.UpdateOwned(ctx, r.owned.hpa, newASC); err != nil {
				return err
			}
		}
	}
	return nil
}

func pluginNeedsUpdate(plg *osv1alpha1.ConsolePlugin, desired *pluginSpec) bool {
	return plg.Spec.Service.Port != desired.Port
}
