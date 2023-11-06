package narrowcache

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Client struct {
	client.Client
	liveClient     kubernetes.Interface
	watchedGVKs    map[string]GVKInfo
	watchedObjects map[string]*watchedObject
	wmut           sync.RWMutex
}

type watchedObject struct {
	cached   client.Object
	handlers []handlerOnQueue
}

type handlerOnQueue struct {
	handler handler.EventHandler
	queue   workqueue.RateLimitingInterface
}

func (c *Client) Get(ctx context.Context, key client.ObjectKey, out client.Object, opts ...client.GetOption) error {
	rlog := log.FromContext(ctx).WithName("narrowcache")
	rlog.WithValues("key", key).Info("Getting object:")
	gvk, err := c.GroupVersionKindFor(out)
	if err != nil {
		return err
	}
	strGVK := gvk.String()
	if info, managed := c.watchedGVKs[strGVK]; managed {
		// Kind is managed by this cache layer => check for watch
		obj, _, err := c.getAndCreateWatchIfNeeded(ctx, info, gvk, key)
		if err != nil {
			return err
		}
		err = copyInto(obj, out)
		if err != nil {
			return err
		}
		return nil
	}

	rlog.WithValues("GVK", strGVK).Info("GVK not managed here: pass it to down layer")
	return c.Client.Get(ctx, key, out, opts...)
}

func (c *Client) getAndCreateWatchIfNeeded(ctx context.Context, info GVKInfo, gvk schema.GroupVersionKind, key client.ObjectKey) (client.Object, string, error) {
	strGVK := gvk.String()
	objKey := strGVK + "|" + key.String()
	rlog := log.FromContext(ctx).WithName("narrowcache").WithValues("objKey", objKey)

	rlog.Info("Managed kind: check cache")
	c.wmut.RLock()
	ca := c.watchedObjects[objKey]
	c.wmut.RUnlock()
	if ca != nil {
		if ca.cached == nil {
			return nil, objKey, errors.NewNotFound(schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind}, key.Name)
		}
		// Return from cache
		rlog.Info("Object found in cache, returning")
		return ca.cached, objKey, nil
	}

	// Live query
	rlog.Info("Object not found in cache, doing live query")
	fetched, err := info.Getter(ctx, c.liveClient, key)
	if err != nil {
		return nil, objKey, err
	}
	rlog.Info("Object fetched")

	// Create watch for later calls
	w, err := info.Watcher(ctx, c.liveClient, key)
	if err != nil {
		return nil, objKey, err
	}

	// Store fetched object
	err = c.setToCache(objKey, fetched)
	if err != nil {
		return nil, objKey, err
	}

	// Start updating goroutine
	go c.updateCache(ctx, objKey, w)

	return fetched.(client.Object), objKey, nil
}

// "Terrible hack" cc directxman12 / sigs.k8s.io/controller-runtime/pkg/cache/internal/cache_reader.go
func copyInto(obj runtime.Object, out client.Object) error {
	// Copy the value of the item in the cache to the returned value
	// TODO(directxman12): this is a terrible hack, pls fix (we should have deepcopyinto)
	outVal := reflect.ValueOf(out)
	objVal := reflect.ValueOf(obj)
	if !objVal.Type().AssignableTo(outVal.Type()) {
		return fmt.Errorf("cache had type %s, but %s was asked for", objVal.Type(), outVal.Type())
	}
	reflect.Indirect(outVal).Set(reflect.Indirect(objVal))
	// if !c.disableDeepCopy {
	// 	out.GetObjectKind().SetGroupVersionKind(c.groupVersionKind)
	// }
	return nil
}

func (c *Client) updateCache(ctx context.Context, key string, watcher watch.Interface) {
	rlog := log.FromContext(ctx).WithName("narrowcache")
	for watchEvent := range watcher.ResultChan() {
		rlog.WithValues("key", key, "event type", watchEvent.Type).Info("Event received")
		if watchEvent.Type == watch.Added || watchEvent.Type == watch.Modified {
			err := c.setToCache(key, watchEvent.Object)
			if err != nil {
				rlog.WithValues("key", key).Error(err, "Error while updating cache")
			}
		} else if watchEvent.Type == watch.Deleted {
			c.removeFromCache(key)
		}
		c.callHandlers(ctx, key, watchEvent)
	}
}

func (c *Client) setToCache(key string, obj runtime.Object) error {
	cObj, ok := obj.(client.Object)
	if !ok {
		return fmt.Errorf("could not convert runtime.Object to client.Object")
	}
	c.wmut.Lock()
	if ca := c.watchedObjects[key]; ca != nil {
		ca.cached = cObj
	} else {
		c.watchedObjects[key] = &watchedObject{cached: cObj}
	}
	c.wmut.Unlock()
	return nil
}

func (c *Client) removeFromCache(key string) {
	c.wmut.Lock()
	if ca := c.watchedObjects[key]; ca != nil {
		ca.cached = nil
	}
	c.wmut.Unlock()
}

func (c *Client) addHandler(key string, hoq handlerOnQueue) {
	c.wmut.Lock()
	if ca := c.watchedObjects[key]; ca != nil {
		ca.handlers = append(ca.handlers, hoq)
	}
	c.wmut.Unlock()
}

func (c *Client) callHandlers(ctx context.Context, key string, ev watch.Event) {
	var fn func(hoq handlerOnQueue)
	switch ev.Type {
	case watch.Added:
		fn = func(hoq handlerOnQueue) {
			createEvent := event.CreateEvent{Object: ev.Object.(client.Object)}
			hoq.handler.Create(ctx, createEvent, hoq.queue)
		}
	case watch.Modified:
		fn = func(hoq handlerOnQueue) {
			// old object unknown (not an issue for us - we just enqueue reconcile requests)
			modEvent := event.UpdateEvent{ObjectOld: ev.Object.(client.Object), ObjectNew: ev.Object.(client.Object)}
			hoq.handler.Update(ctx, modEvent, hoq.queue)
		}
	case watch.Deleted:
		fn = func(hoq handlerOnQueue) {
			delEvent := event.DeleteEvent{Object: ev.Object.(client.Object)}
			hoq.handler.Delete(ctx, delEvent, hoq.queue)
		}
	case watch.Bookmark:
	case watch.Error:
		// Not managed
	}
	if fn == nil {
		return
	}
	c.wmut.RLock()
	if ca := c.watchedObjects[key]; ca != nil {
		for _, hoq := range ca.handlers {
			fn(hoq)
		}
	}
	c.wmut.RUnlock()
}

func (c *Client) GetSource(ctx context.Context, obj client.Object) (source.Source, error) {
	// Prepare a Source and make sure it is associated with a watch
	rlog := log.FromContext(ctx).WithName("narrowcache")
	rlog.WithValues("name", obj.GetName(), "namespace", obj.GetNamespace()).Info("Getting Source:")
	gvk, err := c.GroupVersionKindFor(obj)
	if err != nil {
		return nil, err
	}
	strGVK := gvk.String()
	info, managed := c.watchedGVKs[strGVK]
	if !managed {
		return nil, fmt.Errorf("called 'GetSource' on unmanaged GVK: %s", strGVK)
	}

	_, key, err := c.getAndCreateWatchIfNeeded(ctx, info, gvk, client.ObjectKeyFromObject(obj))
	if err != nil {
		return nil, err
	}

	return &NarrowSource{
		onStart: func(ctx context.Context, h handler.EventHandler, q workqueue.RateLimitingInterface) {
			c.addHandler(key, handlerOnQueue{handler: h, queue: q})
		},
	}, nil
}
