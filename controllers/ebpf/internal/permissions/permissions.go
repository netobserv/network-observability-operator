package permissions

import (
	"context"
	"fmt"
	"reflect"

	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	osv1 "github.com/openshift/api/security/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Vendor int

const (
	VendorUnknown = iota
	VendorVanilla
	VendorOpenShift
)

// Controller reconciles the different resources to enable the privileged operation of the
// Netobserv Agent:
// - Create the privileged namespace with Pod Security annotations
// - Create netobserv-agent service account in the non-privileged namespace
// - For Openshift, apply the required SecurityContextConstraints for privileged Pod operation
type Controller struct {
	Client        reconcilers.ClientHelper
	BaseNamespace string
	// vendor defaults to VendorUnknown and it will be fetched when needed
	vendor Vendor
}

func (c *Controller) Reconcile(ctx context.Context) error {
	log.IntoContext(ctx, log.FromContext(ctx).WithName("permissions"))

	if err := c.reconcileNamespace(ctx); err != nil {
		return fmt.Errorf("reconciling namespace: %w", err)
	}
	if err := c.reconcileServiceAccount(ctx); err != nil {
		return fmt.Errorf("reconciling service account: %w", err)
	}
	if err := c.reconcileVendorPermissions(ctx); err != nil {
		return fmt.Errorf("reconciling vendor permissions: %w", err)
	}
	return nil
}

func (c *Controller) reconcileNamespace(ctx context.Context) error {
	namespace := c.privilegedNamespace()
	rlog := log.FromContext(ctx, "privilegedNamespace", namespace)
	actual := &v1.Namespace{}
	if err := c.Client.Get(ctx, client.ObjectKey{Name: namespace}, actual); err != nil {
		if errors.IsNotFound(err) {
			actual = nil
		} else {
			return fmt.Errorf("can't retrieve current namespace: %w", err)
		}
	}
	desired := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"app":                                "network-observability-operator",
				"pod-security.kubernetes.io/enforce": "privileged",
				"pod-security.kubernetes.io/audit":   "privileged",
			},
		},
	}
	if actual == nil && desired != nil {
		rlog.Info("creating namespace")
		return c.Client.CreateOwned(ctx, desired)
	}
	if actual != nil && desired != nil {
		if !helper.IsSubSet(actual.ObjectMeta.Labels, desired.ObjectMeta.Labels) {
			rlog.Info("updating namespace")
			return c.Client.UpdateOwned(ctx, actual, desired)
		}
	}
	rlog.Info("namespace is already reconciled. Doing nothing")
	return nil
}

func (c *Controller) reconcileServiceAccount(ctx context.Context) error {
	rlog := log.FromContext(ctx, "serviceAccount", constants.EBPFServiceAccount)

	sAcc := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.EBPFServiceAccount,
			Namespace: c.BaseNamespace,
		},
	}
	actual := &v1.ServiceAccount{}
	if err := c.Client.Get(ctx, client.ObjectKeyFromObject(sAcc), actual); err != nil {
		if errors.IsNotFound(err) {
			actual = nil
		} else {
			return fmt.Errorf("can't retrieve current namespace: %w", err)
		}
	}
	if actual == nil {
		rlog.Info("creating service account")
		return c.Client.CreateOwned(ctx, sAcc)
	}
	rlog.Info("service account already reconciled. Doing nothing")
	return nil
}

// reconcileVendorPermissions inspects into the API services to know which mechanism should use to
// reconcile extra permissions (e.g. SecurityContextConstraints in OpenShift)
func (c *Controller) reconcileVendorPermissions(ctx context.Context) error {
	if c.fetchVendor(ctx) == VendorOpenShift {
		return c.reconcileOpenshiftPermissions(ctx)
	}
	return nil
}

func (c *Controller) fetchVendor(ctx context.Context) Vendor {
	if c.vendor != VendorUnknown {
		return c.vendor
	}
	services := apiregv1.APIServiceList{}
	rlog := log.FromContext(ctx)
	if err := c.Client.List(ctx, &services); err != nil {
		rlog.Error(err, "fetching vendor: couldn't retrieve API services")
		return VendorUnknown
	}
	for i := range services.Items {
		if services.Items[i].Name == "v1.security.openshift.io" {
			rlog.Info("fetching vendor: found OpenShift")
			c.vendor = VendorOpenShift
			return c.vendor
		}
	}
	rlog.Info("fetching vendor: any of our registered vendors matched. " +
		"Assuming vanilla Kubernetes")
	c.vendor = VendorVanilla
	return c.vendor
}

func (c *Controller) reconcileOpenshiftPermissions(ctx context.Context) error {
	rlog := log.FromContext(ctx,
		"securityContextConstraints", constants.EBPFSecurityContext)
	scc := &osv1.SecurityContextConstraints{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.EBPFSecurityContext,
		},
		// TODO: replace by individual capabilities
		AllowPrivilegedContainer: true,
		AllowHostNetwork:         true,
		RunAsUser: osv1.RunAsUserStrategyOptions{
			Type: osv1.RunAsUserStrategyRunAsAny,
		},
		SELinuxContext: osv1.SELinuxContextStrategyOptions{
			Type: osv1.SELinuxStrategyRunAsAny,
		},
		Users: []string{
			"system:serviceaccount:" + c.BaseNamespace + ":" + constants.EBPFServiceAccount,
		},
	}
	actual := &osv1.SecurityContextConstraints{}
	if err := c.Client.Get(ctx, client.ObjectKeyFromObject(scc), actual); err != nil {
		if errors.IsNotFound(err) {
			actual = nil
		} else {
			return fmt.Errorf("can't retrieve current namespace: %w", err)
		}
	}
	if actual == nil {
		rlog.Info("creating SecurityContextConstraints")
		return c.Client.CreateOwned(ctx, scc)
	}
	if scc.AllowPrivilegedContainer != actual.AllowPrivilegedContainer ||
		scc.AllowHostNetwork != actual.AllowHostNetwork ||
		scc.RunAsUser != actual.RunAsUser ||
		scc.SELinuxContext != actual.SELinuxContext ||
		!reflect.DeepEqual(scc.Users, actual.Users) {
		rlog.Info("updating SecurityContextConstraints")
		return c.Client.UpdateOwned(ctx, actual, scc)
	}
	rlog.Info("securitycontextconstraints already reconciled. Doing nothing")
	return nil
}

func (c *Controller) privilegedNamespace() string {
	return c.BaseNamespace + constants.EBPFPrivilegedNSSuffix
}
