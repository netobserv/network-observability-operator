package exporter

import (
	"net"

	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/vmware/go-ipfix/pkg/entities"
	ipfixExporter "github.com/vmware/go-ipfix/pkg/exporter"
	"github.com/vmware/go-ipfix/pkg/registry"
)

var ilog = logrus.WithField("component", "exporter/IPFIXProto")

// TODO: encode also the equivalent of the Protobuf's AgentIP field in a format that is binary-
// compatible with OVN-K.

type IPFIX struct {
	hostIP       string
	hostPort     int
	exporter     *ipfixExporter.ExportingProcess
	templateIDv4 uint16
	templateIDv6 uint16
	entitiesV4   []entities.InfoElementWithValue
	entitiesV6   []entities.InfoElementWithValue
}

func addElementToTemplate(log *logrus.Entry, elementName string, value []byte, elements *[]entities.InfoElementWithValue) error {
	element, err := registry.GetInfoElement(elementName, registry.IANAEnterpriseID)
	if err != nil {
		log.WithError(err).Errorf("Did not find the element with name %s", elementName)
		return err
	}
	ie, err := entities.DecodeAndCreateInfoElementWithValue(element, value)
	if err != nil {
		log.WithError(err).Errorf("Failed to decode element %s", elementName)
		return err
	}
	*elements = append(*elements, ie)
	return nil
}

func AddRecordValuesToTemplate(log *logrus.Entry, elements *[]entities.InfoElementWithValue) error {
	err := addElementToTemplate(log, "octetDeltaCount", nil, elements)
	if err != nil {
		return err
	}
	err = addElementToTemplate(log, "tcpControlBits", nil, elements)
	if err != nil {
		return err
	}
	err = addElementToTemplate(log, "flowStartSeconds", nil, elements)
	if err != nil {
		return err
	}
	err = addElementToTemplate(log, "flowStartMilliseconds", nil, elements)
	if err != nil {
		return err
	}
	err = addElementToTemplate(log, "flowEndSeconds", nil, elements)
	if err != nil {
		return err
	}
	err = addElementToTemplate(log, "flowEndMilliseconds", nil, elements)
	if err != nil {
		return err
	}
	err = addElementToTemplate(log, "packetDeltaCount", nil, elements)
	if err != nil {
		return err
	}
	err = addElementToTemplate(log, "interfaceName", nil, elements)
	if err != nil {
		return err
	}
	return nil
}

func SendTemplateRecordv4(log *logrus.Entry, exporter *ipfixExporter.ExportingProcess) (uint16, []entities.InfoElementWithValue, error) {
	templateID := exporter.NewTemplateID()
	templateSet := entities.NewSet(false)
	err := templateSet.PrepareSet(entities.Template, templateID)
	if err != nil {
		return 0, nil, err
	}
	elements := make([]entities.InfoElementWithValue, 0)

	err = addElementToTemplate(log, "ethernetType", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "flowDirection", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "sourceMacAddress", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "destinationMacAddress", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "sourceIPv4Address", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "destinationIPv4Address", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "protocolIdentifier", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "sourceTransportPort", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "destinationTransportPort", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "icmpTypeIPv4", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "icmpCodeIPv4", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = AddRecordValuesToTemplate(log, &elements)
	if err != nil {
		return 0, nil, err
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

func SendTemplateRecordv6(log *logrus.Entry, exporter *ipfixExporter.ExportingProcess) (uint16, []entities.InfoElementWithValue, error) {
	templateID := exporter.NewTemplateID()
	templateSet := entities.NewSet(false)
	err := templateSet.PrepareSet(entities.Template, templateID)
	if err != nil {
		return 0, nil, err
	}
	elements := make([]entities.InfoElementWithValue, 0)

	err = addElementToTemplate(log, "ethernetType", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "flowDirection", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "sourceMacAddress", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "destinationMacAddress", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "sourceIPv6Address", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "destinationIPv6Address", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "nextHeaderIPv6", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "sourceTransportPort", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "destinationTransportPort", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "icmpTypeIPv6", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = addElementToTemplate(log, "icmpCodeIPv6", nil, &elements)
	if err != nil {
		return 0, nil, err
	}
	err = AddRecordValuesToTemplate(log, &elements)
	if err != nil {
		return 0, nil, err
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

// Sends out Template record to the IPFIX collector
func StartIPFIXExporter(hostIP string, hostPort int, transportProto string) (*IPFIX, error) {
	socket := utils.GetSocket(hostIP, hostPort)
	log := ilog.WithField("collector", socket)

	registry.LoadRegistry()
	// Create exporter using local server info
	input := ipfixExporter.ExporterInput{
		CollectorAddress:    socket,
		CollectorProtocol:   transportProto,
		ObservationDomainID: 1,
		TempRefTimeout:      1,
	}
	exporter, err := ipfixExporter.InitExportingProcess(input)
	if err != nil {
		log.Fatalf("Got error when connecting to local server %s: %v", socket, err)
		return nil, err
	}
	log.Infof("Created exporter connecting to local server with address: %s", socket)

	templateIDv4, entitiesV4, err := SendTemplateRecordv4(log, exporter)
	if err != nil {
		log.WithError(err).Error("Failed in send IPFIX template v4 record")
		return nil, err
	}

	templateIDv6, entitiesV6, err := SendTemplateRecordv6(log, exporter)
	if err != nil {
		log.WithError(err).Error("Failed in send IPFIX template v6 record")
		return nil, err
	}
	log.Infof("entities v4 %+v", entitiesV4)
	log.Infof("entities v6 %+v", entitiesV6)

	return &IPFIX{
		hostIP:       hostIP,
		hostPort:     hostPort,
		exporter:     exporter,
		templateIDv4: templateIDv4,
		templateIDv6: templateIDv6,
		entitiesV4:   entitiesV4,
		entitiesV6:   entitiesV6,
	}, nil
}

func setIPv4Address(ieValPtr *entities.InfoElementWithValue, ipAddress net.IP) {
	ieVal := *ieValPtr
	if ipAddress == nil {
		ieVal.SetIPAddressValue(net.ParseIP("0.0.0.0"))
	} else {
		ieVal.SetIPAddressValue(ipAddress)
	}
}
func setIERecordValue(record *model.Record, ieValPtr *entities.InfoElementWithValue) {
	ieVal := *ieValPtr
	switch ieVal.GetName() {
	case "octetDeltaCount":
		ieVal.SetUnsigned64Value(record.Metrics.Bytes)
	case "tcpControlBits":
		ieVal.SetUnsigned16Value(record.Metrics.Flags)
	case "flowStartSeconds":
		ieVal.SetUnsigned32Value(uint32(record.TimeFlowStart.Unix()))
	case "flowStartMilliseconds":
		ieVal.SetUnsigned64Value(uint64(record.TimeFlowStart.UnixMilli()))
	case "flowEndSeconds":
		ieVal.SetUnsigned32Value(uint32(record.TimeFlowEnd.Unix()))
	case "flowEndMilliseconds":
		ieVal.SetUnsigned64Value(uint64(record.TimeFlowEnd.UnixMilli()))
	case "packetDeltaCount":
		ieVal.SetUnsigned64Value(uint64(record.Metrics.Packets))
	case "interfaceName":
		ieVal.SetStringValue(record.Interfaces[0].Interface)
	}
}
func setIEValue(record *model.Record, ieValPtr *entities.InfoElementWithValue) {
	ieVal := *ieValPtr
	switch ieVal.GetName() {
	case "ethernetType":
		ieVal.SetUnsigned16Value(record.Metrics.EthProtocol)
	case "flowDirection":
		ieVal.SetUnsigned8Value(uint8(record.Interfaces[0].Direction))
	case "sourceMacAddress":
		ieVal.SetMacAddressValue(record.Metrics.SrcMac[:])
	case "destinationMacAddress":
		ieVal.SetMacAddressValue(record.Metrics.DstMac[:])
	case "sourceIPv4Address":
		setIPv4Address(ieValPtr, model.IP(record.ID.SrcIp).To4())
	case "destinationIPv4Address":
		setIPv4Address(ieValPtr, model.IP(record.ID.DstIp).To4())
	case "sourceIPv6Address":
		ieVal.SetIPAddressValue(record.ID.SrcIp[:])
	case "destinationIPv6Address":
		ieVal.SetIPAddressValue(record.ID.DstIp[:])
	case "protocolIdentifier", "nextHeaderIPv6":
		ieVal.SetUnsigned8Value(record.ID.TransportProtocol)
	case "sourceTransportPort":
		ieVal.SetUnsigned16Value(record.ID.SrcPort)
	case "destinationTransportPort":
		ieVal.SetUnsigned16Value(record.ID.DstPort)
	case "icmpTypeIPv4", "icmpTypeIPv6":
		ieVal.SetUnsigned8Value(record.ID.IcmpType)
	case "icmpCodeIPv4", "icmpCodeIPv6":
		ieVal.SetUnsigned8Value(record.ID.IcmpCode)
	}
}
func setEntities(record *model.Record, elements *[]entities.InfoElementWithValue) {
	for _, ieVal := range *elements {
		setIEValue(record, &ieVal)
		setIERecordValue(record, &ieVal)
	}
}
func (ipf *IPFIX) sendDataRecord(_ *logrus.Entry, record *model.Record, v6 bool) error {
	dataSet := entities.NewSet(false)
	var templateID uint16
	if v6 {
		templateID = ipf.templateIDv6
		setEntities(record, &ipf.entitiesV6)
	} else {
		templateID = ipf.templateIDv4
		setEntities(record, &ipf.entitiesV4)
	}
	err := dataSet.PrepareSet(entities.Data, templateID)
	if err != nil {
		return err
	}
	if v6 {
		err = dataSet.AddRecord(ipf.entitiesV6, templateID)
		if err != nil {
			return err
		}
	} else {
		err = dataSet.AddRecord(ipf.entitiesV4, templateID)
		if err != nil {
			return err
		}
	}
	_, err = ipf.exporter.SendSet(dataSet)
	if err != nil {
		return err
	}
	return nil
}

// ExportFlows accepts slices of *model.Record by its input channel, converts them
// to IPFIX Records, and submits them to the collector.
func (ipf *IPFIX) ExportFlows(input <-chan []*model.Record) {
	socket := utils.GetSocket(ipf.hostIP, ipf.hostPort)
	log := ilog.WithField("collector", socket)
	for inputRecords := range input {
		for _, record := range inputRecords {
			if record.Metrics.EthProtocol == model.IPv6Type {
				err := ipf.sendDataRecord(log, record, true)
				if err != nil {
					log.WithError(err).Error("Failed in send IPFIX data record")
				}
			} else {
				err := ipf.sendDataRecord(log, record, false)
				if err != nil {
					log.WithError(err).Error("Failed in send IPFIX data record")
				}
			}
		}
	}
	ipf.exporter.CloseConnToCollector()
}
