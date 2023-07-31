package helper

import (
	"strings"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/volumes"
	promConfig "github.com/prometheus/common/config"
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

func LokiNoAuthToken(spec *flowslatest.FlowCollectorLoki) bool {
	return spec.Manual.AuthToken == flowslatest.LokiAuthDisabled
}

func LokiUseHostToken(spec *flowslatest.FlowCollectorLoki) bool {
	return spec.Manual.AuthToken == flowslatest.LokiAuthUseHostToken
}

func LokiForwardUserToken(spec *flowslatest.FlowCollectorLoki) bool {
	return spec.Manual.AuthToken == flowslatest.LokiAuthForwardUserToken
}

func GetLokiStatusTLS(spec *flowslatest.FlowCollectorLoki) flowslatest.ClientTLS {
	if spec.Manual.StatusURL != "" {
		return spec.Manual.StatusTLS
	}
	return spec.Manual.TLS
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

func LokiModeLokiStack(spec *flowslatest.FlowCollectorLoki) bool {
	return spec.Mode == "LOKISTACK"
}

func LokiIngesterURL(spec *flowslatest.FlowCollectorLoki) string {

	switch spec.Mode {
	case "MANUAL":
		{
			return spec.Manual.IngesterURL
		}
	case "LOKISTACK":
		{
			return "https://" + spec.LokiStack.Name + "-gateway-http.netobserv.svc:8080/api/logs/v1/network/"
		}
	default:
		return "http://loki:3100/"
	}
}

func LokiTenantID(spec *flowslatest.FlowCollectorLoki) string {
	switch spec.Mode {
	case "MANUAL":
		{
			return spec.Manual.TenantID
		}
	case "LOKISTACK":
		{
			return "network"
		}
	default:
		return "netobserv"
	}
}

func LokiTLSClient(spec *flowslatest.FlowCollectorLoki, authorization *promConfig.Authorization, vol *volumes.Builder) *promConfig.HTTPClientConfig {

	switch spec.Mode {

	case "MANUAL":
		{
			if spec.Manual.TLS.Enable {
				if spec.Manual.TLS.InsecureSkipVerify {
					return &promConfig.HTTPClientConfig{
						Authorization: authorization,
						TLSConfig: promConfig.TLSConfig{
							InsecureSkipVerify: true,
						},
					}
				}
				caPath := vol.AddCACertificate(&spec.Manual.TLS, "loki-certs")
				return &promConfig.HTTPClientConfig{
					Authorization: authorization,
					TLSConfig: promConfig.TLSConfig{
						CAFile: caPath,
					},
				}

			}
			return &promConfig.HTTPClientConfig{
				Authorization: authorization,
			}
		}
	case "LOKISTACK":
		{
			certRef := flowslatest.CertificateReference{
				Type:     flowslatest.RefTypeConfigMap,
				Name:     spec.LokiStack.Name + "-gateway-ca-bundle",
				CertFile: "service-ca.crt",
			}
			clientTLS := &flowslatest.ClientTLS{
				CACert: certRef,
			}
			caPath := vol.AddCACertificate(clientTLS, "loki-certs")
			return &promConfig.HTTPClientConfig{
				Authorization: authorization,
				TLSConfig: promConfig.TLSConfig{
					CAFile: caPath,
				}}
		}
	}

	return nil
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
