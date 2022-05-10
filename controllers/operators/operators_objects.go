package operators

import (
	sv1b2 "github.com/RedHatInsights/strimzi-client-go/apis/kafka.strimzi.io/v1beta2"

	gv1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	lv1beta1 "github.com/grafana/loki/operator/api/v1beta1"
	"github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	ov1 "github.com/operator-framework/api/pkg/operators/v1"
	ov1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	pv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func operatorGroup(name string, namespace string) *ov1.OperatorGroup {
	return &ov1.OperatorGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       constants.OperatorGroup,
			APIVersion: "operators.coreos.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{"part-of": constants.ObservabilityName},
		},
		Spec: ov1.OperatorGroupSpec{
			TargetNamespaces: []string{namespace},
		},
	}
}

func subscription(name string, namespace string) *ov1alpha1.Subscription {
	return &ov1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       constants.SubscriptionKind,
			APIVersion: "operators.coreos.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{"part-of": constants.ObservabilityName},
		},
	}
}

func grafanaSubscription(name string, namespace string) *ov1alpha1.Subscription {
	s := subscription(name, namespace)
	s.Spec = &ov1alpha1.SubscriptionSpec{
		CatalogSource:          "community-operators",
		CatalogSourceNamespace: "openshift-marketplace",
		Package:                "grafana-operator",
		Channel:                "v4",
		StartingCSV:            "grafana-operator.v4.2.0",
		InstallPlanApproval:    "Automatic",
	}
	return s
}

func strimziSubscription(name string, namespace string) *ov1alpha1.Subscription {
	s := subscription(name, namespace)
	s.Spec = &ov1alpha1.SubscriptionSpec{
		CatalogSource:          "community-operators",
		CatalogSourceNamespace: "openshift-marketplace",
		Package:                "strimzi-kafka-operator",
		Channel:                "stable",
		StartingCSV:            "strimzi-cluster-operator.v0.28.0",
		InstallPlanApproval:    "Automatic",
	}
	return s
}

func lokiSubscription(name string, namespace string) *ov1alpha1.Subscription {
	s := subscription(name, namespace)
	s.Spec = &ov1alpha1.SubscriptionSpec{
		CatalogSource:          "redhat-operators",
		CatalogSourceNamespace: "openshift-marketplace",
		Package:                "loki-operator",
		Channel:                "candidate",
		StartingCSV:            "loki-operator.5.4.0-42",
		InstallPlanApproval:    "Automatic",
	}
	return s
}

func prometheusSubscription(name string, namespace string) *ov1alpha1.Subscription {
	s := subscription(name, namespace)
	s.Spec = &ov1alpha1.SubscriptionSpec{
		CatalogSource:          "community-operators",
		CatalogSourceNamespace: "openshift-marketplace",
		Package:                "prometheus",
		Channel:                "beta",
		StartingCSV:            "prometheusoperator.0.47.0",
		InstallPlanApproval:    "Automatic",
	}
	return s
}

func kafkaInstance(ns string, spec *v1alpha1.KafkaInstanceSpec) *sv1b2.Kafka {
	return &sv1b2.Kafka{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Kafka",
			APIVersion: "kafka.strimzi.io/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KafkaName,
			Namespace: ns,
			Labels:    map[string]string{"part-of": constants.ObservabilityName},
		},
		Spec: &sv1b2.KafkaSpec{
			Kafka: sv1b2.KafkaSpecKafka{
				Config:    spec.Kafka.Config,
				Listeners: spec.Kafka.Listeners,
				Replicas:  spec.Kafka.Replicas,
				Storage:   spec.Kafka.Storage,
				Version:   spec.Kafka.Version,
			},
			Zookeeper: sv1b2.KafkaSpecZookeeper{
				Replicas: spec.Zookeeper.Replicas,
				Storage:  spec.Zookeeper.Storage,
			},
		},
	}
}

func kafkaTopic(ns string, spec *v1alpha1.FlowCollectorKafka) *sv1b2.KafkaTopic {
	topicSpec := &sv1b2.KafkaTopicSpec{}
	if spec.TopicSpec != nil {
		topicSpec.Config = spec.TopicSpec.Config
		topicSpec.Partitions = spec.TopicSpec.Partitions
		topicSpec.Replicas = spec.TopicSpec.Replicas
	}
	return &sv1b2.KafkaTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KafkaTopic",
			APIVersion: "kafka.strimzi.io/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.TopicName, // kafka topic name is set from metadata name
			Namespace: ns,
			Labels:    map[string]string{"part-of": constants.ObservabilityName},
		},
		Spec: topicSpec,
	}
}

func grafanaInstance(ns string, spec *v1alpha1.GrafanaInstanceSpec) *gv1alpha1.Grafana {
	preferService := true
	return &gv1alpha1.Grafana{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Grafana",
			APIVersion: "integreatly.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.GrafanaName,
			Namespace: ns,
			Labels:    map[string]string{"part-of": constants.ObservabilityName},
		},
		Spec: gv1alpha1.GrafanaSpec{
			// fix grafana-service.network-observability.svc.cluster.local:3000 resolution for dashboards
			Client: &gv1alpha1.GrafanaClient{
				PreferService: &preferService,
			},
			Ingress: &gv1alpha1.GrafanaIngress{
				Enabled: spec.Ingress,
			},
			Config: spec.Config,
			// allow dashboards to be discovered https://github.com/grafana-operator/grafana-operator/blob/master/documentation/dashboards.md#dashboard-discovery
			DashboardLabelSelector: []*metav1.LabelSelector{
				{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "part-of",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{constants.ObservabilityName},
						},
					},
				},
			},
		},
	}
}

func grafanaDataSource(ns string, spec *v1alpha1.FlowCollectorLoki) *gv1alpha1.GrafanaDataSource {
	return &gv1alpha1.GrafanaDataSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GrafanaDataSource",
			APIVersion: "integreatly.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.LokiName,
			Namespace: ns,
			Labels:    map[string]string{"part-of": constants.ObservabilityName},
		},
		Spec: gv1alpha1.GrafanaDataSourceSpec{
			//TODO: manage other datasources than loki
			Datasources: []gv1alpha1.GrafanaDataSourceFields{
				{
					Name: "Loki",
					Type: "loki",
					Url:  reconcilers.QuerierURL(spec),
				},
			},
		},
	}
}

func grafanaDashboard(ns string, spec *v1alpha1.GrafanaDashboardSpec) *gv1alpha1.GrafanaDashboard {
	return &gv1alpha1.GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GrafanaDashboard",
			APIVersion: "integreatly.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ObservabilityName,
			Namespace: ns,
			Labels:    map[string]string{"part-of": constants.ObservabilityName},
		},
		Spec: gv1alpha1.GrafanaDashboardSpec{
			CustomFolderName: constants.ObservabilityName,
			//TODO: manage other datasources than loki
			Datasources: []gv1alpha1.GrafanaDashboardDatasource{
				{InputName: "Loki", DatasourceName: "Loki"},
			},
			ConfigMapRef: spec.ConfigMapRef,
		},
	}
}

func lokiInstance(ns string, spec *v1alpha1.LokiInstanceSpec) *lv1beta1.LokiStack {
	return &lv1beta1.LokiStack{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LokiStack",
			APIVersion: "loki.openshift.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.LokiName,
			Namespace: ns,
			Labels:    map[string]string{"part-of": constants.ObservabilityName},
		},
		Spec: lv1beta1.LokiStackSpec{
			Size:              spec.Size,
			ReplicationFactor: spec.ReplicationFactor,
			Storage:           spec.Storage,
			StorageClassName:  spec.StorageClassName,
			Tenants: &lv1beta1.TenantsSpec{
				Mode: lv1beta1.OpenshiftLogging, //TODO: customize gateway https://issues.redhat.com/browse/NETOBSERV-309
			},
		},
	}
}

func prometheusInstance(ns string, spec *v1alpha1.PrometheusInstanceSpec) *pv1.Prometheus {
	return &pv1.Prometheus{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Prometheus",
			APIVersion: "monitoring.coreos.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PrometheusName,
			Namespace: ns,
			Labels:    map[string]string{"part-of": constants.ObservabilityName},
		},
		Spec: pv1.PrometheusSpec{
			Replicas: spec.Replicas,
		},
	}
}
