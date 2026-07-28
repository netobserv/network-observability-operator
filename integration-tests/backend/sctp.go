package e2etests

import (
	"context"
	"fmt"
	"strings"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// enableSCTPModuleOnNode Manual way to enable sctp in a cluster
func enableSCTPModuleOnNode(nodeName, role string) {
	e2e.Logf("This is %s worker node: %s", role, nodeName)
	checkSCTPCmd := "cat /sys/module/sctp/initstate"
	output, err := debugNodeWithCommand(nodeName, checkSCTPCmd)
	var installCmd string
	if err != nil || !strings.Contains(output, "live") {
		e2e.Logf("No sctp module installed, will enable sctp module!!!")
		installCmd = "modprobe sctp"

		// Try 3 times to enable sctp
		o.Eventually(func() error {
			output, installErr := debugNodeWithCommand(nodeName, installCmd)
			if installErr != nil {
				e2e.Logf("modprobe sctp failed on %s node %s: %v, output: %s, retrying...", role, nodeName, installErr, output)
			}
			return installErr
		}, "15s", "5s").ShouldNot(o.HaveOccurred(), fmt.Sprintf("Failed to install sctp module on node %s", nodeName))

		// Wait for sctp applied
		o.Eventually(func() string {
			output, err := debugNodeWithCommand(nodeName, checkSCTPCmd)
			if err != nil {
				e2e.Logf("Wait for sctp applied, %v", err)
			}
			return output
		}, "60s", "10s").Should(o.ContainSubstring("live"), fmt.Sprintf("Failed to load sctp module on node %s", nodeName))
	} else {
		e2e.Logf("sctp module is loaded on node %s\n%s", nodeName, output)
	}
}

func prepareSCTPModule() {
	ctx := context.Background()

	allNodes, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	o.Expect(err).NotTo(o.HaveOccurred())
	for _, node := range allNodes.Items {
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" && cond.Status != "True" {
				g.Skip("There are already some nodes in NotReady or SchedulingDisabled status in cluster, skip the test!!! ")
			}
		}
		if node.Spec.Unschedulable {
			g.Skip("There are already some nodes in NotReady or SchedulingDisabled status in cluster, skip the test!!! ")
		}
	}

	workers, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: "node-role.kubernetes.io/worker"})
	if err != nil || len(workers.Items) == 0 {
		g.Skip("Can not find any woker nodes in the cluster")
	}

	schedulableFound := false
	for _, w := range workers.Items {
		if !w.Spec.Unschedulable {
			schedulableFound = true
			break
		}
	}
	if !schedulableFound {
		g.Skip("Can not find any woker nodes in the cluster")
	}

	rhelWorkers, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: "node-role.kubernetes.io/worker,node.openshift.io/os_id=rhel"})
	o.Expect(err).NotTo(o.HaveOccurred())
	var rhelWorkerNames []string
	for _, n := range rhelWorkers.Items {
		rhelWorkerNames = append(rhelWorkerNames, n.Name)
	}

	rhcosWorkers, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: "node-role.kubernetes.io/worker,node.openshift.io/os_id=rhcos"})
	o.Expect(err).NotTo(o.HaveOccurred())
	var rhcosWorkerNames []string
	for _, n := range rhcosWorkers.Items {
		rhcosWorkerNames = append(rhcosWorkerNames, n.Name)
	}
	e2e.Logf("%v", rhcosWorkerNames)

	if len(rhelWorkerNames) > 0 {
		e2e.Logf("There are %v number rhel workers in this cluster, will use manual way to load sctp module.", len(rhelWorkerNames))
		for _, worker := range rhelWorkerNames {
			enableSCTPModuleOnNode(worker, "rhel")
		}
	}

	e2e.Logf("%v", rhcosWorkerNames)
	if len(rhcosWorkerNames) > 0 {
		for _, worker := range rhcosWorkerNames {
			enableSCTPModuleOnNode(worker, "rhcos")
		}
	}
}
