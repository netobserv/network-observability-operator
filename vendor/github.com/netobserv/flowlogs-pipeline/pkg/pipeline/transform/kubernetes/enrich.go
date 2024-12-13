package kubernetes

import (
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/netobserv/flowlogs-pipeline/pkg/operational"
	inf "github.com/netobserv/flowlogs-pipeline/pkg/pipeline/transform/kubernetes/informers"
	"github.com/sirupsen/logrus"
)

var informers inf.InformersInterface = &inf.Informers{}

// For testing
func MockInformers() {
	informers = inf.NewInformersMock()
}

func InitFromConfig(config api.NetworkTransformKubeConfig, opMetrics *operational.Metrics) error {
	return informers.InitFromConfig(config, opMetrics)
}

func Enrich(outputEntry config.GenericMap, rule *api.K8sRule) {
	ip, ok := outputEntry.LookupString(rule.IPField)
	if !ok {
		return
	}
	potentialKeys := informers.BuildSecondaryNetworkKeys(outputEntry, rule)
	kubeInfo, err := informers.GetInfo(potentialKeys, ip)
	if err != nil {
		logrus.WithError(err).Tracef("can't find kubernetes info for keys %v and IP %s", potentialKeys, ip)
		return
	}
	if rule.Assignee != "otel" {
		// NETOBSERV-666: avoid putting empty namespaces or Loki aggregation queries will
		// differentiate between empty and nil namespaces.
		if kubeInfo.Namespace != "" {
			outputEntry[rule.Output+"_Namespace"] = kubeInfo.Namespace
		}
		outputEntry[rule.Output+"_Name"] = kubeInfo.Name
		outputEntry[rule.Output+"_Type"] = kubeInfo.Type
		outputEntry[rule.Output+"_OwnerName"] = kubeInfo.Owner.Name
		outputEntry[rule.Output+"_OwnerType"] = kubeInfo.Owner.Type
		outputEntry[rule.Output+"_NetworkName"] = kubeInfo.NetworkName
		if rule.LabelsPrefix != "" {
			for labelKey, labelValue := range kubeInfo.Labels {
				outputEntry[rule.LabelsPrefix+"_"+labelKey] = labelValue
			}
		}
		if kubeInfo.HostIP != "" {
			outputEntry[rule.Output+"_HostIP"] = kubeInfo.HostIP
			if kubeInfo.HostName != "" {
				outputEntry[rule.Output+"_HostName"] = kubeInfo.HostName
			}
		}
		fillInK8sZone(outputEntry, rule, kubeInfo, "_Zone")
	} else {
		// NOTE: Some of these fields are taken from opentelemetry specs.
		// See https://opentelemetry.io/docs/specs/semconv/resource/k8s/
		// Other fields (not specified in the specs) are named similarly
		if kubeInfo.Namespace != "" {
			outputEntry[rule.Output+"k8s.namespace.name"] = kubeInfo.Namespace
		}
		switch kubeInfo.Type {
		case inf.TypeNode:
			outputEntry[rule.Output+"k8s.node.name"] = kubeInfo.Name
			outputEntry[rule.Output+"k8s.node.uid"] = kubeInfo.UID
		case inf.TypePod:
			outputEntry[rule.Output+"k8s.pod.name"] = kubeInfo.Name
			outputEntry[rule.Output+"k8s.pod.uid"] = kubeInfo.UID
		case inf.TypeService:
			outputEntry[rule.Output+"k8s.service.name"] = kubeInfo.Name
			outputEntry[rule.Output+"k8s.service.uid"] = kubeInfo.UID
		}
		outputEntry[rule.Output+"k8s.name"] = kubeInfo.Name
		outputEntry[rule.Output+"k8s.type"] = kubeInfo.Type
		outputEntry[rule.Output+"k8s.owner.name"] = kubeInfo.Owner.Name
		outputEntry[rule.Output+"k8s.owner.type"] = kubeInfo.Owner.Type
		if rule.LabelsPrefix != "" {
			for labelKey, labelValue := range kubeInfo.Labels {
				outputEntry[rule.LabelsPrefix+"."+labelKey] = labelValue
			}
		}
		if kubeInfo.HostIP != "" {
			outputEntry[rule.Output+"k8s.host.ip"] = kubeInfo.HostIP
			if kubeInfo.HostName != "" {
				outputEntry[rule.Output+"k8s.host.name"] = kubeInfo.HostName
			}
		}
		fillInK8sZone(outputEntry, rule, kubeInfo, "k8s.zone")
	}
}

const nodeZoneLabelName = "topology.kubernetes.io/zone"

func fillInK8sZone(outputEntry config.GenericMap, rule *api.K8sRule, kubeInfo *inf.Info, zonePrefix string) {
	if !rule.AddZone {
		// Nothing to do
		return
	}
	switch kubeInfo.Type {
	case inf.TypeNode:
		zone, ok := kubeInfo.Labels[nodeZoneLabelName]
		if ok {
			outputEntry[rule.Output+zonePrefix] = zone
		}
		return
	case inf.TypePod:
		nodeInfo, err := informers.GetNodeInfo(kubeInfo.HostName)
		if err != nil {
			logrus.WithError(err).Tracef("can't find nodes info for node %v", kubeInfo.HostName)
			return
		}
		if nodeInfo != nil {
			zone, ok := nodeInfo.Labels[nodeZoneLabelName]
			if ok {
				outputEntry[rule.Output+zonePrefix] = zone
			}
		}
		return

	case inf.TypeService:
		// A service is not assigned to a dedicated zone, skipping
		return
	}
}

func EnrichLayer(outputEntry config.GenericMap, rule *api.K8sInfraRule) {
	outputEntry[rule.Output] = "infra"
	for _, nsnameFields := range rule.NamespaceNameFields {
		if namespace, _ := outputEntry.LookupString(nsnameFields.Namespace); namespace != "" {
			name, _ := outputEntry.LookupString(nsnameFields.Name)
			if objectIsApp(namespace, name, rule) {
				outputEntry[rule.Output] = "app"
				return
			}
		}
	}
}

func objectIsApp(namespace, name string, rule *api.K8sInfraRule) bool {
	for _, prefix := range rule.InfraPrefixes {
		if strings.HasPrefix(namespace, prefix) {
			return false
		}
	}
	for _, ref := range rule.InfraRefs {
		if namespace == ref.Namespace && name == ref.Name {
			return false
		}
	}
	return true
}
