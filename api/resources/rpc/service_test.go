package rpc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/kloudlite/kloudlite/api/resources/events"
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

func TestConnectCreateMarksOnlyObjectDirtyAndListReportsMetadata(t *testing.T) {
	st := store.New()
	st.ReplaceScope("configmaps", "default", []client.Object{
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "existing", Namespace: "default"}},
	})
	server := newTestConnectServer(t, st)
	defer server.Close()

	create := connect.NewClient[MutateResourceRequest, MutateResourceResponse](server.Client(), server.URL+ProcedureCreate, connect.WithCodec(jsonCodec{}))
	_, err := create.CallUnary(context.Background(), connect.NewRequest(&MutateResourceRequest{
		Resource:  "configmaps",
		Namespace: "default",
		Object: jsonRawResource{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]any{
				"name": "created",
			},
		},
	}))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	list := connect.NewClient[ListResourcesRequest, ListResourcesResponse](server.Client(), server.URL+ProcedureList, connect.WithCodec(jsonCodec{}))
	resp, err := list.CallUnary(context.Background(), connect.NewRequest(&ListResourcesRequest{Resource: "configmaps", Namespace: "default"}))
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(resp.Msg.Items) != 1 || resp.Msg.Items[0].JSON["metadata"] == nil {
		t.Fatalf("expected cached list to remain available, got %#v", resp.Msg.Items)
	}
	if len(resp.Msg.Cache.Dirty) != 1 || resp.Msg.Cache.Dirty[0].Name != "created" {
		t.Fatalf("expected only created object to be dirty, got %#v", resp.Msg.Cache.Dirty)
	}
}

func TestWatchObjectStreamsDirtyAndCleanEvents(t *testing.T) {
	st := store.New()
	st.ReplaceScope("configmaps", "default", nil)
	broker := events.NewBroker()
	server := newTestConnectServerWithBroker(t, st, broker)
	defer server.Close()

	watch := connect.NewClient[WatchObjectRequest, ResourceEvent](server.Client(), server.URL+ProcedureWatchObject, connect.WithCodec(jsonCodec{}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := watch.CallServerStream(ctx, connect.NewRequest(&WatchObjectRequest{Resource: "configmaps", Namespace: "default", Name: "app"}))
	if err != nil {
		t.Fatalf("watch failed: %v", err)
	}

	dirty := events.DirtyObject{Resource: "configmaps", Namespace: "default", Name: "app", Reason: events.DirtyReasonPendingPatch}
	if !receiveEvent(t, stream, events.EventDirty, "app", func() {
		broker.Publish(events.Event{Type: events.EventDirty, Resource: "configmaps", Namespace: "default", Name: "app", Dirty: &dirty})
	}) {
		t.Fatal("expected dirty event")
	}
	if !receiveEvent(t, stream, events.EventClean, "app", func() {
		broker.Publish(events.Event{Type: events.EventClean, Resource: "configmaps", Namespace: "default", Name: "app", Dirty: &dirty})
	}) {
		t.Fatal("expected clean event")
	}
}

func newTestConnectServer(t *testing.T, st *store.Store) *httptest.Server {
	t.Helper()
	return newTestConnectServerWithBroker(t, st, events.NewBroker())
}

func newTestConnectServerWithBroker(t *testing.T, st *store.Store, broker *events.Broker) *httptest.Server {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	kube := fake.NewClientBuilder().WithScheme(scheme).Build()
	reg := registry.Default()
	svc := service.NewWithEvents(reg, st, kube, broker)
	mux := http.NewServeMux()
	Register(mux, NewResourceServer(svc, reg, st, broker, nil))
	return httptest.NewServer(mux)
}

func receiveEvent(t *testing.T, stream *connect.ServerStreamForClient[ResourceEvent], eventType events.EventType, name string, publish func()) bool {
	t.Helper()
	result := make(chan bool, 1)
	go func() {
		for stream.Receive() {
			msg := stream.Msg()
			if msg.Type == eventType && msg.Dirty != nil && msg.Dirty.Name == name {
				result <- true
				return
			}
		}
		result <- false
	}()
	time.Sleep(20 * time.Millisecond)
	publish()
	select {
	case ok := <-result:
		if err := stream.Err(); err != nil {
			t.Fatalf("stream error: %v", err)
		}
		return ok
	case <-time.After(time.Second):
		return false
	}
}
