package reconcilers

import (
	"context"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NamespacedObjectManager provides some helpers to manage (fetch, delete) namespace-scoped objects
type NamespacedObjectManager struct {
	client            client.Client
	Namespace         string
	PreviousNamespace string
	managedObjects    []managedObject
}

type managedObject struct {
	name        string
	kind        string
	placeholder client.Object
	found       bool
}

func NewNamespacedObjectManager(cmn *Common) *NamespacedObjectManager {
	return &NamespacedObjectManager{
		client:            cmn.Client,
		Namespace:         cmn.Namespace,
		PreviousNamespace: cmn.PreviousNamespace,
	}
}

// AddManagedObject should be used to register managed objects to be fetched by FetchAll, or deleted when namespace changes
// This is only for namespace-scoped objects that are installed in the desired namespace (in FlowCollector CRD: spec.namespace)
// Cluster-scope objects, or objects installed in a different namespace (e.g. OVS configmap) should not be registered with this function.
func (m *NamespacedObjectManager) AddManagedObject(name string, placeholder client.Object) {
	m.managedObjects = append(m.managedObjects, managedObject{
		name:        name,
		kind:        reflect.TypeOf(placeholder).String(),
		placeholder: placeholder,
	})
}

// FetchAll fetches all managed objects (registered using AddManagedObject) in the current namespace.
// Placeholders are filled with fetched resources. Resources not found are flagged internally.
func (m *NamespacedObjectManager) FetchAll(ctx context.Context) error {
	log := log.FromContext(ctx)
	fetched := []string{}
	notFound := []string{}
	for i, ref := range m.managedObjects {
		m.managedObjects[i].found = false
		objLog := ref.kind + "/" + ref.name
		err := m.client.Get(ctx, types.NamespacedName{Name: ref.name, Namespace: m.Namespace}, ref.placeholder)
		if err != nil {
			if errors.IsNotFound(err) {
				notFound = append(notFound, objLog)
			} else {
				log.Error(err, "Failed to get "+objLog)
				return err
			}
		} else {
			fetched = append(fetched, objLog)
			m.managedObjects[i].found = true
			// On success, placeholder is filled with resource. Caller should keep a pointer to it.
		}
	}
	if len(fetched) > 0 {
		log.Info("FETCHED: " + strings.Join(fetched, ","))
	}
	if len(notFound) > 0 {
		log.Info("(Items not deployed: " + strings.Join(notFound, ",") + ")")
	}
	return nil
}

// CleanupPreviousNamespace removes all managed objects (registered using AddManagedObject) from the previous namespace.
func (m *NamespacedObjectManager) CleanupPreviousNamespace(ctx context.Context) {
	m.cleanup(ctx, m.PreviousNamespace)
}

func (m *NamespacedObjectManager) cleanup(ctx context.Context, namespace string) {
	log := log.FromContext(ctx)
	for _, obj := range m.managedObjects {
		ref := obj.placeholder.DeepCopyObject().(client.Object)
		ref.SetName(obj.name)
		ref.SetNamespace(namespace)
		log.Info("DELETING "+obj.kind, "Namespace", namespace, "Name", obj.name)
		err := m.client.Delete(ctx, ref)
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "Failed to delete old "+obj.kind, "Namespace", namespace, "Name", obj.name)
		}
	}
}

// TryDeleteAll is an helper function that tries to delete all managed objects previously loaded using FetchAll.
func (m *NamespacedObjectManager) TryDeleteAll(ctx context.Context) {
	for _, obj := range m.managedObjects {
		m.TryDelete(ctx, obj.placeholder)
	}
}

// TryDelete is an helper function that tries to delete the provided object previously loaded using FetchAll.
func (m *NamespacedObjectManager) TryDelete(ctx context.Context, obj client.Object) {
	if m.Exists(obj) {
		log := log.FromContext(ctx)
		kind := reflect.TypeOf(obj).String()
		log.Info("DELETING "+kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		err := m.client.Delete(ctx, obj)
		if err != nil {
			log.Error(err, "Failed to delete old "+kind, "Namespace", obj.GetNamespace(), "Name", obj.GetName())
		}
	}
}

// Exists returns true if the provided object isn't nil and was successfully fetched previously with FetchAll
func (m *NamespacedObjectManager) Exists(obj client.Object) bool {
	if obj == nil {
		return false
	}
	for _, managed := range m.managedObjects {
		if obj == managed.placeholder {
			return managed.found
		}
	}
	return false
}
