package controllers

import (
	"fmt"
	"net"
	"time"

	"github.com/netobserv/network-observability-operator/controllers/goflowkube"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	ascv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

var _ = Describe("FlowCollector Controller", func() {

	const timeout = time.Second * 10
	const interval = 50 * time.Millisecond
	const flowCollectorPort = 999
	ipResolver.On("LookupIP", constants.GoflowKubeName+"."+operatorNamespace).
		Return([]net.IP{net.IPv4(11, 22, 33, 44)}, nil)
	expectedSharedTarget := "11.22.33.44:999"
	configMapKey := types.NamespacedName{
		Name:      "ovs-flows-config",
		Namespace: "openshift-network-operator",
	}
	gfKey := types.NamespacedName{
		Name:      constants.GoflowKubeName,
		Namespace: operatorNamespace,
	}
	key := types.NamespacedName{
		Name:      "test-cluster",
		Namespace: operatorNamespace,
	}

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Deployment with autho-scaling", func() {
		var oldGoflowConfigDigest string
		It("Should create successfully", func() {

			created := &flowsv1alpha1.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Name,
				},
				Spec: flowsv1alpha1.FlowCollectorSpec{
					GoflowKube: flowsv1alpha1.FlowCollectorGoflowKube{
						Kind:            "Deployment",
						Port:            flowCollectorPort,
						ImagePullPolicy: "Never",
						LogLevel:        "error",
						Image:           "testimg:latest",
						HPA: &flowsv1alpha1.FlowCollectorHPA{
							MinReplicas:                    helper.Int32Ptr(1),
							MaxReplicas:                    1,
							TargetCPUUtilizationPercentage: helper.Int32Ptr(90),
						},
					},
					IPFIX: flowsv1alpha1.FlowCollectorIPFIX{
						Sampling: 200,
					},
				},
			}

			// Create
			Expect(k8sClient.Create(ctx, created)).Should(Succeed())

			By("Expecting to create the goflow-kube Deployment")
			Eventually(func() interface{} {
				dp := appsv1.Deployment{}
				if err := k8sClient.Get(ctx, gfKey, &dp); err != nil {
					return err
				}
				oldGoflowConfigDigest = dp.Spec.Template.Annotations[goflowkube.PodConfigurationDigest]
				if oldGoflowConfigDigest == "" {
					return fmt.Errorf("%q annotation can't be empty", goflowkube.PodConfigurationDigest)
				}
				return *dp.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(1)))

			svc := v1.Service{}
			By("Expecting to create the goflow-kube Service")
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, gfKey, &svc); err != nil {
					return err
				}
				return svc
			}, timeout, interval).Should(Satisfy(func(svc v1.Service) bool {
				return svc.Labels != nil && svc.Labels["app"] == constants.GoflowKubeName &&
					svc.Spec.Selector != nil && svc.Spec.Selector["app"] == constants.GoflowKubeName &&
					len(svc.Spec.Ports) == 1 &&
					svc.Spec.Ports[0].Protocol == v1.ProtocolUDP &&
					svc.Spec.Ports[0].Port == flowCollectorPort
			}), "unexpected service contents", helper.AsyncJSON{Ptr: svc})

			By("Creating the ovn-flows-configmap with the configuration from the FlowCollector")
			Eventually(func() interface{} {
				ofc := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, configMapKey, &ofc); err != nil {
					return err
				}
				return ofc.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"sampling":           "200",
				"sharedTarget":       expectedSharedTarget,
				"cacheMaxFlows":      "100",
				"cacheActiveTimeout": "10s",
			}))
		})

		It("Should update successfully", func() {
			Eventually(func() error {
				fc := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, key, &fc); err != nil {
					return err
				}
				fc.Spec.IPFIX.CacheActiveTimeout = "30s"
				fc.Spec.IPFIX.Sampling = 1234
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())

			By("Expecting that ovn-flows-configmap is updated accordingly")
			Eventually(func() interface{} {
				ofc := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, configMapKey, &ofc); err != nil {
					return err
				}
				return ofc.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"sampling":           "1234",
				"sharedTarget":       expectedSharedTarget,
				"cacheMaxFlows":      "100",
				"cacheActiveTimeout": "30s",
			}))
		})

		It("Should redeploy if the spec doesn't change but the external goflow-kube-config does", func() {
			Eventually(func() error {
				fc := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, key, &fc); err != nil {
					return err
				}
				fc.Spec.Loki.MaxRetries = 7
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())

			By("Expecting that the goflowkube.PodConfigurationDigest attribute has changed")
			Eventually(func() error {
				dp := appsv1.Deployment{}
				if err := k8sClient.Get(ctx, gfKey, &dp); err != nil {
					return err
				}
				currentGoflowConfigDigest := dp.Spec.Template.Annotations[goflowkube.PodConfigurationDigest]
				if currentGoflowConfigDigest != oldGoflowConfigDigest {
					return fmt.Errorf("annotation %v %q was expected to change",
						goflowkube.PodConfigurationDigest, currentGoflowConfigDigest)
				}
				return nil
			}).Should(Succeed())
		})

		It("Should autoscale when the HPA options change", func() {
			hpa := ascv1.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, gfKey, &hpa)).To(Succeed())
			Expect(*hpa.Spec.MinReplicas).To(Equal(int32(1)))
			Expect(hpa.Spec.MaxReplicas).To(Equal(int32(1)))
			Expect(*hpa.Spec.TargetCPUUtilizationPercentage).To(Equal(int32(90)))
			// update FlowCollector and verify that HPA spec also changed
			fc := flowsv1alpha1.FlowCollector{}
			Expect(k8sClient.Get(ctx, key, &fc)).To(Succeed())
			fc.Spec.GoflowKube.HPA.MinReplicas = helper.Int32Ptr(2)
			fc.Spec.GoflowKube.HPA.MaxReplicas = 2
			Expect(k8sClient.Update(ctx, &fc)).To(Succeed())

			By("Changing the Horizontal Pod Autoscaler instance")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, gfKey, &hpa); err != nil {
					return err
				}
				if *hpa.Spec.MinReplicas != int32(2) || hpa.Spec.MaxReplicas != int32(2) ||
					*hpa.Spec.TargetCPUUtilizationPercentage != int32(90) {
					return fmt.Errorf("expected {2, 2, 90}: Got %v, %v, %v",
						*hpa.Spec.MinReplicas, hpa.Spec.MaxReplicas,
						*hpa.Spec.TargetCPUUtilizationPercentage)
				}
				return nil
			}, timeout, interval).Should(Succeed())
		})

		It("Should delete successfully", func() {
			Eventually(func() error {
				f := &flowsv1alpha1.FlowCollector{}
				_ = k8sClient.Get(ctx, key, f)
				return k8sClient.Delete(ctx, f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete the ovn-flows-configmap")
			Eventually(func() error {
				return k8sClient.Get(ctx, configMapKey, &v1.ConfigMap{})
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
	Context("Deploying as DaemonSet", func() {
		var oldGoflowConfigDigest string
		It("Should create successfully", func() {
			created := &flowsv1alpha1.FlowCollector{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Name,
				},
				Spec: flowsv1alpha1.FlowCollectorSpec{
					GoflowKube: flowsv1alpha1.FlowCollectorGoflowKube{
						Kind:            "DaemonSet",
						Port:            7891,
						ImagePullPolicy: "Never",
						LogLevel:        "error",
						Image:           "testimg:latest",
					},
					IPFIX: flowsv1alpha1.FlowCollectorIPFIX{
						Sampling: 200,
					},
				},
			}
			// Create
			Expect(k8sClient.Create(ctx, created)).Should(Succeed())

			By("Expecting to create the ovn-flows-configmap with the configuration from the FlowCollector")
			Eventually(func() interface{} {
				ofc := v1.ConfigMap{}
				if err := k8sClient.Get(ctx, configMapKey, &ofc); err != nil {
					return err
				}
				return ofc.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"sampling":           "200",
				"nodePort":           "7891",
				"cacheMaxFlows":      "100",
				"cacheActiveTimeout": "10s",
			}))

			ds := appsv1.DaemonSet{}
			Expect(k8sClient.Get(ctx, gfKey, &ds)).To(Succeed())

			oldGoflowConfigDigest = ds.Spec.Template.Annotations[goflowkube.PodConfigurationDigest]
			Expect(oldGoflowConfigDigest).ToNot(BeEmpty())

			By("Creating the required HostPort to access Goflow through the NodeIP", func() {
				var cnt *v1.Container
				for i := range ds.Spec.Template.Spec.Containers {
					if ds.Spec.Template.Spec.Containers[i].Name == constants.GoflowKubeName {
						cnt = &ds.Spec.Template.Spec.Containers[i]
						break
					}
				}
				Expect(cnt).ToNot(BeNil(), "can't find a container named", constants.GoflowKubeName)
				var cp *v1.ContainerPort
				for i := range cnt.Ports {
					if cnt.Ports[i].Name == constants.GoflowKubeName {
						cp = &cnt.Ports[i]
						break
					}
				}
				Expect(cp).
					ToNot(BeNil(), "can't find a container port named", constants.GoflowKubeName)
				Expect(*cp).To(Equal(v1.ContainerPort{
					Name:          constants.GoflowKubeName,
					HostPort:      7891,
					ContainerPort: 7891,
					Protocol:      "UDP",
				}))
			})

			By("Allocating the proper toleration to allow its placement in the master nodes", func() {
				Expect(ds.Spec.Template.Spec.Tolerations).
					To(ContainElement(v1.Toleration{Operator: v1.TolerationOpExists}))
			})
		})
		It("Should redeploy if the spec doesn't change but the external goflow-kube-config does", func() {
			Eventually(func() error {
				fc := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, key, &fc); err != nil {
					return err
				}
				fc.Spec.Loki.MaxRetries = 7
				return k8sClient.Update(ctx, &fc)
			}).Should(Succeed())

			By("Expecting that the goflowkube.PodConfigurationDigest attribute has changed")
			Eventually(func() error {
				dp := appsv1.DaemonSet{}
				if err := k8sClient.Get(ctx, gfKey, &dp); err != nil {
					return err
				}
				currentGoflowConfigDigest := dp.Spec.Template.Annotations[goflowkube.PodConfigurationDigest]
				if currentGoflowConfigDigest != oldGoflowConfigDigest {
					return fmt.Errorf("annotation %v %q was expected to change",
						goflowkube.PodConfigurationDigest, currentGoflowConfigDigest)
				}
				return nil
			}).Should(Succeed())
		})
		Specify("daemonset deletion", func() {
			Eventually(func() error {
				f := &flowsv1alpha1.FlowCollector{}
				_ = k8sClient.Get(ctx, key, f)
				return k8sClient.Delete(ctx, f)
			}, timeout, interval).Should(Succeed())
		})
	})
	Context("configuring the Console plugin", func() {
		dKey := types.NamespacedName{
			Name:      "console-plugin-flowcollector",
			Namespace: operatorNamespace,
		}
		created := &flowsv1alpha1.FlowCollector{
			ObjectMeta: metav1.ObjectMeta{
				Name: dKey.Name,
			},
			Spec: flowsv1alpha1.FlowCollectorSpec{
				GoflowKube: flowsv1alpha1.FlowCollectorGoflowKube{
					Kind:            "Deployment",
					Port:            7891,
					ImagePullPolicy: "Never",
					LogLevel:        "error",
					Image:           "testimg:latest",
				},
				Loki: flowsv1alpha1.FlowCollectorLoki{
					URL: "http://loki:1234",
				},
				ConsolePlugin: flowsv1alpha1.FlowCollectorConsolePlugin{
					Replicas:        1,
					Port:            8888,
					Image:           "console:latest",
					ImagePullPolicy: "Never",
				},
			},
		}
		It("Should configure the Loki URL in the Console plugin backend", func() {
			Expect(k8sClient.Create(ctx, created)).Should(Succeed())
			Eventually(getContainerArgumentAfter("network-observability-plugin", "-loki"),
				timeout, interval).Should(Equal("http://loki:1234"))
		})
		It("Should update the Loki URL in the Console Plugin if it changes in the Spec", func() {
			Expect(func() error {
				upd := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, dKey, &upd); err != nil {
					return err
				}
				upd.Spec.Loki.URL = "http://loki.namespace:8888"
				return k8sClient.Update(ctx, &upd)
			}()).Should(Succeed())
			Eventually(getContainerArgumentAfter("network-observability-plugin", "-loki"),
				timeout, interval).Should(Equal("http://loki.namespace:8888"))
		})
		It("Should use the Loki Querier URL instead of the Loki URL, if the first is defined", func() {
			Expect(func() error {
				upd := flowsv1alpha1.FlowCollector{}
				if err := k8sClient.Get(ctx, dKey, &upd); err != nil {
					return err
				}
				upd.Spec.Loki.QuerierURL = "http://loki-querier:6789"
				return k8sClient.Update(ctx, &upd)
			}()).Should(Succeed())
			Eventually(getContainerArgumentAfter("network-observability-plugin", "-loki"),
				timeout, interval).Should(Equal("http://loki-querier:6789"))
		})
	})
})

func getContainerArgumentAfter(containerName, argName string) func() interface{} {
	pluginDeploymentKey := types.NamespacedName{
		Name:      "network-observability-plugin",
		Namespace: operatorNamespace,
	}
	return func() interface{} {
		deployment := appsv1.Deployment{}
		if err := k8sClient.Get(ctx, pluginDeploymentKey, &deployment); err != nil {
			return err
		}
		for i := range deployment.Spec.Template.Spec.Containers {
			cnt := &deployment.Spec.Template.Spec.Containers[i]
			if cnt.Name == containerName {
				args := cnt.Args
				for len(args) > 0 {
					if args[0] == argName {
						if len(args) < 2 {
							return fmt.Errorf("container %q: arg %v has no value. Actual args: %v",
								containerName, argName, cnt.Args)
						}
						return args[1]
					}
					args = args[1:]
				}
				return fmt.Errorf("container %q: arg %v not found. Actual args: %v",
					containerName, argName, cnt.Args)
			}
		}
		return fmt.Errorf("container not found: %v", containerName)
	}
}
