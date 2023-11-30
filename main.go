/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"

	"go.uber.org/zap/zapcore"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	osv1alpha1 "github.com/openshift/api/console/v1alpha1"
	operatorsv1 "github.com/openshift/api/operator/v1"
	securityv1 "github.com/openshift/api/security/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	ascv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	// nolint:staticcheck
	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	flowsv1beta1 "github.com/netobserv/network-observability-operator/api/v1beta1"
	flowsv1beta2 "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/manager"
	//+kubebuilder:scaffold:imports
)

const app = constants.OperatorName

var (
	buildVersion = "unknown"
	buildDate    = "unknown"
	scheme       = runtime.NewScheme()
	setupLog     = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(flowsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(flowsv1beta1.AddToScheme(scheme))
	utilruntime.Must(flowsv1beta2.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(ascv2.AddToScheme(scheme))
	utilruntime.Must(osv1alpha1.AddToScheme(scheme))
	utilruntime.Must(apiregv1.AddToScheme(scheme))
	utilruntime.Must(securityv1.AddToScheme(scheme))
	utilruntime.Must(operatorsv1.AddToScheme(scheme))
	utilruntime.Must(monitoringv1.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var pprofAddr string
	var enableHTTP2 bool
	var versionFlag bool

	config := manager.Config{}

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&pprofAddr, "profiling-bind-address", "", "The address the profiling endpoint binds to, such as ':6060'. Leave unset to disable profiling.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&config.EBPFAgentImage, "ebpf-agent-image", "quay.io/netobserv/netobserv-ebpf-agent:main", "The image of the eBPF agent")
	flag.StringVar(&config.FlowlogsPipelineImage, "flowlogs-pipeline-image", "quay.io/netobserv/flowlogs-pipeline:main", "The image of Flowlogs Pipeline")
	flag.StringVar(&config.ConsolePluginImage, "console-plugin-image", "quay.io/netobserv/network-observability-console-plugin:main", "The image of the Console Plugin")
	flag.BoolVar(&config.DownstreamDeployment, "downstream-deployment", false, "Either this deployment is a downstream deployment ot not")
	flag.BoolVar(&enableHTTP2, "enable-http2", enableHTTP2, "If HTTP/2 should be enabled for the metrics and webhook servers.")
	flag.BoolVar(&versionFlag, "v", false, "print version")
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	appVersion := fmt.Sprintf("%s [build version: %s, build date: %s]", app, buildVersion, buildDate)
	if versionFlag {
		fmt.Println(appVersion)
		os.Exit(0)
	}
	setupLog.Info("Starting " + appVersion)

	if err := config.Validate(); err != nil {
		setupLog.Error(err, "unable to start the manager")
		os.Exit(1)
	}

	disableHTTP2 := func(c *tls.Config) {
		if enableHTTP2 {
			return
		}
		c.NextProtos = []string{"http/1.1"}
	}

	cfg := ctrl.GetConfigOrDie()

	mgr, err := manager.NewManager(context.Background(), cfg, &config, &ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
			TLSOpts:     []func(*tls.Config){disableHTTP2},
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port:    9443,
			TLSOpts: []func(*tls.Config){disableHTTP2},
		}),
		PprofBindAddress:       pprofAddr,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "7a7ecdcd.netobserv.io",
	}, controllers.Registerers)
	if err != nil {
		setupLog.Error(err, "unable to setup manager")
		os.Exit(1)
	}

	if err = (&flowsv1beta2.FlowCollector{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create v1beta2 webhook", "webhook", "FlowCollector")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
