package tlsconfig

import (
	"context"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FetchAPIServerTLSProfile retrieves the TLS security profile from the OpenShift API Server
func FetchAPIServerTLSProfile(ctx context.Context, cfg *rest.Config) (*configv1.TLSSecurityProfile, error) {
	if cfg == nil {
		return nil, fmt.Errorf("rest config is nil")
	}

	// Create a non-cached client for initial API server queries
	scheme, err := getScheme()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheme: %w", err)
	}

	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Fetch the APIServer configuration
	apiServer := &configv1.APIServer{}
	if err := c.Get(ctx, client.ObjectKey{Name: "cluster"}, apiServer); err != nil {
		if k8serrors.IsNotFound(err) {
			// APIServer object not present (e.g. test environments) — treat as no profile configured
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get APIServer config: %w", err)
	}

	// Return the TLS security profile (may be nil if not set)
	return apiServer.Spec.TLSSecurityProfile, nil
}

// getScheme creates a runtime scheme with necessary types registered
func getScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := configv1.Install(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}
