package e2etests

import (
	"context"
	"fmt"
	"strings"
	"time"

	o "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

type TestVMStaticIPTemplate struct {
	Name        string
	Namespace   string
	Mac         string
	StaticIP    string
	NetworkName string
	RunCmd      string
	Template    string
}

type TestVMUDNTemplate struct {
	Name        string
	Namespace   string
	NetworkName string
	RunCmd      string
	Template    string
}

// check if cluster has baremetal workers
func hasMetalWorkerNodes() bool {
	nodes, err := k8sClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
		LabelSelector: "node-role.kubernetes.io/worker",
	})
	o.Expect(err).NotTo(o.HaveOccurred())

	if len(nodes.Items) == 0 {
		e2e.Logf("Cluster does not have metal worker nodes")
		return false
	}
	for _, node := range nodes.Items {
		instanceType := node.Labels["node.kubernetes.io/instance-type"]
		if !strings.Contains(instanceType, "metal") {
			e2e.Logf("Cluster does not have metal worker nodes")
			return false
		}
	}
	return true
}

func isClusterBareMetal() (bool, error) {
	obj, err := getDynamicResource("infrastructures", "cluster", "")
	if err != nil {
		return false, err
	}

	platformType, _ := getNestedField(obj.Object, ".status.platformStatus.type")
	if !strings.Contains(platformType, "BareMetal") && !strings.Contains(platformType, "None") {
		return false, nil
	}
	return true, nil
}

// wait until hyperconverged is ready
func waitUntilHyperConvergedReady(hc, ns string) {
	err := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, 600*time.Second, false, func(ctx context.Context) (done bool, err error) {
		obj, getErr := getDynamicResource("hyperconverged", hc, ns)
		if getErr != nil {
			if apierrors.IsNotFound(getErr) {
				return false, nil
			}
			return false, getErr
		}
		condStatus, _, _ := getConditionStatus(obj, "Available")
		return condStatus == "True", nil
	})
	assertWaitPollNoErr(err, fmt.Sprintf("HyperConverged %s did not become Available", hc))
}

func (testTemplate *TestVMStaticIPTemplate) createVMStaticIP() error {
	templateParams := []string{"--ignore-unknown-parameters=true", "-f", testTemplate.Template, "-p", "NAME=" + testTemplate.Name, "-p", "NAMESPACE=" + testTemplate.Namespace, "-p", "NETWORK_NAME=" + testTemplate.NetworkName, "-p", "MAC=" + testTemplate.Mac, "-p", "STATIC_IP=" + testTemplate.StaticIP}

	if testTemplate.RunCmd != "" {
		templateParams = append(templateParams, "-p", "RUN_CMD="+testTemplate.RunCmd)
	}

	return applyResourceFromTemplateByAdmin(templateParams...)
}

func (testTemplate *TestVMUDNTemplate) createVMUDN() error {
	templateParams := []string{"--ignore-unknown-parameters=true", "-f", testTemplate.Template, "-p", "NAME=" + testTemplate.Name, "-p", "NAMESPACE=" + testTemplate.Namespace, "-p", "NETWORK_NAME=" + testTemplate.NetworkName}

	if testTemplate.RunCmd != "" {
		templateParams = append(templateParams, "-p", "RUN_CMD="+testTemplate.RunCmd)
	}

	return applyResourceFromTemplateByAdmin(templateParams...)
}

// cleanupCNVWebhooks removes orphaned CNV/HCO webhook configurations
// that can block namespace deletions after the operator is uninstalled.
func cleanupCNVWebhooks() {
	ctx := context.Background()
	cnvPrefixes := []string{"hco", "kubevirt", "virt", "cdi", "kubemacpool", "cnv", "ssp"}

	mutatingList, err := k8sClient.AdmissionregistrationV1().MutatingWebhookConfigurations().List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, wh := range mutatingList.Items {
			for _, prefix := range cnvPrefixes {
				if strings.Contains(strings.ToLower(wh.Name), prefix) {
					e2e.Logf("Cleaning up orphaned MutatingWebhookConfiguration: %s", wh.Name)
					_ = k8sClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(ctx, wh.Name, metav1.DeleteOptions{})
					break
				}
			}
		}
	}

	validatingList, err := k8sClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, wh := range validatingList.Items {
			for _, prefix := range cnvPrefixes {
				if strings.Contains(strings.ToLower(wh.Name), prefix) {
					e2e.Logf("Cleaning up orphaned ValidatingWebhookConfiguration: %s", wh.Name)
					_ = k8sClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Delete(ctx, wh.Name, metav1.DeleteOptions{})
					break
				}
			}
		}
	}
}

// wait until virtual machine is Ready
func waitUntilVMReady(vm, ns string) {
	err := wait.PollUntilContextTimeout(context.Background(), 30*time.Second, 1200*time.Second, false, func(ctx context.Context) (done bool, err error) {
		obj, getErr := getDynamicResource("virtualmachine", vm, ns)
		if getErr != nil {
			if apierrors.IsNotFound(getErr) {
				return false, nil
			}
			return false, getErr
		}
		condStatus, _, _ := getConditionStatus(obj, "Ready")
		return condStatus == "True", nil
	})
	assertWaitPollNoErr(err, fmt.Sprintf("Virtual machine %s did not become Available", vm))
}

// waitForVMIPAssignment waits until the VM has an IP assigned to the specified interface index
func waitForVMIPAssignment(vmName, namespace string, interfaceIndex int) (string, error) {
	var vmIP string
	err := wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 300*time.Second, false, func(ctx context.Context) (done bool, err error) {
		obj, getErr := getDynamicResource("vmi", vmName, namespace)
		if getErr != nil {
			if apierrors.IsNotFound(getErr) {
				return false, nil
			}
			return false, getErr
		}
		interfaces, found, _ := unstructured.NestedSlice(obj.Object, "status", "interfaces")
		if !found || interfaceIndex >= len(interfaces) {
			return false, nil
		}
		iface, ok := interfaces[interfaceIndex].(map[string]interface{})
		if !ok {
			return false, nil
		}
		ip, _ := iface["ipAddress"].(string)
		if ip != "" {
			vmIP = ip
			return true, nil
		}
		return false, nil
	})
	return vmIP, err
}
