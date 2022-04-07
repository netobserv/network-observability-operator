package permissions

import (
	"context"
	"fmt"

	"github.com/netobserv/network-observability-operator/controllers/constants"
	osv1 "github.com/openshift/api/security/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Apply the different resources to enable the privileged operation of the
// Netobserv Agent:
// - Create the privileged namespace with Pod Security annotations
// - Create netobserv-agent service account in the non-privileged namespace
// - For Openshift, apply the required SecurityContextConstraints for privileged Pod operation
// TODO: delete
func Apply(ctx context.Context, k8sClient client.Client, baseNamespace string) error {
	log.IntoContext(ctx, log.FromContext(ctx).WithName("permissions"))

	if err := createOrUpdateNamespace(ctx, k8sClient, baseNamespace); err != nil {
		return err
	}
	if err := createOrUpdateServiceAccount(ctx, k8sClient, baseNamespace); err != nil {
		return err
	}
	return applyExtraPermissions(ctx, k8sClient, baseNamespace)
}

func createOrUpdateNamespace(ctx context.Context, k8sClient client.Client, baseNamespace string) error {
	namespace := baseNamespace + constants.EBPFPrivilegedNSSuffix
	rlog := log.FromContext(ctx)
	rlog.Info("creating or updating agent namespace", "namespace", namespace)
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	result, err := controllerutil.CreateOrUpdate(ctx, k8sClient, &ns,
		func() error {
			if ns.Labels == nil {
				ns.Labels = map[string]string{}
			}
			ns.Labels["pod-security.kubernetes.io/enforce"] = "privileged"
			ns.Labels["pod-security.kubernetes.io/audit"] = "privileged"
			return nil
		})
	if err != nil {
		return fmt.Errorf("can't create/update agent namespace: %w", err)
	}
	rlog.Info("successfully created/updated agent namespace", "result", result)
	return nil
}

func createOrUpdateServiceAccount(ctx context.Context, k8sClient client.Client, baseNamespace string) error {
	res, err := controllerutil.CreateOrUpdate(ctx, k8sClient,
		&v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.EBPFServiceAccount,
				Namespace: baseNamespace,
			},
		},
		func() error {
			return nil
		})
	if err != nil {
		return fmt.Errorf("service account couldn't be created/updated: %w", err)
	}
	log.FromContext(ctx).Info("service account created or updated", "result", res)
	return nil
}

// applyExtraPermissions inspects into the API services to know which mechanism should use to assign
// extra permissions (e.g. SecurityContextConstraints in OpenShift)
func applyExtraPermissions(ctx context.Context, k8sClient client.Client, baseNamespace string) error {
	services := apiregv1.APIServiceList{}
	rlog := log.FromContext(ctx)
	if err := k8sClient.List(ctx, &services); err != nil {
		rlog.Error(err, "couldn't retrieve API services. No extra permissions will be applied")
	} else {
		for i := range services.Items {
			if services.Items[i].Name == "v1.security.openshift.io" {
				rlog.Info("found OpenShift security manger")
				return applyOpenshiftPermissions(ctx, k8sClient, baseNamespace)
			}
		}
	}
	return nil
}

func applyOpenshiftPermissions(ctx context.Context, k8sClient client.Client, baseNamespace string) error {
	namespace := baseNamespace + constants.EBPFPrivilegedNSSuffix
	scc := osv1.SecurityContextConstraints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.EBPFSecurityContext,
			Namespace: namespace,
		},
	}
	res, err := controllerutil.CreateOrUpdate(ctx, k8sClient, &scc,
		func() error {
			scc.ObjectMeta.Name = constants.EBPFSecurityContext
			scc.ObjectMeta.Namespace = namespace
			scc.AllowPrivilegedContainer = true
			scc.AllowHostNetwork = true
			scc.RunAsUser = osv1.RunAsUserStrategyOptions{
				Type: osv1.RunAsUserStrategyRunAsAny,
			}
			scc.SELinuxContext = osv1.SELinuxContextStrategyOptions{
				Type: osv1.SELinuxStrategyRunAsAny,
			}
			scc.Users = []string{
				"system:serviceaccount:" + baseNamespace + ":" + constants.EBPFServiceAccount,
			}
			return nil
		})
	if err != nil {
		return fmt.Errorf("openshift security context constraints"+
			" couldn't be created nor updated: %w", err)
	}
	log.FromContext(ctx).WithName("openShift").
		Info("security context constraint created or updated", "result", res)

	return nil
}
