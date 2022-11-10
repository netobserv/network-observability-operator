package operators

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/conditions"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed embed/*
var content embed.FS

const (
	//environment can either be "k8s" or "openshift" to match embed path
	//subscriptions will differ between each environment
	k8sPath       = "k8s"
	openshiftPath = "openshift"

	operatorGroupPath = "embed/operator_group.yaml"

	lokiInstancePath       = "embed/loki_instance.yaml"
	lokiClusterRole        = "embed/loki_cluster_role.yaml"
	lokiClusterRoleBinding = "embed/loki_cluster_role_binding.yaml"

	kafkaInstancePath = "embed/kafka_instance.yaml"
	kafkaTopicPath    = "embed/kafka_topic.yaml"
	kafkaUserPath     = "embed/kafka_user.yaml"

	//TODO: implement prometheus subscription
	//prometheusInstancePath     = "embed/prometheus_instance.yaml"

	//TODO: implement grafana subscription
	//grafanaInstancePath     = "embed/grafana_instance.yaml"
	//grafanaDashboardPath    = "embed/grafana_dashboard.yaml"
)

type JSONInterface map[string]interface{}
type Reconciler struct {
	client      reconcilers.ClientHelper
	ctx         context.Context
	desiredSpec *flowsv1alpha1.FlowCollectorSpec
	namespace   string
	req         ctrl.Request
	environment string
}

func NewReconciler(ctx context.Context, client reconcilers.ClientHelper, spec *flowsv1alpha1.FlowCollectorSpec, ns string, req ctrl.Request, openshift bool) *Reconciler {
	environment := k8sPath
	if openshift {
		environment = openshiftPath
	}

	return &Reconciler{
		ctx:         ctx,
		client:      client,
		desiredSpec: spec,
		namespace:   ns,
		req:         req,
		environment: environment,
	}
}

// Reconcile the dependent operators by
// creating subscription and instances for each required operator
func (r *Reconciler) Reconcile(client reconcilers.ClientHelper) ([]metav1.Condition, error) {
	// deploy kafka only when operator auto install includes "kafka"
	// and when CRD / Kafka CRD / Owned object has changed
	deployKafka := reconcilers.Contains(r.desiredSpec.OperatorsAutoInstall, constants.KafkaOperator) &&
		reconcilers.Contains(&[]string{constants.Cluster, constants.KafkaCRDName, constants.KafkaTopicCRDName, constants.KafkaUserCRDName}, r.req.Name)

	// deploy loki only when operator auto install includes "loki"
	// and when CRD / Loki CRD / Owned object / Loki secret has changed
	deployLoki := reconcilers.Contains(r.desiredSpec.OperatorsAutoInstall, constants.LokiOperator) &&
		reconcilers.Contains(&[]string{constants.Cluster, constants.LokiCRDName, reconcilers.LokiSecretName(r.desiredSpec)}, r.req.Name)

	conditions := []metav1.Condition{}
	if deployKafka || deployLoki {
		// Operator Group is a prerequisite for subscriptions in namespace
		// check https://olm.operatorframework.io/docs/tasks/install-operator-with-olm/#prerequisites
		err := r.manageOperator([]map[string]*JSONInterface{{
			operatorGroupPath: &JSONInterface{
				"metadata": JSONInterface{
					"namespace": r.namespace,
				},
			},
		}})
		if err != nil {
			return conditions, err
		}

		if deployKafka {
			err = r.deployKafka()
			if err != nil {
				return conditions, err
			}

			condition, err := r.getCondition(kafkaInstancePath, r.namespace)
			if condition != nil {
				conditions = append(conditions, *condition)
			}
			if err != nil {
				return conditions, err
			}
		}

		if deployLoki {
			preReqCondition, err := r.deployLoki()
			if preReqCondition != nil {
				conditions = append(conditions, *preReqCondition)
			}
			if err != nil {
				return conditions, err
			}

			if preReqCondition == nil {
				condition, err := r.getCondition(lokiInstancePath, r.namespace)
				if condition != nil {
					conditions = append(conditions, *condition)
				}
				if err != nil {
					return conditions, err
				}
			}
		}
	}

	return conditions, nil
}

func (r *Reconciler) getCondition(yamlPath string, namespace string) (*metav1.Condition, error) {
	rlog := log.FromContext(r.ctx, "component", "getCondition")

	u, err := loadYaml(yamlPath, &JSONInterface{
		"metadata": JSONInterface{
			"namespace": namespace,
		},
	})
	if err != nil {
		return nil, err
	}

	err = r.client.Get(r.ctx, types.NamespacedName{
		Namespace: u.GetNamespace(),
		Name:      u.GetName(),
	}, u)
	if err != nil {
		message := fmt.Sprintf("%s: '%s' doesn't exist yet in namespace: '%s'. please wait", u.GetKind(), u.GetName(), u.GetNamespace())
		rlog.Info(message)
		return conditions.WaitingDependentOperator(u.GetKind(), message), nil
	}

	var currentCondition *metav1.Condition
	conditions, ok, err := unstructured.NestedSlice(u.Object, "status", "conditions")
	if ok && err == nil {
		for _, i := range conditions {
			condition := i.(map[string]interface{})
			if condition["status"].(string) == "True" {
				rlog.Info(fmt.Sprintf("%s condition is %s", u.GetName(), condition["message"]))
				conditionType := "Unknown"
				if condition["type"] != nil {
					conditionType = condition["type"].(string)
				}
				conditionReason := "Unknown"
				if condition["reason"] != nil {
					conditionReason = condition["reason"].(string)
				}
				conditionMessage := "Unknown"
				if condition["message"] != nil {
					conditionMessage = condition["message"].(string)
				}

				currentCondition = &metav1.Condition{
					Status:  metav1.ConditionTrue,
					Type:    fmt.Sprintf("%s%s", u.GetKind(), conditionType),
					Reason:  conditionReason,
					Message: conditionMessage,
				}
			}
		}
	}

	return currentCondition, nil
}

func (r *Reconciler) deployKafka() error {
	subscription := JSONInterface{
		"metadata": JSONInterface{
			"namespace": r.namespace,
		},
		"spec": JSONInterface{},
	}
	if len(r.desiredSpec.Kafka.AutoInstallSpec.Source) > 0 {
		subscription["spec"].(JSONInterface)["source"] = r.desiredSpec.Loki.AutoInstallSpec.Source
	}
	if len(r.desiredSpec.Kafka.AutoInstallSpec.StartingCSV) > 0 {
		subscription["spec"].(JSONInterface)["installPlanApproval"] = "Manual"
		subscription["spec"].(JSONInterface)["startingCSV"] = r.desiredSpec.Kafka.AutoInstallSpec.StartingCSV
	}

	return r.manageOperator([]map[string]*JSONInterface{{
		fmt.Sprintf("embed/%s/kafka_subscription.yaml", r.environment): &subscription,
		kafkaInstancePath: &JSONInterface{
			"metadata": JSONInterface{
				"namespace": r.namespace,
			},
			"spec": JSONInterface{
				"kafka": JSONInterface{
					"replicas": r.desiredSpec.Kafka.AutoInstallSpec.Replicas,
					"storage": JSONInterface{
						"type":  r.desiredSpec.Kafka.AutoInstallSpec.Storage.Type,
						"size":  r.desiredSpec.Kafka.AutoInstallSpec.Storage.Size,
						"class": r.desiredSpec.Kafka.AutoInstallSpec.Storage.Class,
					},
				},
				"zookeeper": JSONInterface{
					"replicas": r.desiredSpec.Kafka.AutoInstallSpec.ZooKeeperReplicas,
					"storage": JSONInterface{
						"type":  r.desiredSpec.Kafka.AutoInstallSpec.ZooKeeperStorage.Type,
						"size":  r.desiredSpec.Kafka.AutoInstallSpec.ZooKeeperStorage.Size,
						"class": r.desiredSpec.Kafka.AutoInstallSpec.ZooKeeperStorage.Class,
					},
				},
			},
		},
		kafkaTopicPath: &JSONInterface{
			"metadata": JSONInterface{
				"namespace": r.namespace,
			},
			"spec": JSONInterface{
				"partitions": r.desiredSpec.Kafka.AutoInstallSpec.Partitions,
				"replicas":   r.desiredSpec.Kafka.AutoInstallSpec.Replicas,
			},
		},
		kafkaUserPath: &JSONInterface{
			"metadata": JSONInterface{
				"namespace": r.namespace,
			},
		},
	}})
}

func (r *Reconciler) deployLoki() (*metav1.Condition, error) {
	rlog := log.FromContext(r.ctx, "component", "deployLoki")

	//check that loki secret exists
	secretName := reconcilers.LokiSecretName(r.desiredSpec)
	err := r.client.Get(r.ctx, types.NamespacedName{
		Name:      secretName,
		Namespace: r.namespace,
	}, &corev1.Secret{})
	if err != nil {
		if errors.IsNotFound(err) {
			message := fmt.Sprintf("loki secret: '%s' doesn't exist yet in namespace: '%s'. please create it", secretName, r.namespace)
			rlog.Info(message)
			return conditions.MissingLokiSecret(message), nil
		}
		return nil, err
	}

	subscription := JSONInterface{
		"metadata": JSONInterface{
			"namespace": r.namespace,
		},
		"spec": JSONInterface{},
	}
	if len(r.desiredSpec.Loki.AutoInstallSpec.Source) > 0 {
		subscription["spec"].(JSONInterface)["source"] = r.desiredSpec.Loki.AutoInstallSpec.Source
	}
	if len(r.desiredSpec.Loki.AutoInstallSpec.StartingCSV) > 0 {
		subscription["spec"].(JSONInterface)["installPlanApproval"] = "Manual"
		subscription["spec"].(JSONInterface)["startingCSV"] = r.desiredSpec.Loki.AutoInstallSpec.StartingCSV
	}

	//FLP service account can either be flowlogs-pipeline or flowlogs-pipeline-transformer
	flpName := constants.FLPName
	if r.desiredSpec.UseKafka() {
		flpName = flpName + "-transformer"
	}

	//FLP service account always needs role
	roleSubjects := []JSONInterface{
		{
			"kind":      "ServiceAccount",
			"name":      flpName,
			"namespace": r.namespace,
		},
	}
	//netobserv-plugin service account will be used if user token is not forwarded
	if !r.desiredSpec.Loki.ForwardUserToken() {
		roleSubjects = append(roleSubjects, JSONInterface{
			"kind":      "ServiceAccount",
			"name":      constants.PluginName,
			"namespace": r.namespace,
		})
	}

	err = r.manageOperator([]map[string]*JSONInterface{{
		fmt.Sprintf("embed/%s/loki_subscription.yaml", r.environment): &subscription,
		lokiInstancePath: &JSONInterface{
			"metadata": JSONInterface{
				"namespace": r.namespace,
			},
			"spec": JSONInterface{
				"managementState": r.desiredSpec.Loki.AutoInstallSpec.ManagementState,
				"limits": JSONInterface{
					"global": JSONInterface{
						"retention": JSONInterface{
							"days": r.desiredSpec.Loki.AutoInstallSpec.RetentionDays,
						},
					},
				},
				"replicationFactor": r.desiredSpec.Loki.AutoInstallSpec.ReplicationFactor,
				"storage": JSONInterface{
					"secret": JSONInterface{
						"name": r.desiredSpec.Loki.AutoInstallSpec.SecretName,
						"type": r.desiredSpec.Loki.AutoInstallSpec.ObjectStorageType,
					},
				},
				"size":             r.desiredSpec.Loki.AutoInstallSpec.Size,
				"storageClassName": r.desiredSpec.Loki.AutoInstallSpec.StorageClassName,
			},
		},
		lokiClusterRole: nil,
		lokiClusterRoleBinding: &JSONInterface{
			"subjects": roleSubjects,
		},
	}})
	return nil, err
}

func (r *Reconciler) manageOperator(dependencies []map[string]*JSONInterface) error {
	rlog := log.FromContext(r.ctx, "component", "OperatorsController", "function", "manageOperator")

	for _, dMap := range dependencies {
		for path, json := range dMap {
			//load yaml
			yaml, err := loadYaml(path, json)
			if err != nil {
				return err
			}

			//ensure custom resources exists
			crdName := getCRDName(yaml.GetKind())
			if len(crdName) > 0 {
				err := r.client.Get(r.ctx, types.NamespacedName{
					Name: crdName,
				}, &apiextensionsv1.CustomResourceDefinition{})
				if err != nil {
					if errors.IsNotFound(err) {
						rlog.Info(fmt.Sprintf("custom resource definition: '%s' doesn't exist yet. waiting for it", crdName))
						return nil
					}
					return err
				}
			}

			//apply updated yaml
			err = r.client.Apply(r.ctx, yaml)
			if err != nil {
				rlog.Error(err, fmt.Sprintf("can't apply '%s' yaml", path))
				return err
			}
		}
	}
	return nil
}

func getCRDName(kind string) string {
	switch kind {
	case "Kafka":
		return constants.KafkaCRDName
	case "KafkaTopic":
		return constants.KafkaTopicCRDName
	case "KafkaUser":
		return constants.KafkaUserCRDName
	case "LokiStack":
		return constants.LokiCRDName
	default:
		return ""
	}
}

func loadYAMLToJSON(name string) (JSONInterface, error) {
	var result JSONInterface
	yamlBytes, err := content.ReadFile(name)
	if err != nil {
		return result, err
	}
	err = yaml.Unmarshal(yamlBytes, &result)
	return result, err
}

func convertJSONToUnstructured(jsonMap JSONInterface) (*unstructured.Unstructured, error) {
	u := unstructured.Unstructured{}

	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		return &u, err
	}
	err = json.Unmarshal(jsonBytes, &u.Object)
	return &u, err
}

func loadYaml(name string, fv *JSONInterface) (*unstructured.Unstructured, error) {
	jsonMap, err := loadYAMLToJSON(name)
	if err != nil {
		return nil, err
	}

	if fv != nil {
		jsonMap = helper.Merge(jsonMap, *fv)
	}

	u, err := convertJSONToUnstructured(jsonMap)
	return u, err
}
