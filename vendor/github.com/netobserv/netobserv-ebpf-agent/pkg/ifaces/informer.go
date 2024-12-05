package ifaces

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// EventType for an interface: added, deleted
type EventType int

const (
	EventAdded EventType = iota
	EventDeleted
)

func (e EventType) String() string {
	switch e {
	case EventAdded:
		return "Added"
	case EventDeleted:
		return "Deleted"
	default:
		return fmt.Sprintf("Unknown (%d)", e)
	}
}

var ilog = logrus.WithField("component", "ifaces.Informer")

// Event of a network interface, given the type (added, removed) and the interface name
type Event struct {
	Type      EventType
	Interface Interface
}

type Interface struct {
	Name  string
	Index int
	NetNS netns.NsHandle
}

// Informer provides notifications about each network interface that is added or removed
// from the host. Production implementations: Poller and Watcher.
type Informer interface {
	// Subscribe returns a channel that sends Event instances.
	Subscribe(ctx context.Context) (<-chan Event, error)
}

func netInterfaces(nsh netns.NsHandle) ([]Interface, error) {
	handle, err := netlink.NewHandleAt(nsh)
	if err != nil {
		return nil, fmt.Errorf("failed to create handle for netns (%s): %w", nsh.String(), err)
	}
	defer handle.Close()

	// Get a list of interfaces in the namespace
	links, err := handle.LinkList()
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces in netns (%s): %w", nsh.String(), err)
	}

	names := make([]Interface, len(links))
	for i, link := range links {
		names[i] = Interface{Name: link.Attrs().Name, Index: link.Attrs().Index, NetNS: nsh}
	}
	return names, nil
}
