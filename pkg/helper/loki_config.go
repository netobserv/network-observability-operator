package helper

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
)

type LokiConfig struct {
	flowslatest.LokiManualParams
	BatchWait    *metav1.Duration
	BatchSize    int64
	Timeout      *metav1.Duration
	MinBackoff   *metav1.Duration
	MaxBackoff   *metav1.Duration
	MaxRetries   *int32
	StaticLabels map[string]string
}

func NewLokiConfig(spec *flowslatest.FlowCollectorLoki) LokiConfig {
	loki := LokiConfig{
		BatchWait:    spec.BatchWait,
		BatchSize:    spec.BatchSize,
		Timeout:      spec.Timeout,
		MinBackoff:   spec.MinBackoff,
		MaxBackoff:   spec.MaxBackoff,
		MaxRetries:   spec.MaxRetries,
		StaticLabels: spec.StaticLabels,
	}
	switch spec.Mode {
	case flowslatest.LokiModeLokiStack:
		dotNamespace := ""
		if len(spec.LokiStack.Namespace) > 0 {
			dotNamespace = "." + spec.LokiStack.Namespace
		}
		gatewayURL := fmt.Sprintf("https://%s-gateway-http%s.svc:8080/api/logs/v1/network/", spec.LokiStack.Name, dotNamespace)
		loki.LokiManualParams = flowslatest.LokiManualParams{
			QuerierURL:  gatewayURL,
			IngesterURL: gatewayURL,
			StatusURL:   fmt.Sprintf("https://%s-query-frontend-http%s.svc:3100/", spec.LokiStack.Name, dotNamespace),
			TenantID:    "network",
			AuthToken:   flowslatest.LokiAuthForwardUserToken,
			TLS: flowslatest.ClientTLS{
				Enable: true,
				CACert: flowslatest.CertificateReference{
					Type:      flowslatest.RefTypeConfigMap,
					Name:      fmt.Sprintf("%s-gateway-ca-bundle", spec.LokiStack.Name),
					Namespace: spec.LokiStack.Namespace,
					CertFile:  "service-ca.crt",
				},
			},
			StatusTLS: flowslatest.ClientTLS{
				Enable: true,
				CACert: flowslatest.CertificateReference{
					Type:      flowslatest.RefTypeConfigMap,
					Name:      fmt.Sprintf("%s-ca-bundle", spec.LokiStack.Name),
					Namespace: spec.LokiStack.Namespace,
					CertFile:  "service-ca.crt",
				},
				UserCert: flowslatest.CertificateReference{
					Type:      flowslatest.RefTypeSecret,
					Name:      fmt.Sprintf("%s-query-frontend-http", spec.LokiStack.Name),
					Namespace: spec.LokiStack.Namespace,
					CertFile:  "tls.crt",
					CertKey:   "tls.key",
				},
			},
		}
	case flowslatest.LokiModeMonolithic:
		loki.LokiManualParams = flowslatest.LokiManualParams{
			QuerierURL:  spec.Monolithic.URL,
			IngesterURL: spec.Monolithic.URL,
			StatusURL:   spec.Monolithic.URL,
			TenantID:    spec.Monolithic.TenantID,
			TLS:         spec.Monolithic.TLS,
		}
	case flowslatest.LokiModeMicroservices:
		loki.LokiManualParams = flowslatest.LokiManualParams{
			QuerierURL:  spec.Microservices.QuerierURL,
			IngesterURL: spec.Microservices.IngesterURL,
			StatusURL:   spec.Microservices.QuerierURL,
			TenantID:    spec.Microservices.TenantID,
			TLS:         spec.Microservices.TLS,
		}
	case flowslatest.LokiModeManual:
		loki.LokiManualParams = spec.Manual
	default:
		// Default / fallback => manual
		loki.LokiManualParams = spec.Manual
	}
	return loki
}

func (l *LokiConfig) UseForwardToken() bool {
	return l.LokiManualParams.AuthToken == flowslatest.LokiAuthForwardUserToken
}

func (l *LokiConfig) UseHostToken() bool {
	return l.LokiManualParams.AuthToken == flowslatest.LokiAuthUseHostToken
}
