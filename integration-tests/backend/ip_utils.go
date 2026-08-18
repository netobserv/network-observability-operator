package e2etests

import (
	"context"
	"fmt"
	"net"
	"strings"

	o "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	e2e "k8s.io/kubernetes/test/e2e/framework"
	e2eoutput "k8s.io/kubernetes/test/e2e/framework/pod/output"
	netutils "k8s.io/utils/net"
)

func checkIPStackType() string {
	obj, err := getDynamicResource("network.operator", "cluster", "")
	o.Expect(err).NotTo(o.HaveOccurred())

	rawNetworks, found, _ := unstructured.NestedSlice(obj.Object, "spec", "serviceNetwork")
	o.Expect(found).To(o.BeTrue())
	svcNetwork := fmt.Sprintf("%v", rawNetworks)
	if strings.Count(svcNetwork, ":") >= 2 && strings.Count(svcNetwork, ".") >= 2 {
		return "dualstack"
	} else if strings.Count(svcNetwork, ":") >= 2 {
		return "ipv6single"
	} else if strings.Count(svcNetwork, ".") >= 2 {
		return "ipv4single"
	}
	return ""
}

func getServiceIPv4(namespace, serviceName string) string {
	svc, err := k8sClient.CoreV1().Services(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	o.Expect(err).NotTo(o.HaveOccurred())
	serviceIPv4 := svc.Spec.ClusterIP
	e2e.Logf("The service %s IP in namespace %s is %q", serviceName, namespace, serviceIPv4)
	return serviceIPv4
}

// getPodIP returns IPv6 and IPv4 in vars in order on dual stack respectively and main IP in case of single stack (v4 or v6) in 1st var, and nil in 2nd var
func getPodIP(namespace, podName, ipStack string) (string, string) {
	pod, err := k8sClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	o.Expect(err).NotTo(o.HaveOccurred())

	if ipStack == "ipv6single" || ipStack == "ipv4single" {
		o.Expect(len(pod.Status.PodIPs)).To(o.BeNumerically(">", 0))
		podIP := pod.Status.PodIPs[0].IP
		e2e.Logf("The pod %s IP in namespace %s is %q", podName, namespace, podIP)
		return podIP, ""
	} else if ipStack == "dualstack" {
		o.Expect(len(pod.Status.PodIPs)).To(o.BeNumerically(">", 1))
		podIP1 := pod.Status.PodIPs[1].IP
		podIP2 := pod.Status.PodIPs[0].IP
		e2e.Logf("The pod's %s 1st IP in namespace %s is %q", podName, namespace, podIP1)
		e2e.Logf("The pod's %s 2nd IP in namespace %s is %q", podName, namespace, podIP2)
		if netutils.IsIPv6String(podIP1) {
			e2e.Logf("This is IPv4 primary dual stack cluster with IP %s", podIP1)
			return podIP1, podIP2
		}
		e2e.Logf("This is IPv6 primary dual stack cluster with IP %s", podIP2)
		return podIP2, podIP1
	}
	return "", ""
}

// CurlPod2PodFail ensures no connectivity from a pod to pod regardless of network addressing type on cluster
func CurlPod2PodFail(namespaceSrc, podNameSrc, namespaceDst, podNameDst, ipStackType string) {
	podIP1, podIP2 := getPodIP(namespaceDst, podNameDst, ipStackType)
	if podIP2 != "" {
		_, err := e2eoutput.RunHostCmd(namespaceSrc, podNameSrc, "curl --connect-timeout 5 -s "+net.JoinHostPort(podIP1, "8080"))
		o.Expect(err).To(o.HaveOccurred())
		_, err = e2eoutput.RunHostCmd(namespaceSrc, podNameSrc, "curl --connect-timeout 5 -s "+net.JoinHostPort(podIP2, "8080"))
		o.Expect(err).To(o.HaveOccurred())
	} else {
		_, err := e2eoutput.RunHostCmd(namespaceSrc, podNameSrc, "curl --connect-timeout 5 -s "+net.JoinHostPort(podIP1, "8080"))
		o.Expect(err).To(o.HaveOccurred())
	}
}
