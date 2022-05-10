package operators

import (
	"context"
	"fmt"

	sv1b2 "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta2" //TODO replace ViaQ repo by grafana/loki/operator one when released
	"github.com/go-logr/logr"
	gv1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	lv1beta1 "github.com/grafana/loki/operator/api/v1beta1"
	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	ov1 "github.com/operator-framework/api/pkg/operators/v1"
	ov1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	pv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Reconciler struct {
	ns          string
	client      reconcilers.ClientHelper
	log         logr.Logger
	desiredSpec *flowsv1alpha1.FlowCollectorSpec
}

func NewReconciler(ctx context.Context, client reconcilers.ClientHelper, namespace string, spec *flowsv1alpha1.FlowCollectorSpec) *Reconciler {
	rlog := log.FromContext(ctx, "component", "Operators")

	return &Reconciler{
		client:      client,
		ns:          namespace,
		desiredSpec: spec,
		log:         rlog,
	}
}

// Reconcile the required installed operators
func (r *Reconciler) Reconcile(ctx context.Context, target *flowsv1alpha1.FlowCollector) error {
	operators := []string{}

	// install operators if main instance is required
	// dependent objects will be ignored if InstanceSpec is nil
	if target.Spec.Loki.InstanceSpec != nil && target.Spec.Loki.InstanceSpec.Enable {
		operators = append(operators, constants.LokiOperator)
	}
	if target.Spec.Kafka.InstanceSpec != nil && target.Spec.Kafka.InstanceSpec.Enable {
		operators = append(operators, constants.StrimziOperator)
	}
	if target.Spec.Grafana.InstanceSpec != nil && target.Spec.Grafana.InstanceSpec.Enable {
		operators = append(operators, constants.GrafanaOperator)
	}
	if target.Spec.Prometheus.InstanceSpec != nil && target.Spec.Prometheus.InstanceSpec.Enable {
		operators = append(operators, constants.PrometheusOperator)
	}

	if len(operators) == 0 {
		r.log.Info("no operators required")
		return nil
	}

	// Operator Group is a prerequisite for subscriptions in namespace
	// check https://olm.operatorframework.io/docs/tasks/install-operator-with-olm/#prerequisites
	name := constants.ObservabilityName
	r.log.Info("checking operator group " + name)
	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: r.ns,
		Name:      name,
	}, &ov1.OperatorGroup{})

	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(ctx, operatorGroup(name, r.ns))
			if err != nil {
				r.log.Error(err, fmt.Sprintf("Failed to create operator group %s", name))
			} else {
				r.log.Info(fmt.Sprintf("sucessfully created operator group %s", name))
			}
		} else {
			r.log.Error(err, fmt.Sprintf("Failed to get operator group %s", name))
		}
	} else {
		r.log.Info(fmt.Sprintf("operator group %s already exists", name))
	}

	// Create subscription and instances for each required operator
	for _, oName := range operators {
		if err := r.manageOperator(ctx, oName); err != nil {
			return err
		}
		if err := r.manageInstance(ctx, oName); err != nil {
			return err
		}
	}

	r.log.Info("reconcile operators done")
	return nil
}

func (r *Reconciler) manageOperator(ctx context.Context, name string) error {
	s := r.getSubscription(name, r.ns)

	// subscription will create operator objects from a CatalogSource and keep them up to date
	r.log.Info("checking subscription " + s.Name)
	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: s.Namespace,
		Name:      s.Name,
	}, &ov1alpha1.Subscription{})
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.client.Create(ctx, s)
			if err != nil {
				r.log.Error(err, fmt.Sprintf("Failed to create subscription %s", s.Name))
				return err
			}
			r.log.Info(fmt.Sprintf("sucessfully created subscription %s", s.Name))
			return nil
		}
		r.log.Error(err, fmt.Sprintf("Failed to get subscription %s", s.Name))
		return err
	}
	r.log.Info(fmt.Sprintf("subscription %s already exists", s.Name))
	return nil
}

func (r *Reconciler) manageInstance(ctx context.Context, operatorName string) error {
	instances := r.getInstances(operatorName)

	for t, i := range instances {
		err := r.client.Get(ctx, types.NamespacedName{
			Namespace: i.GetNamespace(),
			Name:      i.GetName(),
		}, t)

		if err != nil {
			if errors.IsNotFound(err) {
				err = r.client.CreateOwned(ctx, i)
				if err != nil {
					r.log.Error(err, fmt.Sprintf("Failed to create %v", i))
					return err
				}
				r.log.Info(fmt.Sprintf("sucessfully created %v", i))
			} else {
				r.log.Error(err, fmt.Sprintf("Failed to get %v for %s in namespace %s", t, operatorName, r.ns))
				return err
			}
		} else {
			r.log.Info(fmt.Sprintf("%v already exists in namespace %s", t, r.ns))
		}

	}
	return nil
}

func (r *Reconciler) getInstances(operatorName string) map[client.Object]client.Object {
	switch operatorName {
	case constants.LokiOperator:
		return map[client.Object]client.Object{
			&lv1beta1.LokiStack{}: lokiInstance(r.ns, r.desiredSpec.Loki.InstanceSpec),
		}
	case constants.StrimziOperator:
		return map[client.Object]client.Object{
			&sv1b2.Kafka{}:      kafkaInstance(r.ns, r.desiredSpec.Kafka.InstanceSpec),
			&sv1b2.KafkaTopic{}: kafkaTopic(r.ns, &r.desiredSpec.Kafka),
		}
	case constants.GrafanaOperator:
		grafanaMap := map[client.Object]client.Object{
			&gv1alpha1.Grafana{}:           grafanaInstance(r.ns, r.desiredSpec.Grafana.InstanceSpec),
			&gv1alpha1.GrafanaDataSource{}: grafanaDataSource(r.ns, &r.desiredSpec.Loki),
		}
		if r.desiredSpec.Grafana.DashboardSpec != nil {
			grafanaMap[&gv1alpha1.GrafanaDashboard{}] = grafanaDashboard(r.ns, r.desiredSpec.Grafana.DashboardSpec)
		}
		return grafanaMap
	case constants.PrometheusOperator:
		return map[client.Object]client.Object{
			&pv1.Prometheus{}: prometheusInstance(r.ns, r.desiredSpec.Prometheus.InstanceSpec),
		}
	default:
		r.log.Error(fmt.Errorf("getInstances for operator %s is not yet implemented", operatorName), "operators implementation error")
		return nil
	}
}

func (r *Reconciler) getSubscription(operatorName string, operatorNamespace string) *ov1alpha1.Subscription {
	switch operatorName {
	case constants.LokiOperator:
		return lokiSubscription(operatorName, operatorNamespace)
	case constants.StrimziOperator:
		return strimziSubscription(operatorName, operatorNamespace)
	case constants.GrafanaOperator:
		return grafanaSubscription(operatorName, operatorNamespace)
	case constants.PrometheusOperator:
		return prometheusSubscription(operatorName, operatorNamespace)
	default:
		r.log.Error(fmt.Errorf("subscription for operator %s is not yet implemented", operatorName), "operators implementation error")
		return nil
	}
}
