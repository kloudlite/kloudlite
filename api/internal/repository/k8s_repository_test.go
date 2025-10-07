package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestK8sClusterRepository_Create(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("successful create", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		repo := NewK8sClusterRepository[*corev1.Namespace, *corev1.NamespaceList](
			k8sClient,
			func() *corev1.Namespace { return &corev1.Namespace{} },
			func() *corev1.NamespaceList { return &corev1.NamespaceList{} },
		)

		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace",
			},
		}

		err := repo.Create(context.Background(), ns)
		assert.NoError(t, err)
	})

	t.Run("create already exists", func(t *testing.T) {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "existing-namespace",
			},
		}

		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()
		repo := NewK8sClusterRepository[*corev1.Namespace, *corev1.NamespaceList](
			k8sClient,
			func() *corev1.Namespace { return &corev1.Namespace{} },
			func() *corev1.NamespaceList { return &corev1.NamespaceList{} },
		)

		newNs := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "existing-namespace",
			},
		}

		err := repo.Create(context.Background(), newNs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestK8sClusterRepository_Get(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("get not found", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		repo := NewK8sClusterRepository[*corev1.Namespace, *corev1.NamespaceList](
			k8sClient,
			func() *corev1.Namespace { return &corev1.Namespace{} },
			func() *corev1.NamespaceList { return &corev1.NamespaceList{} },
		)

		_, err := repo.Get(context.Background(), "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestK8sClusterRepository_Update(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("update not found", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		repo := NewK8sClusterRepository[*corev1.Namespace, *corev1.NamespaceList](
			k8sClient,
			func() *corev1.Namespace { return &corev1.Namespace{} },
			func() *corev1.NamespaceList { return &corev1.NamespaceList{} },
		)

		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nonexistent",
			},
		}

		err := repo.Update(context.Background(), ns)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("update conflict", func(t *testing.T) {
		// Fake client DOES simulate conflicts with resource version mismatches
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "test-ns",
				ResourceVersion: "1",
			},
		}

		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()
		repo := NewK8sClusterRepository[*corev1.Namespace, *corev1.NamespaceList](
			k8sClient,
			func() *corev1.Namespace { return &corev1.Namespace{} },
			func() *corev1.NamespaceList { return &corev1.NamespaceList{} },
		)

		// Update with wrong version
		updatedNs := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "test-ns",
				ResourceVersion: "999",
			},
		}

		err := repo.Update(context.Background(), updatedNs)
		// Fake client simulates conflict errors
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "conflict")
	})
}

func TestK8sClusterRepository_Delete(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("delete not found", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		repo := NewK8sClusterRepository[*corev1.Namespace, *corev1.NamespaceList](
			k8sClient,
			func() *corev1.Namespace { return &corev1.Namespace{} },
			func() *corev1.NamespaceList { return &corev1.NamespaceList{} },
		)

		err := repo.Delete(context.Background(), "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestK8sClusterRepository_Patch(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("patch not found", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		repo := NewK8sClusterRepository[*corev1.Namespace, *corev1.NamespaceList](
			k8sClient,
			func() *corev1.Namespace { return &corev1.Namespace{} },
			func() *corev1.NamespaceList { return &corev1.NamespaceList{} },
		)

		patchData := map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"test": "value",
				},
			},
		}

		_, err := repo.Patch(context.Background(), "nonexistent", patchData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("patch successful", func(t *testing.T) {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-ns",
			},
		}

		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()
		repo := NewK8sClusterRepository[*corev1.Namespace, *corev1.NamespaceList](
			k8sClient,
			func() *corev1.Namespace { return &corev1.Namespace{} },
			func() *corev1.NamespaceList { return &corev1.NamespaceList{} },
		)

		patchData := map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"test": "value",
				},
			},
		}

		patched, err := repo.Patch(context.Background(), "test-ns", patchData)
		assert.NoError(t, err)
		assert.NotNil(t, patched)
	})
}

func TestK8sClusterRepository_List(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("list with label selector", func(t *testing.T) {
		ns1 := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ns1",
				Labels: map[string]string{
					"env": "prod",
				},
			},
		}

		ns2 := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ns2",
				Labels: map[string]string{
					"env": "dev",
				},
			},
		}

		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns1, ns2).Build()
		repo := NewK8sClusterRepository[*corev1.Namespace, *corev1.NamespaceList](
			k8sClient,
			func() *corev1.Namespace { return &corev1.Namespace{} },
			func() *corev1.NamespaceList { return &corev1.NamespaceList{} },
		)

		list, err := repo.List(context.Background(), WithLabelSelector("env=prod"))
		assert.NoError(t, err)
		assert.Len(t, list.Items, 1)
		assert.Equal(t, "ns1", list.Items[0].Name)
	})

	t.Run("list with limit", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		repo := NewK8sClusterRepository[*corev1.Namespace, *corev1.NamespaceList](
			k8sClient,
			func() *corev1.Namespace { return &corev1.Namespace{} },
			func() *corev1.NamespaceList { return &corev1.NamespaceList{} },
		)

		_, err := repo.List(context.Background(), WithLimit(10))
		assert.NoError(t, err)
	})

	t.Run("list with continue", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		repo := NewK8sClusterRepository[*corev1.Namespace, *corev1.NamespaceList](
			k8sClient,
			func() *corev1.Namespace { return &corev1.Namespace{} },
			func() *corev1.NamespaceList { return &corev1.NamespaceList{} },
		)

		_, err := repo.List(context.Background(), WithContinue("token-123"))
		assert.NoError(t, err)
	})
}

func TestK8sNamespacedRepository_Create(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("create already exists", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "existing-cm",
				Namespace: "default",
			},
		}

		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm).Build()
		repo := NewK8sNamespacedRepository[*corev1.ConfigMap, *corev1.ConfigMapList](
			k8sClient,
			func() *corev1.ConfigMap { return &corev1.ConfigMap{} },
			func() *corev1.ConfigMapList { return &corev1.ConfigMapList{} },
		)

		newCm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "existing-cm",
				Namespace: "default",
			},
		}

		err := repo.Create(context.Background(), newCm)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestK8sNamespacedRepository_Get(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("get not found", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		repo := NewK8sNamespacedRepository[*corev1.ConfigMap, *corev1.ConfigMapList](
			k8sClient,
			func() *corev1.ConfigMap { return &corev1.ConfigMap{} },
			func() *corev1.ConfigMapList { return &corev1.ConfigMapList{} },
		)

		_, err := repo.Get(context.Background(), "default", "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestK8sNamespacedRepository_Update(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("update not found", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		repo := NewK8sNamespacedRepository[*corev1.ConfigMap, *corev1.ConfigMapList](
			k8sClient,
			func() *corev1.ConfigMap { return &corev1.ConfigMap{} },
			func() *corev1.ConfigMapList { return &corev1.ConfigMapList{} },
		)

		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nonexistent",
				Namespace: "default",
			},
		}

		err := repo.Update(context.Background(), cm)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestK8sNamespacedRepository_Delete(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("delete not found", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		repo := NewK8sNamespacedRepository[*corev1.ConfigMap, *corev1.ConfigMapList](
			k8sClient,
			func() *corev1.ConfigMap { return &corev1.ConfigMap{} },
			func() *corev1.ConfigMapList { return &corev1.ConfigMapList{} },
		)

		err := repo.Delete(context.Background(), "default", "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestK8sNamespacedRepository_Patch(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("patch not found", func(t *testing.T) {
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		repo := NewK8sNamespacedRepository[*corev1.ConfigMap, *corev1.ConfigMapList](
			k8sClient,
			func() *corev1.ConfigMap { return &corev1.ConfigMap{} },
			func() *corev1.ConfigMapList { return &corev1.ConfigMapList{} },
		)

		patchData := map[string]interface{}{
			"data": map[string]interface{}{
				"key": "value",
			},
		}

		_, err := repo.Patch(context.Background(), "default", "nonexistent", patchData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("patch successful", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: "default",
			},
			Data: map[string]string{
				"key1": "value1",
			},
		}

		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm).Build()
		repo := NewK8sNamespacedRepository[*corev1.ConfigMap, *corev1.ConfigMapList](
			k8sClient,
			func() *corev1.ConfigMap { return &corev1.ConfigMap{} },
			func() *corev1.ConfigMapList { return &corev1.ConfigMapList{} },
		)

		patchData := map[string]interface{}{
			"data": map[string]interface{}{
				"key2": "value2",
			},
		}

		patched, err := repo.Patch(context.Background(), "default", "test-cm", patchData)
		assert.NoError(t, err)
		assert.NotNil(t, patched)
	})
}

func TestK8sNamespacedRepository_List(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	t.Run("list with options", func(t *testing.T) {
		cm1 := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cm1",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test",
				},
			},
		}

		cm2 := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cm2",
				Namespace: "default",
				Labels: map[string]string{
					"app": "prod",
				},
			},
		}

		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm1, cm2).Build()
		repo := NewK8sNamespacedRepository[*corev1.ConfigMap, *corev1.ConfigMapList](
			k8sClient,
			func() *corev1.ConfigMap { return &corev1.ConfigMap{} },
			func() *corev1.ConfigMapList { return &corev1.ConfigMapList{} },
		)

		list, err := repo.List(context.Background(), "default", WithLabelSelector("app=test"))
		assert.NoError(t, err)
		assert.Len(t, list.Items, 1)
		assert.Equal(t, "cm1", list.Items[0].Name)
	})
}

func TestErrorTypeCheckers(t *testing.T) {
	t.Run("IsNotFound", func(t *testing.T) {
		err := apierrors.NewNotFound(schema.GroupResource{}, "test")
		assert.True(t, apierrors.IsNotFound(err))

		err = apierrors.NewAlreadyExists(schema.GroupResource{}, "test")
		assert.False(t, apierrors.IsNotFound(err))
	})

	t.Run("IsAlreadyExists", func(t *testing.T) {
		err := apierrors.NewAlreadyExists(schema.GroupResource{}, "test")
		assert.True(t, apierrors.IsAlreadyExists(err))

		err = apierrors.NewNotFound(schema.GroupResource{}, "test")
		assert.False(t, apierrors.IsAlreadyExists(err))
	})

	t.Run("IsConflict", func(t *testing.T) {
		err := apierrors.NewConflict(schema.GroupResource{}, "test", nil)
		assert.True(t, apierrors.IsConflict(err))

		err = apierrors.NewNotFound(schema.GroupResource{}, "test")
		assert.False(t, apierrors.IsConflict(err))
	})
}
