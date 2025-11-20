package networkpolicy

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	kubernetesServiceName      = "kubernetes"
	kubernetesServiceNamespace = "default"
)

// GetAPIServerEndpointIPs retrieves the API server endpoint IP addresses.
// It first tries to use EndpointSlice API (v1), and falls back to Endpoints API if unavailable.
func GetAPIServerEndpointIPs(ctx context.Context, cl client.Client) ([]string, error) {
	logger := log.FromContext(ctx)

	// Try EndpointSlice first (discovery.k8s.io/v1, available since k8s 1.21)
	ips, err := getEndpointIPsFromEndpointSlice(ctx, cl)
	if err == nil && len(ips) > 0 {
		logger.V(1).Info("Retrieved API server endpoint IPs from EndpointSlice", "ips", ips)
		return ips, nil
	}

	if err != nil {
		logger.V(1).Info("Failed to get EndpointSlice, falling back to Endpoints API", "error", err)
	}

	// Fallback to Endpoints API (core/v1, deprecated but widely available)
	ips, err = getEndpointIPsFromEndpoints(ctx, cl)
	if err != nil {
		return nil, fmt.Errorf("failed to get API server endpoint IPs: %w", err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no API server endpoint IPs found")
	}

	logger.V(1).Info("Retrieved API server endpoint IPs from Endpoints", "ips", ips)
	return ips, nil
}

// getEndpointIPsFromEndpointSlice retrieves endpoint IPs using the EndpointSlice API
func getEndpointIPsFromEndpointSlice(ctx context.Context, cl client.Client) ([]string, error) {
	// List all EndpointSlices for the kubernetes service
	endpointSliceList := &discoveryv1.EndpointSliceList{}
	err := cl.List(ctx, endpointSliceList, client.InNamespace(kubernetesServiceNamespace), client.MatchingLabels{
		discoveryv1.LabelServiceName: kubernetesServiceName,
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("EndpointSlice for kubernetes service not found")
		}
		return nil, err
	}

	if len(endpointSliceList.Items) == 0 {
		return nil, fmt.Errorf("no EndpointSlices found for kubernetes service")
	}

	var ips []string
	for i := range endpointSliceList.Items {
		for j := range endpointSliceList.Items[i].Endpoints {
			endpoint := &endpointSliceList.Items[i].Endpoints[j]
			// Only use ready endpoints
			if endpoint.Conditions.Ready != nil && *endpoint.Conditions.Ready {
				ips = append(ips, endpoint.Addresses...)
			}
		}
	}

	return ips, nil
}

// getEndpointIPsFromEndpoints retrieves endpoint IPs using the legacy Endpoints API
func getEndpointIPsFromEndpoints(ctx context.Context, cl client.Client) ([]string, error) {
	//nolint:staticcheck // SA1019: Endpoints is deprecated but used as fallback for k8s < 1.21
	endpoints := &corev1.Endpoints{}
	err := cl.Get(ctx, types.NamespacedName{
		Name:      kubernetesServiceName,
		Namespace: kubernetesServiceNamespace,
	}, endpoints)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("endpoints for kubernetes service not found")
		}
		return nil, err
	}

	var ips []string
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			ips = append(ips, address.IP)
		}
	}

	return ips, nil
}
