package cleanup

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	// Add to this list any object that we used to generate in past versions, and stopped doing so.
	// For instance, any object that was renamed between two releases of the operator and that could therefore
	// remain on the cluster after an upgrade.
	cleanupList = []cleanupItem{
		{
			ref:         client.ObjectKey{Name: "grafana-dashboard-netobserv", Namespace: "openshift-config-managed"},
			placeholder: &corev1.ConfigMap{},
		},
	}
	// Need to run only once, at operator startup, this is not part of the reconcile loop
	didRun = false
)

type cleanupItem struct {
	ref         client.ObjectKey
	placeholder client.Object
}

func CleanPastReferences(ctx context.Context, cl client.Client, defaultNamespace string) error {
	if didRun {
		return nil
	}
	log := log.FromContext(ctx)
	log.Info("Check and clean old objects")
	// Search for all past references to clean up. If one is found, delete it.
	for _, item := range cleanupList {
		if item.ref.Namespace == "" {
			item.ref.Namespace = defaultNamespace
		}
		if err := cl.Get(ctx, item.ref, item.placeholder); err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return err
		}
		log.
			WithValues("name", item.ref.Name).
			WithValues("namespace", item.ref.Namespace).
			Info("Deleting old object")
		if err := cl.Delete(ctx, item.placeholder); err != nil {
			return err
		}
	}
	didRun = true
	return nil
}
