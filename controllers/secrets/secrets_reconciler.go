package secrets

import (
	"context"
	"fmt"
	"reflect"

	"github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconciler reconciles the kafka secrets needed by the Netobserv Agent:
type Reconciler struct {
	client              reconcilers.ClientHelper
	baseNamespace       string
	privilegedNamespace string
	req                 ctrl.Request
}

func NewReconciler(
	client reconcilers.ClientHelper,
	baseNamespace string,
	req ctrl.Request,
) Reconciler {
	return Reconciler{
		client:              client,
		baseNamespace:       baseNamespace,
		privilegedNamespace: baseNamespace + constants.EBPFPrivilegedNSSuffix,
		req:                 req,
	}
}

func (c *Reconciler) Reconcile(ctx context.Context, desired *v1alpha1.FlowCollectorSpec) error {

	kafkaTLS := reconcilers.KafkaTLS(&desired.Kafka, desired.OperatorsAutoInstall)
	if desired.UseKafka() && kafkaTLS != nil && kafkaTLS.Enable {

		if (c.req.Name == kafkaTLS.UserCert.Name && c.req.Namespace == c.baseNamespace) || c.req.Name == constants.Cluster {
			if err := c.reconcileSecret(ctx, kafkaTLS.UserCert.Name); err != nil {
				return fmt.Errorf("reconciling kafka user secret: %w", err)
			}
		}

		if (c.req.Name == kafkaTLS.CACert.Name && c.req.Namespace == c.baseNamespace) || c.req.Name == constants.Cluster {
			if err := c.reconcileSecret(ctx, kafkaTLS.CACert.Name); err != nil {
				return fmt.Errorf("reconciling kafka ca secret: %w", err)
			}
		}
	}
	return nil
}

func (c *Reconciler) getSecrets(ctx context.Context, name string) (*v1.Secret, *v1.Secret, error) {
	rlog := log.FromContext(ctx, "component", "deployLoki")

	secret := &v1.Secret{}
	privilegedSecret := &v1.Secret{}

	if err := c.client.Get(ctx, client.ObjectKey{Name: name, Namespace: c.baseNamespace}, secret); err != nil {
		if errors.IsNotFound(err) {
			rlog.Info(fmt.Sprintf("secret %s is not yet available in namespace %s", name, c.baseNamespace))
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("can't retrieve secret: %w", err)
	}

	if err := c.client.Get(ctx, client.ObjectKey{Name: name, Namespace: c.privilegedNamespace}, privilegedSecret); err != nil {
		if errors.IsNotFound(err) {
			return secret, nil, nil
		}
		return nil, nil, fmt.Errorf("can't retrieve privileged secret: %w", err)
	}

	return secret, privilegedSecret, nil
}

func (c *Reconciler) reconcileSecret(ctx context.Context, name string) error {
	rlog := log.FromContext(ctx, "Name", name, "Namespace", c.baseNamespace, "PrivilegedNamespace", c.privilegedNamespace)

	secret, privilegedSecret, err := c.getSecrets(ctx, name)
	if err != nil {
		return err
	} else if secret == nil {
		// don't throw error when secret is not found
		return nil
	}

	if privilegedSecret == nil {
		rlog.Info(fmt.Sprintf("creating secret %s in namespace %s", name, c.privilegedNamespace))
		secret.ObjectMeta = metav1.ObjectMeta{
			Name:      name,
			Namespace: c.privilegedNamespace,
		}
		return c.client.CreateOwned(ctx, secret)
	} else if !reflect.DeepEqual(privilegedSecret.Data, secret.Data) {
		rlog.Info(fmt.Sprintf("updating secret %s in namespace %s", name, c.privilegedNamespace))
		privilegedSecret.Data = secret.Data
		return c.client.UpdateOwned(ctx, privilegedSecret, privilegedSecret)
	}

	rlog.Info(fmt.Sprintf("secret %s is up to date", name))
	return nil
}
