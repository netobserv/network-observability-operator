package helper

import (
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetSampling(spec *flowslatest.FlowCollectorSpec) int {
	if UseEBPF(spec) {
		return int(*spec.Agent.EBPF.Sampling)
	}
	return int(spec.Agent.IPFIX.Sampling)
}

func UseEBPF(spec *flowslatest.FlowCollectorSpec) bool {
	return spec.Agent.Type == flowslatest.AgentEBPF
}

func UseIPFIX(spec *flowslatest.FlowCollectorSpec) bool {
	return spec.Agent.Type == flowslatest.AgentIPFIX
}

func UseKafka(spec *flowslatest.FlowCollectorSpec) bool {
	return spec.DeploymentModel == flowslatest.DeploymentModelKafka
}

func HasKafkaExporter(spec *flowslatest.FlowCollectorSpec) bool {
	for _, ex := range spec.Exporters {
		if ex.Type == flowslatest.KafkaExporter {
			return true
		}
	}
	return false
}

func HPADisabled(spec *flowslatest.FlowCollectorHPA) bool {
	return spec.Status == flowslatest.HPAStatusDisabled
}

func HPAEnabled(spec *flowslatest.FlowCollectorHPA) bool {
	return spec.Status == flowslatest.HPAStatusEnabled
}

func GetRecordTypes(processor *flowslatest.FlowCollectorFLP) []string {
	outputRecordTypes := []string{constants.FlowLogType}
	if processor.LogTypes != nil {
		switch *processor.LogTypes {
		case flowslatest.LogTypeFlows:
			outputRecordTypes = []string{
				constants.FlowLogType,
			}
		case flowslatest.LogTypeConversations:
			outputRecordTypes = []string{
				constants.NewConnectionType,
				constants.HeartbeatType,
				constants.EndConnectionType,
			}
		case flowslatest.LogTypeEndedConversations:
			outputRecordTypes = []string{
				constants.EndConnectionType,
			}
		case flowslatest.LogTypeAll:
			outputRecordTypes = []string{
				constants.FlowLogType,
				constants.NewConnectionType,
				constants.HeartbeatType,
				constants.EndConnectionType,
			}
		}
	}
	return outputRecordTypes
}

func UseSASL(cfg *flowslatest.SASLConfig) bool {
	return cfg.Type == flowslatest.SASLPlain || cfg.Type == flowslatest.SASLScramSHA512
}

func UseLoki(spec *flowslatest.FlowCollectorSpec) bool {
	// nil should fallback to default value, which is "true"
	return spec.Loki.Enable == nil || *spec.Loki.Enable
}

func LokiNoAuthToken(spec *flowslatest.FlowCollectorLoki) bool {
	switch spec.Mode {
	case flowslatest.LokiModeLokiStack:
		return false
	default:
		return spec.Manual.AuthToken == flowslatest.LokiAuthDisabled
	}
}

func LokiUseHostToken(spec *flowslatest.FlowCollectorLoki) bool {
	switch spec.Mode {
	case flowslatest.LokiModeLokiStack:
		return false
	default:
		return spec.Manual.AuthToken == flowslatest.LokiAuthUseHostToken
	}
}

func LokiForwardUserToken(spec *flowslatest.FlowCollectorLoki) bool {
	switch spec.Mode {
	case flowslatest.LokiModeLokiStack:
		return true
	default:
		return spec.Manual.AuthToken == flowslatest.LokiAuthForwardUserToken
	}
}

func getLokiStackNameAndNamespace(spec *flowslatest.LokiStack) (string, string) {
	if spec != nil {
		return spec.Name, spec.Namespace
	}
	return "loki", "netobserv"
}

func lokiStackGatewayURL(spec *flowslatest.FlowCollectorLoki) string {
	name, namespace := getLokiStackNameAndNamespace(spec.LokiStack)
	return "https://" + name + "-gateway-http." + namespace + ".svc:8080/api/logs/v1/network/"
}

func lokiStackStatusURL(spec *flowslatest.FlowCollectorLoki) string {
	name, namespace := getLokiStackNameAndNamespace(spec.LokiStack)
	return "https://" + name + "-query-frontend-http." + namespace + ".svc:3100/"
}

func LokiIngesterURL(spec *flowslatest.FlowCollectorLoki) string {
	switch spec.Mode {
	case flowslatest.LokiModeLokiStack:
		return lokiStackGatewayURL(spec)
	default:
		return spec.Manual.IngesterURL
	}
}

func LokiQuerierURL(spec *flowslatest.FlowCollectorLoki) string {
	switch spec.Mode {
	case flowslatest.LokiModeLokiStack:
		return lokiStackGatewayURL(spec)
	default:
		if spec.Manual.QuerierURL != "" {
			return spec.Manual.QuerierURL
		}
		return spec.Manual.IngesterURL
	}
}

func LokiStatusURL(spec *flowslatest.FlowCollectorLoki) string {
	switch spec.Mode {
	case flowslatest.LokiModeLokiStack:
		return lokiStackStatusURL(spec)
	default:
		if spec.Manual.StatusURL != "" {
			return spec.Manual.StatusURL
		}
		return LokiQuerierURL(spec)
	}
}

func LokiTenantID(spec *flowslatest.FlowCollectorLoki) string {
	switch spec.Mode {
	case flowslatest.LokiModeLokiStack:
		return "network"
	default:
		return spec.Manual.TenantID
	}
}

func LokiTLS(spec *flowslatest.FlowCollectorLoki) *flowslatest.ClientTLS {
	switch spec.Mode {
	case flowslatest.LokiModeLokiStack:
		name, _ := getLokiStackNameAndNamespace(spec.LokiStack)
		clientTLS := &flowslatest.ClientTLS{
			Enable: true,
			CACert: flowslatest.CertificateReference{
				Type:     flowslatest.RefTypeConfigMap,
				Name:     name + "-gateway-ca-bundle",
				CertFile: "service-ca.crt",
			},
			InsecureSkipVerify: false,
		}
		return clientTLS
	default:
		return &spec.Manual.TLS
	}
}

func LokiStatusTLS(spec *flowslatest.FlowCollectorLoki) *flowslatest.ClientTLS {
	switch spec.Mode {
	case flowslatest.LokiModeLokiStack:
		name, _ := getLokiStackNameAndNamespace(spec.LokiStack)
		clientTLS := &flowslatest.ClientTLS{
			Enable: true,
			CACert: flowslatest.CertificateReference{
				Type:     flowslatest.RefTypeConfigMap,
				Name:     name + "-ca-bundle",
				CertFile: "service-ca.crt",
			},
			InsecureSkipVerify: false,
			UserCert: flowslatest.CertificateReference{
				Type:     flowslatest.RefTypeSecret,
				Name:     name + "-query-frontend-http",
				CertFile: "tls.crt",
				CertKey:  "tls.key",
			},
		}
		return clientTLS
	default:
		if spec.Manual.StatusURL != "" {
			return &spec.Manual.StatusTLS
		}
		return &spec.Manual.TLS
	}
}

func UseConsolePlugin(spec *flowslatest.FlowCollectorSpec) bool {
	return UseLoki(spec) &&
		// nil should fallback to default value, which is "true"
		(spec.ConsolePlugin.Enable == nil || *spec.ConsolePlugin.Enable)
}

func IsFeatureEnabled(spec *flowslatest.FlowCollectorEBPF, feature flowslatest.AgentFeature) bool {
	for _, f := range spec.Features {
		if f == feature {
			return true
		}
	}
	return false
}

func IsPrivileged(spec *flowslatest.FlowCollectorEBPF) bool {
	return spec.Privileged
}

func IsPktDropEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	if IsPrivileged(spec) && IsFeatureEnabled(spec, flowslatest.PacketDrop) {
		return true
	}
	return false
}

func IsDNSTrackingEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	if IsPrivileged(spec) && IsFeatureEnabled(spec, flowslatest.DNSTracking) {
		return true
	}
	return false
}

func IsFlowRTTEnabled(spec *flowslatest.FlowCollectorEBPF) bool {
	return IsFeatureEnabled(spec, flowslatest.FlowRTT)
}

func PtrBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func PtrInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func IsOwned(obj client.Object) bool {
	refs := obj.GetOwnerReferences()
	return len(refs) > 0 && strings.HasPrefix(refs[0].APIVersion, flowslatest.GroupVersion.Group)
}
