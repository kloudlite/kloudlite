package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/service"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
)

type Handler struct {
	service  *service.Service
	registry *registry.Registry
	ensurer  NamespaceEnsurer
}

type NamespaceEnsurer interface {
	EnsureNamespaced(ctx context.Context, alias string, namespace string) error
}

func New(svc *service.Service, reg *registry.Registry, ensurer NamespaceEnsurer) *Handler {
	if reg == nil {
		reg = registry.Default()
	}
	return &Handler{service: svc, registry: reg, ensurer: ensurer}
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
	if err := h.requireScope(alias, scope); err != nil {
		h.error(c, err)
		return
	}
	if err := validatePathParams(scope, namespace, ""); err != nil {
		h.error(c, err)
		return
	}
	selector, err := parseSelector(c)
	if err != nil {
		h.error(c, err)
		return
	}
	if err := h.ensureNamespaced(c, alias, scope, namespace); err != nil {
		h.error(c, err)
		return
	}

	items, err := h.service.List(c.Request.Context(), alias, namespace, selector)
	if err != nil {
		h.error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) get(c *gin.Context, scope registry.Scope, namespace string) {
	alias := c.Param("resource")
	if err := h.requireScope(alias, scope); err != nil {
		h.error(c, err)
		return
	}
	if err := validatePathParams(scope, namespace, c.Param("name")); err != nil {
		h.error(c, err)
		return
	}
	if err := h.ensureNamespaced(c, alias, scope, namespace); err != nil {
		h.error(c, err)
		return
	}

	object, err := h.service.Get(c.Request.Context(), alias, namespace, c.Param("name"))
	if err != nil {
		h.error(c, err)
		return
	}
	c.JSON(http.StatusOK, object)
}

func (h *Handler) create(c *gin.Context, scope registry.Scope, namespace string) {
	alias := c.Param("resource")
	if err := h.requireScope(alias, scope); err != nil {
		h.error(c, err)
		return
	}
	if err := validatePathParams(scope, namespace, ""); err != nil {
		h.error(c, err)
		return
	}
	if err := h.ensureNamespaced(c, alias, scope, namespace); err != nil {
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

	created, err := h.service.Create(c.Request.Context(), alias, namespace, object)
	if err != nil {
		h.error(c, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) patch(c *gin.Context, scope registry.Scope, namespace string) {
	alias := c.Param("resource")
	if err := h.requireScope(alias, scope); err != nil {
		h.error(c, err)
		return
	}
	if err := validatePathParams(scope, namespace, c.Param("name")); err != nil {
		h.error(c, err)
		return
	}
	if err := h.ensureNamespaced(c, alias, scope, namespace); err != nil {
		h.error(c, err)
		return
	}

	object, err := decodeObject(c)
	if err != nil {
		h.error(c, err)
		return
	}
	patched, err := h.service.Patch(c.Request.Context(), alias, namespace, c.Param("name"), object)
	if err != nil {
		h.error(c, err)
		return
	}
	c.JSON(http.StatusOK, patched)
}

func (h *Handler) delete(c *gin.Context, scope registry.Scope, namespace string) {
	alias := c.Param("resource")
	if err := h.requireScope(alias, scope); err != nil {
		h.error(c, err)
		return
	}
	if err := validatePathParams(scope, namespace, c.Param("name")); err != nil {
		h.error(c, err)
		return
	}
	if err := h.ensureNamespaced(c, alias, scope, namespace); err != nil {
		h.error(c, err)
		return
	}

	if err := h.service.Delete(c.Request.Context(), alias, namespace, c.Param("name")); err != nil {
		h.error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) requireScope(alias string, scope registry.Scope) error {
	resource, ok := h.registry.Get(alias)
	if !ok {
		return service.UnknownResource(alias)
	}
	if resource.Scope != scope {
		return service.NewError(service.ErrWrongScope, "resource \""+alias+"\" is "+string(resource.Scope)+"-scoped, not "+string(scope)+"-scoped", nil)
	}
	return nil
}

func (h *Handler) ensureNamespaced(c *gin.Context, alias string, scope registry.Scope, namespace string) error {
	if scope != registry.Namespaced || h.ensurer == nil {
		return nil
	}
	return h.ensurer.EnsureNamespaced(c.Request.Context(), alias, namespace)
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
