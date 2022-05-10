package permissions

import (
	"context"
	"fmt"

	"github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/discover"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	osv1 "github.com/openshift/api/security/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var AllowedCapabilities = []v1.Capability{"BPF", "PERFMON", "NET_ADMIN", "SYS_RESOURCE"}

// Reconciler reconciles the different resources to enable the privileged operation of the
// Netobserv Agent:
// - Create the privileged namespace with Pod Permissions annotations (for Vanilla K8s)
// - Create netobserv-ebpf-agent service account in the privileged namespace
// - For Openshift, apply the required SecurityContextConstraints for privileged Pod operation
type Reconciler struct {
	client              reconcilers.ClientHelper
	privilegedNamespace string
	vendor              *discover.Permissions
}

func NewReconciler(
	client reconcilers.ClientHelper,
	privilegedNamespace string,
	permissionsVendor *discover.Permissions,
) Reconciler {
	return Reconciler{
		client:              client,
		privilegedNamespace: privilegedNamespace,
		vendor:              permissionsVendor,
	}
}

func (c *Reconciler) Reconcile(ctx context.Context, desired *v1alpha1.FlowCollectorEBPF) error {
	log.IntoContext(ctx, log.FromContext(ctx).WithName("permissions"))

	if err := c.reconcileNamespace(ctx); err != nil {
		return fmt.Errorf("reconciling namespace: %w", err)
	}
	if err := c.reconcileServiceAccount(ctx); err != nil {
		return fmt.Errorf("reconciling service account: %w", err)
	}
	if err := c.reconcileVendorPermissions(ctx, desired); err != nil {
		return fmt.Errorf("reconciling vendor permissions: %w", err)
	}
	return nil
}

func (c *Reconciler) reconcileNamespace(ctx context.Context) error {
	rlog := log.FromContext(ctx, "PrivilegedNamespace", c.privilegedNamespace)
	actual := &v1.Namespace{}
	if err := c.client.Get(ctx, client.ObjectKey{Name: c.privilegedNamespace}, actual); err != nil {
		if errors.IsNotFound(err) {
			actual = nil
		} else {
			return fmt.Errorf("can't retrieve current namespace: %w", err)
		}
	}
	desired := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.privilegedNamespace,
			Labels: map[string]string{
				"app":                                "network-observability-operator",
				"pod-security.kubernetes.io/enforce": "privileged",
				"pod-security.kubernetes.io/audit":   "privileged",
			},
		},
	}
	if actual == nil && desired != nil {
		rlog.Info("creating namespace")
		return c.client.CreateOwned(ctx, desired)
	}
	if actual != nil && desired != nil {
		if !helper.IsSubSet(actual.ObjectMeta.Labels, desired.ObjectMeta.Labels) {
			rlog.Info("updating namespace")
			return c.client.UpdateOwned(ctx, actual, desired)
		}
	}
	rlog.Info("namespace is already reconciled. Doing nothing")
	return nil
}

func (c *Reconciler) reconcileServiceAccount(ctx context.Context) error {
	rlog := log.FromContext(ctx, "serviceAccount", constants.EBPFServiceAccount)

	sAcc := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.EBPFServiceAccount,
			Namespace: c.privilegedNamespace,
		},
	}
	actual := &v1.ServiceAccount{}
	if err := c.client.Get(ctx, client.ObjectKeyFromObject(sAcc), actual); err != nil {
		if errors.IsNotFound(err) {
			actual = nil
		} else {
			return fmt.Errorf("can't retrieve current namespace: %w", err)
		}
	}
	if actual == nil {
		rlog.Info("creating service account")
		return c.client.CreateOwned(ctx, sAcc)
	}
	rlog.Info("service account already reconciled. Doing nothing")
	return nil
}

func (c *Reconciler) reconcileVendorPermissions(
	ctx context.Context, desired *v1alpha1.FlowCollectorEBPF,
) error {
	if c.vendor.Vendor(ctx) == discover.VendorOpenShift {
		return c.reconcileOpenshiftPermissions(ctx, desired)
	}
	return nil
}

func (c *Reconciler) reconcileOpenshiftPermissions(
	ctx context.Context, desired *v1alpha1.FlowCollectorEBPF,
) error {
	rlog := log.FromContext(ctx,
		"securityContextConstraints", constants.EBPFSecurityContext)
	scc := &osv1.SecurityContextConstraints{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.EBPFSecurityContext,
		},
		AllowHostNetwork: true,
		RunAsUser: osv1.RunAsUserStrategyOptions{
			Type: osv1.RunAsUserStrategyRunAsAny,
		},
		SELinuxContext: osv1.SELinuxContextStrategyOptions{
			Type: osv1.SELinuxStrategyRunAsAny,
		},
		Users: []string{
			"system:serviceaccount:" + c.privilegedNamespace + ":" + constants.EBPFServiceAccount,
		},
	}
	if desired.Privileged {
		scc.AllowPrivilegedContainer = true
	} else {
		scc.AllowedCapabilities = AllowedCapabilities
	}
	actual := &osv1.SecurityContextConstraints{}
	if err := c.client.Get(ctx, client.ObjectKeyFromObject(scc), actual); err != nil {
		if errors.IsNotFound(err) {
			actual = nil
		} else {
			return fmt.Errorf("can't retrieve current namespace: %w", err)
		}
	}
	if actual == nil {
		rlog.Info("creating SecurityContextConstraints")
		return c.client.CreateOwned(ctx, scc)
	}
	if !equality.Semantic.DeepDerivative(&scc, &actual) {
		rlog.Info("updating SecurityContextConstraints")
		return c.client.UpdateOwned(ctx, actual, scc)
	}
	rlog.Info("SecurityContextConstraints already reconciled. Doing nothing")
	return nil
}
