package reconcilers

import (
	"fmt"
	"strings"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	corev1 "k8s.io/api/core/v1"
)

func KafkaAddress(spec *flowsv1alpha1.FlowCollectorKafka, namespace string, operatorsAutoInstall *[]string) string {
	if len(namespace) > 0 && Contains(operatorsAutoInstall, constants.KafkaOperator) {
		return fmt.Sprintf("kafka-cluster-kafka-bootstrap.%s.svc.cluster.local", namespace)
	}
	return spec.Address
}

func KafkaTopic(spec *flowsv1alpha1.FlowCollectorKafka, operatorsAutoInstall *[]string) string {
	if Contains(operatorsAutoInstall, constants.KafkaOperator) {
		return "network-flows"
	}
	return spec.Topic
}

func lokistackGatewayURL(namespace string) string {
	return fmt.Sprintf("https://lokistack-gateway-http.%s.svc.cluster.local:8080/api/logs/v1/network/", namespace)
}

func LokiURL(flowCollector *flowsv1alpha1.FlowCollectorSpec, namespace string) string {
	// force loki url to loki gateway from operator if requested
	if Contains(flowCollector.OperatorsAutoInstall, constants.LokiOperator) {
		return lokistackGatewayURL(namespace)
	}
	return flowCollector.Loki.URL
}

func LokiQuerierURL(flowCollector *flowsv1alpha1.FlowCollectorSpec, namespace string) string {
	// force loki url to loki gateway from operator if requested
	if Contains(flowCollector.OperatorsAutoInstall, constants.LokiOperator) {
		return lokistackGatewayURL(namespace)
	} else if flowCollector.Loki.QuerierURL != "" {
		return flowCollector.Loki.QuerierURL
	}
	return LokiURL(flowCollector, namespace)
}

func LokiStatusURL(flowCollector *flowsv1alpha1.FlowCollectorSpec, namespace string) string {
	// force loki url to loki query front end from operator if requested
	if Contains(flowCollector.OperatorsAutoInstall, constants.LokiOperator) {
		return fmt.Sprintf("https://lokistack-query-frontend-http.%s.svc.cluster.local:3100", namespace)
	} else if flowCollector.Loki.StatusURL != "" {
		return flowCollector.Loki.StatusURL
	}
	return LokiQuerierURL(flowCollector, namespace)
}

func LokiSecretName(flowCollector *flowsv1alpha1.FlowCollectorSpec) string {
	if flowCollector.Loki.AutoInstallSpec != nil {
		return flowCollector.Loki.AutoInstallSpec.SecretName
	}
	return "loki-secret"
}

func SendAuthToken(flowCollector *flowsv1alpha1.FlowCollectorSpec) bool {
	if Contains(flowCollector.OperatorsAutoInstall, constants.LokiOperator) {
		return true
	}
	return flowCollector.Loki.UseHostToken() || flowCollector.Loki.ForwardUserToken()
}

func SendHostToken(flowCollector *flowsv1alpha1.FlowCollectorSpec) bool {
	if Contains(flowCollector.OperatorsAutoInstall, constants.LokiOperator) {
		return true
	}
	return flowCollector.Loki.UseHostToken()
}

func TenantID(flowCollector *flowsv1alpha1.FlowCollectorSpec) string {
	if Contains(flowCollector.OperatorsAutoInstall, constants.LokiOperator) {
		return "network"
	}
	return flowCollector.Loki.TenantID
}

func LokiTLS(flowCollector *flowsv1alpha1.FlowCollectorSpec) *flowsv1alpha1.ClientTLS {
	if Contains(flowCollector.OperatorsAutoInstall, constants.LokiOperator) {
		return &flowsv1alpha1.ClientTLS{
			Enable:             true,
			InsecureSkipVerify: false,
			CACert: flowsv1alpha1.CertificateReference{
				Type:     "configmap",
				Name:     "lokistack-ca-bundle",
				CertFile: "service-ca.crt",
			},
		}
	}
	return &flowCollector.Loki.TLS
}

func KafkaTLS(spec *flowsv1alpha1.FlowCollectorKafka, operatorsAutoInstall *[]string) *flowsv1alpha1.ClientTLS {
	if Contains(operatorsAutoInstall, constants.KafkaOperator) {
		return &flowsv1alpha1.ClientTLS{
			Enable:             true,
			InsecureSkipVerify: false,
			CACert: flowsv1alpha1.CertificateReference{
				Type:     "secret",
				Name:     "kafka-cluster-cluster-ca-cert",
				CertFile: "ca.crt",
			},
			UserCert: flowsv1alpha1.CertificateReference{
				Type:     "secret",
				Name:     "flp-kafka",
				CertFile: "user.crt",
				CertKey:  "user.key",
			},
		}
	}
	return &spec.TLS
}

func MatchingSecret(secret *corev1.Secret) bool {
	match := false
	for _, v := range secret.Labels {
		vLowerCase := strings.ToLower(v)
		match = match ||
			strings.Contains(vLowerCase, constants.LokiOperator) ||
			strings.Contains(vLowerCase, constants.KafkaOperator)
	}

	return match ||
		strings.Contains(secret.Name, constants.LokiOperator) ||
		strings.Contains(secret.Name, constants.KafkaOperator)
}

func MatchingCRD(crdName string) bool {
	return Contains(&[]string{
		constants.KafkaCRDName,
		constants.KafkaTopicCRDName,
		constants.KafkaUserCRDName,
		constants.LokiCRDName,
	}, crdName)
}

func Contains(array *[]string, value string) bool {
	if array == nil {
		return false
	}

	for _, v := range *array {
		if v == value {
			return true
		}
	}
	return false
}
