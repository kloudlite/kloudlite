package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/dto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupServiceHandlerTest() (*ServiceHandlers, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	// Add apps/v1 scheme for Deployments
	_ = appsv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	logger, _ := zap.NewDevelopment()

	// Pass nil for clientset since we can't easily fake it and GetServiceLogs won't be tested here
	handlers := NewServiceHandlers(k8sClient, nil, logger)
	router := gin.New()

	return handlers, router
}

func TestListServices(t *testing.T) {
	handlers, router := setupServiceHandlerTest()

	// Create test deployments (service handler looks for deployments with kloudlite.io/managed label)
	deployment1 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-1",
			Namespace: "test-ns",
			Labels: map[string]string{
				"kloudlite.io/managed": "true",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "web",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "web",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "web",
							Image: "nginx",
						},
					},
				},
			},
		},
	}

	deployment2 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-2",
			Namespace: "test-ns",
			Labels: map[string]string{
				"kloudlite.io/managed": "true",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "api",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "api",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "api",
							Image: "api:latest",
						},
					},
				},
			},
		},
	}

	// Create matching services
	service1 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-1",
			Namespace: "test-ns",
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: "10.0.0.1",
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: map[string]string{
				"app": "web",
			},
		},
	}

	service2 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-2",
			Namespace: "test-ns",
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeLoadBalancer,
			ClusterIP: "10.0.0.2",
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Protocol:   corev1.ProtocolTCP,
					Port:       443,
					TargetPort: intstr.FromInt(8443),
				},
			},
			Selector: map[string]string{
				"app": "api",
			},
		},
	}

	_ = handlers.k8sClient.Create(context.Background(), deployment1)
	_ = handlers.k8sClient.Create(context.Background(), deployment2)
	_ = handlers.k8sClient.Create(context.Background(), service1)
	_ = handlers.k8sClient.Create(context.Background(), service2)

	t.Run("should list all services in namespace", func(t *testing.T) {
		router.GET("/namespaces/:namespace/services", handlers.ListServices)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns/services", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ServiceListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, response.Count)
		assert.Len(t, response.Services, 2)

		// Check first service
		assert.Equal(t, "service-1", response.Services[0].Name)
		assert.Equal(t, "test-ns", response.Services[0].Namespace)
		assert.Equal(t, "ClusterIP", response.Services[0].Type)
		assert.Equal(t, "10.0.0.1", response.Services[0].ClusterIP)
		assert.Len(t, response.Services[0].Ports, 1)
		assert.Equal(t, "http", response.Services[0].Ports[0].Name)
		assert.Equal(t, "TCP", response.Services[0].Ports[0].Protocol)
		assert.Equal(t, int32(80), response.Services[0].Ports[0].Port)
		assert.Equal(t, "8080", response.Services[0].Ports[0].TargetPort)
		assert.Equal(t, "web", response.Services[0].Selector["app"])

		// Check second service
		assert.Equal(t, "service-2", response.Services[1].Name)
		assert.Equal(t, "LoadBalancer", response.Services[1].Type)
	})

	t.Run("should return 400 when namespace is empty", func(t *testing.T) {
		router.GET("/namespaces/:namespace/services-empty", handlers.ListServices)

		req := httptest.NewRequest(http.MethodGet, "/namespaces//services-empty", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dto.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Namespace is required", response.Error)
	})

	t.Run("should return empty list for namespace with no services", func(t *testing.T) {
		router.GET("/namespaces/:namespace/services-empty-ns", handlers.ListServices)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/empty-ns/services-empty-ns", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ServiceListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.Count)
		assert.Len(t, response.Services, 0)
	})

	t.Run("should handle service with multiple ports", func(t *testing.T) {
		multiPortDeployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "multi-port-service",
				Namespace: "test-ns-2",
				Labels: map[string]string{
					"kloudlite.io/managed": "true",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":  "multi",
						"tier": "backend",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app":  "multi",
							"tier": "backend",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "multi",
								Image: "multi:latest",
							},
						},
					},
				},
			},
		}

		multiPortService := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "multi-port-service",
				Namespace: "test-ns-2",
			},
			Spec: corev1.ServiceSpec{
				Type:      corev1.ServiceTypeNodePort,
				ClusterIP: "10.0.0.3",
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						Protocol:   corev1.ProtocolTCP,
						Port:       80,
						TargetPort: intstr.FromInt(8080),
					},
					{
						Name:       "https",
						Protocol:   corev1.ProtocolTCP,
						Port:       443,
						TargetPort: intstr.FromInt(8443),
					},
					{
						Name:       "metrics",
						Protocol:   corev1.ProtocolTCP,
						Port:       9090,
						TargetPort: intstr.FromString("metrics"),
					},
				},
				Selector: map[string]string{
					"app":  "multi",
					"tier": "backend",
				},
			},
		}

		_ = handlers.k8sClient.Create(context.Background(), multiPortDeployment)
		_ = handlers.k8sClient.Create(context.Background(), multiPortService)

		router.GET("/namespaces/:namespace/services-multi", handlers.ListServices)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns-2/services-multi", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ServiceListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 1, response.Count)
		assert.Len(t, response.Services, 1)

		service := response.Services[0]
		assert.Equal(t, "multi-port-service", service.Name)
		assert.Equal(t, "NodePort", service.Type)
		assert.Len(t, service.Ports, 3)

		// Check ports
		assert.Equal(t, "http", service.Ports[0].Name)
		assert.Equal(t, int32(80), service.Ports[0].Port)
		assert.Equal(t, "8080", service.Ports[0].TargetPort)

		assert.Equal(t, "https", service.Ports[1].Name)
		assert.Equal(t, int32(443), service.Ports[1].Port)
		assert.Equal(t, "8443", service.Ports[1].TargetPort)

		assert.Equal(t, "metrics", service.Ports[2].Name)
		assert.Equal(t, int32(9090), service.Ports[2].Port)
		assert.Equal(t, "metrics", service.Ports[2].TargetPort)

		// Check selectors
		assert.Equal(t, "multi", service.Selector["app"])
		assert.Equal(t, "backend", service.Selector["tier"])
	})
}
