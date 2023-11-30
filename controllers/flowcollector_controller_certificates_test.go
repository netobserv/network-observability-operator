package controllers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta2"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	. "github.com/netobserv/network-observability-operator/controllers/controllerstest"
	"github.com/netobserv/network-observability-operator/controllers/flp"
	"github.com/netobserv/network-observability-operator/pkg/watchers"
)

var cmw watchers.ConfigWatchable
var sw watchers.SecretWatchable

// nolint:cyclop
func flowCollectorCertificatesSpecs() {
	const operatorNamespace = "main-namespace"
	crKey := types.NamespacedName{
		Name: "cluster",
	}
	flpKey := types.NamespacedName{
		Name:      constants.FLPName + flp.FlpConfSuffix[flp.ConfKafkaTransformer],
		Namespace: operatorNamespace,
	}
	pluginKey := types.NamespacedName{
		Name:      constants.PluginName,
		Namespace: operatorNamespace,
	}
	agentKey := types.NamespacedName{
		Name:      constants.EBPFAgentName,
		Namespace: operatorNamespace + "-privileged",
	}
	lokiCert := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "loki-gateway-ca-bundle",
			Namespace: "loki-namespace",
		},
		Data: map[string]string{
			"service-ca.crt": "--- LOKI CA CERT ---",
			"other":          "any",
		},
	}
	expectedLokiHash, _ := cmw.GetDigest(&lokiCert, []string{"service-ca.crt"})
	kafkaCert := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafka-ca",
			Namespace: operatorNamespace,
		},
		Data: map[string]string{
			"cert.crt": "--- KAFKA CA CERT ---",
			"other":    "any",
		},
	}
	expectedKafkaHash, _ := cmw.GetDigest(&kafkaCert, []string{"cert.crt"})
	kafkaUserCert := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafka-user",
			Namespace: operatorNamespace,
		},
		Data: map[string][]byte{
			"user.crt": []byte("--- KAFKA USER CERT ---"),
			"user.key": []byte("--- KAFKA USER KEY ---"),
			"other":    []byte("any"),
		},
	}
	expectedKafkaUserHash, _ := sw.GetDigest(&kafkaUserCert, []string{"user.crt", "user.key"})
	kafka2Cert := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafka-exporter-ca",
			Namespace: "kafka-exporter-namespace",
		},
		Data: map[string]string{
			"cert.crt": "--- KAFKA 2 CA CERT ---",
			"other":    "any",
		},
	}
	expectedKafka2Hash, _ := cmw.GetDigest(&kafka2Cert, []string{"cert.crt"})
	kafka2Sasl := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kafka-exporter-sasl",
			Namespace: operatorNamespace,
		},
		Data: map[string][]byte{
			"username": []byte("aiapaec"),
			"password": []byte("azerty"),
		},
	}
	expectedKafkaSaslHash1, _ := sw.GetDigest(&kafka2Sasl, []string{"username"})
	expectedKafkaSaslHash2, _ := sw.GetDigest(&kafka2Sasl, []string{"password"})

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	agent := appsv1.DaemonSet{}
	flp := appsv1.Deployment{}
	plugin := appsv1.Deployment{}
	var lastAgentAnnots map[string]string
	var lastFLPAnnots map[string]string
	var lastPluginAnnots map[string]string

	Context("Verify expectations are sane", func() {
		It("Expected hashes should all be different", func() {
			cmEmpty, _ := cmw.GetDigest(&v1.ConfigMap{}, []string{"any"})
			sEmpty, _ := sw.GetDigest(&v1.Secret{}, []string{"any"})
			allKeys := map[string]interface{}{}
			for _, hash := range []string{"", cmEmpty, sEmpty, expectedLokiHash, expectedKafkaHash, expectedKafka2Hash, expectedKafkaUserHash, expectedKafkaSaslHash1, expectedKafkaSaslHash2} {
				allKeys[hash] = nil
			}
			Expect(allKeys).To(HaveLen(9))
		})
	})

	Context("Deploying with Loki and Kafka certificates", func() {
		It("Should create all certs successfully", func() {
			By("Creating Loki certificate")
			Eventually(func() interface{} { return k8sClient.Create(ctx, &lokiCert) }, timeout, interval).Should(Succeed())
			By("Creating Kafka CA certificate")
			Eventually(func() interface{} { return k8sClient.Create(ctx, &kafkaCert) }, timeout, interval).Should(Succeed())
			By("Creating Kafka User certificate")
			Eventually(func() interface{} { return k8sClient.Create(ctx, &kafkaUserCert) }, timeout, interval).Should(Succeed())
			By("Creating Kafka-export CA certificate")
			Eventually(func() interface{} { return k8sClient.Create(ctx, &kafka2Cert) }, timeout, interval).Should(Succeed())
			By("Creating Kafka-export SASL key")
			Eventually(func() interface{} { return k8sClient.Create(ctx, &kafka2Sasl) }, timeout, interval).Should(Succeed())
		})

		flowSpec := flowslatest.FlowCollectorSpec{
			Namespace:       operatorNamespace,
			DeploymentModel: flowslatest.DeploymentModelKafka,
			Agent: flowslatest.FlowCollectorAgent{
				Type: "EBPF",
			},
			Loki: flowslatest.FlowCollectorLoki{
				Enable: ptr.To(true),
				Mode:   flowslatest.LokiModeLokiStack,
				LokiStack: flowslatest.LokiStackRef{
					Name:      "loki",
					Namespace: "loki-namespace",
				},
			},
			Kafka: flowslatest.FlowCollectorKafka{
				TLS: flowslatest.ClientTLS{
					Enable: true,
					CACert: flowslatest.CertificateReference{
						Type:     flowslatest.RefTypeConfigMap,
						Name:     kafkaCert.Name,
						CertFile: "cert.crt",
						// No namespace means operator's namespace
					},
					UserCert: flowslatest.CertificateReference{
						Type:     flowslatest.RefTypeSecret,
						Name:     kafkaUserCert.Name,
						CertFile: "user.crt",
						CertKey:  "user.key",
						// No namespace means operator's namespace
					},
				},
			},
			Exporters: []*flowslatest.FlowCollectorExporter{{
				Type: flowslatest.KafkaExporter,
				Kafka: flowslatest.FlowCollectorKafka{
					TLS: flowslatest.ClientTLS{
						Enable: true,
						CACert: flowslatest.CertificateReference{
							Type:      flowslatest.RefTypeConfigMap,
							Name:      kafka2Cert.Name,
							Namespace: kafka2Cert.Namespace,
							CertFile:  "cert.crt",
						},
					},
					SASL: flowslatest.SASLConfig{
						Type: "PLAIN",
						ClientIDReference: flowslatest.FileReference{
							Type: flowslatest.RefTypeSecret,
							Name: kafka2Sasl.Name,
							File: "username",
						},
						ClientSecretReference: flowslatest.FileReference{
							Type: flowslatest.RefTypeSecret,
							Name: kafka2Sasl.Name,
							File: "password",
						},
					},
				},
			}},
		}

		It("Should create CR successfully", func() {
			Eventually(func() interface{} {
				return k8sClient.Create(ctx, &flowslatest.FlowCollector{
					ObjectMeta: metav1.ObjectMeta{
						Name: crKey.Name,
					},
					Spec: flowSpec,
				})
			}, timeout, interval).Should(Succeed())
		})

		It("Should copy certificates when necessary", func() {
			By("Expecting Loki CA cert copied to operator namespace")
			Eventually(func() interface{} {
				var cm v1.ConfigMap
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: lokiCert.Name, Namespace: operatorNamespace}, &cm); err != nil {
					return err
				}
				return cm.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"service-ca.crt": "--- LOKI CA CERT ---",
				"other":          "any",
			}))
			By("Expecting Kafka CA cert copied to privileged namespace")
			Eventually(func() interface{} {
				var cm v1.ConfigMap
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: kafkaCert.Name, Namespace: operatorNamespace + "-privileged"}, &cm); err != nil {
					return err
				}
				return cm.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"cert.crt": "--- KAFKA CA CERT ---",
				"other":    "any",
			}))
			By("Expecting Kafka User cert copied to privileged namespace")
			Eventually(func() interface{} {
				var s v1.Secret
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: kafkaUserCert.Name, Namespace: operatorNamespace + "-privileged"}, &s); err != nil {
					return err
				}
				return s.Data
			}, timeout, interval).Should(Equal(map[string][]byte{
				"user.crt": []byte("--- KAFKA USER CERT ---"),
				"user.key": []byte("--- KAFKA USER KEY ---"),
				"other":    []byte("any"),
			}))
		})

		It("Should have all certificates mounted", func() {
			By("Expecting Kafka certificates for Agent mounted")
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, agentKey, &agent); err != nil {
					return err
				}
				return agent.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(HaveLen(2))
			Expect(agent.Spec.Template.Annotations).To(HaveLen(2))
			Expect(agent.Spec.Template.Annotations["flows.netobserv.io/watched-kafka-ca"]).To(Equal(expectedKafkaHash))
			Expect(agent.Spec.Template.Annotations["flows.netobserv.io/watched-kafka-user"]).To(Equal(expectedKafkaUserHash))
			Expect(agent.Spec.Template.Spec.Volumes[0].Name).To(Equal("kafka-certs-ca"))
			Expect(agent.Spec.Template.Spec.Volumes[1].Name).To(Equal("kafka-certs-user"))
			lastAgentAnnots = agent.Spec.Template.Annotations

			By("Expecting Loki certificate for Plugin mounted")
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, pluginKey, &plugin); err != nil {
					return err
				}
				return plugin.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(HaveLen(5))
			Expect(plugin.Spec.Template.Annotations).To(HaveLen(1))
			Expect(plugin.Spec.Template.Spec.Volumes[0].Name).To(Equal("console-serving-cert"))
			Expect(plugin.Spec.Template.Spec.Volumes[1].Name).To(Equal("config-volume"))
			Expect(plugin.Spec.Template.Spec.Volumes[2].Name).To(Equal("loki-certs-ca"))
			Expect(plugin.Spec.Template.Spec.Volumes[3].Name).To(Equal("loki-status-certs-ca"))
			Expect(plugin.Spec.Template.Spec.Volumes[4].Name).To(Equal("loki-status-certs-user"))
			lastPluginAnnots = plugin.Spec.Template.Annotations

			By("Expecting Loki and Kafka certificates for FLP mounted")
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, flpKey, &flp); err != nil {
					return err
				}
				return flp.Spec.Template.Spec.Volumes
			}, timeout, interval).Should(HaveLen(8))
			Expect(flp.Spec.Template.Annotations).To(HaveLen(8))
			Expect(flp.Spec.Template.Annotations["flows.netobserv.io/watched-kafka-ca"]).To(Equal(expectedKafkaHash))
			Expect(flp.Spec.Template.Annotations["flows.netobserv.io/watched-kafka-user"]).To(Equal(expectedKafkaUserHash))
			Expect(flp.Spec.Template.Annotations["flows.netobserv.io/watched-kafka-export-0-ca"]).To(Equal(expectedKafka2Hash))
			Expect(flp.Spec.Template.Annotations["flows.netobserv.io/watched-kafka-export-0-sd1"]).To(Equal(expectedKafkaSaslHash1))
			Expect(flp.Spec.Template.Annotations["flows.netobserv.io/watched-kafka-export-0-sd2"]).To(Equal(expectedKafkaSaslHash2))
			Expect(flp.Spec.Template.Spec.Volumes[0].Name).To(Equal("config-volume"))
			Expect(flp.Spec.Template.Spec.Volumes[1].Name).To(Equal("kafka-cert-ca"))
			Expect(flp.Spec.Template.Spec.Volumes[2].Name).To(Equal("kafka-cert-user"))
			Expect(flp.Spec.Template.Spec.Volumes[3].Name).To(Equal("flowlogs-pipeline")) // token
			Expect(flp.Spec.Template.Spec.Volumes[4].Name).To(Equal("loki-certs-ca"))
			Expect(flp.Spec.Template.Spec.Volumes[5].Name).To(Equal("kafka-export-0-ca"))
			Expect(flp.Spec.Template.Spec.Volumes[6].Name).To(Equal("kafka-export-0-sasl-id"))
			Expect(flp.Spec.Template.Spec.Volumes[7].Name).To(Equal("kafka-export-0-sasl-secret"))
			lastFLPAnnots = flp.Spec.Template.Annotations
		})
	})

	Context("Updating Kafka certificates", func() {
		It("Should update Kafka certificate", func() {
			By("Updating Kafka CA certificate")
			kafkaCert.Data["cert.crt"] = "--- KAFKA CA CERT MODIFIED ---"
			Eventually(func() interface{} { return k8sClient.Update(ctx, &kafkaCert) }, timeout, interval).Should(Succeed())
			By("Updating Kafka User certificate")
			kafkaUserCert.Data["user.crt"] = []byte("--- KAFKA USER CERT MODIFIED ---")
			Eventually(func() interface{} { return k8sClient.Update(ctx, &kafkaUserCert) }, timeout, interval).Should(Succeed())
		})

		It("Should copy certificates when necessary", func() {
			By("Expecting Kafka CA cert updated to privileged namespace")
			Eventually(func() interface{} {
				var cm v1.ConfigMap
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: kafkaCert.Name, Namespace: operatorNamespace + "-privileged"}, &cm); err != nil {
					return err
				}
				return cm.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"cert.crt": "--- KAFKA CA CERT MODIFIED ---",
				"other":    "any",
			}))
			By("Expecting Kafka User cert updated to privileged namespace")
			Eventually(func() interface{} {
				var s v1.Secret
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: kafkaUserCert.Name, Namespace: operatorNamespace + "-privileged"}, &s); err != nil {
					return err
				}
				return s.Data
			}, timeout, interval).Should(Equal(map[string][]byte{
				"user.crt": []byte("--- KAFKA USER CERT MODIFIED ---"),
				"user.key": []byte("--- KAFKA USER KEY ---"),
				"other":    []byte("any"),
			}))
		})

		It("Should redeploy eBPF Agent", func() {
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, agentKey, &agent); err != nil {
					return err
				}
				return agent.Spec.Template.Annotations
			}, timeout, interval).Should(Not(Equal(lastAgentAnnots)))
			lastAgentAnnots = agent.Spec.Template.Annotations
		})

		It("Should redeploy FLP", func() {
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, flpKey, &flp); err != nil {
					return err
				}
				return flp.Spec.Template.Annotations
			}, timeout, interval).Should(Not(Equal(lastFLPAnnots)))
			lastFLPAnnots = flp.Spec.Template.Annotations
		})
	})

	Context("Updating Loki certificate", func() {
		It("Should update Loki certificate", func() {
			By("Updating Loki CA certificate")
			lokiCert.Data["service-ca.crt"] = "--- LOKI CA CERT MODIFIED ---"
			Eventually(func() interface{} { return k8sClient.Update(ctx, &lokiCert) }, timeout, interval).Should(Succeed())
		})

		It("Should copy certificates when necessary", func() {
			By("Expecting Loki CA cert updated to operator namespace")
			Eventually(func() interface{} {
				var cm v1.ConfigMap
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: lokiCert.Name, Namespace: operatorNamespace}, &cm); err != nil {
					return err
				}
				return cm.Data
			}, timeout, interval).Should(Equal(map[string]string{
				"service-ca.crt": "--- LOKI CA CERT MODIFIED ---",
				"other":          "any",
			}))
		})

		// Console plugin is not restarted, as Loki certificate is always read from file
		It("Should not redeploy Console plugin", func() {
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, pluginKey, &plugin); err != nil {
					return err
				}
				return plugin.Spec.Template.Annotations
			}, timeout, interval).Should(Equal(lastPluginAnnots))
			lastPluginAnnots = plugin.Spec.Template.Annotations
		})

		// FLP is not restarted, as Loki certificate is always read from file
		It("Should not redeploy FLP", func() {
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, flpKey, &flp); err != nil {
					return err
				}
				return flp.Spec.Template.Annotations
			}, timeout, interval).Should(Equal(lastFLPAnnots))
			lastFLPAnnots = flp.Spec.Template.Annotations
		})
	})

	Context("Dummy update of Kafka Secret/CM without cert change", func() {
		It("Should update Kafka Secret/CM", func() {
			By("Updating Kafka CM")
			kafkaCert.Annotations = map[string]string{"hey": "new annotation"}
			kafkaCert.Data["other"] = "any MODIFIED"
			Eventually(func() interface{} { return k8sClient.Update(ctx, &kafkaCert) }, timeout, interval).Should(Succeed())
			By("Updating Kafka Secret")
			kafkaUserCert.Annotations = map[string]string{"ho": "new annotation"}
			kafkaUserCert.Data["other"] = []byte("any MODIFIED")
			Eventually(func() interface{} { return k8sClient.Update(ctx, &kafkaUserCert) }, timeout, interval).Should(Succeed())
		})

		It("Should not redeploy Agent", func() {
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, agentKey, &agent); err != nil {
					return err
				}
				return agent.Spec.Template.Annotations
			}, timeout, interval).Should(Equal(lastAgentAnnots))
			lastAgentAnnots = agent.Spec.Template.Annotations
		})

		It("Should not redeploy FLP", func() {
			Eventually(func() interface{} {
				if err := k8sClient.Get(ctx, flpKey, &flp); err != nil {
					return err
				}
				return flp.Spec.Template.Annotations
			}, timeout, interval).Should(Equal(lastFLPAnnots))
			lastFLPAnnots = flp.Spec.Template.Annotations
		})
	})

	Context("Cleanup", func() {
		// Retrieve CR to get its UID
		flowCR := flowslatest.FlowCollector{}
		It("Should get CR", func() {
			Eventually(func() error {
				return k8sClient.Get(ctx, crKey, &flowCR)
			}, timeout, interval).Should(Succeed())
		})

		It("Should delete CR", func() {
			Eventually(func() error {
				return k8sClient.Delete(ctx, &flowCR)
			}, timeout, interval).Should(Succeed())
		})

		It("Should be garbage collected", func() {
			By("Expecting flowlogs-pipeline deployment to be garbage collected")
			Eventually(func() interface{} {
				d := appsv1.Deployment{}
				_ = k8sClient.Get(ctx, flpKey, &d)
				return &d
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))

			By("Expecting agent daemonset to be garbage collected")
			Eventually(func() interface{} {
				d := appsv1.DaemonSet{}
				_ = k8sClient.Get(ctx, agentKey, &d)
				return &d
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))

			By("Expecting console plugin deployment to be garbage collected")
			Eventually(func() interface{} {
				d := appsv1.Deployment{}
				_ = k8sClient.Get(ctx, pluginKey, &d)
				return &d
			}, timeout, interval).Should(BeGarbageCollectedBy(&flowCR))
		})

		It("Should not get CR", func() {
			Eventually(func() bool {
				err := k8sClient.Get(ctx, crKey, &flowCR)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
}
