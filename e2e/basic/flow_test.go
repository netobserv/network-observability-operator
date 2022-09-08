package basic

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/mariomac/guara/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/netobserv/network-observability-operator/e2e/cluster"
	"github.com/netobserv/network-observability-operator/e2e/cluster/tester"
)

const (
	clusterNamePrefix = "netobserv-e2e-basic-"
	testTimeout       = 10 * time.Minute
	namespace         = "default"
)

var (
	testCluster *cluster.Kind
	log         = tester.InitLogger()
)

func TestMain(m *testing.M) {
	// if os.Getenv("ACTIONS_RUNNER_DEBUG") == "true" {
	// 	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	// }
	testCluster = cluster.NewKind(
		clusterNamePrefix+time.Now().Format("20060102-150405"),
		path.Join("..", ".."),
		cluster.Deploy(cluster.Deployment{
			Order: cluster.AfterAgent, ManifestFile: "manifests/pods.yml"}),
	)
	testCluster.Run(m)
}

// TestBasicFlowCapture checks that the agent is correctly capturing the request/response flows
// between the pods/service deployed from the manifests/pods.yml file
func TestBasicFlowCapture(t *testing.T) {
	bt := FlowCaptureTester{
		Cluster:   testCluster,
		Namespace: namespace,
		Timeout:   testTimeout,
	}
	bt.DoTest(t)
}

// TestSinglePacketFlows uses a known packet size and number to check that,
// (1) packets are aggregated only once,
// (2) once packets are evicted, no more flows are aggregated on top of them.
func TestSinglePacketFlows(t *testing.T) {
	var pingerIP, serverPodIP string
	testCluster.TestEnv().Test(t, features.New("single-packet flow capture").Setup(
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			kclient, err := kubernetes.NewForConfig(cfg.Client().RESTConfig())
			require.NoError(t, err)
			// extract pinger Pod information from kubernetes
			test.Eventually(t, testTimeout, func(t require.TestingT) {
				client, err := kclient.CoreV1().Pods(namespace).
					Get(ctx, "pinger", metav1.GetOptions{})
				require.NoError(t, err)
				require.NotEmpty(t, client.Status.PodIP)
				pingerIP = client.Status.PodIP
			}, test.Interval(time.Second))
			// extract server (ping destination) pod information from kubernetes
			test.Eventually(t, testTimeout, func(t require.TestingT) {
				server, err := kclient.CoreV1().Pods(namespace).
					List(ctx, metav1.ListOptions{LabelSelector: "app=server"})
				require.NoError(t, err)
				require.Len(t, server.Items, 1)
				require.NotEmpty(t, server.Items)
				require.NotEmpty(t, server.Items[0].Status.PodIP)
				serverPodIP = server.Items[0].Status.PodIP
			}, test.Interval(time.Second))
			return ctx
		},
	).Assess("correctness of single, sequential small ICMP packets from pinger to server",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			pods, err := tester.NewPods(cfg)
			require.NoError(t, err)

			const ipIcmpHeadersLen = 42
			latestFlowMS := time.Now().Add(-time.Minute)
			for pktLen := 50; pktLen <= 200; pktLen++ {
				log.Info("Sending ICMP packet", "destinationIP", serverPodIP)
				stdOut, stdErr, err := pods.Execute(ctx, namespace, "pinger",
					"ping", "-s", strconv.Itoa(pktLen), "-c", "1", serverPodIP)
				require.NoError(t, err)
				log.Info("ping sent", "stdOut", stdOut, "stdErr", stdErr)

				sent, recv := getPingFlows(t, latestFlowMS)
				log.Info(fmt.Sprintf("ping request flow: %#v", sent))
				log.Info(fmt.Sprintf("ping response flow: %#v", recv))

				assert.Equal(t, pingerIP, sent["SrcAddr"])
				assert.Equal(t, serverPodIP, sent["DstAddr"])
				assert.EqualValues(t, pktLen+ipIcmpHeadersLen, sent["Bytes"])
				assert.EqualValues(t, 1, sent["Packets"])
				assert.Equal(t, pingerIP, recv["DstAddr"])
				assert.Equal(t, serverPodIP, recv["SrcAddr"])
				assert.EqualValues(t, pktLen+ipIcmpHeadersLen, recv["Bytes"])
				assert.EqualValues(t, 1, recv["Packets"])

				if t.Failed() {
					log.Info(fmt.Sprintf("latestFlowMS: %v (vs received %d)", latestFlowMS.UnixMilli(),
						recv["TimeFlowEndMs"]))
					return ctx
				}
				latestFlowMS = asTime(recv["TimeFlowEndMs"])

			}

			return ctx
		},
	).Feature())
}

func getPingFlows(t *testing.T, newerThan time.Time) (sent, recv map[string]interface{}) {
	log.Info("Verifying that the request/return ICMP packets have been captured individually")
	var query *tester.LokiQueryResponse
	var err error
	test.Eventually(t, testTimeout, func(t require.TestingT) {
		query, err = testCluster.Loki().
			Query(1, `{SrcK8S_OwnerName="pinger",DstK8S_OwnerName="server"}|="\"Proto\":1,"`) // Proto 1 == ICMP
		require.NoError(t, err)
		require.NotNil(t, query)
		require.NotEmpty(t, query.Data.Result)
		if len(query.Data.Result) > 0 {
			sent, err = query.Data.Result[0].Values[0].FlowData()
			require.NoError(t, err)
			require.Less(t, newerThan.UnixMilli(),
				asTime(sent["TimeFlowStartMs"]).UnixMilli())
		}
	}, test.Interval(time.Second))

	test.Eventually(t, testTimeout, func(t require.TestingT) {
		query, err = testCluster.Loki().
			Query(1, `{DstK8S_OwnerName="pinger",SrcK8S_OwnerName="server"}|="\"Proto\":1,"`) // Proto 1 == ICMP
		require.NoError(t, err)
		require.NotNil(t, query)
		require.Len(t, query.Data.Result, 1)
		if len(query.Data.Result) > 0 {
			recv, err = query.Data.Result[0].Values[0].FlowData()
			require.NoError(t, err)
			require.Less(t, newerThan.UnixMilli(),
				asTime(recv["TimeFlowStartMs"]).UnixMilli())
		}
	}, test.Interval(time.Second))
	return sent, recv
}
