package flow

import (
	"container/list"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/netobserv/netobserv-ebpf-agent/pkg/ebpf"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/metrics"
	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"
)

var dlog = logrus.WithField("component", "flow/Deduper")
var timeNow = time.Now

// deduperCache implement a LRU cache whose elements are evicted if they haven't been accessed
// during the expire duration.
// It is not safe for concurrent access.
type deduperCache struct {
	expire time.Duration
	// key: ebpf.BpfFlowId with the interface and MACs erased, to detect duplicates
	// value: listElement pointing to a struct entry
	ifaces map[ebpf.BpfFlowId]*list.Element
	// element: entry structs of the ifaces map ordered by expiry time
	entries *list.List
}

type entry struct {
	key           *ebpf.BpfFlowId
	dnsRecord     *ebpf.BpfDnsRecordT
	flowRTT       *uint64
	networkEvents *[4][8]uint8
	ifIndex       uint32
	expiryTime    time.Time
	dupList       *[]map[string]uint8
}

// Dedupe receives flows and filters these belonging to duplicate interfaces. It will forward
// the flows from the first interface coming to it, until that flow expires in the cache
// (no activity for it during the expiration time)
// The justMark argument tells that the deduper should not drop the duplicate flows but
// set their Duplicate field.
func Dedupe(expireTime time.Duration, justMark, mergeDup bool, ifaceNamer InterfaceNamer, m *metrics.Metrics) func(in <-chan []*model.Record, out chan<- []*model.Record) {
	cache := &deduperCache{
		expire:  expireTime,
		entries: list.New(),
		ifaces:  map[ebpf.BpfFlowId]*list.Element{},
	}
	return func(in <-chan []*model.Record, out chan<- []*model.Record) {
		for records := range in {
			cache.removeExpired()
			fwd := make([]*model.Record, 0, len(records))
			for _, record := range records {
				cache.checkDupe(record, justMark, mergeDup, &fwd, ifaceNamer)
			}
			if len(fwd) > 0 {
				out <- fwd
				m.EvictionCounter.WithSource("deduper").Inc()
				m.EvictedFlowsCounter.WithSource("deduper").Add(float64(len(fwd)))
			}
			m.BufferSizeGauge.WithBufferName("deduper-list").Set(float64(cache.entries.Len()))
			m.BufferSizeGauge.WithBufferName("deduper-map").Set(float64(len(cache.ifaces)))
		}
	}
}

// checkDupe check current record if its already available nad if not added to fwd records list
func (c *deduperCache) checkDupe(r *model.Record, justMark, mergeDup bool, fwd *[]*model.Record, ifaceNamer InterfaceNamer) {
	mergeEntry := make(map[string]uint8)
	rk := r.Id
	// zeroes fields from key that should be ignored from the flow comparison
	rk.IfIndex = 0
	rk.SrcMac = [model.MacLen]uint8{0, 0, 0, 0, 0, 0}
	rk.DstMac = [model.MacLen]uint8{0, 0, 0, 0, 0, 0}
	rk.Direction = 0
	// If a flow has been accounted previously, whatever its interface was,
	// it updates the expiry time for that flow
	if ele, ok := c.ifaces[rk]; ok {
		fEntry := ele.Value.(*entry)
		fEntry.expiryTime = timeNow().Add(c.expire)
		c.entries.MoveToFront(ele)
		// The input flow is duplicate if its interface is different to the interface
		// of the non-duplicate flow that was first registered in the cache
		// except if the new flow has DNS enrichment in this case will enrich the flow in the cache
		// with DNS info and mark the current flow as duplicate
		if r.Metrics.DnsRecord.Latency != 0 && fEntry.dnsRecord.Latency == 0 {
			// copy DNS record to the cached entry and mark it as duplicate
			fEntry.dnsRecord.Flags = r.Metrics.DnsRecord.Flags
			fEntry.dnsRecord.Id = r.Metrics.DnsRecord.Id
			fEntry.dnsRecord.Latency = r.Metrics.DnsRecord.Latency
			fEntry.dnsRecord.Errno = r.Metrics.DnsRecord.Errno
		}
		// If the new flow has flowRTT then enrich the flow in the case with the same RTT and mark it duplicate
		if r.Metrics.FlowRtt != 0 && *fEntry.flowRTT == 0 {
			*fEntry.flowRTT = r.Metrics.FlowRtt
		}
		// If the new flows have network events, then enrich the flow in the cache and mark the flow as duplicate
		for i, md := range r.Metrics.NetworkEvents {
			if !model.AllZerosMetaData(md) && model.AllZerosMetaData(fEntry.networkEvents[i]) {
				copy(fEntry.networkEvents[i][:], md[:])
			}
		}
		if fEntry.ifIndex != r.Id.IfIndex {
			if justMark {
				r.Duplicate = true
				*fwd = append(*fwd, r)
			}
			if mergeDup {
				ifName := ifaceNamer(int(r.Id.IfIndex))
				mergeEntry[ifName] = r.Id.Direction
				if dupEntryNew(*fEntry.dupList, mergeEntry) {
					*fEntry.dupList = append(*fEntry.dupList, mergeEntry)
					dlog.Debugf("merge list entries dump:")
					for _, entry := range *fEntry.dupList {
						for k, v := range entry {
							dlog.Debugf("interface %s dir %d", k, v)
						}
					}
				}
			}
			return
		}
		*fwd = append(*fwd, r)
		return
	}
	// The flow has not been accounted previously (or was forgotten after expiration)
	// so we register it for that concrete interface
	e := entry{
		key:           &rk,
		dnsRecord:     &r.Metrics.DnsRecord,
		flowRTT:       &r.Metrics.FlowRtt,
		networkEvents: &r.Metrics.NetworkEvents,
		ifIndex:       r.Id.IfIndex,
		expiryTime:    timeNow().Add(c.expire),
	}
	if mergeDup {
		ifName := ifaceNamer(int(r.Id.IfIndex))
		mergeEntry[ifName] = r.Id.Direction
		r.DupList = append(r.DupList, mergeEntry)
		e.dupList = &r.DupList
	}
	c.ifaces[rk] = c.entries.PushFront(&e)
	*fwd = append(*fwd, r)
}

func dupEntryNew(dupList []map[string]uint8, mergeEntry map[string]uint8) bool {
	for _, entry := range dupList {
		if reflect.DeepEqual(entry, mergeEntry) {
			return false
		}
	}
	return true
}

func (c *deduperCache) removeExpired() {
	now := timeNow()
	ele := c.entries.Back()
	evicted := 0
	for ele != nil && now.After(ele.Value.(*entry).expiryTime) {
		evicted++
		c.entries.Remove(ele)
		fEntry := ele.Value.(*entry)
		fEntry.dupList = nil
		delete(c.ifaces, *fEntry.key)
		ele = c.entries.Back()
	}
	if evicted > 0 {
		dlog.WithFields(logrus.Fields{
			"current":    c.entries.Len(),
			"evicted":    evicted,
			"expiryTime": c.expire,
		}).Debug("entries evicted from the deduper cache")
	}
}
