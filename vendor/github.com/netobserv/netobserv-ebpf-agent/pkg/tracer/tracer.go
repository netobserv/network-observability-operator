package tracer

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/netobserv/netobserv-ebpf-agent/pkg/ebpf"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/ifaces"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/kernel"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/metrics"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"
	"github.com/prometheus/client_golang/prometheus"

	cilium "github.com/cilium/ebpf"
	"github.com/cilium/ebpf/btf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"github.com/gavv/monotime"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

const (
	qdiscType = "clsact"
	// ebpf map names as defined in bpf/maps_definition.h
	aggregatedFlowsMap    = "aggregated_flows"
	additionalFlowMetrics = "additional_flow_metrics"
	dnsLatencyMap         = "dns_flows"
	flowFilterMap         = "filter_map"
	flowPeerFilterMap     = "peer_filter_map"
	// constants defined in flows.c as "volatile const"
	constSampling                       = "sampling"
	constHasFilterSampling              = "has_filter_sampling"
	constTraceMessages                  = "trace_messages"
	constEnableRtt                      = "enable_rtt"
	constEnableDNSTracking              = "enable_dns_tracking"
	constDNSTrackingPort                = "dns_port"
	dnsDefaultPort                      = 53
	constEnableFlowFiltering            = "enable_flows_filtering"
	constEnableNetworkEventsMonitoring  = "enable_network_events_monitoring"
	constNetworkEventsMonitoringGroupID = "network_events_monitoring_groupid"
	constEnablePktTranslation           = "enable_pkt_translation_tracking"
	pktDropHook                         = "kfree_skb"
	constPcaEnable                      = "enable_pca"
	pcaRecordsMap                       = "packet_record"
	tcEgressFilterName                  = "tc/tc_egress_flow_parse"
	tcIngressFilterName                 = "tc/tc_ingress_flow_parse"
	tcpFentryHook                       = "tcp_rcv_fentry"
	tcpRcvKprobe                        = "tcp_rcv_kprobe"
	rhNetworkEventsMonitoringHook       = "rh_psample_sample_packet"
	networkEventsMonitoringHook         = "psample_sample_packet"
	defaultNetworkEventsGroupID         = 10
)

var log = logrus.WithField("component", "ebpf.FlowFetcher")
var plog = logrus.WithField("component", "ebpf.PacketFetcher")

// FlowFetcher reads and forwards the Flows from the Traffic Control hooks in the eBPF kernel space.
// It provides access both to flows that are aggregated in the kernel space (via PerfCPU hashmap)
// and to flows that are forwarded by the kernel via ringbuffer because could not be aggregated
// in the map
type FlowFetcher struct {
	objects                     *ebpf.BpfObjects
	qdiscs                      map[ifaces.Interface]*netlink.GenericQdisc
	egressFilters               map[ifaces.Interface]*netlink.BpfFilter
	ingressFilters              map[ifaces.Interface]*netlink.BpfFilter
	ringbufReader               *ringbuf.Reader
	cacheMaxSize                int
	enableIngress               bool
	enableEgress                bool
	pktDropsTracePoint          link.Link
	rttFentryLink               link.Link
	rttKprobeLink               link.Link
	egressTCXLink               map[ifaces.Interface]link.Link
	ingressTCXLink              map[ifaces.Interface]link.Link
	networkEventsMonitoringLink link.Link
	nfNatManIPLink              link.Link
	lookupAndDeleteSupported    bool
	useEbpfManager              bool
	pinDir                      string
}

type FlowFetcherConfig struct {
	EnableIngress                  bool
	EnableEgress                   bool
	Debug                          bool
	Sampling                       int
	CacheMaxSize                   int
	EnablePktDrops                 bool
	EnableDNSTracker               bool
	DNSTrackerPort                 uint16
	EnableRTT                      bool
	EnableNetworkEventsMonitoring  bool
	NetworkEventsMonitoringGroupID int
	EnableFlowFilter               bool
	EnablePCA                      bool
	EnablePktTranslation           bool
	UseEbpfManager                 bool
	BpfManBpfFSPath                string
	FilterConfig                   []*FilterConfig
}

// nolint:golint,cyclop
func NewFlowFetcher(cfg *FlowFetcherConfig) (*FlowFetcher, error) {
	var pktDropsLink, networkEventsMonitoringLink, rttFentryLink, rttKprobeLink link.Link
	var nfNatManIPLink link.Link
	var err error
	objects := ebpf.BpfObjects{}
	var pinDir string
	var filter *Filter
	if cfg.EnableFlowFilter {
		filter = NewFilter(cfg.FilterConfig)
	}

	if !cfg.UseEbpfManager {
		if err := rlimit.RemoveMemlock(); err != nil {
			log.WithError(err).
				Warn("can't remove mem lock. The agent will not be able to start eBPF programs")
		}
		spec, err := ebpf.LoadBpf()
		if err != nil {
			return nil, fmt.Errorf("loading BPF data: %w", err)
		}

		// Resize maps according to user-provided configuration
		spec.Maps[aggregatedFlowsMap].MaxEntries = uint32(cfg.CacheMaxSize)
		spec.Maps[additionalFlowMetrics].MaxEntries = uint32(cfg.CacheMaxSize)

		// remove pinning from all maps
		maps2Name := []string{"aggregated_flows", "additional_flow_metrics", "direct_flows", "dns_flows", "filter_map", "peer_filter_map", "global_counters", "packet_record"}
		for _, m := range maps2Name {
			spec.Maps[m].Pinning = 0
		}

		traceMsgs := 0
		if cfg.Debug {
			traceMsgs = 1
		}

		enableRtt := 0
		if cfg.EnableRTT {
			enableRtt = 1
		}

		enableDNSTracking := 0
		dnsTrackerPort := uint16(dnsDefaultPort)
		if cfg.EnableDNSTracker {
			enableDNSTracking = 1
			if cfg.DNSTrackerPort != 0 {
				dnsTrackerPort = cfg.DNSTrackerPort
			}
		}

		if enableDNSTracking == 0 {
			spec.Maps[dnsLatencyMap].MaxEntries = 1
		}

		enableFlowFiltering := 0
		hasFilterSampling := uint8(0)
		if filter != nil {
			enableFlowFiltering = 1
			hasFilterSampling = filter.hasSampling()
		} else {
			spec.Maps[flowFilterMap].MaxEntries = 1
			spec.Maps[flowPeerFilterMap].MaxEntries = 1
		}
		enableNetworkEventsMonitoring := 0
		if cfg.EnableNetworkEventsMonitoring {
			enableNetworkEventsMonitoring = 1
		}
		networkEventsMonitoringGroupID := defaultNetworkEventsGroupID
		if cfg.NetworkEventsMonitoringGroupID > 0 {
			networkEventsMonitoringGroupID = cfg.NetworkEventsMonitoringGroupID
		}
		enablePktTranslation := 0
		if cfg.EnablePktTranslation {
			enablePktTranslation = 1
		}
		if err := spec.RewriteConstants(map[string]interface{}{
			// When adding constants here, remember to delete them in NewPacketFetcher
			constSampling:                       uint32(cfg.Sampling),
			constHasFilterSampling:              hasFilterSampling,
			constTraceMessages:                  uint8(traceMsgs),
			constEnableRtt:                      uint8(enableRtt),
			constEnableDNSTracking:              uint8(enableDNSTracking),
			constDNSTrackingPort:                dnsTrackerPort,
			constEnableFlowFiltering:            uint8(enableFlowFiltering),
			constEnableNetworkEventsMonitoring:  uint8(enableNetworkEventsMonitoring),
			constNetworkEventsMonitoringGroupID: uint8(networkEventsMonitoringGroupID),
			constEnablePktTranslation:           uint8(enablePktTranslation),
		}); err != nil {
			return nil, fmt.Errorf("rewriting BPF constants definition: %w", err)
		}

		oldKernel := kernel.IsKernelOlderThan("5.14.0")
		if oldKernel {
			log.Infof("kernel older than 5.14.0 detected: not all hooks are supported")
		}
		rtOldKernel := kernel.IsRealTimeKernel() && kernel.IsKernelOlderThan("5.14.0-292")
		if rtOldKernel {
			log.Infof("kernel is realtime and older than 5.14.0-292 not all hooks are supported")
		}
		supportNetworkEvents := !kernel.IsKernelOlderThan("5.14.0-427")
		objects, err = kernelSpecificLoadAndAssign(oldKernel, rtOldKernel, supportNetworkEvents, spec, pinDir)
		if err != nil {
			return nil, err
		}

		log.Debugf("Deleting specs for PCA")
		// Deleting specs for PCA
		// Always set pcaRecordsMap to the minimum in FlowFetcher - PCA and Flow Fetcher are mutually exclusive.
		spec.Maps[pcaRecordsMap].MaxEntries = 1

		objects.TcxEgressPcaParse = nil
		objects.TcIngressPcaParse = nil
		delete(spec.Programs, constPcaEnable)

		if cfg.EnablePktDrops && !oldKernel && !rtOldKernel {
			pktDropsLink, err = link.Tracepoint("skb", pktDropHook, objects.KfreeSkb, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to attach the BPF program to kfree_skb tracepoint: %w", err)
			}
		}

		if cfg.EnableNetworkEventsMonitoring {
			if supportNetworkEvents {
				// Enable the following logic with RHEL9.6 when its available
				if !kernel.IsKernelOlderThan("5.16.0") {
					//revive:disable
					/*
						networkEventsMonitoringLink, err = link.Kprobe(networkEventsMonitoringHook, objects.NetworkEventsMonitoring, nil)
						if err != nil {
							return nil, fmt.Errorf("failed to attach the BPF program network events monitoring kprobe: %w", err)
						}
					*/
				} else {
					log.Infof("kernel older than 5.16.0 detected: use custom network_events_monitoring hook")
					networkEventsMonitoringLink, err = link.Kprobe(rhNetworkEventsMonitoringHook, objects.RhNetworkEventsMonitoring, nil)
					if err != nil {
						return nil, fmt.Errorf("failed to attach the BPF program network events monitoring kprobe: %w", err)
					}
				}
			} else {
				log.Infof("kernel older than 5.14.0-427 detected: it does not support network_events_monitoring hook, skip")
			}
		}

		if cfg.EnableRTT {
			if !oldKernel {
				rttFentryLink, err = link.AttachTracing(link.TracingOptions{
					Program: objects.BpfPrograms.TcpRcvFentry,
				})
				if err == nil {
					goto next
				}
				if err != nil {
					log.Warningf("failed to attach the BPF program to tcpReceiveFentry: %v fallback to use kprobe", err)
					// Fall through to use kprobe
				}
			}
			// try to use kprobe for older kernels
			if !rtOldKernel {
				rttKprobeLink, err = link.Kprobe("tcp_rcv_established", objects.TcpRcvKprobe, nil)
				if err != nil {
					log.Warningf("failed to attach the BPF program to kprobe: %v", err)
					return nil, fmt.Errorf("failed to attach the BPF program to tcpReceiveKprobe: %w", err)
				}
			}
		}
	next:
		if cfg.EnablePktTranslation {
			nfNatManIPLink, err = link.Kprobe("nf_nat_manip_pkt", objects.TrackNatManipPkt, nil)
			if err != nil {
				log.Warningf("failed to attach the BPF program to nat_manip kprobe: %v", err)
				return nil, fmt.Errorf("failed to attach the BPF program to nat_manip kprobe: %w", err)
			}
		}
	} else {
		pinDir = cfg.BpfManBpfFSPath
		opts := &cilium.LoadPinOptions{
			ReadOnly:  false,
			WriteOnly: false,
			Flags:     0,
		}

		log.Info("BPFManager mode: loading aggregated flows pinned maps")
		mPath := path.Join(pinDir, "aggregated_flows")
		objects.BpfMaps.AggregatedFlows, err = cilium.LoadPinnedMap(mPath, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", mPath, err)
		}
		log.Info("BPFManager mode: loading additional flow metrics pinned maps")
		mPath = path.Join(pinDir, "additional_flow_metrics")
		objects.BpfMaps.AdditionalFlowMetrics, err = cilium.LoadPinnedMap(mPath, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", mPath, err)
		}
		log.Info("BPFManager mode: loading direct flows pinned maps")
		mPath = path.Join(pinDir, "direct_flows")
		objects.BpfMaps.DirectFlows, err = cilium.LoadPinnedMap(mPath, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", mPath, err)
		}
		log.Infof("BPFManager mode: loading DNS flows pinned maps")
		mPath = path.Join(pinDir, "dns_flows")
		objects.BpfMaps.DnsFlows, err = cilium.LoadPinnedMap(mPath, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", mPath, err)
		}
		log.Infof("BPFManager mode: loading filter pinned maps")
		mPath = path.Join(pinDir, "filter_map")
		objects.BpfMaps.FilterMap, err = cilium.LoadPinnedMap(mPath, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", mPath, err)
		}
		objects.BpfMaps.PeerFilterMap, err = cilium.LoadPinnedMap(mPath, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", mPath, err)
		}
		log.Infof("BPFManager mode: loading global counters pinned maps")
		mPath = path.Join(pinDir, "global_counters")
		objects.BpfMaps.GlobalCounters, err = cilium.LoadPinnedMap(mPath, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", mPath, err)
		}
		log.Infof("BPFManager mode: loading packet record pinned maps")
		mPath = path.Join(pinDir, "packet_record")
		objects.BpfMaps.PacketRecord, err = cilium.LoadPinnedMap(mPath, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", mPath, err)
		}
	}

	if filter != nil {
		if err := filter.ProgramFilter(&objects); err != nil {
			return nil, fmt.Errorf("programming flow filter: %w", err)
		}
	}

	flows, err := ringbuf.NewReader(objects.BpfMaps.DirectFlows)
	if err != nil {
		return nil, fmt.Errorf("accessing to ringbuffer: %w", err)
	}

	return &FlowFetcher{
		objects:                     &objects,
		ringbufReader:               flows,
		egressFilters:               map[ifaces.Interface]*netlink.BpfFilter{},
		ingressFilters:              map[ifaces.Interface]*netlink.BpfFilter{},
		qdiscs:                      map[ifaces.Interface]*netlink.GenericQdisc{},
		cacheMaxSize:                cfg.CacheMaxSize,
		enableIngress:               cfg.EnableIngress,
		enableEgress:                cfg.EnableEgress,
		pktDropsTracePoint:          pktDropsLink,
		rttFentryLink:               rttFentryLink,
		rttKprobeLink:               rttKprobeLink,
		nfNatManIPLink:              nfNatManIPLink,
		egressTCXLink:               map[ifaces.Interface]link.Link{},
		ingressTCXLink:              map[ifaces.Interface]link.Link{},
		networkEventsMonitoringLink: networkEventsMonitoringLink,
		lookupAndDeleteSupported:    true, // this will be turned off later if found to be not supported
		useEbpfManager:              cfg.UseEbpfManager,
		pinDir:                      pinDir,
	}, nil
}

func (m *FlowFetcher) AttachTCX(iface ifaces.Interface) error {
	ilog := log.WithField("iface", iface)
	if iface.NetNS != netns.None() {
		originalNs, err := netns.Get()
		if err != nil {
			return fmt.Errorf("failed to get current netns: %w", err)
		}
		defer func() {
			if err := netns.Set(originalNs); err != nil {
				ilog.WithError(err).Error("failed to set netns back")
			}
			originalNs.Close()
		}()
		if err := unix.Setns(int(iface.NetNS), unix.CLONE_NEWNET); err != nil {
			return fmt.Errorf("failed to setns to %s: %w", iface.NetNS, err)
		}
	}

	if m.enableEgress {
		egrLink, err := link.AttachTCX(link.TCXOptions{
			Program:   m.objects.BpfPrograms.TcxEgressFlowParse,
			Attach:    cilium.AttachTCXEgress,
			Interface: iface.Index,
		})
		if err != nil {
			if errors.Is(err, fs.ErrExist) {
				// The interface already has a TCX egress hook
				log.WithField("iface", iface.Name).Debug("interface already has a TCX egress hook ignore")
			} else {
				return fmt.Errorf("failed to attach TCX egress: %w", err)
			}
		}
		m.egressTCXLink[iface] = egrLink
		ilog.WithField("interface", iface.Name).Debug("successfully attach egressTCX hook")
	}

	if m.enableIngress {
		ingLink, err := link.AttachTCX(link.TCXOptions{
			Program:   m.objects.BpfPrograms.TcxIngressFlowParse,
			Attach:    cilium.AttachTCXIngress,
			Interface: iface.Index,
		})
		if err != nil {
			if errors.Is(err, fs.ErrExist) {
				// The interface already has a TCX ingress hook
				log.WithField("iface", iface.Name).Debug("interface already has a TCX ingress hook ignore")
			} else {
				return fmt.Errorf("failed to attach TCX ingress: %w", err)
			}
		}
		m.ingressTCXLink[iface] = ingLink
		ilog.WithField("interface", iface.Name).Debug("successfully attach ingressTCX hook")
	}

	return nil
}

func (m *FlowFetcher) DetachTCX(iface ifaces.Interface) error {
	ilog := log.WithField("iface", iface)
	if iface.NetNS != netns.None() {
		originalNs, err := netns.Get()
		if err != nil {
			return fmt.Errorf("failed to get current netns: %w", err)
		}
		defer func() {
			if err := netns.Set(originalNs); err != nil {
				ilog.WithError(err).Error("failed to set netns back")
			}
			originalNs.Close()
		}()
		if err := unix.Setns(int(iface.NetNS), unix.CLONE_NEWNET); err != nil {
			return fmt.Errorf("failed to setns to %s: %w", iface.NetNS, err)
		}
	}
	if m.enableEgress {
		if l := m.egressTCXLink[iface]; l != nil {
			if err := l.Close(); err != nil {
				return fmt.Errorf("TCX: failed to close egress link: %w", err)
			}
			ilog.WithField("interface", iface.Name).Debug("successfully detach egressTCX hook")
		} else {
			return fmt.Errorf("egress link does not have a TCX egress hook")
		}
	}

	if m.enableIngress {
		if l := m.ingressTCXLink[iface]; l != nil {
			if err := l.Close(); err != nil {
				return fmt.Errorf("TCX: failed to close ingress link: %w", err)
			}
			ilog.WithField("interface", iface.Name).Debug("successfully detach ingressTCX hook")
		} else {
			return fmt.Errorf("ingress link does not have a TCX ingress hook")
		}
	}

	return nil
}

func removeTCFilters(ifName string, tcDir uint32) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}

	filters, err := netlink.FilterList(link, tcDir)
	if err != nil {
		return err
	}
	var errs []error
	for _, f := range filters {
		if err := netlink.FilterDel(f); err != nil {
			errs = append(errs, err)
		}
	}

	return kerrors.NewAggregate(errs)
}

func unregister(iface ifaces.Interface) error {
	ilog := log.WithField("iface", iface)
	ilog.Debugf("looking for previously installed TC filters on %s", iface.Name)
	links, err := netlink.LinkList()
	if err != nil {
		return fmt.Errorf("retrieving all netlink devices: %w", err)
	}

	egressDevs := []netlink.Link{}
	ingressDevs := []netlink.Link{}
	for _, l := range links {
		if l.Attrs().Name != iface.Name {
			continue
		}
		ingressFilters, err := netlink.FilterList(l, netlink.HANDLE_MIN_INGRESS)
		if err != nil {
			return fmt.Errorf("listing ingress filters: %w", err)
		}
		for _, filter := range ingressFilters {
			if bpfFilter, ok := filter.(*netlink.BpfFilter); ok {
				if strings.HasPrefix(bpfFilter.Name, tcIngressFilterName) {
					ingressDevs = append(ingressDevs, l)
				}
			}
		}

		egressFilters, err := netlink.FilterList(l, netlink.HANDLE_MIN_EGRESS)
		if err != nil {
			return fmt.Errorf("listing egress filters: %w", err)
		}
		for _, filter := range egressFilters {
			if bpfFilter, ok := filter.(*netlink.BpfFilter); ok {
				if strings.HasPrefix(bpfFilter.Name, tcEgressFilterName) {
					egressDevs = append(egressDevs, l)
				}
			}
		}
	}

	for _, dev := range ingressDevs {
		ilog.Debugf("removing ingress stale tc filters from %s", dev.Attrs().Name)
		err = removeTCFilters(dev.Attrs().Name, netlink.HANDLE_MIN_INGRESS)
		if err != nil {
			ilog.WithError(err).Errorf("couldn't remove ingress tc filters from %s", dev.Attrs().Name)
		}
	}

	for _, dev := range egressDevs {
		ilog.Debugf("removing egress stale tc filters from %s", dev.Attrs().Name)
		err = removeTCFilters(dev.Attrs().Name, netlink.HANDLE_MIN_EGRESS)
		if err != nil {
			ilog.WithError(err).Errorf("couldn't remove egress tc filters from %s", dev.Attrs().Name)
		}
	}

	return nil
}

func (m *FlowFetcher) UnRegister(iface ifaces.Interface) error {
	// qdiscs, ingress and egress filters are automatically deleted so we don't need to
	// specifically detach them from the ebpfFetcher
	return unregister(iface)
}

// Register and links the eBPF fetcher into the system. The program should invoke Unregister
// before exiting.
func (m *FlowFetcher) Register(iface ifaces.Interface) error {
	ilog := log.WithField("iface", iface)
	handle, err := netlink.NewHandleAt(iface.NetNS)
	if err != nil {
		return fmt.Errorf("failed to create handle for netns (%s): %w", iface.NetNS.String(), err)
	}
	defer handle.Close()

	// Load pre-compiled programs and maps into the kernel, and rewrites the configuration
	ipvlan, err := handle.LinkByIndex(iface.Index)
	if err != nil {
		return fmt.Errorf("failed to lookup ipvlan device %d (%s): %w", iface.Index, iface.Name, err)
	}
	qdiscAttrs := netlink.QdiscAttrs{
		LinkIndex: ipvlan.Attrs().Index,
		Handle:    netlink.MakeHandle(0xffff, 0),
		Parent:    netlink.HANDLE_CLSACT,
	}
	qdisc := &netlink.GenericQdisc{
		QdiscAttrs: qdiscAttrs,
		QdiscType:  qdiscType,
	}
	if err := handle.QdiscDel(qdisc); err == nil {
		ilog.Warn("qdisc clsact already existed. Deleted it")
	}
	if err := handle.QdiscAdd(qdisc); err != nil {
		if errors.Is(err, fs.ErrExist) {
			ilog.WithError(err).Warn("qdisc clsact already exists. Ignoring")
		} else {
			return fmt.Errorf("failed to create clsact qdisc on %d (%s): %w", iface.Index, iface.Name, err)
		}
	}
	m.qdiscs[iface] = qdisc

	// Remove previously installed filters
	if err := unregister(iface); err != nil {
		return fmt.Errorf("failed to remove previous filters: %w", err)
	}

	if err := m.registerEgress(iface, ipvlan, handle); err != nil {
		return err
	}

	return m.registerIngress(iface, ipvlan, handle)
}

func (m *FlowFetcher) registerEgress(iface ifaces.Interface, ipvlan netlink.Link, handle *netlink.Handle) error {
	ilog := log.WithField("iface", iface)
	if !m.enableEgress {
		ilog.Debug("ignoring egress traffic, according to user configuration")
		return nil
	}
	// Fetch events on egress
	egressAttrs := netlink.FilterAttrs{
		LinkIndex: ipvlan.Attrs().Index,
		Parent:    netlink.HANDLE_MIN_EGRESS,
		Handle:    netlink.MakeHandle(0, 1),
		Protocol:  3,
		Priority:  1,
	}
	egressFilter := &netlink.BpfFilter{
		FilterAttrs:  egressAttrs,
		Fd:           m.objects.TcEgressFlowParse.FD(),
		Name:         tcEgressFilterName,
		DirectAction: true,
	}
	if err := handle.FilterDel(egressFilter); err == nil {
		ilog.Warn("egress filter already existed. Deleted it")
	}
	if err := handle.FilterAdd(egressFilter); err != nil {
		if errors.Is(err, fs.ErrExist) {
			ilog.WithError(err).Warn("egress filter already exists. Ignoring")
		} else {
			return fmt.Errorf("failed to create egress filter: %w", err)
		}
	}
	m.egressFilters[iface] = egressFilter
	return nil
}

func (m *FlowFetcher) registerIngress(iface ifaces.Interface, ipvlan netlink.Link, handle *netlink.Handle) error {
	ilog := log.WithField("iface", iface)
	if !m.enableIngress {
		ilog.Debug("ignoring ingress traffic, according to user configuration")
		return nil
	}
	// Fetch events on ingress
	ingressAttrs := netlink.FilterAttrs{
		LinkIndex: ipvlan.Attrs().Index,
		Parent:    netlink.HANDLE_MIN_INGRESS,
		Handle:    netlink.MakeHandle(0, 1),
		Protocol:  unix.ETH_P_ALL,
		Priority:  1,
	}
	ingressFilter := &netlink.BpfFilter{
		FilterAttrs:  ingressAttrs,
		Fd:           m.objects.TcIngressFlowParse.FD(),
		Name:         tcIngressFilterName,
		DirectAction: true,
	}
	if err := handle.FilterDel(ingressFilter); err == nil {
		ilog.Warn("ingress filter already existed. Deleted it")
	}
	if err := handle.FilterAdd(ingressFilter); err != nil {
		if errors.Is(err, fs.ErrExist) {
			ilog.WithError(err).Warn("ingress filter already exists. Ignoring")
		} else {
			return fmt.Errorf("failed to create ingress filter: %w", err)
		}
	}
	m.ingressFilters[iface] = ingressFilter
	return nil
}

// Close the eBPF fetcher from the system.
// We don't need a "Close(iface)" method because the filters and qdiscs
// are automatically removed when the interface is down
// nolint:cyclop
func (m *FlowFetcher) Close() error {
	log.Debug("unregistering eBPF objects")

	var errs []error

	if m.pktDropsTracePoint != nil {
		if err := m.pktDropsTracePoint.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if m.rttFentryLink != nil {
		if err := m.rttFentryLink.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if m.rttKprobeLink != nil {
		if err := m.rttKprobeLink.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if m.networkEventsMonitoringLink != nil {
		if err := m.networkEventsMonitoringLink.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if m.nfNatManIPLink != nil {
		if err := m.nfNatManIPLink.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	// m.ringbufReader.Read is a blocking operation, so we need to close the ring buffer
	// from another goroutine to avoid the system not being able to exit if there
	// isn't traffic in a given interface
	if m.ringbufReader != nil {
		if err := m.ringbufReader.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if m.objects != nil {
		if err := m.objects.TcEgressFlowParse.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := m.objects.TcIngressFlowParse.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := m.objects.TcxEgressFlowParse.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := m.objects.TcxIngressFlowParse.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := m.objects.AggregatedFlows.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := m.objects.AdditionalFlowMetrics.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := m.objects.DirectFlows.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := m.objects.DnsFlows.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := m.objects.GlobalCounters.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := m.objects.FilterMap.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := m.objects.PeerFilterMap.Close(); err != nil {
			errs = append(errs, err)
		}
		if len(errs) == 0 {
			m.objects = nil
		}
	}

	for iface, ef := range m.egressFilters {
		log := log.WithField("interface", iface)
		log.Debug("deleting egress filter")
		if err := doIgnoreNoDev(netlink.FilterDel, netlink.Filter(ef), log); err != nil {
			errs = append(errs, fmt.Errorf("deleting egress filter: %w", err))
		}
	}
	m.egressFilters = map[ifaces.Interface]*netlink.BpfFilter{}
	for iface, igf := range m.ingressFilters {
		log := log.WithField("interface", iface)
		log.Debug("deleting ingress filter")
		if err := doIgnoreNoDev(netlink.FilterDel, netlink.Filter(igf), log); err != nil {
			errs = append(errs, fmt.Errorf("deleting ingress filter: %w", err))
		}
	}
	m.ingressFilters = map[ifaces.Interface]*netlink.BpfFilter{}
	for iface, qd := range m.qdiscs {
		log := log.WithField("interface", iface)
		log.Debug("deleting Qdisc")
		if err := doIgnoreNoDev(netlink.QdiscDel, netlink.Qdisc(qd), log); err != nil {
			errs = append(errs, fmt.Errorf("deleting qdisc: %w", err))
		}
	}
	m.qdiscs = map[ifaces.Interface]*netlink.GenericQdisc{}

	for iface, l := range m.egressTCXLink {
		log := log.WithField("interface", iface)
		log.Debug("detach egress TCX hook")
		l.Close()
	}
	m.egressTCXLink = map[ifaces.Interface]link.Link{}
	for iface, l := range m.ingressTCXLink {
		log := log.WithField("interface", iface)
		log.Debug("detach ingress TCX hook")
		l.Close()
	}
	m.ingressTCXLink = map[ifaces.Interface]link.Link{}

	if !m.useEbpfManager {
		if err := m.removeAllPins(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	var errStrings []string
	for _, err := range errs {
		errStrings = append(errStrings, err.Error())
	}
	return errors.New(`errors: "` + strings.Join(errStrings, `", "`) + `"`)
}

// removeAllPins removes all pins.
func (m *FlowFetcher) removeAllPins() error {
	files, err := os.ReadDir(m.pinDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, file := range files {
		if err := os.Remove(path.Join(m.pinDir, file.Name())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.Remove(m.pinDir); err != nil {
		return err
	}
	return nil
}

// doIgnoreNoDev runs the provided syscall over the provided device and ignores the error
// if the cause is a non-existing device (just logs the error as debug).
// If the agent is deployed as part of the Network Observability pipeline, normally
// undeploying the FlowCollector could cause the agent to try to remove resources
// from Pods that have been removed immediately before (e.g. flowlogs-pipeline or the
// console plugin), so we avoid logging some errors that would unnecessarily raise the
// user's attention.
// This function uses generics because the set of provided functions accept different argument
// types.
func doIgnoreNoDev[T any](sysCall func(T) error, dev T, log *logrus.Entry) error {
	if err := sysCall(dev); err != nil {
		if errors.Is(err, unix.ENODEV) {
			log.WithError(err).Error("can't delete. Ignore this error if other pods or interfaces " +
				" are also being deleted at this moment. For example, if you are undeploying " +
				" a FlowCollector or Deployment where this agent is part of")
		} else {
			return err
		}
	}
	return nil
}

func (m *FlowFetcher) ReadRingBuf() (ringbuf.Record, error) {
	return m.ringbufReader.Read()
}

// LookupAndDeleteMap reads all the entries from the eBPF map and removes them from it.
// TODO: detect whether BatchLookupAndDelete is supported (Kernel>=5.6) and use it selectively
// Supported Lookup/Delete operations by kernel: https://github.com/iovisor/bcc/blob/master/docs/kernel-versions.md
func (m *FlowFetcher) LookupAndDeleteMap(met *metrics.Metrics) map[ebpf.BpfFlowId]model.BpfFlowContent {
	if !m.lookupAndDeleteSupported {
		return m.legacyLookupAndDeleteMap(met)
	}

	flowMap := m.objects.AggregatedFlows
	var flows = make(map[ebpf.BpfFlowId]model.BpfFlowContent, m.cacheMaxSize)
	var ids []ebpf.BpfFlowId
	var id ebpf.BpfFlowId
	var baseMetrics ebpf.BpfFlowMetrics

	// First, get all ids and don't care about metrics (we need lookup+delete to be atomic)
	iterator := flowMap.Iterate()
	for iterator.Next(&id, &baseMetrics) {
		ids = append(ids, id)
	}

	countMain := 0
	// Run the atomic Lookup+Delete; if new ids have been inserted in the meantime, they'll be fetched next time
	for i, id := range ids {
		countMain++
		if err := flowMap.LookupAndDelete(&id, &baseMetrics); err != nil {
			if i == 0 && errors.Is(err, cilium.ErrNotSupported) {
				log.WithError(err).Warnf("switching to legacy mode")
				m.lookupAndDeleteSupported = false
				return m.legacyLookupAndDeleteMap(met)
			}
			log.WithError(err).WithField("flowId", id).Warnf("couldn't lookup/delete flow entry")
			met.Errors.WithErrorName("flow-fetcher", "CannotDeleteFlows").Inc()
			continue
		}
		flows[id] = model.NewBpfFlowContent(baseMetrics)
	}

	// Reiterate on additional metrics
	var additionalMetrics []ebpf.BpfAdditionalMetrics
	ids = []ebpf.BpfFlowId{}
	addtlIterator := m.objects.AdditionalFlowMetrics.Iterate()
	for addtlIterator.Next(&id, &additionalMetrics) {
		ids = append(ids, id)
	}

	countAdditional := 0
	for i, id := range ids {
		countAdditional++
		if err := m.objects.AdditionalFlowMetrics.LookupAndDelete(&id, &additionalMetrics); err != nil {
			if i == 0 && errors.Is(err, cilium.ErrNotSupported) {
				log.WithError(err).Warnf("switching to legacy mode")
				m.lookupAndDeleteSupported = false
				return m.legacyLookupAndDeleteMap(met)
			}
			log.WithError(err).WithField("flowId", id).Warnf("couldn't lookup/delete additional metrics entry")
			met.Errors.WithErrorName("flow-fetcher", "CannotDeleteAdditionalMetric").Inc()
			continue
		}
		flow, found := flows[id]
		if !found {
			flow = model.BpfFlowContent{BpfFlowMetrics: &ebpf.BpfFlowMetrics{}}
		}
		for iMet := range additionalMetrics {
			flow.AccumulateAdditional(&additionalMetrics[iMet])
		}
		m.increaseEnrichmentStats(met, &flow)
		flows[id] = flow
	}
	met.BufferSizeGauge.WithBufferName("additionalmap").Set(float64(countAdditional))
	met.BufferSizeGauge.WithBufferName("flowmap").Set(float64(countMain))
	met.BufferSizeGauge.WithBufferName("merged-maps").Set(float64(len(flows)))

	m.ReadGlobalCounter(met)
	return flows
}

func (m *FlowFetcher) increaseEnrichmentStats(met *metrics.Metrics, flow *model.BpfFlowContent) {
	if flow.AdditionalMetrics != nil {
		met.FlowEnrichmentCounter.Increase(
			flow.AdditionalMetrics.DnsRecord.Id != 0,
			flow.AdditionalMetrics.FlowRtt != 0,
			flow.AdditionalMetrics.PktDrops.Packets != 0,
			!model.AllZerosMetaData(flow.AdditionalMetrics.NetworkEvents[0]),
			!model.AllZeroIP(model.IP(flow.AdditionalMetrics.TranslatedFlow.Daddr)),
		)
	} else {
		met.FlowEnrichmentCounter.Increase(false, false, false, false, false)
	}
}

// ReadGlobalCounter reads the global counter and updates drop flows counter metrics
func (m *FlowFetcher) ReadGlobalCounter(met *metrics.Metrics) {
	var allCPUValue []uint32
	globalCounters := map[ebpf.BpfGlobalCountersKeyT]prometheus.Counter{
		ebpf.BpfGlobalCountersKeyTHASHMAP_FLOWS_DROPPED:               met.DroppedFlowsCounter.WithSourceAndReason("flow-fetcher", "CannotUpdateFlowsHashMap"),
		ebpf.BpfGlobalCountersKeyTHASHMAP_FAIL_UPDATE_DNS:             met.DroppedFlowsCounter.WithSourceAndReason("flow-fetcher", "CannotUpdateDNSHashMap"),
		ebpf.BpfGlobalCountersKeyTFILTER_REJECT:                       met.FilteredFlowsCounter.WithSourceAndReason("flow-filtering", "FilterReject"),
		ebpf.BpfGlobalCountersKeyTFILTER_ACCEPT:                       met.FilteredFlowsCounter.WithSourceAndReason("flow-filtering", "FilterAccept"),
		ebpf.BpfGlobalCountersKeyTFILTER_NOMATCH:                      met.FilteredFlowsCounter.WithSourceAndReason("flow-filtering", "FilterNoMatch"),
		ebpf.BpfGlobalCountersKeyTNETWORK_EVENTS_ERR:                  met.NetworkEventsCounter.WithSourceAndReason("network-events", "NetworkEventsErrors"),
		ebpf.BpfGlobalCountersKeyTNETWORK_EVENTS_ERR_GROUPID_MISMATCH: met.NetworkEventsCounter.WithSourceAndReason("network-events", "NetworkEventsErrorsGroupIDMismatch"),
		ebpf.BpfGlobalCountersKeyTNETWORK_EVENTS_ERR_UPDATE_MAP_FLOWS: met.NetworkEventsCounter.WithSourceAndReason("network-events", "NetworkEventsErrorsFlowMapUpdate"),
		ebpf.BpfGlobalCountersKeyTNETWORK_EVENTS_GOOD:                 met.NetworkEventsCounter.WithSourceAndReason("network-events", "NetworkEventsGoodEvent"),
		ebpf.BpfGlobalCountersKeyTOBSERVED_INTF_MISSED:                met.Errors.WithErrorName("flow-fetcher", "MaxObservedInterfacesReached"),
	}
	zeroCounters := make([]uint32, cilium.MustPossibleCPU())
	for key := ebpf.BpfGlobalCountersKeyT(0); key < ebpf.BpfGlobalCountersKeyTMAX_COUNTERS; key++ {
		if err := m.objects.GlobalCounters.Lookup(key, &allCPUValue); err != nil {
			log.WithError(err).Warnf("couldn't read global counter")
			return
		}
		metric := globalCounters[key]
		if metric != nil {
			// aggregate all the counters
			for _, counter := range allCPUValue {
				metric.Add(float64(counter))
			}
		}
		// reset the global counter-map entry
		if err := m.objects.GlobalCounters.Put(key, zeroCounters); err != nil {
			log.WithError(err).Warnf("coudn't reset global counter")
			return
		}
	}
}

// DeleteMapsStaleEntries Look for any stale entries in the features maps and delete them
func (m *FlowFetcher) DeleteMapsStaleEntries(timeOut time.Duration) {
	m.lookupAndDeleteDNSMap(timeOut)
}

// lookupAndDeleteDNSMap iterate over DNS queries map and delete any stale DNS requests
// entries which never get responses for.
func (m *FlowFetcher) lookupAndDeleteDNSMap(timeOut time.Duration) {
	monotonicTimeNow := monotime.Now()
	dnsMap := m.objects.DnsFlows
	var dnsKey ebpf.BpfDnsFlowId
	var keysToDelete []ebpf.BpfDnsFlowId
	var dnsVal uint64

	if dnsMap != nil {
		// Ideally the Lookup + Delete should be atomic, however we cannot use LookupAndDelete since the deletion is conditional
		// Do not delete while iterating, as it causes severe performance degradation
		iterator := dnsMap.Iterate()
		for iterator.Next(&dnsKey, &dnsVal) {
			if time.Duration(uint64(monotonicTimeNow)-dnsVal) >= timeOut {
				keysToDelete = append(keysToDelete, dnsKey)
			}
		}
		for _, dnsKey = range keysToDelete {
			if err := dnsMap.Delete(dnsKey); err != nil {
				log.WithError(err).WithField("dnsKey", dnsKey).Warnf("couldn't delete DNS record entry")
			}
		}
	}
}

// kernelSpecificLoadAndAssign based on a kernel version, it will load only the supported eBPF hooks
func kernelSpecificLoadAndAssign(oldKernel, rtKernel, supportNetworkEvents bool, spec *cilium.CollectionSpec, pinDir string) (ebpf.BpfObjects, error) {
	objects := ebpf.BpfObjects{}

	// Helper to remove common hooks
	removeCommonHooks := func() {
		delete(spec.Programs, pktDropHook)
		delete(spec.Programs, rhNetworkEventsMonitoringHook)
	}

	// Helper to load and assign BPF objects
	loadAndAssign := func(objects interface{}) error {
		if err := spec.LoadAndAssign(objects, &cilium.CollectionOptions{Maps: cilium.MapOptions{PinPath: pinDir}}); err != nil {
			var ve *cilium.VerifierError
			if errors.As(err, &ve) {
				log.Infof("Verifier error: %+v", ve)
			}
			return fmt.Errorf("loading and assigning BPF objects: %w", err)
		}
		return nil
	}

	// Configure BPF programs based on the kernel type
	switch {
	case oldKernel && rtKernel:
		type newBpfPrograms struct {
			TcEgressFlowParse   *cilium.Program `ebpf:"tc_egress_flow_parse"`
			TcIngressFlowParse  *cilium.Program `ebpf:"tc_ingress_flow_parse"`
			TcxEgressFlowParse  *cilium.Program `ebpf:"tcx_egress_flow_parse"`
			TcxIngressFlowParse *cilium.Program `ebpf:"tcx_ingress_flow_parse"`
			TcEgressPcaParse    *cilium.Program `ebpf:"tc_egress_pca_parse"`
			TcIngressPcaParse   *cilium.Program `ebpf:"tc_ingress_pca_parse"`
			TcxEgressPcaParse   *cilium.Program `ebpf:"tcx_egress_pca_parse"`
			TcxIngressPcaParse  *cilium.Program `ebpf:"tcx_ingress_pca_parse"`
			TrackNatManipPkt    *cilium.Program `ebpf:"track_nat_manip_pkt"`
		}
		type newBpfObjects struct {
			newBpfPrograms
			ebpf.BpfMaps
		}
		var newObjects newBpfObjects
		removeCommonHooks()
		delete(spec.Programs, tcpRcvKprobe)
		delete(spec.Programs, tcpFentryHook)

		if err := loadAndAssign(&newObjects); err != nil {
			return objects, err
		}

		objects = ebpf.BpfObjects{
			BpfPrograms: ebpf.BpfPrograms{
				TcEgressFlowParse:         newObjects.TcEgressFlowParse,
				TcIngressFlowParse:        newObjects.TcIngressFlowParse,
				TcxEgressFlowParse:        newObjects.TcxEgressFlowParse,
				TcxIngressFlowParse:       newObjects.TcxIngressFlowParse,
				TcEgressPcaParse:          newObjects.TcEgressPcaParse,
				TcIngressPcaParse:         newObjects.TcIngressPcaParse,
				TcxEgressPcaParse:         newObjects.TcxEgressPcaParse,
				TcxIngressPcaParse:        newObjects.TcxIngressPcaParse,
				TrackNatManipPkt:          newObjects.TrackNatManipPkt,
				TcpRcvKprobe:              nil,
				TcpRcvFentry:              nil,
				KfreeSkb:                  nil,
				RhNetworkEventsMonitoring: nil,
			},
			BpfMaps: ebpf.BpfMaps{
				DirectFlows:           newObjects.DirectFlows,
				AggregatedFlows:       newObjects.AggregatedFlows,
				AdditionalFlowMetrics: newObjects.AdditionalFlowMetrics,
				DnsFlows:              newObjects.DnsFlows,
				FilterMap:             newObjects.FilterMap,
				PeerFilterMap:         newObjects.PeerFilterMap,
				GlobalCounters:        newObjects.GlobalCounters,
			},
		}

	case oldKernel:
		type newBpfPrograms struct {
			TcEgressFlowParse   *cilium.Program `ebpf:"tc_egress_flow_parse"`
			TcIngressFlowParse  *cilium.Program `ebpf:"tc_ingress_flow_parse"`
			TcxEgressFlowParse  *cilium.Program `ebpf:"tcx_egress_flow_parse"`
			TcxIngressFlowParse *cilium.Program `ebpf:"tcx_ingress_flow_parse"`
			TcEgressPcaParse    *cilium.Program `ebpf:"tc_egress_pca_parse"`
			TcIngressPcaParse   *cilium.Program `ebpf:"tc_ingress_pca_parse"`
			TcxEgressPcaParse   *cilium.Program `ebpf:"tcx_egress_pca_parse"`
			TcxIngressPcaParse  *cilium.Program `ebpf:"tcx_ingress_pca_parse"`
			TCPRcvKprobe        *cilium.Program `ebpf:"tcp_rcv_kprobe"`
			TrackNatManipPkt    *cilium.Program `ebpf:"track_nat_manip_pkt"`
		}
		type newBpfObjects struct {
			newBpfPrograms
			ebpf.BpfMaps
		}
		var newObjects newBpfObjects
		removeCommonHooks()
		delete(spec.Programs, tcpFentryHook)

		if err := loadAndAssign(&newObjects); err != nil {
			return objects, err
		}

		objects = ebpf.BpfObjects{
			BpfPrograms: ebpf.BpfPrograms{
				TcEgressFlowParse:         newObjects.TcEgressFlowParse,
				TcIngressFlowParse:        newObjects.TcIngressFlowParse,
				TcxEgressFlowParse:        newObjects.TcxEgressFlowParse,
				TcxIngressFlowParse:       newObjects.TcxIngressFlowParse,
				TcEgressPcaParse:          newObjects.TcEgressPcaParse,
				TcIngressPcaParse:         newObjects.TcIngressPcaParse,
				TcxEgressPcaParse:         newObjects.TcxEgressPcaParse,
				TcxIngressPcaParse:        newObjects.TcxIngressPcaParse,
				TcpRcvKprobe:              newObjects.TCPRcvKprobe,
				TrackNatManipPkt:          newObjects.TrackNatManipPkt,
				TcpRcvFentry:              nil,
				KfreeSkb:                  nil,
				RhNetworkEventsMonitoring: nil,
			},
			BpfMaps: ebpf.BpfMaps{
				DirectFlows:           newObjects.DirectFlows,
				AggregatedFlows:       newObjects.AggregatedFlows,
				AdditionalFlowMetrics: newObjects.AdditionalFlowMetrics,
				DnsFlows:              newObjects.DnsFlows,
				FilterMap:             newObjects.FilterMap,
				PeerFilterMap:         newObjects.PeerFilterMap,
				GlobalCounters:        newObjects.GlobalCounters,
			},
		}

	case rtKernel:
		type newBpfPrograms struct {
			TcEgressFlowParse   *cilium.Program `ebpf:"tc_egress_flow_parse"`
			TcIngressFlowParse  *cilium.Program `ebpf:"tc_ingress_flow_parse"`
			TcxEgressFlowParse  *cilium.Program `ebpf:"tcx_egress_flow_parse"`
			TcxIngressFlowParse *cilium.Program `ebpf:"tcx_ingress_flow_parse"`
			TcEgressPcaParse    *cilium.Program `ebpf:"tc_egress_pca_parse"`
			TcIngressPcaParse   *cilium.Program `ebpf:"tc_ingress_pca_parse"`
			TcxEgressPcaParse   *cilium.Program `ebpf:"tcx_egress_pca_parse"`
			TcxIngressPcaParse  *cilium.Program `ebpf:"tcx_ingress_pca_parse"`
			TCPRcvFentry        *cilium.Program `ebpf:"tcp_rcv_fentry"`
			TrackNatManipPkt    *cilium.Program `ebpf:"track_nat_manip_pkt"`
		}
		type newBpfObjects struct {
			newBpfPrograms
			ebpf.BpfMaps
		}
		var newObjects newBpfObjects
		removeCommonHooks()
		delete(spec.Programs, tcpRcvKprobe)

		if err := loadAndAssign(&newObjects); err != nil {
			return objects, err
		}

		objects = ebpf.BpfObjects{
			BpfPrograms: ebpf.BpfPrograms{
				TcEgressFlowParse:         newObjects.TcEgressFlowParse,
				TcIngressFlowParse:        newObjects.TcIngressFlowParse,
				TcxEgressFlowParse:        newObjects.TcxEgressFlowParse,
				TcxIngressFlowParse:       newObjects.TcxIngressFlowParse,
				TcEgressPcaParse:          newObjects.TcEgressPcaParse,
				TcIngressPcaParse:         newObjects.TcIngressPcaParse,
				TcxEgressPcaParse:         newObjects.TcxEgressPcaParse,
				TcxIngressPcaParse:        newObjects.TcxIngressPcaParse,
				TcpRcvFentry:              newObjects.TCPRcvFentry,
				TrackNatManipPkt:          newObjects.TrackNatManipPkt,
				TcpRcvKprobe:              nil,
				KfreeSkb:                  nil,
				RhNetworkEventsMonitoring: nil,
			},
			BpfMaps: ebpf.BpfMaps{
				DirectFlows:           newObjects.DirectFlows,
				AggregatedFlows:       newObjects.AggregatedFlows,
				AdditionalFlowMetrics: newObjects.AdditionalFlowMetrics,
				DnsFlows:              newObjects.DnsFlows,
				FilterMap:             newObjects.FilterMap,
				PeerFilterMap:         newObjects.PeerFilterMap,
				GlobalCounters:        newObjects.GlobalCounters,
			},
		}

	case !supportNetworkEvents:
		type newBpfPrograms struct {
			TcEgressFlowParse   *cilium.Program `ebpf:"tc_egress_flow_parse"`
			TcIngressFlowParse  *cilium.Program `ebpf:"tc_ingress_flow_parse"`
			TcxEgressFlowParse  *cilium.Program `ebpf:"tcx_egress_flow_parse"`
			TcxIngressFlowParse *cilium.Program `ebpf:"tcx_ingress_flow_parse"`
			TcEgressPcaParse    *cilium.Program `ebpf:"tc_egress_pca_parse"`
			TcIngressPcaParse   *cilium.Program `ebpf:"tc_ingress_pca_parse"`
			TcxEgressPcaParse   *cilium.Program `ebpf:"tcx_egress_pca_parse"`
			TcxIngressPcaParse  *cilium.Program `ebpf:"tcx_ingress_pca_parse"`
			TCPRcvFentry        *cilium.Program `ebpf:"tcp_rcv_fentry"`
			TCPRcvKprobe        *cilium.Program `ebpf:"tcp_rcv_kprobe"`
			KfreeSkb            *cilium.Program `ebpf:"kfree_skb"`
			TrackNatManipPkt    *cilium.Program `ebpf:"track_nat_manip_pkt"`
		}
		type newBpfObjects struct {
			newBpfPrograms
			ebpf.BpfMaps
		}
		var newObjects newBpfObjects
		delete(spec.Programs, rhNetworkEventsMonitoringHook)

		if err := loadAndAssign(&newObjects); err != nil {
			return objects, err
		}

		objects = ebpf.BpfObjects{
			BpfPrograms: ebpf.BpfPrograms{
				TcEgressFlowParse:         newObjects.TcEgressFlowParse,
				TcIngressFlowParse:        newObjects.TcIngressFlowParse,
				TcxEgressFlowParse:        newObjects.TcxEgressFlowParse,
				TcxIngressFlowParse:       newObjects.TcxIngressFlowParse,
				TcEgressPcaParse:          newObjects.TcEgressPcaParse,
				TcIngressPcaParse:         newObjects.TcIngressPcaParse,
				TcxEgressPcaParse:         newObjects.TcxEgressPcaParse,
				TcxIngressPcaParse:        newObjects.TcxIngressPcaParse,
				TcpRcvFentry:              newObjects.TCPRcvFentry,
				TcpRcvKprobe:              newObjects.TCPRcvKprobe,
				KfreeSkb:                  newObjects.KfreeSkb,
				TrackNatManipPkt:          newObjects.TrackNatManipPkt,
				RhNetworkEventsMonitoring: nil,
			},
			BpfMaps: ebpf.BpfMaps{
				DirectFlows:           newObjects.DirectFlows,
				AggregatedFlows:       newObjects.AggregatedFlows,
				AdditionalFlowMetrics: newObjects.AdditionalFlowMetrics,
				DnsFlows:              newObjects.DnsFlows,
				FilterMap:             newObjects.FilterMap,
				PeerFilterMap:         newObjects.PeerFilterMap,
				GlobalCounters:        newObjects.GlobalCounters,
			},
		}

	default:
		if err := loadAndAssign(&objects); err != nil {
			return objects, err
		}
	}

	// Release cached kernel BTF memory
	btf.FlushKernelSpec()

	return objects, nil
}

// It provides access to packets from  the kernel space (via PerfCPU hashmap)
type PacketFetcher struct {
	objects                  *ebpf.BpfObjects
	qdiscs                   map[ifaces.Interface]*netlink.GenericQdisc
	egressFilters            map[ifaces.Interface]*netlink.BpfFilter
	ingressFilters           map[ifaces.Interface]*netlink.BpfFilter
	perfReader               *perf.Reader
	cacheMaxSize             int
	enableIngress            bool
	enableEgress             bool
	egressTCXLink            map[ifaces.Interface]link.Link
	ingressTCXLink           map[ifaces.Interface]link.Link
	lookupAndDeleteSupported bool
}

func NewPacketFetcher(cfg *FlowFetcherConfig) (*PacketFetcher, error) {
	if err := rlimit.RemoveMemlock(); err != nil {
		log.WithError(err).
			Warn("can't remove mem lock. The agent could not be able to start eBPF programs")
	}

	objects := ebpf.BpfObjects{}
	spec, err := ebpf.LoadBpf()
	if err != nil {
		return nil, err
	}
	pcaEnable := 0
	if cfg.EnablePCA {
		pcaEnable = 1
	}

	if err := spec.RewriteConstants(map[string]interface{}{
		constSampling:  uint32(cfg.Sampling),
		constPcaEnable: uint8(pcaEnable),
	}); err != nil {
		return nil, fmt.Errorf("rewriting BPF constants definition: %w", err)
	}

	// remove pinning from all maps
	maps2Name := []string{"aggregated_flows", "additional_flow_metrics", "direct_flows", "dns_flows", "filter_map", "global_counters", "packet_record"}
	for _, m := range maps2Name {
		spec.Maps[m].Pinning = 0
	}

	type pcaBpfPrograms struct {
		TcEgressPcaParse   *cilium.Program `ebpf:"tc_egress_pca_parse"`
		TcIngressPcaParse  *cilium.Program `ebpf:"tc_ingress_pca_parse"`
		TcxEgressPcaParse  *cilium.Program `ebpf:"tcx_egress_pca_parse"`
		TcxIngressPcaParse *cilium.Program `ebpf:"tcx_ingress_pca_parse"`
	}
	type newBpfObjects struct {
		pcaBpfPrograms
		ebpf.BpfMaps
	}
	var newObjects newBpfObjects
	delete(spec.Programs, pktDropHook)
	delete(spec.Programs, rhNetworkEventsMonitoringHook)
	delete(spec.Programs, tcpRcvKprobe)
	delete(spec.Programs, tcpFentryHook)
	delete(spec.Programs, aggregatedFlowsMap)
	delete(spec.Programs, additionalFlowMetrics)
	delete(spec.Programs, constSampling)
	delete(spec.Programs, constHasFilterSampling)
	delete(spec.Programs, constTraceMessages)
	delete(spec.Programs, constEnableDNSTracking)
	delete(spec.Programs, constDNSTrackingPort)
	delete(spec.Programs, constEnableRtt)
	delete(spec.Programs, constEnableFlowFiltering)
	delete(spec.Programs, constEnableNetworkEventsMonitoring)
	delete(spec.Programs, constNetworkEventsMonitoringGroupID)

	if err := spec.LoadAndAssign(&newObjects, &cilium.CollectionOptions{Maps: cilium.MapOptions{PinPath: ""}}); err != nil {
		var ve *cilium.VerifierError
		if errors.As(err, &ve) {
			// Using %+v will print the whole verifier error, not just the last
			// few lines.
			plog.Infof("Verifier error: %+v", ve)
		}
		return nil, fmt.Errorf("loading and assigning BPF objects: %w", err)
	}

	objects = ebpf.BpfObjects{
		BpfPrograms: ebpf.BpfPrograms{
			TcEgressPcaParse:          newObjects.TcEgressPcaParse,
			TcIngressPcaParse:         newObjects.TcIngressPcaParse,
			TcxEgressPcaParse:         newObjects.TcxEgressPcaParse,
			TcxIngressPcaParse:        newObjects.TcxIngressPcaParse,
			TcEgressFlowParse:         nil,
			TcIngressFlowParse:        nil,
			TcxEgressFlowParse:        nil,
			TcxIngressFlowParse:       nil,
			TcpRcvFentry:              nil,
			TcpRcvKprobe:              nil,
			KfreeSkb:                  nil,
			RhNetworkEventsMonitoring: nil,
		},
		BpfMaps: ebpf.BpfMaps{
			PacketRecord:  newObjects.PacketRecord,
			FilterMap:     newObjects.FilterMap,
			PeerFilterMap: newObjects.PeerFilterMap,
		},
	}

	f := NewFilter(cfg.FilterConfig)
	if err := f.ProgramFilter(&objects); err != nil {
		return nil, fmt.Errorf("programming flow filter: %w", err)
	}

	// read packets from igress+egress perf array
	packets, err := perf.NewReader(objects.PacketRecord, os.Getpagesize())
	if err != nil {
		return nil, fmt.Errorf("accessing to perf: %w", err)
	}

	return &PacketFetcher{
		objects:                  &objects,
		perfReader:               packets,
		egressFilters:            map[ifaces.Interface]*netlink.BpfFilter{},
		ingressFilters:           map[ifaces.Interface]*netlink.BpfFilter{},
		qdiscs:                   map[ifaces.Interface]*netlink.GenericQdisc{},
		cacheMaxSize:             cfg.CacheMaxSize,
		enableIngress:            cfg.EnableIngress,
		enableEgress:             cfg.EnableEgress,
		egressTCXLink:            map[ifaces.Interface]link.Link{},
		ingressTCXLink:           map[ifaces.Interface]link.Link{},
		lookupAndDeleteSupported: true, // this will be turned off later if found to be not supported
	}, nil
}

func registerInterface(iface ifaces.Interface) (*netlink.GenericQdisc, netlink.Link, error) {
	ilog := plog.WithField("iface", iface)
	handle, err := netlink.NewHandleAt(iface.NetNS)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create handle for netns (%s): %w", iface.NetNS.String(), err)
	}
	defer handle.Close()

	// Load pre-compiled programs and maps into the kernel, and rewrites the configuration
	ipvlan, err := handle.LinkByIndex(iface.Index)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup ipvlan device %d (%s): %w", iface.Index, iface.Name, err)
	}
	qdiscAttrs := netlink.QdiscAttrs{
		LinkIndex: ipvlan.Attrs().Index,
		Handle:    netlink.MakeHandle(0xffff, 0),
		Parent:    netlink.HANDLE_CLSACT,
	}
	qdisc := &netlink.GenericQdisc{
		QdiscAttrs: qdiscAttrs,
		QdiscType:  qdiscType,
	}
	if err := handle.QdiscDel(qdisc); err == nil {
		ilog.Warn("qdisc clsact already existed. Deleted it")
	}
	if err := handle.QdiscAdd(qdisc); err != nil {
		if errors.Is(err, fs.ErrExist) {
			ilog.WithError(err).Warn("qdisc clsact already exists. Ignoring")
		} else {
			return nil, nil, fmt.Errorf("failed to create clsact qdisc on %d (%s): %w", iface.Index, iface.Name, err)
		}
	}
	return qdisc, ipvlan, nil
}

func (p *PacketFetcher) UnRegister(iface ifaces.Interface) error {
	// qdiscs, ingress and egress filters are automatically deleted so we don't need to
	// specifically detach them from the ebpfFetcher
	return unregister(iface)
}

func (p *PacketFetcher) Register(iface ifaces.Interface) error {
	qdisc, ipvlan, err := registerInterface(iface)
	if err != nil {
		return err
	}
	p.qdiscs[iface] = qdisc

	if err := p.registerEgress(iface, ipvlan); err != nil {
		return err
	}
	return p.registerIngress(iface, ipvlan)
}

func (p *PacketFetcher) DetachTCX(iface ifaces.Interface) error {
	ilog := log.WithField("iface", iface)
	if iface.NetNS != netns.None() {
		originalNs, err := netns.Get()
		if err != nil {
			return fmt.Errorf("PCA failed to get current netns: %w", err)
		}
		defer func() {
			if err := netns.Set(originalNs); err != nil {
				ilog.WithError(err).Error("PCA failed to set netns back")
			}
			originalNs.Close()
		}()
		if err := unix.Setns(int(iface.NetNS), unix.CLONE_NEWNET); err != nil {
			return fmt.Errorf("PCA failed to setns to %s: %w", iface.NetNS, err)
		}
	}
	if p.enableEgress {
		if l := p.egressTCXLink[iface]; l != nil {
			if err := l.Close(); err != nil {
				return fmt.Errorf("TCX: failed to close egress link: %w", err)
			}
			ilog.WithField("interface", iface.Name).Debug("successfully detach egressTCX hook")
		} else {
			return fmt.Errorf("egress link does not support TCX hook")
		}
	}

	if p.enableIngress {
		if l := p.ingressTCXLink[iface]; l != nil {
			if err := l.Close(); err != nil {
				return fmt.Errorf("TCX: failed to close ingress link: %w", err)
			}
			ilog.WithField("interface", iface.Name).Debug("successfully detach ingressTCX hook")
		} else {
			return fmt.Errorf("ingress link does not support TCX hook")
		}
	}
	return nil
}

func (p *PacketFetcher) AttachTCX(iface ifaces.Interface) error {
	ilog := log.WithField("iface", iface)
	if iface.NetNS != netns.None() {
		originalNs, err := netns.Get()
		if err != nil {
			return fmt.Errorf("PCA failed to get current netns: %w", err)
		}
		defer func() {
			if err := netns.Set(originalNs); err != nil {
				ilog.WithError(err).Error("PCA failed to set netns back")
			}
			originalNs.Close()
		}()
		if err := unix.Setns(int(iface.NetNS), unix.CLONE_NEWNET); err != nil {
			return fmt.Errorf("PCA failed to setns to %s: %w", iface.NetNS, err)
		}
	}

	if p.enableEgress {
		egrLink, err := link.AttachTCX(link.TCXOptions{
			Program:   p.objects.BpfPrograms.TcxEgressPcaParse,
			Attach:    cilium.AttachTCXEgress,
			Interface: iface.Index,
		})
		if err != nil {
			if errors.Is(err, fs.ErrExist) {
				// The interface already has a TCX egress hook
				log.WithField("iface", iface.Name).Debug("interface already has a TCX PCA egress hook ignore")
			} else {
				return fmt.Errorf("failed to attach PCA TCX egress: %w", err)
			}
		}
		p.egressTCXLink[iface] = egrLink
		ilog.WithField("interface", iface.Name).Debug("successfully attach PCA egressTCX hook")
	}

	if p.enableIngress {
		ingLink, err := link.AttachTCX(link.TCXOptions{
			Program:   p.objects.BpfPrograms.TcxIngressPcaParse,
			Attach:    cilium.AttachTCXIngress,
			Interface: iface.Index,
		})
		if err != nil {
			if errors.Is(err, fs.ErrExist) {
				// The interface already has a TCX ingress hook
				log.WithField("iface", iface.Name).Debug("interface already has a TCX PCA ingress hook ignore")
			} else {
				return fmt.Errorf("failed to attach PCA TCX ingress: %w", err)
			}
		}
		p.ingressTCXLink[iface] = ingLink
		ilog.WithField("interface", iface.Name).Debug("successfully attach PCA ingressTCX hook")
	}

	return nil
}

func fetchEgressEvents(iface ifaces.Interface, ipvlan netlink.Link, parser *cilium.Program, name string) (*netlink.BpfFilter, error) {
	ilog := plog.WithField("iface", iface)
	egressAttrs := netlink.FilterAttrs{
		LinkIndex: ipvlan.Attrs().Index,
		Parent:    netlink.HANDLE_MIN_EGRESS,
		Handle:    netlink.MakeHandle(0, 1),
		Protocol:  3,
		Priority:  1,
	}
	egressFilter := &netlink.BpfFilter{
		FilterAttrs:  egressAttrs,
		Fd:           parser.FD(),
		Name:         "tc/" + name,
		DirectAction: true,
	}
	if err := netlink.FilterDel(egressFilter); err == nil {
		ilog.Warn("egress filter already existed. Deleted it")
	}
	if err := netlink.FilterAdd(egressFilter); err != nil {
		if errors.Is(err, fs.ErrExist) {
			ilog.WithError(err).Warn("egress filter already exists. Ignoring")
		} else {
			return nil, fmt.Errorf("failed to create egress filter: %w", err)
		}
	}
	return egressFilter, nil

}

func (p *PacketFetcher) registerEgress(iface ifaces.Interface, ipvlan netlink.Link) error {
	egressFilter, err := fetchEgressEvents(iface, ipvlan, p.objects.TcEgressPcaParse, "tc_egress_pca_parse")
	if err != nil {
		return err
	}

	p.egressFilters[iface] = egressFilter
	return nil
}

func fetchIngressEvents(iface ifaces.Interface, ipvlan netlink.Link, parser *cilium.Program, name string) (*netlink.BpfFilter, error) {
	ilog := plog.WithField("iface", iface)
	ingressAttrs := netlink.FilterAttrs{
		LinkIndex: ipvlan.Attrs().Index,
		Parent:    netlink.HANDLE_MIN_INGRESS,
		Handle:    netlink.MakeHandle(0, 1),
		Protocol:  3,
		Priority:  1,
	}
	ingressFilter := &netlink.BpfFilter{
		FilterAttrs:  ingressAttrs,
		Fd:           parser.FD(),
		Name:         "tc/" + name,
		DirectAction: true,
	}
	if err := netlink.FilterDel(ingressFilter); err == nil {
		ilog.Warn("egress filter already existed. Deleted it")
	}
	if err := netlink.FilterAdd(ingressFilter); err != nil {
		if errors.Is(err, fs.ErrExist) {
			ilog.WithError(err).Warn("ingress filter already exists. Ignoring")
		} else {
			return nil, fmt.Errorf("failed to create egress filter: %w", err)
		}
	}
	return ingressFilter, nil

}

func (p *PacketFetcher) registerIngress(iface ifaces.Interface, ipvlan netlink.Link) error {
	ingressFilter, err := fetchIngressEvents(iface, ipvlan, p.objects.TcIngressPcaParse, "tc_ingress_pca_parse")
	if err != nil {
		return err
	}

	p.ingressFilters[iface] = ingressFilter
	return nil
}

// Close the eBPF fetcher from the system.
// We don't need an "Close(iface)" method because the filters and qdiscs
// are automatically removed when the interface is down
func (p *PacketFetcher) Close() error {
	log.Debug("unregistering eBPF objects")

	var errs []error
	if p.perfReader != nil {
		if err := p.perfReader.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if p.objects != nil {
		if err := p.objects.TcEgressPcaParse.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := p.objects.TcIngressPcaParse.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := p.objects.TcxEgressPcaParse.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := p.objects.TcxIngressPcaParse.Close(); err != nil {
			errs = append(errs, err)
		}
		if err := p.objects.PacketRecord.Close(); err != nil {
			errs = append(errs, err)
		}
		p.objects = nil
	}
	for iface, ef := range p.egressFilters {
		log.WithField("interface", iface).Debug("deleting egress filter")
		if err := netlink.FilterDel(ef); err != nil {
			errs = append(errs, fmt.Errorf("deleting egress filter: %w", err))
		}
	}
	p.egressFilters = map[ifaces.Interface]*netlink.BpfFilter{}
	for iface, igf := range p.ingressFilters {
		log.WithField("interface", iface).Debug("deleting ingress filter")
		if err := netlink.FilterDel(igf); err != nil {
			errs = append(errs, fmt.Errorf("deleting ingress filter: %w", err))
		}
	}
	p.ingressFilters = map[ifaces.Interface]*netlink.BpfFilter{}
	for iface, qd := range p.qdiscs {
		log.WithField("interface", iface).Debug("deleting Qdisc")
		if err := netlink.QdiscDel(qd); err != nil {
			errs = append(errs, fmt.Errorf("deleting qdisc: %w", err))
		}
	}
	p.qdiscs = map[ifaces.Interface]*netlink.GenericQdisc{}
	if len(errs) == 0 {
		return nil
	}

	for iface, l := range p.egressTCXLink {
		log := log.WithField("interface", iface)
		log.Debug("detach egress TCX hook")
		l.Close()

	}
	p.egressTCXLink = map[ifaces.Interface]link.Link{}
	for iface, l := range p.ingressTCXLink {
		log := log.WithField("interface", iface)
		log.Debug("detach ingress TCX hook")
		l.Close()
	}
	p.ingressTCXLink = map[ifaces.Interface]link.Link{}

	var errStrings []string
	for _, err := range errs {
		errStrings = append(errStrings, err.Error())
	}
	return errors.New(`errors: "` + strings.Join(errStrings, `", "`) + `"`)
}

func (p *PacketFetcher) ReadPerf() (perf.Record, error) {
	return p.perfReader.Read()
}

func (p *PacketFetcher) LookupAndDeleteMap(met *metrics.Metrics) map[int][]*byte {
	if !p.lookupAndDeleteSupported {
		return p.legacyLookupAndDeleteMap(met)
	}

	packetMap := p.objects.PacketRecord
	iterator := packetMap.Iterate()
	packets := make(map[int][]*byte, p.cacheMaxSize)
	var id int
	var ids []int
	var packet []*byte

	// First, get all ids and ignore content (we need lookup+delete to be atomic)
	for iterator.Next(&id, &packet) {
		ids = append(ids, id)
	}

	// Run the atomic Lookup+Delete; if new ids have been inserted in the meantime, they'll be fetched next time
	for i, id := range ids {
		if err := packetMap.LookupAndDelete(&id, &packet); err != nil {
			if i == 0 && errors.Is(err, cilium.ErrNotSupported) {
				log.WithError(err).Warnf("switching to legacy mode")
				p.lookupAndDeleteSupported = false
				return p.legacyLookupAndDeleteMap(met)
			}
			log.WithError(err).WithField("packetID", id).Warnf("couldn't delete entry")
			met.Errors.WithErrorName("pkt-fetcher", "CannotDeleteEntry").Inc()
		}
		packets[id] = packet
	}

	return packets
}
