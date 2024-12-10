package flow

import (
	"time"

	"github.com/netobserv/netobserv-ebpf-agent/pkg/model"
	"github.com/sirupsen/logrus"
)

var plog = logrus.WithField("component", "packet/PerfBuffer")

type PerfBuffer struct {
	maxEntries   int
	evictTimeout time.Duration
	entries      [](*model.PacketRecord)
}

func NewPerfBuffer(
	maxEntries int, evictTimeout time.Duration,
) *PerfBuffer {
	return &PerfBuffer{
		maxEntries:   maxEntries,
		evictTimeout: evictTimeout,
		entries:      []*model.PacketRecord{},
	}
}

func (c *PerfBuffer) PBuffer(in <-chan *model.PacketRecord, out chan<- []*model.PacketRecord) {
	evictTick := time.NewTicker(c.evictTimeout)
	defer evictTick.Stop()
	ind := 0
	for {
		select {
		case <-evictTick.C:
			if len(c.entries) == 0 {
				break
			}
			evictingEntries := c.entries
			c.entries = []*model.PacketRecord{}
			logrus.WithField("packets", len(evictingEntries)).
				Debug("evicting packets from userspace  on timeout")
			c.evict(evictingEntries, out)
		case packet, ok := <-in:
			if !ok {
				plog.Debug("input channel closed. Evicting entries")
				c.evict(c.entries, out)
				plog.Debug("exiting perfbuffer routine")
				return
			}
			if len(c.entries) >= c.maxEntries {
				evictingEntries := c.entries
				c.entries = []*model.PacketRecord{}
				logrus.WithField("packets", len(evictingEntries)).
					Debug("evicting packets from userspace accounter after reaching cache max length")
				c.evict(evictingEntries, out)
			}
			c.entries = append(c.entries, model.NewPacketRecord(packet.Stream, (uint32)(len(packet.Stream)), packet.Time))
			ind++
		}
	}
}

func (c *PerfBuffer) evict(entries [](*model.PacketRecord), evictor chan<- []*model.PacketRecord) {
	packets := make([]*model.PacketRecord, 0, len(entries))
	for _, payload := range entries {
		packets = append(packets, model.NewPacketRecord(payload.Stream, (uint32)(len(payload.Stream)), payload.Time))
	}
	alog.WithField("numEntries", len(packets)).Debug("packets evicted from userspace accounter")
	evictor <- packets
}
