// Copyright 2020 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exporter

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/pion/dtls/v2"
	"k8s.io/klog/v2"

	"github.com/vmware/go-ipfix/pkg/entities"
)

const startTemplateID uint16 = 255
const defaultCheckConnInterval = 10 * time.Second
const defaultJSONBufferLen = 5000

type templateValue struct {
	elements      []*entities.InfoElement
	minDataRecLen uint16
}

//  1. Tested one exportingProcess process per exporter. Can support multiple collector scenario by
//     creating different instances of exporting process. Need to be tested
//  2. Only one observation point per observation domain is supported,
//     so observation point ID not defined.
//  3. Supports only TCP and UDP; one session at a time. SCTP is not supported.
//  4. UDP needs to send MTU size packets as per RFC7011. We are not honoring that,
//     and relying on IP fragmentation and assuming data loss in the network is minimal.
//     We will revisit this if there are any issues, and get PathMTU from the user
//     as part of exporter input.
type ExportingProcess struct {
	connToCollector net.Conn
	obsDomainID     uint32
	seqNumber       uint32
	templateID      uint16
	templatesMap    map[uint16]templateValue
	templateRefCh   chan struct{}
	templateMutex   sync.Mutex
	sendJSONRecord  bool
	jsonBufferLen   int
}

type ExporterTLSClientConfig struct {
	// ServerName is passed to the server for SNI and is used in the client to check server
	// certificates against. If ServerName is empty, the hostname used to contact the
	// server is used.
	ServerName string
	// CAData holds PEM-encoded bytes for trusted root certificates for server.
	CAData []byte
	// CertData holds PEM-encoded bytes.
	CertData []byte
	// KeyData holds PEM-encoded bytes.
	KeyData []byte
}

type ExporterInput struct {
	// CollectorAddress needs to be provided in hostIP:port format.
	CollectorAddress string
	// CollectorProtocol needs to be provided in lower case format.
	// We support "tcp" and "udp" protocols.
	CollectorProtocol   string
	ObservationDomainID uint32
	TempRefTimeout      uint32
	// TLSClientConfig is set to use an encrypted connection to the collector.
	TLSClientConfig   *ExporterTLSClientConfig
	IsIPv6            bool
	SendJSONRecord    bool
	JSONBufferLen     int
	CheckConnInterval time.Duration
}

// InitExportingProcess takes in collector address(net.Addr format), obsID(observation ID)
// and tempRefTimeout(template refresh timeout). tempRefTimeout is applicable only
// for collectors listening over UDP; unit is seconds. For TCP, you can pass any
// value. For UDP, if 0 is passed, consider 1800s as default.
//
// PathMTU is recommended for UDP transport. If not given a valid value, i.e., either
// 0 or a value more than 1500, we consider a default value of 512B as per RFC7011.
// PathMTU is optional for TCP as we use max socket buffer size of 65535. It can
// be provided as 0.
// JSONBufferLen is recommended for sending json record. If not given a valid value,
// we consider a default 5000B.
func InitExportingProcess(input ExporterInput) (*ExportingProcess, error) {
	var conn net.Conn
	var err error
	if input.TLSClientConfig != nil {
		tlsConfig := input.TLSClientConfig
		if input.CollectorProtocol == "tcp" { // use TLS
			config, configErr := createClientConfig(tlsConfig)
			if configErr != nil {
				return nil, configErr
			}
			conn, err = tls.Dial(input.CollectorProtocol, input.CollectorAddress, config)
			if err != nil {
				klog.Errorf("Cannot the create the tls connection to the Collector %s: %v", input.CollectorAddress, err)
				return nil, err
			}
		} else if input.CollectorProtocol == "udp" { // use DTLS
			// TODO: support client authentication
			if len(tlsConfig.CertData) > 0 || len(tlsConfig.KeyData) > 0 {
				klog.Error("Client-authentication is not supported yet for DTLS, cert and key data will be ignored")
			}
			roots := x509.NewCertPool()
			ok := roots.AppendCertsFromPEM(tlsConfig.CAData)
			if !ok {
				return nil, fmt.Errorf("failed to parse root certificate")
			}
			config := &dtls.Config{
				RootCAs:              roots,
				ExtendedMasterSecret: dtls.RequireExtendedMasterSecret,
				ServerName:           tlsConfig.ServerName,
			}
			udpAddr, err := net.ResolveUDPAddr(input.CollectorProtocol, input.CollectorAddress)
			if err != nil {
				return nil, err
			}
			conn, err = dtls.Dial(udpAddr.Network(), udpAddr, config)
			if err != nil {
				klog.Errorf("Cannot the create the dtls connection to the Collector %s: %v", udpAddr.String(), err)
				return nil, err
			}
		}
	} else {
		conn, err = net.Dial(input.CollectorProtocol, input.CollectorAddress)
		if err != nil {
			klog.Errorf("Cannot the create the connection to the Collector %s: %v", input.CollectorAddress, err)
			return nil, err
		}
	}
	expProc := &ExportingProcess{
		connToCollector: conn,
		obsDomainID:     input.ObservationDomainID,
		seqNumber:       0,
		templateID:      startTemplateID,
		templatesMap:    make(map[uint16]templateValue),
		templateRefCh:   make(chan struct{}),
		sendJSONRecord:  input.SendJSONRecord,
	}

	// Start a goroutine for checking whether connection to collector is still open
	if input.CollectorProtocol == "tcp" {
		interval := input.CheckConnInterval
		if interval == 0 {
			interval = defaultCheckConnInterval
		}
		go func() {
			ticker := time.NewTicker(interval)
			oneByteForRead := make([]byte, 1)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					isConnected := expProc.checkConnToCollector(oneByteForRead)
					if !isConnected {
						expProc.CloseConnToCollector()
						klog.Error("Error when connecting to collector because connection is closed.")
						return
					}
				}
			}
		}()
	}

	// Template refresh logic is only for UDP transport.
	if input.CollectorProtocol == "udp" {
		if input.TempRefTimeout == 0 {
			// Default value
			input.TempRefTimeout = entities.TemplateRefreshTimeOut
		}
		go func() {
			ticker := time.NewTicker(time.Duration(input.TempRefTimeout) * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-expProc.templateRefCh:
					break
				case <-ticker.C:
					err := expProc.sendRefreshedTemplates()
					if err != nil {
						// Other option is sending messages through channel to library consumers
						klog.Errorf("Error when sending refreshed templates: %v. Closing the connection to IPFIX controller", err)
						expProc.CloseConnToCollector()
					}
				}
			}
		}()
	}
	if expProc.sendJSONRecord {
		if input.JSONBufferLen <= 0 {
			expProc.jsonBufferLen = defaultJSONBufferLen
		} else {
			expProc.jsonBufferLen = input.JSONBufferLen
		}
	}
	return expProc, nil
}

func (ep *ExportingProcess) SendSet(set entities.Set) (int, error) {
	// Iterate over all records in the set.
	setType := set.GetSetType()
	if setType == entities.Undefined {
		return 0, fmt.Errorf("set type is not properly defined")
	}
	for _, record := range set.GetRecords() {
		if setType == entities.Template {
			ep.updateTemplate(record.GetTemplateID(), record.GetOrderedElementList(), record.GetMinDataRecordLen())
		} else if setType == entities.Data {
			err := ep.dataRecSanityCheck(record)
			if err != nil {
				return 0, fmt.Errorf("error when doing sanity check:%v", err)
			}
		}
	}
	// Update the length in set header before sending the message.
	set.UpdateLenInHeader()

	var bytesSent int
	var err error
	if !ep.sendJSONRecord {
		bytesSent, err = ep.createAndSendIPFIXMsg(set)
	} else {
		if setType == entities.Data {
			bytesSent, err = ep.createAndSendJSONMsg(set)
		}
	}
	if err != nil {
		return bytesSent, err
	}
	return bytesSent, nil
}

func (ep *ExportingProcess) GetMsgSizeLimit() int {
	return entities.MaxSocketMsgSize
}

func (ep *ExportingProcess) CloseConnToCollector() {
	if !isChanClosed(ep.templateRefCh) {
		close(ep.templateRefCh) // Close template refresh channel
	}
	err := ep.connToCollector.Close()
	// Just log the error that happened when closing the connection. Not returning error as we do not expect library
	// consumers to exit their programs with this error.
	if err != nil {
		klog.Errorf("Error when closing connection to collector: %v", err)
	}
}

// checkConnToCollector checks whether the connection from exporter is still open
// by trying to read from connection. Closed connection will return EOF from read.
func (ep *ExportingProcess) checkConnToCollector(oneByteForRead []byte) bool {
	ep.connToCollector.SetReadDeadline(time.Now().Add(time.Millisecond))
	if _, err := ep.connToCollector.Read(oneByteForRead); err == io.EOF {
		return false
	}
	return true
}

// NewTemplateID is called to get ID when creating new template record.
func (ep *ExportingProcess) NewTemplateID() uint16 {
	ep.templateID++
	return ep.templateID
}

// createAndSendIPFIXMsg takes in a set as input, creates the IPFIX message, and sends it out.
// TODO: This method will change when we support sending multiple sets.
func (ep *ExportingProcess) createAndSendIPFIXMsg(set entities.Set) (int, error) {
	if set.GetSetType() == entities.Data {
		ep.seqNumber = ep.seqNumber + set.GetNumberOfRecords()
	}
	bytesSlice, err := CreateIPFIXMsg(set, ep.obsDomainID, ep.seqNumber, time.Now())
	if err != nil {
		return 0, err
	}

	// Send the message on the exporter connection.
	bytesSent, err := ep.connToCollector.Write(bytesSlice)

	if err != nil {
		return bytesSent, fmt.Errorf("error when sending message on the connection: %v", err)
	} else if bytesSent != len(bytesSlice) {
		return bytesSent, fmt.Errorf("could not send the complete message on the connection")
	}

	return bytesSent, nil
}

// createAndSendJSONMsg takes in a set as input, creates the JSON record, and sends it out.
func (ep *ExportingProcess) createAndSendJSONMsg(set entities.Set) (int, error) {
	var bytesSent int
	for _, record := range set.GetRecords() {
		elements := make(map[string]interface{})
		orderedElements := record.GetOrderedElementList()
		for _, element := range orderedElements {
			switch element.GetDataType() {
			case entities.Unsigned8:
				elements[element.GetName()] = element.GetUnsigned8Value()
			case entities.Unsigned16:
				elements[element.GetName()] = element.GetUnsigned16Value()
			case entities.Unsigned32:
				elements[element.GetName()] = element.GetUnsigned32Value()
			case entities.Unsigned64:
				elements[element.GetName()] = element.GetUnsigned64Value()
			case entities.Signed8:
				elements[element.GetName()] = element.GetSigned8Value()
			case entities.Signed16:
				elements[element.GetName()] = element.GetSigned16Value()
			case entities.Signed32:
				elements[element.GetName()] = element.GetSigned32Value()
			case entities.Signed64:
				elements[element.GetName()] = element.GetSigned64Value()
			case entities.Float32:
				elements[element.GetName()] = element.GetFloat32Value()
			case entities.Float64:
				elements[element.GetName()] = element.GetFloat64Value()
			case entities.Boolean:
				elements[element.GetName()] = element.GetBooleanValue()
			case entities.DateTimeSeconds:
				elements[element.GetName()] = element.GetUnsigned32Value()
			case entities.DateTimeMilliseconds:
				elements[element.GetName()] = element.GetUnsigned64Value()
			case entities.DateTimeMicroseconds, entities.DateTimeNanoseconds:
				return bytesSent, fmt.Errorf("API does not support micro and nano seconds types yet")
			case entities.MacAddress:
				elements[element.GetName()] = element.GetMacAddressValue()
			case entities.Ipv4Address, entities.Ipv6Address:
				elements[element.GetName()] = element.GetIPAddressValue()
			case entities.String:
				elements[element.GetName()] = element.GetStringValue()
			default:
				return bytesSent, fmt.Errorf("API supports only valid information elements with datatypes given in RFC7011")
			}
		}
		message := make(map[string]interface{}, 2)
		message["ipfix"] = elements
		message["@timestamp"] = time.Now().Format(time.RFC3339)
		writer := bytes.NewBuffer(make([]byte, 0, ep.jsonBufferLen))
		encoder := json.NewEncoder(writer)
		err := encoder.Encode(message)
		if err != nil {
			return bytesSent, fmt.Errorf("error when encoding message to JSON: %v", err)
		}
		// Send the message on the exporter connection.
		bytes, err := ep.connToCollector.Write(writer.Bytes())
		if err != nil {
			return bytes, fmt.Errorf("error when sending message on the connection: %v", err)
		}
		bytesSent += bytes
	}
	return bytesSent, nil
}

func (ep *ExportingProcess) updateTemplate(id uint16, elements []entities.InfoElementWithValue, minDataRecLen uint16) {
	ep.templateMutex.Lock()
	defer ep.templateMutex.Unlock()

	if _, exist := ep.templatesMap[id]; exist {
		return
	}
	ep.templatesMap[id] = templateValue{
		make([]*entities.InfoElement, len(elements)),
		minDataRecLen,
	}
	for i, elem := range elements {
		ep.templatesMap[id].elements[i] = elem.GetInfoElement()
	}
	return
}

//nolint:unused // Keeping this function for reference.
func (ep *ExportingProcess) deleteTemplate(id uint16) error {
	ep.templateMutex.Lock()
	defer ep.templateMutex.Unlock()

	if _, exist := ep.templatesMap[id]; !exist {
		return fmt.Errorf("template %d does not exist in exporting process", id)
	}
	delete(ep.templatesMap, id)
	return nil
}

func (ep *ExportingProcess) sendRefreshedTemplates() error {
	// Send refreshed template for every template in template map
	templateSets := make([]entities.Set, 0)

	ep.templateMutex.Lock()
	for templateID, tempValue := range ep.templatesMap {
		tempSet := entities.NewSet(false)
		if err := tempSet.PrepareSet(entities.Template, templateID); err != nil {
			return err
		}
		elements := make([]entities.InfoElementWithValue, len(tempValue.elements))
		var err error
		for i, element := range tempValue.elements {
			if elements[i], err = entities.DecodeAndCreateInfoElementWithValue(element, nil); err != nil {
				return err
			}
		}
		err = tempSet.AddRecord(elements, templateID)
		if err != nil {
			return err
		}
		templateSets = append(templateSets, tempSet)
	}
	ep.templateMutex.Unlock()

	for _, templateSet := range templateSets {
		if _, err := ep.SendSet(templateSet); err != nil {
			return err
		}
	}
	return nil
}

func (ep *ExportingProcess) dataRecSanityCheck(rec entities.Record) error {
	templateID := rec.GetTemplateID()

	ep.templateMutex.Lock()
	defer ep.templateMutex.Unlock()

	if _, exist := ep.templatesMap[templateID]; !exist {
		return fmt.Errorf("process: templateID %d does not exist in exporting process", templateID)
	}
	if rec.GetFieldCount() != uint16(len(ep.templatesMap[templateID].elements)) {
		return fmt.Errorf("process: field count of data does not match templateID %d", templateID)
	}
	if len(rec.GetBuffer()) < int(ep.templatesMap[templateID].minDataRecLen) {
		return fmt.Errorf("process: Data Record does not pass the min required length (%d) check for template ID %d", ep.templatesMap[templateID].minDataRecLen, templateID)
	}
	return nil
}

func isChanClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

func createClientConfig(config *ExporterTLSClientConfig) (*tls.Config, error) {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(config.CAData)
	if !ok {
		return nil, fmt.Errorf("failed to parse root certificate")
	}
	if config.CertData == nil {
		return &tls.Config{
			RootCAs:    roots,
			MinVersion: tls.VersionTLS12,
			ServerName: config.ServerName,
		}, nil
	}
	cert, err := tls.X509KeyPair(config.CertData, config.KeyData)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      roots,
		MinVersion:   tls.VersionTLS12,
		ServerName:   config.ServerName,
	}, nil
}
