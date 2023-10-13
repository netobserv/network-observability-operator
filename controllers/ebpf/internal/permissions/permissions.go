package permissions

import (
	"context"
	"fmt"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
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
	reconcilers.Common
}

func NewReconciler(cmn *reconcilers.Common) Reconciler {
	return Reconciler{Common: *cmn}
}

func (c *Reconciler) Reconcile(ctx context.Context, desired *flowslatest.FlowCollectorEBPF) error {
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
	ns := c.PrivilegedNamespace()
	if ns != c.PreviousPrivilegedNamespace() {
		if err := c.cleanupPreviousNamespace(ctx); err != nil {
			return err
		}
	}
	rlog := log.FromContext(ctx, "PrivilegedNamespace", ns)
	actual := &v1.Namespace{}
	if err := c.Get(ctx, client.ObjectKey{Name: ns}, actual); err != nil {
		if errors.IsNotFound(err) {
			actual = nil
		} else {
			return fmt.Errorf("can't retrieve current namespace: %w", err)
		}
	}
	desired := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
			Labels: map[string]string{
				"app":                                constants.OperatorName,
				"pod-security.kubernetes.io/enforce": "privileged",
				"pod-security.kubernetes.io/audit":   "privileged",
			},
		},
	}
	if actual == nil && desired != nil {
		rlog.Info("creating namespace")
		return c.CreateOwned(ctx, desired)
	}
	if actual != nil && desired != nil {
		// We noticed that audit labels are automatically removed
		// in some configurations of K8s, so to avoid an infinite update loop, we just ignore
		// it (if the user removes it manually, it's at their own risk)
		if !helper.IsSubSet(actual.ObjectMeta.Labels,
			map[string]string{
				"app":                                constants.OperatorName,
				"pod-security.kubernetes.io/enforce": "privileged",
			}) {
			rlog.Info("updating namespace")
			return c.UpdateOwned(ctx, actual, desired)
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
			Namespace: c.PrivilegedNamespace(),
		},
	}
	actual := &v1.ServiceAccount{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(sAcc), actual); err != nil {
		if errors.IsNotFound(err) {
			actual = nil
		} else {
			return fmt.Errorf("can't retrieve current namespace: %w", err)
		}
	}
	if actual == nil {
		rlog.Info("creating service account")
		return c.CreateOwned(ctx, sAcc)
	}
	rlog.Info("service account already reconciled. Doing nothing")
	return nil
}

func (c *Reconciler) reconcileVendorPermissions(
	ctx context.Context, desired *flowslatest.FlowCollectorEBPF,
) error {
	if c.UseOpenShiftSCC {
		return c.reconcileOpenshiftPermissions(ctx, desired)
	}
	return nil
}

func (c *Reconciler) reconcileOpenshiftPermissions(
	ctx context.Context, desired *flowslatest.FlowCollectorEBPF,
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
			"system:serviceaccount:" + c.PrivilegedNamespace() + ":" + constants.EBPFServiceAccount,
		},
	}
	if desired.Privileged {
		scc.AllowPrivilegedContainer = true
		scc.AllowHostDirVolumePlugin = true
	} else {
		scc.AllowedCapabilities = AllowedCapabilities
	}
	actual := &osv1.SecurityContextConstraints{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(scc), actual); err != nil {
		if errors.IsNotFound(err) {
			actual = nil
		} else {
			return fmt.Errorf("can't retrieve current namespace: %w", err)
		}
	}
	if actual == nil {
		rlog.Info("creating SecurityContextConstraints")
		return c.CreateOwned(ctx, scc)
	}
	if scc.AllowHostNetwork != actual.AllowHostNetwork ||
		!equality.Semantic.DeepDerivative(&scc.RunAsUser, &actual.RunAsUser) ||
		!equality.Semantic.DeepDerivative(&scc.SELinuxContext, &actual.SELinuxContext) ||
		!equality.Semantic.DeepDerivative(&scc.Users, &actual.Users) ||
		scc.AllowPrivilegedContainer != actual.AllowPrivilegedContainer ||
		scc.AllowHostDirVolumePlugin != actual.AllowHostDirVolumePlugin ||
		!equality.Semantic.DeepDerivative(&scc.AllowedCapabilities, &actual.AllowedCapabilities) {

		rlog.Info("updating SecurityContextConstraints")
		return c.UpdateOwned(ctx, actual, scc)
	}
	rlog.Info("SecurityContextConstraints already reconciled. Doing nothing")
	return nil
}

func (c *Reconciler) cleanupPreviousNamespace(ctx context.Context) error {
	rlog := log.FromContext(ctx, "PreviousPrivilegedNamespace", c.PreviousPrivilegedNamespace())

	// Delete service account
	if err := c.Delete(ctx, &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.EBPFServiceAccount,
			Namespace: c.PreviousPrivilegedNamespace(),
		},
	}); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("deleting eBPF agent ServiceAccount: %w", err)
	}
	// Do not delete SCC as it's not namespace-scoped (it will be reconciled "as usual")

	previous := &v1.Namespace{}
	if err := c.Get(ctx, client.ObjectKey{Name: c.PreviousPrivilegedNamespace()}, previous); err != nil {
		if errors.IsNotFound(err) {
			// Not found => return without error
			rlog.Info("Previous privileged namespace not found, skipping cleanup")
			return nil
		}
		return fmt.Errorf("can't retrieve previous namespace: %w", err)
	}
	// Make sure we own that namespace
	if helper.IsOwned(previous) {
		rlog.Info("Owning previous privileged namespace: deleting it")
		if err := c.Delete(ctx, previous); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("deleting privileged namespace: %w", err)
		}
	} else {
		rlog.Info("Not owning previous privileged namespace: delete related content only")
	}
	return nil
}
