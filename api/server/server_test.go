package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kloudlite/kloudlite/api/config"
	"github.com/kloudlite/kloudlite/api/k8s"
	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/service"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"github.com/kloudlite/kloudlite/api/resources/watch"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	packagesv1 "github.com/kloudlite/kloudlite/types/packages/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/types/snapshot/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/types/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/types/workspace/v1"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestWaitForHTTPSStartupReturnsImmediateError(t *testing.T) {
	want := errors.New("bind failed")
	errCh := make(chan error, 1)
	errCh <- want

	if err := waitForHTTPSStartup(errCh, time.Second); !errors.Is(err, want) {
		t.Fatalf("expected immediate startup error %v, got %v", want, err)
	}
}

func TestWaitForHTTPSStartupContinuesAfterDelay(t *testing.T) {
	errCh := make(chan error, 1)

	if err := waitForHTTPSStartup(errCh, time.Nanosecond); err != nil {
		t.Fatalf("expected nil after startup delay, got %v", err)
	}
}

func TestWebhookRouterRequiresAuthForResourceRoutes(t *testing.T) {
	router := newWebhookTestRouter(t, config.AuthConfig{JWTSecret: "secret"})

	response := performServerRequest(router, http.MethodGet, "/api/v1/namespaces/default/resources/configmaps", nil, "")

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnauthorized, response.Code, response.Body.String())
	}
}

func TestWebhookRouterAllowsAuthenticatedResourceRoutesToReachHandler(t *testing.T) {
	secret := "secret"
	router := newWebhookTestRouter(t, config.AuthConfig{JWTSecret: secret})

	response := performServerRequest(router, http.MethodGet, "/api/v1/namespaces/default/resources/configmaps", nil, "Bearer "+signedServerTestToken(t, secret))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func TestWebhookRouterSkipsResourceAuthWhenConfigured(t *testing.T) {
	router := newWebhookTestRouter(t, config.AuthConfig{SkipAuthentication: true})

	response := performServerRequest(router, http.MethodGet, "/api/v1/namespaces/default/resources/configmaps", nil, "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func TestWebhookRouterLeavesHealthWebhookAndInfoUnauthenticated(t *testing.T) {
	router := newWebhookTestRouter(t, config.AuthConfig{JWTSecret: "secret"})

	tests := []struct {
		name                    string
		method                  string
		path                    string
		body                    []byte
		wantStatus              int
		allowAnyNonUnauthorized bool
	}{
		{name: "health", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK},
		{name: "ready", method: http.MethodGet, path: "/ready", wantStatus: http.StatusOK},
		{name: "info", method: http.MethodGet, path: "/api/v1/info", wantStatus: http.StatusOK},
		{name: "webhook", method: http.MethodPost, path: "/webhooks/validate/users", body: webhookUserAdmissionBody(t), allowAnyNonUnauthorized: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performServerRequest(router, tt.method, tt.path, tt.body, "")
			if tt.allowAnyNonUnauthorized {
				if response.Code == http.StatusUnauthorized {
					t.Fatalf("expected unauthenticated webhook to avoid auth middleware, got %d: %s", response.Code, response.Body.String())
				}
				return
			}
			if response.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", tt.wantStatus, response.Code, response.Body.String())
			}
		})
	}
}

func newWebhookTestRouter(t *testing.T, auth config.AuthConfig) http.Handler {
	t.Helper()
	kube := newServerFakeClient(t)
	st := store.New()
	reg := registry.Default()
	svc := service.New(reg, st, kube)
	watchManager := watch.NewManager(reg, st, kube, zap.NewNop())
	return setupWebhookRouter(&config.Config{Auth: auth}, zap.NewNop(), &k8s.Client{RuntimeClient: kube}, svc, reg, watchManager)
}

func newServerFakeClient(t *testing.T, objects ...client.Object) client.WithWatch {
	t.Helper()
	scheme := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		clientgoscheme.AddToScheme,
		platformv1alpha1.AddToScheme,
		environmentsv1.AddToScheme,
		machinesv1.AddToScheme,
		workspacesv1.AddToScheme,
		packagesv1.AddToScheme,
		snapshotv1.AddToScheme,
	} {
		if err := add(scheme); err != nil {
			t.Fatal(err)
		}
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func performServerRequest(router http.Handler, method string, path string, body []byte, authHeader string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func signedServerTestToken(t *testing.T, secret string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func webhookUserAdmissionBody(t *testing.T) []byte {
	t.Helper()
	active := true
	user := platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &active,
			Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		},
	}
	userBytes, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshal user: %v", err)
	}
	body, err := json.Marshal(admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object:    runtime.RawExtension{Raw: userBytes},
		},
	})
	if err != nil {
		t.Fatalf("marshal admission review: %v", err)
	}
	return body
}
