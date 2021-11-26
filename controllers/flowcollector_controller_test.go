package controllers

import (
	"fmt"
	"net"
	"time"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/pkg/helper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	ascv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
		Namespace: cnoNamespace,
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
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name: constants.GoflowKubeName, Namespace: operatorNamespace,
			}, &ds)).To(Succeed())

			By("Creating the required HostPort to access Goflow through the NodeIP", func() {
				var cnt *v1.Container
				for _, c := range ds.Spec.Template.Spec.Containers {
					if c.Name == constants.GoflowKubeName {
						cnt = &c
						break
					}
				}
				Expect(cnt).ToNot(BeNil(), "can't find a container named", constants.GoflowKubeName)
				var cp *v1.ContainerPort
				for _, p := range cnt.Ports {
					if p.Name == constants.GoflowKubeName {
						cp = &p
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
		Specify("daemonset deletion", func() {
			Eventually(func() error {
				f := &flowsv1alpha1.FlowCollector{}
				_ = k8sClient.Get(ctx, key, f)
				return k8sClient.Delete(ctx, f)
			}, timeout, interval).Should(Succeed())
		})
	})
})
