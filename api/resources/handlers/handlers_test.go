package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/service"
	"github.com/kloudlite/kloudlite/api/resources/store"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestListNamespacedResourceReturnsServiceUnavailableBeforeCacheReadiness(t *testing.T) {
	router, _ := newTestRouter(t, store.New())

	response := performRequest(router, http.MethodGet, "/api/v1/namespaces/default/resources/configmaps", nil)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d: %s", http.StatusServiceUnavailable, response.Code, response.Body.String())
	}
}

func TestClusterRouteRejectsNamespacedResource(t *testing.T) {
	router, _ := newTestRouter(t, store.New())

	response := performRequest(router, http.MethodGet, "/api/v1/resources/workspaces", nil)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestUnknownResourceReturnsNotFound(t *testing.T) {
	router, _ := newTestRouter(t, store.New())

	response := performRequest(router, http.MethodGet, "/api/v1/resources/does-not-exist", nil)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, response.Code, response.Body.String())
	}
}

func TestCreateNamespacedResourceWritesThroughKubernetesAndReturnsCreated(t *testing.T) {
	router, kube := newTestRouter(t, store.New())
	body := []byte(`{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"app"},"data":{"key":"value"}}`)

	response := performRequest(router, http.MethodPost, "/api/v1/namespaces/default/resources/configmaps", body)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var created corev1.ConfigMap
	if err := kube.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: "app"}, &created); err != nil {
		t.Fatalf("expected created configmap in fake client: %v", err)
	}
	if created.Data["key"] != "value" {
		t.Fatalf("expected created data to be preserved, got %#v", created.Data)
	}
}

func TestGetNamespacedResourceReturnsCachedObjectWhenReady(t *testing.T) {
	st := store.New()
	st.ReplaceScope("configmaps", "default", []client.Object{
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}, Data: map[string]string{"key": "value"}},
	})
	ensurer := &recordingEnsurer{}
	router, _ := newTestRouter(t, st, ensurer)

	response := performRequest(router, http.MethodGet, "/api/v1/namespaces/default/resources/configmaps/app", nil)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var object corev1.ConfigMap
	if err := json.Unmarshal(response.Body.Bytes(), &object); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if object.Name != "app" || object.Data["key"] != "value" {
		t.Fatalf("unexpected object response: %#v", object)
	}
	ensurer.expectCalled(t, "configmaps", "default")
}

func TestListNamespacedResourceReturnsItemsWhenReady(t *testing.T) {
	st := store.New()
	st.ReplaceScope("configmaps", "default", []client.Object{
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}},
	})
	ensurer := &recordingEnsurer{}
	router, _ := newTestRouter(t, st, ensurer)

	response := performRequest(router, http.MethodGet, "/api/v1/namespaces/default/resources/configmaps", nil)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload struct {
		Items []corev1.ConfigMap `json:"items"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) != 1 || payload.Items[0].Name != "app" {
		t.Fatalf("unexpected list response: %#v", payload.Items)
	}
	ensurer.expectCalled(t, "configmaps", "default")
}

func TestListNamespacedResourceSurfacesEnsurerError(t *testing.T) {
	router, _ := newTestRouter(t, store.New(), &recordingEnsurer{err: errors.New("sync failed")})

	response := performRequest(router, http.MethodGet, "/api/v1/namespaces/default/resources/configmaps", nil)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d: %s", http.StatusInternalServerError, response.Code, response.Body.String())
	}
}

func TestGetNamespacedResourceSurfacesTypedEnsurerError(t *testing.T) {
	ensurer := &recordingEnsurer{err: service.NewError(service.ErrBadRequest, "bad ensure", nil)}
	router, _ := newTestRouter(t, store.New(), ensurer)

	response := performRequest(router, http.MethodGet, "/api/v1/namespaces/default/resources/configmaps/app", nil)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	ensurer.expectCalled(t, "configmaps", "default")
}

func TestCreateNamespacedResourceForcesPathNamespace(t *testing.T) {
	ensurer := &recordingEnsurer{}
	router, kube := newTestRouter(t, store.New(), ensurer)
	body := []byte(`{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"app","namespace":"body"}}`)

	response := performRequest(router, http.MethodPost, "/api/v1/namespaces/default/resources/configmaps", body)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var created corev1.ConfigMap
	if err := kube.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: "app"}, &created); err != nil {
		t.Fatalf("expected configmap in path namespace: %v", err)
	}
	var bodyNamespace corev1.ConfigMap
	if err := kube.Get(context.Background(), client.ObjectKey{Namespace: "body", Name: "app"}, &bodyNamespace); err == nil {
		t.Fatal("did not expect configmap to be created in body namespace")
	}
	ensurer.expectCalled(t, "configmaps", "default")
}

func TestCreateRejectsInvalidJSONBody(t *testing.T) {
	router, _ := newTestRouter(t, store.New())

	response := performRequest(router, http.MethodPost, "/api/v1/namespaces/default/resources/configmaps", []byte(`{"metadata":`))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestCreateRejectsTrailingJSONBody(t *testing.T) {
	router, _ := newTestRouter(t, store.New())
	body := []byte(`{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"app"}}{"metadata":{"name":"other"}}`)

	response := performRequest(router, http.MethodPost, "/api/v1/namespaces/default/resources/configmaps", body)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestPatchAndDeleteReturnNotImplemented(t *testing.T) {
	ensurer := &recordingEnsurer{}
	router, _ := newTestRouter(t, store.New(), ensurer)
	patchBody := []byte(`{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"app"}}`)

	patchResponse := performRequest(router, http.MethodPatch, "/api/v1/namespaces/default/resources/configmaps/app", patchBody)
	if patchResponse.Code != http.StatusNotImplemented {
		t.Fatalf("expected patch status %d, got %d: %s", http.StatusNotImplemented, patchResponse.Code, patchResponse.Body.String())
	}
	ensurer.expectCalled(t, "configmaps", "default")
	ensurer.reset()

	deleteResponse := performRequest(router, http.MethodDelete, "/api/v1/namespaces/default/resources/configmaps/app", nil)
	if deleteResponse.Code != http.StatusNotImplemented {
		t.Fatalf("expected delete status %d, got %d: %s", http.StatusNotImplemented, deleteResponse.Code, deleteResponse.Body.String())
	}
	ensurer.expectCalled(t, "configmaps", "default")
}

func newTestRouter(t *testing.T, st *store.Store, ensurers ...NamespaceEnsurer) (*gin.Engine, client.Client) {
	t.Helper()
	ginMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(ginMode) })

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	kube := fake.NewClientBuilder().WithScheme(scheme).Build()
	svc := service.New(registry.Default(), st, kube)
	router := gin.New()
	var ensurer NamespaceEnsurer
	if len(ensurers) > 0 {
		ensurer = ensurers[0]
	}
	New(svc, registry.Default(), ensurer).RegisterRoutes(router.Group("/api/v1"))
	return router, kube
}

type recordingEnsurer struct {
	calls     int
	alias     string
	namespace string
	err       error
}

func (e *recordingEnsurer) EnsureNamespaced(ctx context.Context, alias string, namespace string) error {
	_ = ctx
	e.calls++
	e.alias = alias
	e.namespace = namespace
	return e.err
}

func (e *recordingEnsurer) expectCalled(t *testing.T, alias string, namespace string) {
	t.Helper()
	if e.calls != 1 || e.alias != alias || e.namespace != namespace {
		t.Fatalf("expected one ensure call for %s/%s, got calls=%d alias=%q namespace=%q", alias, namespace, e.calls, e.alias, e.namespace)
	}
}

func (e *recordingEnsurer) reset() {
	e.calls = 0
	e.alias = ""
	e.namespace = ""
}

func performRequest(router http.Handler, method string, path string, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
