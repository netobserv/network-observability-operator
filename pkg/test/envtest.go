package test

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	//nolint:revive,stylecheck
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive,stylecheck
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	operatorsv1 "github.com/openshift/api/operator/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	// nolint:staticcheck
	flowsv1beta1 "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta1"
	flowsv1beta2 "github.com/netobserv/network-observability-operator/apis/flowcollector/v1beta2"
	metricsv1alpha1 "github.com/netobserv/network-observability-operator/apis/flowmetrics/v1alpha1"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	"github.com/netobserv/network-observability-operator/pkg/manager"
	"github.com/netobserv/network-observability-operator/pkg/manager/status"
)

const (
	Timeout              = time.Second * 10
	Interval             = 1 * time.Second
	ConsistentlyTimeout  = 2 * time.Second
	ConsistentlyInterval = 500 * time.Millisecond
)

func PrepareEnvTest(controllers []manager.Registerer, namespaces []string, basePath string) (context.Context, client.Client, *envtest.Environment, context.CancelFunc) {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	ctx, cancel := context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv := &envtest.Environment{
		Scheme: scheme.Scheme,
		CRDInstallOptions: envtest.CRDInstallOptions{
			Paths: []string{
				// Hack to reintroduce when the API stored version != latest version: comment-out config/crd/bases and use hack instead; see also Makefile "hack-crd-for-test"
				filepath.Join(basePath, "..", "config", "crd", "bases"),
				// filepath.Join(basePath, "..", "hack"),
				// We need to install the ConsolePlugin CRD to test setup of our Network Console Plugin
				filepath.Join(basePath, "..", "vendor", "github.com", "openshift", "api", "console", "v1alpha1"),
				filepath.Join(basePath, "..", "vendor", "github.com", "openshift", "api", "config", "v1"),
				filepath.Join(basePath, "..", "vendor", "github.com", "openshift", "api", "operator", "v1"),
				filepath.Join(basePath, "..", "test-assets"),
			},
			CleanUpAfterUse: true,
			WebhookOptions: envtest.WebhookInstallOptions{
				Paths: []string{
					filepath.Join(basePath, "..", "config", "webhook"),
				},
			},
		},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = flowsv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = flowsv1beta2.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = metricsv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = osv1alpha1.Install(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = configv1.Install(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = apiregv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = ascv2.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = operatorsv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = monitoringv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	for _, ns := range namespaces {
		err := k8sClient.Create(ctx, &corev1.Namespace{
			TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		})
		Expect(err).NotTo(HaveOccurred())
	}

	k8sManager, err := manager.NewManager(
		context.Background(),
		cfg,
		&manager.Config{
			EBPFAgentImage:        "registry-proxy.engineering.redhat.com/rh-osbs/network-observability-ebpf-agent@sha256:6481481ba23375107233f8d0a4f839436e34e50c2ec550ead0a16c361ae6654e",
			FlowlogsPipelineImage: "registry-proxy.engineering.redhat.com/rh-osbs/network-observability-flowlogs-pipeline@sha256:6481481ba23375107233f8d0a4f839436e34e50c2ec550ead0a16c361ae6654e",
			ConsolePluginImage:    "registry-proxy.engineering.redhat.com/rh-osbs/network-observability-console-plugin@sha256:6481481ba23375107233f8d0a4f839436e34e50c2ec550ead0a16c361ae6654e",
			DownstreamDeployment:  false,
		},
		&ctrl.Options{
			Scheme: scheme.Scheme,
			Metrics: server.Options{
				BindAddress: "0", // disable
			},
		},
		controllers,
	)

	Expect(err).ToNot(HaveOccurred())
	Expect(k8sManager).NotTo(BeNil())

	err = helper.SetCRDForTests(filepath.Join(basePath, ".."))
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

	return ctx, k8sClient, testEnv, cancel
}

func TeardownEnvTest(testEnv *envtest.Environment, cancel context.CancelFunc) {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
}

func GetCR(ctx context.Context, k8sClient client.Client, key types.NamespacedName) *flowsv1beta2.FlowCollector {
	cr := flowsv1beta2.FlowCollector{}
	Eventually(func() error {
		return k8sClient.Get(ctx, key, &cr)
	}).Should(Succeed())
	return &cr
}

func UpdateCR(ctx context.Context, k8sClient client.Client, key types.NamespacedName, updater func(*flowsv1beta2.FlowCollector)) {
	Eventually(func() error {
		cr := GetCR(ctx, k8sClient, key)
		updater(cr)
		return k8sClient.Update(ctx, cr)
	}, Timeout, Interval).Should(Succeed())
}

func CleanupCR(ctx context.Context, k8sClient client.Client, key types.NamespacedName) {
	By("Getting the CR")
	flowCR := GetCR(ctx, k8sClient, key)

	By("Deleting CR")
	Eventually(func() error {
		return k8sClient.Delete(ctx, flowCR)
	}, Timeout, Interval).Should(Succeed())

	By("Getting (no) CR")
	Eventually(func() error {
		err := k8sClient.Get(ctx, key, flowCR)
		if err == nil && flowCR.GetDeletionTimestamp() == nil {
			err = fmt.Errorf("CR is still present and not marked for deletion. Status: %s", status.ConditionsToString(flowCR.Status.Conditions))
		}
		return err
	}, Timeout, Interval).Should(Or(BeNil(), MatchError(`flowcollectors.flows.netobserv.io "cluster" not found`)))
}

func VolumeNames(vols []corev1.Volume) []string {
	var volNames []string
	for iv := range vols {
		volNames = append(volNames, vols[iv].Name)
	}
	return volNames
}

func Annotations(annots map[string]string) []string {
	var kv []string
	for k, v := range annots {
		kv = append(kv, k+"="+v)
	}
	return kv
}
