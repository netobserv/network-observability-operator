package flowcollector

import (
	"github.com/netobserv/network-observability-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

// TODO: make configurable
const (
	eventsChBuffer = 5
informerResync = 10*time.Minute
)

const(
	resourceName = "flowcollector"
)

type eventHandler struct {
	events chan<- Event
}

func (e *eventHandler) OnAdd(obj interface{}) {
	panic("implement me")
}

func (e *eventHandler) OnUpdate(oldObj, newObj interface{}) {
	panic("implement me")
}

func (e *eventHandler) OnDelete(obj interface{}) {
	panic("implement me")
}

// Informer starts the FlowCollector's Informer in background and sends any
// update through the returned channel
func Informer(client kubernetes.Interface) (<-chan Event, error) {
	informer, err := informers.NewSharedInformerFactory(client, informerResync).
		ForResource(schema.GroupVersionResource{
			Group: v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Resource: resourceName,
		})
	if err != nil {
		return nil, err
	}
	events := make(chan Event, eventsChBuffer)
	go watch(informer, events)
	return events, nil
}

func watch(informer informers.GenericInformer, events chan<- Event) {
	informer.Informer().AddEventHandler()
}

func test() {
	var client kubernetes.Interface

	//gi.Informer().AddEventHandler()
}
