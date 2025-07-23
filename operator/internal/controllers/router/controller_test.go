package router

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	fn "github.com/kloudlite/kloudlite/operator/toolkit/functions"
	v1 "github.com/kloudlite/operator/api/v1"
	"github.com/kloudlite/operator/internal/controllers/router/templates"
)

// Unit tests for helper functions
func TestGenNginxIngressAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		router   *v1.Router
		expected map[string]string
	}{
		{
			name: "basic router without features",
			router: &v1.Router{
				Spec: v1.RouterSpec{},
			},
			expected: map[string]string{
				"nginx.ingress.kubernetes.io/preserve-trailing-slash": "true",
				"nginx.ingress.kubernetes.io/rewrite-target":          "/$1",
				"nginx.ingress.kubernetes.io/from-to-www-redirect":    "true",
			},
		},
		{
			name: "router with basic auth enabled",
			router: &v1.Router{
				Spec: v1.RouterSpec{
					BasicAuth: &v1.RouterBasicAuth{
						Enabled:    ptr.To(true),
						SecretName: "test-secret",
					},
				},
			},
			expected: map[string]string{
				"nginx.ingress.kubernetes.io/preserve-trailing-slash": "true",
				"nginx.ingress.kubernetes.io/rewrite-target":          "/$1",
				"nginx.ingress.kubernetes.io/from-to-www-redirect":    "true",
				"nginx.ingress.kubernetes.io/auth-type":               "basic",
				"nginx.ingress.kubernetes.io/auth-secret":             "test-secret",
				"nginx.ingress.kubernetes.io/auth-realm":              "route is protected by basic auth",
			},
		},
		{
			name: "router with HTTPS enabled",
			router: &v1.Router{
				Spec: v1.RouterSpec{
					Https: &v1.RouterHttps{
						Enabled:       ptr.To(true),
						ForceRedirect: true,
					},
				},
			},
			expected: map[string]string{
				"nginx.ingress.kubernetes.io/preserve-trailing-slash":  "true",
				"nginx.ingress.kubernetes.io/rewrite-target":           "/$1",
				"nginx.ingress.kubernetes.io/from-to-www-redirect":     "true",
				"nginx.kubernetes.io/ssl-redirect":                     "true",
				"nginx.ingress.kubernetes.io/force-ssl-redirect":       "true",
			},
		},
		{
			name: "router with rate limiting",
			router: &v1.Router{
				Spec: v1.RouterSpec{
					RateLimit: &v1.RouterRateLimit{
						Enabled:     ptr.To(true),
						Rps:         10,
						Rpm:         600,
						Connections: 100,
					},
				},
			},
			expected: map[string]string{
				"nginx.ingress.kubernetes.io/preserve-trailing-slash": "true",
				"nginx.ingress.kubernetes.io/rewrite-target":          "/$1",
				"nginx.ingress.kubernetes.io/from-to-www-redirect":    "true",
				"nginx.ingress.kubernetes.io/limit-rps":               "10",
				"nginx.ingress.kubernetes.io/limit-rpm":               "600",
				"nginx.ingress.kubernetes.io/limit-connections":       "100",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenNginxIngressAnnotations(tt.router)
			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, result[key])
				}
			}
		})
	}
}

func TestIsHttpsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		router   *v1.Router
		expected bool
	}{
		{
			name:     "no HTTPS config",
			router:   &v1.Router{Spec: v1.RouterSpec{}},
			expected: false,
		},
		{
			name: "HTTPS explicitly enabled",
			router: &v1.Router{
				Spec: v1.RouterSpec{
					Https: &v1.RouterHttps{Enabled: ptr.To(true)},
				},
			},
			expected: true,
		},
		{
			name: "HTTPS explicitly disabled",
			router: &v1.Router{
				Spec: v1.RouterSpec{
					Https: &v1.RouterHttps{Enabled: ptr.To(false)},
				},
			},
			expected: false,
		},
		{
			name: "HTTPS config without enabled field (defaults to true)",
			router: &v1.Router{
				Spec: v1.RouterSpec{
					Https: &v1.RouterHttps{},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHttpsEnabled(tt.router)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Integration tests with Ginkgo
var _ = Describe("Router Controller", func() {
	Context("When reconciling a Router resource", func() {
		var (
			ctx           context.Context
			router        *v1.Router
			reconciler    *Reconciler
			namespaceName string
		)

		BeforeEach(func() {
			ctx = context.Background()
			namespaceName = "test-" + fn.CleanerNanoid(8)

			// Create test namespace
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
				},
			}
			Expect(k8sClient.Create(ctx, ns)).To(Succeed())

			// Create test ingress class
			ingressClass := &networkingv1.IngressClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-ingress-class",
				},
				Spec: networkingv1.IngressClassSpec{
					Controller: "nginx.org/ingress-controller",
				},
			}
			err := k8sClient.Create(ctx, ingressClass)
			if err != nil && !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			// Setup reconciler with template
			reconciler = &Reconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Name:   "test-router-controller",
			}

			var templateErr error
			reconciler.templateIngress, templateErr = templates.Read(templates.IngressTemplate)
			Expect(templateErr).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			// Cleanup namespace
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
				},
			}
			_ = k8sClient.Delete(ctx, ns)
		})

		Context("Basic Router without features", func() {
			BeforeEach(func() {
				router = &v1.Router{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-router",
						Namespace: namespaceName,
					},
					Spec: v1.RouterSpec{
						IngressClass: "test-ingress-class",
						Routes: []v1.RouterRoute{
							{
								Host:    "example.com",
								Path:    "/api",
								Service: "api-service",
								Port:    8080,
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, router)).To(Succeed())
			})

			It("should create an ingress resource", func() {
				By("Reconciling the router")
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      router.Name,
						Namespace: router.Namespace,
					},
				})
				Expect(err).NotTo(HaveOccurred())

				By("Checking if ingress was created")
				ingress := &networkingv1.Ingress{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      router.Name,
					Namespace: router.Namespace,
				}, ingress)
				Expect(err).NotTo(HaveOccurred())

				By("Verifying ingress spec")
				Expect(ingress.Spec.IngressClassName).To(Equal(ptr.To("test-ingress-class")))
				Expect(ingress.Spec.Rules).To(HaveLen(1))
				Expect(ingress.Spec.Rules[0].Host).To(Equal("example.com"))
			})
		})

		Context("Router with Basic Auth", func() {
			BeforeEach(func() {
				router = &v1.Router{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-router-auth",
						Namespace: namespaceName,
					},
					Spec: v1.RouterSpec{
						IngressClass: "test-ingress-class",
						BasicAuth: &v1.RouterBasicAuth{
							Enabled: ptr.To(true),
						},
						Routes: []v1.RouterRoute{
							{
								Host:    "secure.example.com",
								Path:    "/api",
								Service: "api-service",
								Port:    8080,
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, router)).To(Succeed())
			})

			It("should create a basic auth secret and ingress with auth annotations", func() {
				By("Reconciling the router")
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      router.Name,
						Namespace: router.Namespace,
					},
				})
				Expect(err).NotTo(HaveOccurred())

				By("Checking if basic auth secret was created")
				secret := &corev1.Secret{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      router.Name + "-basic-auth",
					Namespace: router.Namespace,
				}, secret)
				Expect(err).NotTo(HaveOccurred())
				Expect(secret.Data).To(HaveKey("auth"))
				Expect(secret.Data).To(HaveKey("username"))
				Expect(secret.Data).To(HaveKey("password"))

				By("Checking if ingress has auth annotations")
				ingress := &networkingv1.Ingress{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      router.Name,
					Namespace: router.Namespace,
				}, ingress)
				Expect(err).NotTo(HaveOccurred())
				Expect(ingress.Annotations).To(HaveKeyWithValue("nginx.ingress.kubernetes.io/auth-type", "basic"))
			})
		})

		Context("Router with HTTPS", func() {
			BeforeEach(func() {
				router = &v1.Router{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-router-https",
						Namespace: namespaceName,
					},
					Spec: v1.RouterSpec{
						IngressClass: "test-ingress-class",
						Https: &v1.RouterHttps{
							Enabled: ptr.To(true),
						},
						Routes: []v1.RouterRoute{
							{
								Host:    "secure.example.com",
								Path:    "/",
								Service: "web-service",
								Port:    80,
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, router)).To(Succeed())
			})

			It("should create ingress with TLS configuration", func() {
				By("Reconciling the router")
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      router.Name,
						Namespace: router.Namespace,
					},
				})
				Expect(err).NotTo(HaveOccurred())

				By("Checking if ingress has TLS configuration")
				ingress := &networkingv1.Ingress{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      router.Name,
					Namespace: router.Namespace,
				}, ingress)
				Expect(err).NotTo(HaveOccurred())
				
				Expect(ingress.Spec.TLS).To(HaveLen(1))
				Expect(ingress.Spec.TLS[0].Hosts).To(ContainElement("secure.example.com"))
				Expect(ingress.Spec.TLS[0].SecretName).To(ContainSubstring("secure-example-com-tls"))
			})
		})
	})
})
