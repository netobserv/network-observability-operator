package helper

import (
	"fmt"

	flowslatest "github.com/netobserv/network-observability-operator/api/flowcollector/v1beta2"
)

type LokiConfig struct {
	flowslatest.LokiManualParams
}

func NewLokiConfig(spec *flowslatest.FlowCollectorLoki, namespace string, useGRPC bool) LokiConfig {
	loki := LokiConfig{}
	switch spec.Mode {
	case flowslatest.LokiModeLokiStack:
		ns := namespace
		if len(spec.LokiStack.Namespace) > 0 {
			ns = spec.LokiStack.Namespace
		}
		gatewayURL := fmt.Sprintf("https://%s-gateway-http.%s.svc:8080/api/logs/v1/network/", spec.LokiStack.Name, ns)
		ingesterURL := gatewayURL
		if useGRPC {
			ingesterURL = fmt.Sprintf("%s-distributor-grpc.%s.svc:9095", spec.LokiStack.Name, ns)
		}
		// Configure TLS based on client type
		tlsConfig := flowslatest.ClientTLS{
			Enable: true,
		}

		// Set TLS certificates based on the connection type
		if useGRPC {
			// For gRPC ingester connections: use the Loki signing CA
			tlsConfig.CACert = flowslatest.CertificateReference{
				Type:      flowslatest.RefTypeConfigMap,
				Name:      fmt.Sprintf("%s-ca-bundle", spec.LokiStack.Name),
				Namespace: spec.LokiStack.Namespace,
				CertFile:  "service-ca.crt",
			}
			tlsConfig.UserCert = flowslatest.CertificateReference{
				Type:      flowslatest.RefTypeSecret,
				Name:      fmt.Sprintf("%s-distributor-grpc", spec.LokiStack.Name),
				Namespace: spec.LokiStack.Namespace,
				CertFile:  "tls.crt",
				CertKey:   "tls.key",
			}
		} else {
			// For HTTP gateway connections: use the OpenShift service serving CA
			tlsConfig.CACert = flowslatest.CertificateReference{
				Type:      flowslatest.RefTypeConfigMap,
				Name:      fmt.Sprintf("%s-gateway-ca-bundle", spec.LokiStack.Name),
				Namespace: spec.LokiStack.Namespace,
				CertFile:  "service-ca.crt",
			}
		}

		loki.LokiManualParams = flowslatest.LokiManualParams{
			QuerierURL:  gatewayURL,
			IngesterURL: ingesterURL,
			StatusURL:   fmt.Sprintf("https://%s-query-frontend-http.%s.svc:3100/", spec.LokiStack.Name, ns),
			TenantID:    "network",
			AuthToken:   flowslatest.LokiAuthForwardUserToken,
			TLS:         tlsConfig,
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
