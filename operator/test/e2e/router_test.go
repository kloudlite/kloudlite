package e2e_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/kloudlite/operator/api/v1"
	"github.com/kloudlite/operator/test/utils"
)

var _ = Describe("Router Controller E2E Tests", Ordered, func() {
	const (
		timeout  = time.Minute * 2
		interval = time.Second * 5
	)

	var (
		testNamespace string
		k8sClient     client.Client
		testService   *corev1.Service
	)

	BeforeAll(func() {
		// Create a unique namespace for tests
		testNamespace = fmt.Sprintf("router-test-%d", time.Now().Unix())
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}
		Expect(utils.CreateResource(ns)).To(Succeed())

		// Initialize Kubernetes client
		k8sClient = utils.GetK8sClient()

		// Create a test service to route to
		testService = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-service",
				Namespace: testNamespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": "test-app",
				},
				Ports: []corev1.ServicePort{
					{
						Name:     "http",
						Port:     80,
						Protocol: corev1.ProtocolTCP,
					},
				},
			},
		}
		Expect(utils.CreateResource(testService)).To(Succeed())
	})

	AfterAll(func() {
		// Clean up test namespace
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}
		Expect(utils.DeleteResource(ns)).To(Succeed())
	})

	Context("Basic Router Functionality", func() {
		It("should create a simple router with single route", func() {
			routerName := "simple-router"
			router := &v1.Router{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerName,
					Namespace: testNamespace,
				},
				Spec: v1.RouterSpec{
					Routes: []v1.RouterRoute{
						{
							Host:        "test.example.com",
							ServiceName: testService.Name,
							Path:        "/",
							Port:        80,
						},
					},
				},
			}

			By("Creating the Router resource")
			Expect(utils.CreateResource(router)).To(Succeed())

			By("Verifying Router is ready")
			Eventually(func() bool {
				var r v1.Router
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName,
					Namespace: testNamespace,
				}, &r)
				return err == nil && r.Status.IsReady
			}, timeout, interval).Should(BeTrue())

			By("Verifying Ingress is created")
			Eventually(func() bool {
				var ingress networkingv1.Ingress
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName,
					Namespace: testNamespace,
				}, &ingress)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying Ingress configuration")
			var ingress networkingv1.Ingress
			Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
				Name:      routerName,
				Namespace: testNamespace,
			}, &ingress)).To(Succeed())

			Expect(ingress.Spec.Rules).To(HaveLen(1))
			Expect(ingress.Spec.Rules[0].Host).To(Equal("test.example.com"))
			Expect(ingress.Spec.Rules[0].HTTP.Paths).To(HaveLen(1))
			Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Path).To(Equal("/?(.*)"))
			Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(Equal(testService.Name))

			By("Cleaning up")
			Expect(utils.DeleteResource(router)).To(Succeed())
		})

		It("should handle multiple routes to different services", func() {
			routerName := "multi-route-router"
			
			// Create additional test service
			service2 := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service-2",
					Namespace: testNamespace,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "test-app-2",
					},
					Ports: []corev1.ServicePort{
						{
							Name:     "http",
							Port:     8080,
							Protocol: corev1.ProtocolTCP,
						},
					},
				},
			}
			Expect(utils.CreateResource(service2)).To(Succeed())
			defer utils.DeleteResource(service2)

			router := &v1.Router{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerName,
					Namespace: testNamespace,
				},
				Spec: v1.RouterSpec{
					Routes: []v1.RouterRoute{
						{
							Host:        "app1.example.com",
							ServiceName: testService.Name,
							Path:        "/",
							Port:        80,
						},
						{
							Host:        "app2.example.com",
							ServiceName: service2.Name,
							Path:        "/api",
							Port:        8080,
						},
					},
				},
			}

			By("Creating the Router resource")
			Expect(utils.CreateResource(router)).To(Succeed())

			By("Verifying Ingress has multiple rules")
			Eventually(func() bool {
				var ingress networkingv1.Ingress
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName,
					Namespace: testNamespace,
				}, &ingress)
				return err == nil && len(ingress.Spec.Rules) == 2
			}, timeout, interval).Should(BeTrue())

			By("Verifying route configurations")
			var ingress networkingv1.Ingress
			Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
				Name:      routerName,
				Namespace: testNamespace,
			}, &ingress)).To(Succeed())

			// Verify first route
			rule1 := findIngressRule(ingress.Spec.Rules, "app1.example.com")
			Expect(rule1).NotTo(BeNil())
			Expect(rule1.HTTP.Paths[0].Backend.Service.Name).To(Equal(testService.Name))
			Expect(rule1.HTTP.Paths[0].Backend.Service.Port.Number).To(Equal(int32(80)))

			// Verify second route
			rule2 := findIngressRule(ingress.Spec.Rules, "app2.example.com")
			Expect(rule2).NotTo(BeNil())
			Expect(rule2.HTTP.Paths[0].Path).To(Equal("/api?(.*)"))
			Expect(rule2.HTTP.Paths[0].Backend.Service.Name).To(Equal(service2.Name))
			Expect(rule2.HTTP.Paths[0].Backend.Service.Port.Number).To(Equal(int32(8080)))

			By("Cleaning up")
			Expect(utils.DeleteResource(router)).To(Succeed())
		})
	})

	Context("Basic Authentication", func() {
		It("should setup basic authentication when enabled", func() {
			routerName := "basic-auth-router"
			router := &v1.Router{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerName,
					Namespace: testNamespace,
				},
				Spec: v1.RouterSpec{
					Routes: []v1.RouterRoute{
						{
							Host:        "auth.example.com",
							ServiceName: testService.Name,
							Path:        "/",
							Port:        80,
						},
					},
					BasicAuth: &v1.RouterBasicAuth{
						Enabled: ptr.To(true),
					},
				},
			}

			By("Creating the Router resource")
			Expect(utils.CreateResource(router)).To(Succeed())

			By("Verifying basic auth secret is created")
			Eventually(func() bool {
				var secret corev1.Secret
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName + "-basic-auth",
					Namespace: testNamespace,
				}, &secret)
				if err != nil {
					return false
				}
				// Check secret contains required fields
				_, hasAuth := secret.Data["auth"]
				_, hasUsername := secret.Data["username"]
				_, hasPassword := secret.Data["password"]
				return hasAuth && hasUsername && hasPassword
			}, timeout, interval).Should(BeTrue())

			By("Verifying ingress has basic auth annotations")
			var ingress networkingv1.Ingress
			Eventually(func() bool {
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName,
					Namespace: testNamespace,
				}, &ingress)
				if err != nil {
					return false
				}
				authType, hasAuthType := ingress.Annotations["nginx.ingress.kubernetes.io/auth-type"]
				authSecret, hasAuthSecret := ingress.Annotations["nginx.ingress.kubernetes.io/auth-secret"]
				return hasAuthType && authType == "basic" && hasAuthSecret && authSecret == routerName+"-basic-auth"
			}, timeout, interval).Should(BeTrue())

			By("Cleaning up")
			Expect(utils.DeleteResource(router)).To(Succeed())
		})
	})

	Context("HTTPS Configuration", func() {
		It("should configure HTTPS when enabled", func() {
			routerName := "https-router"
			router := &v1.Router{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerName,
					Namespace: testNamespace,
				},
				Spec: v1.RouterSpec{
					Routes: []v1.RouterRoute{
						{
							Host:        "secure.example.com",
							ServiceName: testService.Name,
							Path:        "/",
							Port:        80,
						},
					},
					Https: &v1.RouterHttps{
						Enabled:       ptr.To(true),
						ForceRedirect: true,
						ClusterIssuer: "test-issuer",
					},
				},
			}

			By("Creating the Router resource")
			Expect(utils.CreateResource(router)).To(Succeed())

			By("Verifying ingress has TLS configuration")
			Eventually(func() bool {
				var ingress networkingv1.Ingress
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName,
					Namespace: testNamespace,
				}, &ingress)
				return err == nil && len(ingress.Spec.TLS) > 0
			}, timeout, interval).Should(BeTrue())

			By("Verifying HTTPS annotations")
			var ingress networkingv1.Ingress
			Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
				Name:      routerName,
				Namespace: testNamespace,
			}, &ingress)).To(Succeed())

			Expect(ingress.Annotations["nginx.kubernetes.io/ssl-redirect"]).To(Equal("true"))
			Expect(ingress.Annotations["nginx.ingress.kubernetes.io/force-ssl-redirect"]).To(Equal("true"))
			
			// Verify TLS configuration
			Expect(ingress.Spec.TLS).To(HaveLen(1))
			Expect(ingress.Spec.TLS[0].Hosts).To(ContainElement("secure.example.com"))

			By("Cleaning up")
			Expect(utils.DeleteResource(router)).To(Succeed())
		})
	})

	Context("Rate Limiting", func() {
		It("should configure rate limiting when enabled", func() {
			routerName := "rate-limit-router"
			router := &v1.Router{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerName,
					Namespace: testNamespace,
				},
				Spec: v1.RouterSpec{
					Routes: []v1.RouterRoute{
						{
							Host:        "limited.example.com",
							ServiceName: testService.Name,
							Path:        "/",
							Port:        80,
						},
					},
					RateLimit: &v1.RouterRateLimit{
						Enabled:     ptr.To(true),
						Rps:         10,
						Rpm:         600,
						Connections: 5,
					},
				},
			}

			By("Creating the Router resource")
			Expect(utils.CreateResource(router)).To(Succeed())

			By("Verifying rate limit annotations")
			Eventually(func() bool {
				var ingress networkingv1.Ingress
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName,
					Namespace: testNamespace,
				}, &ingress)
				if err != nil {
					return false
				}
				
				rps, hasRps := ingress.Annotations["nginx.ingress.kubernetes.io/limit-rps"]
				rpm, hasRpm := ingress.Annotations["nginx.ingress.kubernetes.io/limit-rpm"]
				conn, hasConn := ingress.Annotations["nginx.ingress.kubernetes.io/limit-connections"]
				
				return hasRps && rps == "10" && hasRpm && rpm == "600" && hasConn && conn == "5"
			}, timeout, interval).Should(BeTrue())

			By("Cleaning up")
			Expect(utils.DeleteResource(router)).To(Succeed())
		})
	})

	Context("CORS Configuration", func() {
		It("should configure CORS when enabled", func() {
			routerName := "cors-router"
			router := &v1.Router{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerName,
					Namespace: testNamespace,
				},
				Spec: v1.RouterSpec{
					Routes: []v1.RouterRoute{
						{
							Host:        "api.example.com",
							ServiceName: testService.Name,
							Path:        "/api",
							Port:        80,
						},
					},
					Cors: &v1.RouterCors{
						Enabled:          ptr.To(true),
						Origins:          []string{"https://app.example.com", "https://web.example.com"},
						AllowCredentials: true,
					},
				},
			}

			By("Creating the Router resource")
			Expect(utils.CreateResource(router)).To(Succeed())

			By("Verifying CORS annotations")
			Eventually(func() bool {
				var ingress networkingv1.Ingress
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName,
					Namespace: testNamespace,
				}, &ingress)
				if err != nil {
					return false
				}
				
				enabled, hasEnabled := ingress.Annotations["nginx.ingress.kubernetes.io/enable-cors"]
				origins, hasOrigins := ingress.Annotations["nginx.ingress.kubernetes.io/cors-allow-origin"]
				creds, hasCreds := ingress.Annotations["nginx.ingress.kubernetes.io/cors-allow-credentials"]
				
				return hasEnabled && enabled == "true" && 
					hasOrigins && origins == "https://app.example.com,https://web.example.com" &&
					hasCreds && creds == "true"
			}, timeout, interval).Should(BeTrue())

			By("Cleaning up")
			Expect(utils.DeleteResource(router)).To(Succeed())
		})
	})

	Context("Advanced Features", func() {
		It("should handle max body size configuration", func() {
			routerName := "body-size-router"
			router := &v1.Router{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerName,
					Namespace: testNamespace,
				},
				Spec: v1.RouterSpec{
					Routes: []v1.RouterRoute{
						{
							Host:        "upload.example.com",
							ServiceName: testService.Name,
							Path:        "/upload",
							Port:        80,
						},
					},
					MaxBodySizeInMB: ptr.To(int32(100)),
				},
			}

			By("Creating the Router resource")
			Expect(utils.CreateResource(router)).To(Succeed())

			By("Verifying body size annotation")
			Eventually(func() string {
				var ingress networkingv1.Ingress
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName,
					Namespace: testNamespace,
				}, &ingress)
				if err != nil {
					return ""
				}
				return ingress.Annotations["nginx.ingress.kubernetes.io/proxy-body-size"]
			}, timeout, interval).Should(Equal("100m"))

			By("Cleaning up")
			Expect(utils.DeleteResource(router)).To(Succeed())
		})

		It("should handle backend protocol configuration", func() {
			routerName := "backend-protocol-router"
			router := &v1.Router{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerName,
					Namespace: testNamespace,
				},
				Spec: v1.RouterSpec{
					Routes: []v1.RouterRoute{
						{
							Host:        "grpc.example.com",
							ServiceName: testService.Name,
							Path:        "/",
							Port:        80,
						},
					},
					BackendProtocol: ptr.To("GRPC"),
				},
			}

			By("Creating the Router resource")
			Expect(utils.CreateResource(router)).To(Succeed())

			By("Verifying backend protocol annotation")
			Eventually(func() string {
				var ingress networkingv1.Ingress
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName,
					Namespace: testNamespace,
				}, &ingress)
				if err != nil {
					return ""
				}
				return ingress.Annotations["nginx.ingress.kubernetes.io/backend-protocol"]
			}, timeout, interval).Should(Equal("GRPC"))

			By("Cleaning up")
			Expect(utils.DeleteResource(router)).To(Succeed())
		})
	})

	Context("Resource Cleanup", func() {
		It("should clean up owned resources when router is deleted", func() {
			routerName := "cleanup-test-router"
			router := &v1.Router{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerName,
					Namespace: testNamespace,
				},
				Spec: v1.RouterSpec{
					Routes: []v1.RouterRoute{
						{
							Host:        "cleanup.example.com",
							ServiceName: testService.Name,
							Path:        "/",
							Port:        80,
						},
					},
					BasicAuth: &v1.RouterBasicAuth{
						Enabled: ptr.To(true),
					},
				},
			}

			By("Creating the Router resource")
			Expect(utils.CreateResource(router)).To(Succeed())

			By("Waiting for resources to be created")
			Eventually(func() bool {
				var ingress networkingv1.Ingress
				var secret corev1.Secret
				
				err1 := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName,
					Namespace: testNamespace,
				}, &ingress)
				
				err2 := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName + "-basic-auth",
					Namespace: testNamespace,
				}, &secret)
				
				return err1 == nil && err2 == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the Router resource")
			Expect(utils.DeleteResource(router)).To(Succeed())

			By("Verifying owned resources are deleted")
			Eventually(func() bool {
				var ingress networkingv1.Ingress
				var secret corev1.Secret
				
				err1 := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName,
					Namespace: testNamespace,
				}, &ingress)
				
				err2 := k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      routerName + "-basic-auth",
					Namespace: testNamespace,
				}, &secret)
				
				return client.IgnoreNotFound(err1) == nil && client.IgnoreNotFound(err2) == nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})

// Helper function to find ingress rule by host
func findIngressRule(rules []networkingv1.IngressRule, host string) *networkingv1.IngressRule {
	for _, rule := range rules {
		if rule.Host == host {
			return &rule
		}
	}
	return nil
}