package rpc

import "github.com/kloudlite/kloudlite/api/resources/events"

const ServiceName = "kloudlite.resources.v1.ResourceService"

const (
	ProcedureGet         = "/" + ServiceName + "/Get"
	ProcedureList        = "/" + ServiceName + "/List"
	ProcedureCreate      = "/" + ServiceName + "/Create"
	ProcedurePatch       = "/" + ServiceName + "/Patch"
	ProcedureDelete      = "/" + ServiceName + "/Delete"
	ProcedureWatchObject = "/" + ServiceName + "/WatchObject"
	ProcedureWatchList   = "/" + ServiceName + "/WatchList"
)

type GetResourceRequest struct {
	Resource  string `json:"resource"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
}

type ListResourcesRequest struct {
	Resource  string            `json:"resource"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Hash      string            `json:"hash,omitempty"`
}

type MutateResourceRequest struct {
	Resource  string          `json:"resource"`
	Namespace string          `json:"namespace,omitempty"`
	Name      string          `json:"name,omitempty"`
	Object    jsonRawResource `json:"object,omitempty"`
}

type DeleteResourceRequest struct {
	Resource  string `json:"resource"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
}

type WatchObjectRequest = GetResourceRequest
type WatchListRequest struct {
	Resource  string `json:"resource"`
	Namespace string `json:"namespace,omitempty"`
}

type ResourceObject struct {
	JSON            jsonRawResource `json:"json"`
	ResourceVersion string          `json:"resourceVersion,omitempty"`
}

type CacheInfo struct {
	Dirty []events.DirtyObject `json:"dirty,omitempty"`
}

type GetResourceResponse struct {
	Object ResourceObject `json:"object"`
	Cache  CacheInfo      `json:"cache"`
}

type ListResourcesResponse struct {
	Items []ResourceObject `json:"items"`
	Cache CacheInfo        `json:"cache"`
}

type MutateResourceResponse struct {
	Object ResourceObject `json:"object,omitempty"`
	Cache  CacheInfo      `json:"cache"`
}

type DeleteResourceResponse struct {
	Cache CacheInfo `json:"cache"`
}

type ResourceEvent struct {
	Type   events.EventType       `json:"type"`
	Object *ResourceObject        `json:"object,omitempty"`
	List   *ListResourcesResponse `json:"list,omitempty"`
	Dirty  *events.DirtyObject    `json:"dirty,omitempty"`
}

type jsonRawResource map[string]any
