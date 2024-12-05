package ifaces

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
)

// Poller periodically looks for the network interfaces in the system and forwards Event
// notifications when interfaces are added or deleted.
type Poller struct {
	period     time.Duration
	current    map[Interface]struct{}
	interfaces func(handle netns.NsHandle) ([]Interface, error)
	bufLen     int
}

func NewPoller(period time.Duration, bufLen int) *Poller {
	return &Poller{
		period:     period,
		bufLen:     bufLen,
		interfaces: netInterfaces,
		current:    map[Interface]struct{}{},
	}
}

func (np *Poller) Subscribe(ctx context.Context) (<-chan Event, error) {

	out := make(chan Event, np.bufLen)
	netns, err := getNetNS()
	if err != nil {
		go np.pollForEvents(ctx, "", out)
	} else {
		for _, n := range netns {
			go np.pollForEvents(ctx, n, out)
		}
	}
	return out, nil
}

func (np *Poller) pollForEvents(ctx context.Context, ns string, out chan Event) {
	log := logrus.WithField("component", "ifaces.Poller")
	log.WithField("period", np.period).Debug("subscribing to Interface events")
	ticker := time.NewTicker(np.period)
	var netnsHandle netns.NsHandle
	var err error

	if ns == "" {
		netnsHandle = netns.None()
	} else {
		netnsHandle, err = netns.GetFromName(ns)
		if err != nil {
			return
		}
	}

	defer ticker.Stop()
	for {
		if ifaces, err := np.interfaces(netnsHandle); err != nil {
			log.WithError(err).Warn("fetching interface names")
		} else {
			log.WithField("names", ifaces).Debug("fetched interface names")
			np.diffNames(out, ifaces)
		}
		select {
		case <-ctx.Done():
			log.Debug("stopped")
			close(out)
			return
		case <-ticker.C:
			// continue after a period
		}
	}
}

// diffNames compares and updates the internal account of interfaces with the latest list of
// polled interfaces. It forwards Events for any detected addition or removal of interfaces.
func (np *Poller) diffNames(events chan Event, ifaces []Interface) {
	// Check for new interfaces
	acquired := map[Interface]struct{}{}
	for _, iface := range ifaces {
		acquired[iface] = struct{}{}
		if _, ok := np.current[iface]; !ok {
			ilog.WithField("interface", iface).Debug("added network interface")
			np.current[iface] = struct{}{}
			events <- Event{
				Type:      EventAdded,
				Interface: iface,
			}
		}
	}
	// Check for deleted interfaces
	for iface := range np.current {
		if _, ok := acquired[iface]; !ok {
			delete(np.current, iface)
			ilog.WithField("interface", iface).Debug("deleted network interface")
			events <- Event{
				Type:      EventDeleted,
				Interface: iface,
			}
		}
	}
}
