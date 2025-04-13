/*
 * Copyright (C) 2024 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package write

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/vmware/go-ipfix/pkg/entities"
	ipfixExporter "github.com/vmware/go-ipfix/pkg/exporter"
	"github.com/vmware/go-ipfix/pkg/registry"
)

type writeIpfix struct {
	hostPort           string
	transport          string
	templateIDv4       uint16
	templateIDv6       uint16
	enrichEnterpriseID uint32
	exporter           *ipfixExporter.ExportingProcess
	entitiesV4         []entities.InfoElementWithValue
	entitiesV6         []entities.InfoElementWithValue
}

type FieldMap struct {
	Key      string
	Getter   func(entities.InfoElementWithValue) any
	Setter   func(entities.InfoElementWithValue, any)
	Matcher  func(entities.InfoElementWithValue, any) bool
	Optional bool
}

// IPv6Type value as defined in IEEE 802: https://www.iana.org/assignments/ieee-802-numbers/ieee-802-numbers.xhtml
const IPv6Type uint16 = 0x86DD

var (
	ilog       = logrus.WithField("component", "write.Ipfix")
	IANAFields = []string{
		"ethernetType",
		"flowDirection",
		"sourceMacAddress",
		"destinationMacAddress",
		"protocolIdentifier",
		"sourceTransportPort",
		"destinationTransportPort",
		"octetDeltaCount",
		"flowStartMilliseconds",
		"flowEndMilliseconds",
		"packetDeltaCount",
		"interfaceName",
	}
	IPv4IANAFields = append([]string{
		"sourceIPv4Address",
		"destinationIPv4Address",
	}, IANAFields...)
	IPv6IANAFields = append([]string{
		"sourceIPv6Address",
		"destinationIPv6Address",
		"nextHeaderIPv6",
	}, IANAFields...)
	KubeFields = []string{
		"sourcePodNamespace",
		"sourcePodName",
		"destinationPodNamespace",
		"destinationPodName",
		"sourceNodeName",
		"destinationNodeName",
	}
	CustomNetworkFields = []string{
		"timeFlowRttNs",
		"interfaces",
		"directions",
	}

	MapIPFIXKeys = map[string]FieldMap{
		"sourceIPv4Address": {
			Key:    "SrcAddr",
			Getter: func(elt entities.InfoElementWithValue) any { return elt.GetIPAddressValue().String() },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetIPAddressValue(net.ParseIP(rec.(string))) },
		},
		"destinationIPv4Address": {
			Key:    "DstAddr",
			Getter: func(elt entities.InfoElementWithValue) any { return elt.GetIPAddressValue().String() },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetIPAddressValue(net.ParseIP(rec.(string))) },
		},
		"sourceIPv6Address": {
			Key:    "SrcAddr",
			Getter: func(elt entities.InfoElementWithValue) any { return elt.GetIPAddressValue().String() },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetIPAddressValue(net.ParseIP(rec.(string))) },
		},
		"destinationIPv6Address": {
			Key:    "DstAddr",
			Getter: func(elt entities.InfoElementWithValue) any { return elt.GetIPAddressValue().String() },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetIPAddressValue(net.ParseIP(rec.(string))) },
		},
		"nextHeaderIPv6": {
			Key:    "Proto",
			Getter: func(elt entities.InfoElementWithValue) any { return elt.GetUnsigned8Value() },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetUnsigned8Value(rec.(uint8)) },
		},
		"sourceMacAddress": {
			Key:    "SrcMac",
			Getter: func(elt entities.InfoElementWithValue) any { return elt.GetMacAddressValue().String() },
			Setter: func(elt entities.InfoElementWithValue, rec any) {
				mac, _ := net.ParseMAC(rec.(string))
				elt.SetMacAddressValue(mac)
			},
		},
		"destinationMacAddress": {
			Key:    "DstMac",
			Getter: func(elt entities.InfoElementWithValue) any { return elt.GetMacAddressValue().String() },
			Setter: func(elt entities.InfoElementWithValue, rec any) {
				mac, _ := net.ParseMAC(rec.(string))
				elt.SetMacAddressValue(mac)
			},
		},
		"ethernetType": {
			Key:    "Etype",
			Getter: func(elt entities.InfoElementWithValue) any { return elt.GetUnsigned16Value() },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetUnsigned16Value(rec.(uint16)) },
		},
		"flowDirection": {
			Key: "IfDirections",
			Setter: func(elt entities.InfoElementWithValue, rec any) {
				if dirs, ok := rec.([]int); ok && len(dirs) > 0 {
					elt.SetUnsigned8Value(uint8(dirs[0]))
				}
			},
			Matcher: func(elt entities.InfoElementWithValue, expected any) bool {
				ifdirs := expected.([]int)
				return int(elt.GetUnsigned8Value()) == ifdirs[0]
			},
		},
		"directions": {
			Key: "IfDirections",
			Getter: func(elt entities.InfoElementWithValue) any {
				var dirs []int
				for _, dir := range strings.Split(elt.GetStringValue(), ",") {
					d, _ := strconv.Atoi(dir)
					dirs = append(dirs, d)
				}
				return dirs
			},
			Setter: func(elt entities.InfoElementWithValue, rec any) {
				if dirs, ok := rec.([]int); ok && len(dirs) > 0 {
					var asStr []string
					for _, dir := range dirs {
						asStr = append(asStr, strconv.Itoa(dir))
					}
					elt.SetStringValue(strings.Join(asStr, ","))
				}
			},
		},
		"protocolIdentifier": {
			Key:    "Proto",
			Getter: func(elt entities.InfoElementWithValue) any { return elt.GetUnsigned8Value() },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetUnsigned8Value(rec.(uint8)) },
		},
		"sourceTransportPort": {
			Key:    "SrcPort",
			Getter: func(elt entities.InfoElementWithValue) any { return elt.GetUnsigned16Value() },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetUnsigned16Value(rec.(uint16)) },
		},
		"destinationTransportPort": {
			Key:    "DstPort",
			Getter: func(elt entities.InfoElementWithValue) any { return elt.GetUnsigned16Value() },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetUnsigned16Value(rec.(uint16)) },
		},
		"octetDeltaCount": {
			Key:    "Bytes",
			Getter: func(elt entities.InfoElementWithValue) any { return elt.GetUnsigned64Value() },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetUnsigned64Value(rec.(uint64)) },
		},
		"flowStartMilliseconds": {
			Key:    "TimeFlowStartMs",
			Getter: func(elt entities.InfoElementWithValue) any { return int64(elt.GetUnsigned64Value()) },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetUnsigned64Value(uint64(rec.(int64))) },
		},
		"flowEndMilliseconds": {
			Key:    "TimeFlowEndMs",
			Getter: func(elt entities.InfoElementWithValue) any { return int64(elt.GetUnsigned64Value()) },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetUnsigned64Value(uint64(rec.(int64))) },
		},
		"packetDeltaCount": {
			Key:    "Packets",
			Getter: func(elt entities.InfoElementWithValue) any { return uint32(elt.GetUnsigned64Value()) },
			Setter: func(elt entities.InfoElementWithValue, rec any) { elt.SetUnsigned64Value(uint64(rec.(uint32))) },
		},
		"interfaceName": {
			Key: "Interfaces",
			Setter: func(elt entities.InfoElementWithValue, rec any) {
				if ifs, ok := rec.([]string); ok && len(ifs) > 0 {
					elt.SetStringValue(ifs[0])
				}
			},
			Matcher: func(elt entities.InfoElementWithValue, expected any) bool {
				ifs := expected.([]string)
				return elt.GetStringValue() == ifs[0]
			},
		},
		"interfaces": {
			Key:    "Interfaces",
			Getter: func(elt entities.InfoElementWithValue) any { return strings.Split(elt.GetStringValue(), ",") },
			Setter: func(elt entities.InfoElementWithValue, rec any) {
				if ifs, ok := rec.([]string); ok {
					elt.SetStringValue(strings.Join(ifs, ","))
				}
			},
		},
		"sourcePodNamespace": {
			Key:      "SrcK8S_Namespace",
			Getter:   func(elt entities.InfoElementWithValue) any { return elt.GetStringValue() },
			Setter:   func(elt entities.InfoElementWithValue, rec any) { elt.SetStringValue(rec.(string)) },
			Optional: true,
		},
		"sourcePodName": {
			Key:      "SrcK8S_Name",
			Getter:   func(elt entities.InfoElementWithValue) any { return elt.GetStringValue() },
			Setter:   func(elt entities.InfoElementWithValue, rec any) { elt.SetStringValue(rec.(string)) },
			Optional: true,
		},
		"destinationPodNamespace": {
			Key:      "DstK8S_Namespace",
			Getter:   func(elt entities.InfoElementWithValue) any { return elt.GetStringValue() },
			Setter:   func(elt entities.InfoElementWithValue, rec any) { elt.SetStringValue(rec.(string)) },
			Optional: true,
		},
		"destinationPodName": {
			Key:      "DstK8S_Name",
			Getter:   func(elt entities.InfoElementWithValue) any { return elt.GetStringValue() },
			Setter:   func(elt entities.InfoElementWithValue, rec any) { elt.SetStringValue(rec.(string)) },
			Optional: true,
		},
		"sourceNodeName": {
			Key:      "SrcK8S_HostName",
			Getter:   func(elt entities.InfoElementWithValue) any { return elt.GetStringValue() },
			Setter:   func(elt entities.InfoElementWithValue, rec any) { elt.SetStringValue(rec.(string)) },
			Optional: true,
		},
		"destinationNodeName": {
			Key:      "DstK8S_HostName",
			Getter:   func(elt entities.InfoElementWithValue) any { return elt.GetStringValue() },
			Setter:   func(elt entities.InfoElementWithValue, rec any) { elt.SetStringValue(rec.(string)) },
			Optional: true,
		},
		"timeFlowRttNs": {
			Key:      "TimeFlowRttNs",
			Getter:   func(elt entities.InfoElementWithValue) any { return int64(elt.GetUnsigned64Value()) },
			Setter:   func(elt entities.InfoElementWithValue, rec any) { elt.SetUnsigned64Value(uint64(rec.(int64))) },
			Optional: true,
		},
	}
)

func addElementToTemplate(elementName string, value []byte, elements *[]entities.InfoElementWithValue, registryID uint32) error {
	element, err := registry.GetInfoElement(elementName, registryID)
	if err != nil {
		ilog.WithError(err).Errorf("Did not find the element with name %s", elementName)
		return err
	}
	ie, err := entities.DecodeAndCreateInfoElementWithValue(element, value)
	if err != nil {
		ilog.WithError(err).Errorf("Failed to decode element %s", elementName)
		return err
	}
	*elements = append(*elements, ie)
	return nil
}

func addNetworkEnrichmentToTemplate(elements *[]entities.InfoElementWithValue, registryID uint32) error {
	for _, field := range CustomNetworkFields {
		if err := addElementToTemplate(field, nil, elements, registryID); err != nil {
			return err
		}
	}
	return nil
}

func addKubeContextToTemplate(elements *[]entities.InfoElementWithValue, registryID uint32) error {
	for _, field := range KubeFields {
		if err := addElementToTemplate(field, nil, elements, registryID); err != nil {
			return err
		}
	}
	return nil
}

func loadCustomRegistry(enterpriseID uint32) error {
	err := registry.InitNewRegistry(enterpriseID)
	if err != nil {
		ilog.WithError(err).Errorf("Failed to initialize registry")
		return err
	}
	err = registry.PutInfoElement((*entities.NewInfoElement("sourcePodNamespace", 7733, entities.String, enterpriseID, 65535)), enterpriseID)
	if err != nil {
		ilog.WithError(err).Errorf("Failed to register element")
		return err
	}
	err = registry.PutInfoElement((*entities.NewInfoElement("sourcePodName", 7734, entities.String, enterpriseID, 65535)), enterpriseID)
	if err != nil {
		ilog.WithError(err).Errorf("Failed to register element")
		return err
	}
	err = registry.PutInfoElement((*entities.NewInfoElement("destinationPodNamespace", 7735, entities.String, enterpriseID, 65535)), enterpriseID)
	if err != nil {
		ilog.WithError(err).Errorf("Failed to register element")
		return err
	}
	err = registry.PutInfoElement((*entities.NewInfoElement("destinationPodName", 7736, entities.String, enterpriseID, 65535)), enterpriseID)
	if err != nil {
		ilog.WithError(err).Errorf("Failed to register element")
		return err
	}
	err = registry.PutInfoElement((*entities.NewInfoElement("sourceNodeName", 7737, entities.String, enterpriseID, 65535)), enterpriseID)
	if err != nil {
		ilog.WithError(err).Errorf("Failed to register element")
		return err
	}
	err = registry.PutInfoElement((*entities.NewInfoElement("destinationNodeName", 7738, entities.String, enterpriseID, 65535)), enterpriseID)
	if err != nil {
		ilog.WithError(err).Errorf("Failed to register element")
		return err
	}
	err = registry.PutInfoElement((*entities.NewInfoElement("timeFlowRttNs", 7740, entities.Unsigned64, enterpriseID, 8)), enterpriseID)
	if err != nil {
		ilog.WithError(err).Errorf("Failed to register element")
		return err
	}
	err = registry.PutInfoElement((*entities.NewInfoElement("interfaces", 7741, entities.String, enterpriseID, 65535)), enterpriseID)
	if err != nil {
		ilog.WithError(err).Errorf("Failed to register element")
		return err
	}
	err = registry.PutInfoElement((*entities.NewInfoElement("directions", 7742, entities.String, enterpriseID, 65535)), enterpriseID)
	if err != nil {
		ilog.WithError(err).Errorf("Failed to register element")
		return err
	}
	return nil
}

func SendTemplateRecordv4(exporter *ipfixExporter.ExportingProcess, enrichEnterpriseID uint32) (uint16, []entities.InfoElementWithValue, error) {
	templateID := exporter.NewTemplateID()
	templateSet := entities.NewSet(false)
	err := templateSet.PrepareSet(entities.Template, templateID)
	if err != nil {
		ilog.WithError(err).Error("Failed in PrepareSet")
		return 0, nil, err
	}
	elements := make([]entities.InfoElementWithValue, 0)

	for _, field := range IPv4IANAFields {
		err = addElementToTemplate(field, nil, &elements, registry.IANAEnterpriseID)
		if err != nil {
			return 0, nil, err
		}
	}
	if enrichEnterpriseID != 0 {
		err = addKubeContextToTemplate(&elements, enrichEnterpriseID)
		if err != nil {
			return 0, nil, err
		}
		err = addNetworkEnrichmentToTemplate(&elements, enrichEnterpriseID)
		if err != nil {
			return 0, nil, err
		}
	}
	err = templateSet.AddRecord(elements, templateID)
	if err != nil {
		ilog.WithError(err).Error("Failed in Add Record")
		return 0, nil, err
	}
	_, err = exporter.SendSet(templateSet)
	if err != nil {
		ilog.WithError(err).Error("Failed to send template record")
		return 0, nil, err
	}

	return templateID, elements, nil
}

func SendTemplateRecordv6(exporter *ipfixExporter.ExportingProcess, enrichEnterpriseID uint32) (uint16, []entities.InfoElementWithValue, error) {
	templateID := exporter.NewTemplateID()
	templateSet := entities.NewSet(false)
	err := templateSet.PrepareSet(entities.Template, templateID)
	if err != nil {
		return 0, nil, err
	}
	elements := make([]entities.InfoElementWithValue, 0)

	for _, field := range IPv6IANAFields {
		err = addElementToTemplate(field, nil, &elements, registry.IANAEnterpriseID)
		if err != nil {
			return 0, nil, err
		}
	}
	if enrichEnterpriseID != 0 {
		err = addKubeContextToTemplate(&elements, enrichEnterpriseID)
		if err != nil {
			return 0, nil, err
		}
		err = addNetworkEnrichmentToTemplate(&elements, enrichEnterpriseID)
		if err != nil {
			return 0, nil, err
		}
	}

	err = templateSet.AddRecord(elements, templateID)
	if err != nil {
		return 0, nil, err
	}
	_, err = exporter.SendSet(templateSet)
	if err != nil {
		return 0, nil, err
	}

	return templateID, elements, nil
}

//nolint:cyclop
func setElementValue(record config.GenericMap, ieValPtr *entities.InfoElementWithValue) error {
	ieVal := *ieValPtr
	name := ieVal.GetName()
	mapping, ok := MapIPFIXKeys[name]
	if !ok {
		return nil
	}
	if value := record[mapping.Key]; value != nil {
		mapping.Setter(ieVal, value)
	} else if !mapping.Optional {
		return fmt.Errorf("unable to find %s (%s) in record", name, mapping.Key)
	}
	return nil
}

func setEntities(record config.GenericMap, elements *[]entities.InfoElementWithValue) error {
	for _, ieVal := range *elements {
		err := setElementValue(record, &ieVal)
		if err != nil {
			return err
		}
	}
	return nil
}
func (t *writeIpfix) sendDataRecord(record config.GenericMap, v6 bool) error {
	dataSet := entities.NewSet(false)
	var templateID uint16
	if v6 {
		templateID = t.templateIDv6
		err := setEntities(record, &t.entitiesV6)
		if err != nil {
			return err
		}
	} else {
		templateID = t.templateIDv4
		err := setEntities(record, &t.entitiesV4)
		if err != nil {
			return err
		}
	}
	err := dataSet.PrepareSet(entities.Data, templateID)
	if err != nil {
		return err
	}
	if v6 {
		err = dataSet.AddRecord(t.entitiesV6, templateID)
		if err != nil {
			return err
		}
	} else {
		err = dataSet.AddRecord(t.entitiesV4, templateID)
		if err != nil {
			return err
		}
	}
	_, err = t.exporter.SendSet(dataSet)
	if err != nil {
		return err
	}
	return nil
}

// Write writes a flow before being stored
func (t *writeIpfix) Write(entry config.GenericMap) {
	ilog.Tracef("entering writeIpfix Write")
	if IPv6Type == entry["Etype"].(uint16) {
		err := t.sendDataRecord(entry, true)
		if err != nil {
			ilog.WithError(err).Error("Failed in send v6 IPFIX record")
		}
	} else {
		err := t.sendDataRecord(entry, false)
		if err != nil {
			ilog.WithError(err).Error("Failed in send v4 IPFIX record")
		}
	}
}

// NewWriteIpfix creates a new write
func NewWriteIpfix(params config.StageParam) (Writer, error) {
	ilog.Debugf("entering NewWriteIpfix")

	ipfixConfigIn := api.WriteIpfix{}
	if params.Write != nil && params.Write.Ipfix != nil {
		ipfixConfigIn = *params.Write.Ipfix
	}
	// need to combine defaults with parameters that are provided in the config yaml file
	ipfixConfigIn.SetDefaults()

	if err := ipfixConfigIn.Validate(); err != nil {
		return nil, fmt.Errorf("the provided config is not valid: %w", err)
	}
	writeIpfix := &writeIpfix{}
	if params.Write != nil && params.Write.Ipfix != nil {
		writeIpfix.transport = params.Write.Ipfix.Transport
		writeIpfix.hostPort = fmt.Sprintf("%s:%d", params.Write.Ipfix.TargetHost, params.Write.Ipfix.TargetPort)
		writeIpfix.enrichEnterpriseID = uint32(params.Write.Ipfix.EnterpriseID)
	}
	// Initialize IPFIX registry and send templates
	registry.LoadRegistry()
	var err error
	if params.Write != nil && params.Write.Ipfix != nil && params.Write.Ipfix.EnterpriseID != 0 {
		err = loadCustomRegistry(writeIpfix.enrichEnterpriseID)
		if err != nil {
			ilog.Fatalf("Failed to load Custom(%d) Registry", writeIpfix.enrichEnterpriseID)
		}
	}

	// Create exporter using local server info
	input := ipfixExporter.ExporterInput{
		CollectorAddress:    writeIpfix.hostPort,
		CollectorProtocol:   writeIpfix.transport,
		ObservationDomainID: 1,
		TempRefTimeout:      1,
	}
	writeIpfix.exporter, err = ipfixExporter.InitExportingProcess(input)
	if err != nil {
		ilog.Fatalf("Got error when connecting to server %s: %v", writeIpfix.hostPort, err)
		return nil, err
	}
	ilog.Infof("Created exporter connecting to server with address: %s", writeIpfix.hostPort)

	writeIpfix.templateIDv4, writeIpfix.entitiesV4, err = SendTemplateRecordv4(writeIpfix.exporter, writeIpfix.enrichEnterpriseID)
	if err != nil {
		ilog.WithError(err).Error("Failed in send IPFIX template v4 record")
		return nil, err
	}

	writeIpfix.templateIDv6, writeIpfix.entitiesV6, err = SendTemplateRecordv6(writeIpfix.exporter, writeIpfix.enrichEnterpriseID)
	if err != nil {
		ilog.WithError(err).Error("Failed in send IPFIX template v6 record")
		return nil, err
	}
	return writeIpfix, nil
}
