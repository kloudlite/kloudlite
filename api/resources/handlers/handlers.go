package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/resources/operations"
	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/service"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
)

type Handler struct {
	ops *operations.Operations
}

type NamespaceEnsurer = operations.NamespaceEnsurer

func New(svc *service.Service, reg *registry.Registry, st *store.Store, ensurer NamespaceEnsurer) *Handler {
	if reg == nil {
		reg = registry.Default()
	}
	return &Handler{ops: operations.New(svc, reg, st, ensurer)}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/resources/:resource", h.listCluster)
	group.GET("/resources/:resource/:name", h.getCluster)
	group.POST("/resources/:resource", h.createCluster)
	group.PATCH("/resources/:resource/:name", h.patchCluster)
	group.DELETE("/resources/:resource/:name", h.deleteCluster)

	group.GET("/namespaces/:namespace/resources/:resource", h.listNamespaced)
	group.GET("/namespaces/:namespace/resources/:resource/:name", h.getNamespaced)
	group.POST("/namespaces/:namespace/resources/:resource", h.createNamespaced)
	group.PATCH("/namespaces/:namespace/resources/:resource/:name", h.patchNamespaced)
	group.DELETE("/namespaces/:namespace/resources/:resource/:name", h.deleteNamespaced)
}

func (h *Handler) listCluster(c *gin.Context) {
	h.list(c, registry.Cluster, "")
}

func (h *Handler) getCluster(c *gin.Context) {
	h.get(c, registry.Cluster, "")
}

func (h *Handler) createCluster(c *gin.Context) {
	h.create(c, registry.Cluster, "")
}

func (h *Handler) patchCluster(c *gin.Context) {
	h.patch(c, registry.Cluster, "")
}

func (h *Handler) deleteCluster(c *gin.Context) {
	h.delete(c, registry.Cluster, "")
}

func (h *Handler) listNamespaced(c *gin.Context) {
	h.list(c, registry.Namespaced, c.Param("namespace"))
}

func (h *Handler) getNamespaced(c *gin.Context) {
	h.get(c, registry.Namespaced, c.Param("namespace"))
}

func (h *Handler) createNamespaced(c *gin.Context) {
	h.create(c, registry.Namespaced, c.Param("namespace"))
}

func (h *Handler) patchNamespaced(c *gin.Context) {
	h.patch(c, registry.Namespaced, c.Param("namespace"))
}

func (h *Handler) deleteNamespaced(c *gin.Context) {
	h.delete(c, registry.Namespaced, c.Param("namespace"))
}

func (h *Handler) list(c *gin.Context, scope registry.Scope, namespace string) {
	alias := c.Param("resource")
	if err := validatePathParams(scope, namespace, ""); err != nil {
		h.error(c, err)
		return
	}
	selector, err := parseSelector(c)
	if err != nil {
		h.error(c, err)
		return
	}
	result, err := h.ops.List(c.Request.Context(), operations.ListRequest{Resource: alias, Namespace: namespace, Scope: scope, Selector: selector})
	if err != nil {
		h.error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": result.Items})
}

func (h *Handler) get(c *gin.Context, scope registry.Scope, namespace string) {
	alias := c.Param("resource")
	if err := validatePathParams(scope, namespace, c.Param("name")); err != nil {
		h.error(c, err)
		return
	}
	result, err := h.ops.Get(c.Request.Context(), operations.GetRequest{Resource: alias, Namespace: namespace, Name: c.Param("name"), Scope: scope})
	if err != nil {
		h.error(c, err)
		return
	}
	c.JSON(http.StatusOK, result.Object)
}

func (h *Handler) create(c *gin.Context, scope registry.Scope, namespace string) {
	alias := c.Param("resource")
	if err := validatePathParams(scope, namespace, ""); err != nil {
		h.error(c, err)
		return
	}

	object, err := decodeObject(c)
	if err != nil {
		h.error(c, err)
		return
	}
	if scope == registry.Namespaced {
		object.SetNamespace(namespace)
	}

	created, err := h.ops.Create(c.Request.Context(), operations.MutateRequest{Resource: alias, Namespace: namespace, Scope: scope, Object: object})
	if err != nil {
		h.error(c, err)
		return
	}
	c.JSON(http.StatusCreated, created.Object)
}

func (h *Handler) patch(c *gin.Context, scope registry.Scope, namespace string) {
	alias := c.Param("resource")
	if err := validatePathParams(scope, namespace, c.Param("name")); err != nil {
		h.error(c, err)
		return
	}

	object, err := decodeObject(c)
	if err != nil {
		h.error(c, err)
		return
	}
	patched, err := h.ops.Patch(c.Request.Context(), operations.MutateRequest{Resource: alias, Namespace: namespace, Name: c.Param("name"), Scope: scope, Object: object})
	if err != nil {
		h.error(c, err)
		return
	}
	c.JSON(http.StatusOK, patched.Object)
}

func (h *Handler) delete(c *gin.Context, scope registry.Scope, namespace string) {
	alias := c.Param("resource")
	if err := validatePathParams(scope, namespace, c.Param("name")); err != nil {
		h.error(c, err)
		return
	}
	if _, err := h.ops.Delete(c.Request.Context(), operations.DeleteRequest{Resource: alias, Namespace: namespace, Name: c.Param("name"), Scope: scope}); err != nil {
		h.error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) error(c *gin.Context, err error) {
	c.JSON(service.HTTPStatus(err), gin.H{"error": err.Error()})
}

func parseSelector(c *gin.Context) (store.Selector, error) {
	selector := store.Selector{}
	if hash := c.Query("hash"); hash != "" {
		selector.Hash = hash
	} else if hash := c.Query("kloudlite.io/hash"); hash != "" {
		selector.Hash = hash
	}

	labelSelector := c.Query("labelSelector")
	if labelSelector == "" {
		return selector, nil
	}
	parsed, err := labels.Parse(labelSelector)
	if err != nil {
		return selector, service.NewError(service.ErrBadRequest, "invalid labelSelector", err)
	}
	requirements, selectable := parsed.Requirements()
	if !selectable {
		return selector, service.NewError(service.ErrBadRequest, "unsupported labelSelector", nil)
	}
	selector.Labels = map[string]string{}
	for _, requirement := range requirements {
		if requirement.Operator() != selection.Equals && requirement.Operator() != selection.DoubleEquals {
			return selector, service.NewError(service.ErrBadRequest, fmt.Sprintf("unsupported labelSelector requirement %q", requirement.String()), nil)
		}
		values := requirement.ValuesUnsorted()
		if len(values) != 1 {
			return selector, service.NewError(service.ErrBadRequest, fmt.Sprintf("unsupported labelSelector requirement %q", requirement.String()), nil)
		}
		selector.Labels[requirement.Key()] = values[0]
	}
	return selector, nil
}

func validatePathParams(scope registry.Scope, namespace string, name string) error {
	if scope == registry.Namespaced {
		if namespace == "" {
			return service.NewError(service.ErrBadRequest, "namespace is required", nil)
		}
		if errs := utilvalidation.IsDNS1123Label(namespace); len(errs) > 0 {
			return service.NewError(service.ErrBadRequest, "invalid namespace: "+errs[0], nil)
		}
	}
	if name == "" {
		return nil
	}
	if errs := utilvalidation.IsDNS1123Subdomain(name); len(errs) > 0 {
		return service.NewError(service.ErrBadRequest, "invalid name: "+errs[0], nil)
	}
	return nil
}

func decodeObject(c *gin.Context) (*unstructured.Unstructured, error) {
	object := &unstructured.Unstructured{}
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(object); err != nil {
		return nil, service.NewError(service.ErrBadRequest, "invalid JSON body", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, service.NewError(service.ErrBadRequest, "invalid JSON body", err)
	}
	return object, nil
}
