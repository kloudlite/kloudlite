package rpc

import (
	"context"
	"encoding/json"
	"net/http"

	"connectrpc.com/connect"
	"github.com/kloudlite/kloudlite/api/resources/events"
	"github.com/kloudlite/kloudlite/api/resources/operations"
	"github.com/kloudlite/kloudlite/api/resources/projection"
	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/service"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceServer struct {
	service  *service.Service
	registry *registry.Registry
	store    *store.Store
	broker   *events.Broker
	ensurer  namespaceEnsurer
	ops      *operations.Operations
}

type namespaceEnsurer interface {
	EnsureNamespaced(ctx context.Context, alias string, namespace string) error
}

func NewResourceServer(svc *service.Service, reg *registry.Registry, st *store.Store, broker *events.Broker, ensurer namespaceEnsurer) *ResourceServer {
	if broker == nil {
		broker = events.NewBroker()
	}
	if reg == nil {
		reg = registry.Default()
	}
	return &ResourceServer{service: svc, registry: reg, store: st, broker: broker, ensurer: ensurer, ops: operations.New(svc, reg, st, ensurer)}
}

func Register(mux interface{ Handle(string, http.Handler) }, server *ResourceServer, opts ...connect.HandlerOption) {
	options := append([]connect.HandlerOption{connect.WithCodec(jsonCodec{})}, opts...)
	mux.Handle(ProcedureGet, connect.NewUnaryHandler(ProcedureGet, server.Get, options...))
	mux.Handle(ProcedureList, connect.NewUnaryHandler(ProcedureList, server.List, options...))
	mux.Handle(ProcedureCreate, connect.NewUnaryHandler(ProcedureCreate, server.Create, options...))
	mux.Handle(ProcedurePatch, connect.NewUnaryHandler(ProcedurePatch, server.Patch, options...))
	mux.Handle(ProcedureDelete, connect.NewUnaryHandler(ProcedureDelete, server.Delete, options...))
	mux.Handle(ProcedureWatchObject, connect.NewServerStreamHandler(ProcedureWatchObject, server.WatchObject, options...))
	mux.Handle(ProcedureWatchList, connect.NewServerStreamHandler(ProcedureWatchList, server.WatchList, options...))
}

func Handler(procedure string, server *ResourceServer, opts ...connect.HandlerOption) http.Handler {
	options := append([]connect.HandlerOption{connect.WithCodec(jsonCodec{})}, opts...)
	switch procedure {
	case ProcedureGet:
		return connect.NewUnaryHandler(ProcedureGet, server.Get, options...)
	case ProcedureList:
		return connect.NewUnaryHandler(ProcedureList, server.List, options...)
	case ProcedureCreate:
		return connect.NewUnaryHandler(ProcedureCreate, server.Create, options...)
	case ProcedurePatch:
		return connect.NewUnaryHandler(ProcedurePatch, server.Patch, options...)
	case ProcedureDelete:
		return connect.NewUnaryHandler(ProcedureDelete, server.Delete, options...)
	case ProcedureWatchObject:
		return connect.NewServerStreamHandler(ProcedureWatchObject, server.WatchObject, options...)
	case ProcedureWatchList:
		return connect.NewServerStreamHandler(ProcedureWatchList, server.WatchList, options...)
	default:
		return http.NotFoundHandler()
	}
}

func (s *ResourceServer) Get(ctx context.Context, req *connect.Request[GetResourceRequest]) (*connect.Response[GetResourceResponse], error) {
	if req.Msg.Resource == "machinetypes" {
		result, err := s.ops.Get(ctx, operations.GetRequest{Resource: req.Msg.Resource, Name: req.Msg.Name, Scope: registry.Cluster})
		if err != nil {
			return nil, connectError(err)
		}
		return connect.NewResponse(&GetResourceResponse{Object: mustResourceObject(result.Object)}), nil
	}
	if err := s.ensure(ctx, req.Msg.Resource, req.Msg.Namespace); err != nil {
		return nil, connectError(err)
	}
	object, err := s.service.Get(ctx, req.Msg.Resource, req.Msg.Namespace, req.Msg.Name)
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&GetResourceResponse{Object: mustResourceObject(object), Cache: s.cacheForObject(req.Msg.Resource, req.Msg.Namespace, req.Msg.Name)}), nil
}

func (s *ResourceServer) List(ctx context.Context, req *connect.Request[ListResourcesRequest]) (*connect.Response[ListResourcesResponse], error) {
	if req.Msg.Resource == "machinetypes" {
		result, err := s.ops.List(ctx, operations.ListRequest{Resource: req.Msg.Resource, Scope: registry.Cluster})
		if err != nil {
			return nil, connectError(err)
		}
		items := make([]ResourceObject, 0, len(result.Items))
		for _, item := range result.Items {
			items = append(items, mustResourceObject(item))
		}
		return connect.NewResponse(&ListResourcesResponse{Items: items}), nil
	}
	if err := s.ensure(ctx, req.Msg.Resource, req.Msg.Namespace); err != nil {
		return nil, connectError(err)
	}
	selector := store.Selector{Labels: req.Msg.Labels, Hash: req.Msg.Hash}
	items, err := s.service.List(ctx, req.Msg.Resource, req.Msg.Namespace, selector)
	if err != nil {
		return nil, connectError(err)
	}
	objects := make([]ResourceObject, 0, len(items))
	for _, item := range items {
		objects = append(objects, mustResourceObject(item))
	}
	return connect.NewResponse(&ListResourcesResponse{Items: objects, Cache: CacheInfo{Dirty: s.store.DirtyForList(req.Msg.Resource, req.Msg.Namespace, selector)}}), nil
}

func (s *ResourceServer) Create(ctx context.Context, req *connect.Request[MutateResourceRequest]) (*connect.Response[MutateResourceResponse], error) {
	if req.Msg.Resource == "machinetypes" {
		_, err := s.ops.Create(ctx, operations.MutateRequest{Resource: req.Msg.Resource, Scope: registry.Cluster, Object: unstructuredFromRaw(req.Msg.Object)})
		return nil, connectError(err)
	}
	if err := s.ensure(ctx, req.Msg.Resource, req.Msg.Namespace); err != nil {
		return nil, connectError(err)
	}
	created, err := s.service.Create(ctx, req.Msg.Resource, req.Msg.Namespace, unstructuredFromRaw(req.Msg.Object))
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&MutateResourceResponse{Object: mustResourceObject(created), Cache: s.cacheForObject(req.Msg.Resource, created.GetNamespace(), created.GetName())}), nil
}

func (s *ResourceServer) Patch(ctx context.Context, req *connect.Request[MutateResourceRequest]) (*connect.Response[MutateResourceResponse], error) {
	if req.Msg.Resource == "machinetypes" {
		_, err := s.ops.Patch(ctx, operations.MutateRequest{Resource: req.Msg.Resource, Name: req.Msg.Name, Scope: registry.Cluster, Object: unstructuredFromRaw(req.Msg.Object)})
		return nil, connectError(err)
	}
	if err := s.ensure(ctx, req.Msg.Resource, req.Msg.Namespace); err != nil {
		return nil, connectError(err)
	}
	patched, err := s.service.Patch(ctx, req.Msg.Resource, req.Msg.Namespace, req.Msg.Name, unstructuredFromRaw(req.Msg.Object))
	if err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&MutateResourceResponse{Object: mustResourceObject(patched), Cache: s.cacheForObject(req.Msg.Resource, patched.GetNamespace(), patched.GetName())}), nil
}

func (s *ResourceServer) Delete(ctx context.Context, req *connect.Request[DeleteResourceRequest]) (*connect.Response[DeleteResourceResponse], error) {
	if req.Msg.Resource == "machinetypes" {
		_, err := s.ops.Delete(ctx, operations.DeleteRequest{Resource: req.Msg.Resource, Name: req.Msg.Name, Scope: registry.Cluster})
		return nil, connectError(err)
	}
	if err := s.ensure(ctx, req.Msg.Resource, req.Msg.Namespace); err != nil {
		return nil, connectError(err)
	}
	if err := s.service.Delete(ctx, req.Msg.Resource, req.Msg.Namespace, req.Msg.Name); err != nil {
		return nil, connectError(err)
	}
	return connect.NewResponse(&DeleteResourceResponse{Cache: s.cacheForObject(req.Msg.Resource, req.Msg.Namespace, req.Msg.Name)}), nil
}

func (s *ResourceServer) WatchObject(ctx context.Context, req *connect.Request[WatchObjectRequest], stream *connect.ServerStream[ResourceEvent]) error {
	if req.Msg.Resource == "machinetypes" {
		result, err := s.ops.Get(ctx, operations.GetRequest{Resource: req.Msg.Resource, Name: req.Msg.Name, Scope: registry.Cluster})
		if err != nil && !service.IsKind(err, service.ErrNotFound) {
			return connectError(err)
		}
		event := ResourceEvent{Type: events.EventInitialState}
		if err == nil {
			object := mustResourceObject(result.Object)
			event.Object = &object
		}
		if err := stream.Send(&event); err != nil {
			return err
		}
		<-ctx.Done()
		return ctx.Err()
	}
	if err := s.ensure(ctx, req.Msg.Resource, req.Msg.Namespace); err != nil {
		return connectError(err)
	}
	ch := s.broker.Subscribe(ctx)
	object, err := s.service.Get(ctx, req.Msg.Resource, req.Msg.Namespace, req.Msg.Name)
	if err == nil {
		var dirtyPtr *events.DirtyObject
		if dirty, ok := s.store.DirtyForObject(req.Msg.Resource, req.Msg.Namespace, req.Msg.Name); ok {
			dirtyPtr = &dirty
		}
		initial := resourceEventFromObjectState(projection.ObjectInitialState(object, dirtyPtr))
		if err := stream.Send(&initial); err != nil {
			return err
		}
	} else if !service.IsKind(err, service.ErrNotFound) {
		return connectError(err)
	} else if err := stream.Send(&ResourceEvent{Type: events.EventInitialState}); err != nil {
		return err
	}
	return s.streamEventsFrom(ctx, stream, ch, func(event events.Event) bool {
		return event.Resource == req.Msg.Resource && event.Namespace == req.Msg.Namespace && event.Name == req.Msg.Name
	})
}

func (s *ResourceServer) WatchList(ctx context.Context, req *connect.Request[WatchListRequest], stream *connect.ServerStream[ResourceEvent]) error {
	if req.Msg.Resource == "machinetypes" {
		result, err := s.ops.List(ctx, operations.ListRequest{Resource: req.Msg.Resource, Scope: registry.Cluster})
		if err != nil {
			return connectError(err)
		}
		items := make([]ResourceObject, 0, len(result.Items))
		for _, item := range result.Items {
			items = append(items, mustResourceObject(item))
		}
		if err := sendInitialList(stream, &ListResourcesResponse{Items: items}); err != nil {
			return err
		}
		<-ctx.Done()
		return ctx.Err()
	}
	if err := s.ensure(ctx, req.Msg.Resource, req.Msg.Namespace); err != nil {
		return connectError(err)
	}
	ch := s.broker.Subscribe(ctx)
	list, err := s.List(ctx, connect.NewRequest(&ListResourcesRequest{Resource: req.Msg.Resource, Namespace: req.Msg.Namespace}))
	if err != nil {
		return err
	}
	if err := sendInitialList(stream, list.Msg); err != nil {
		return err
	}
	for _, dirty := range list.Msg.Cache.Dirty {
		if err := stream.Send(&ResourceEvent{Type: events.EventDirty, Dirty: &dirty}); err != nil {
			return err
		}
	}
	return s.streamEventsFrom(ctx, stream, ch, func(event events.Event) bool {
		return event.Resource == req.Msg.Resource && event.Namespace == req.Msg.Namespace
	})
}

func (s *ResourceServer) ensure(ctx context.Context, alias string, namespace string) error {
	resource, ok := s.registry.Get(alias)
	if !ok {
		return service.UnknownResource(alias)
	}
	if resource.Scope != registry.Namespaced {
		return nil
	}
	if namespace == "" {
		return service.NewError(service.ErrBadRequest, "namespace is required", nil)
	}
	if s.ensurer == nil {
		return nil
	}
	return s.ensurer.EnsureNamespaced(ctx, alias, namespace)
}

func (s *ResourceServer) streamEvents(ctx context.Context, stream *connect.ServerStream[ResourceEvent], matches func(events.Event) bool) error {
	return s.streamEventsFrom(ctx, stream, s.broker.Subscribe(ctx), matches)
}

func (s *ResourceServer) streamEventsFrom(ctx context.Context, stream *connect.ServerStream[ResourceEvent], ch <-chan events.Event, matches func(events.Event) bool) error {
	for event := range ch {
		if !matches(event) {
			continue
		}
		rpcEvent := ResourceEvent{Type: event.Type, Dirty: event.Dirty}
		if object, ok := event.Object.(client.Object); ok {
			resourceObject := mustResourceObject(object)
			rpcEvent.Object = &resourceObject
		}
		if err := stream.Send(&rpcEvent); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func (s *ResourceServer) cacheForObject(resource string, namespace string, name string) CacheInfo {
	if dirty, ok := s.store.DirtyForObject(resource, namespace, name); ok {
		return CacheInfo{Dirty: []events.DirtyObject{dirty}}
	}
	return CacheInfo{}
}

func resourceEventFromObjectState(state projection.ObjectState) ResourceEvent {
	event := ResourceEvent{Type: events.EventInitialState, Dirty: state.Dirty}
	if state.Object != nil {
		object := mustResourceObject(state.Object)
		event.Object = &object
	}
	return event
}

func sendInitialList(stream *connect.ServerStream[ResourceEvent], list *ListResourcesResponse) error {
	return stream.Send(&ResourceEvent{Type: events.EventInitialState, List: list})
}

func mustResourceObject(object client.Object) ResourceObject {
	data, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	var raw jsonRawResource
	if err := json.Unmarshal(data, &raw); err != nil {
		panic(err)
	}
	return ResourceObject{JSON: raw, ResourceVersion: object.GetResourceVersion()}
}

func unstructuredFromRaw(raw jsonRawResource) *unstructured.Unstructured {
	if raw == nil {
		raw = jsonRawResource{}
	}
	return &unstructured.Unstructured{Object: map[string]any(raw)}
}

func connectError(err error) error {
	switch service.HTTPStatus(err) {
	case http.StatusBadRequest:
		return connect.NewError(connect.CodeInvalidArgument, err)
	case http.StatusUnauthorized:
		return connect.NewError(connect.CodeUnauthenticated, err)
	case http.StatusNotFound:
		return connect.NewError(connect.CodeNotFound, err)
	case http.StatusConflict:
		return connect.NewError(connect.CodeAlreadyExists, err)
	case http.StatusServiceUnavailable:
		return connect.NewError(connect.CodeUnavailable, err)
	case http.StatusNotImplemented:
		return connect.NewError(connect.CodeUnimplemented, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
