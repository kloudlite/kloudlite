package manifests

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/kloudlite/kloudlite/api/k8s"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

//go:embed local-path-storageclass.yaml
var storageClassYAML []byte

//go:embed local-path-provisioner-config.yaml
var provisionerConfigYAML []byte

// Install applies cluster-scoped manifests (StorageClasses, ConfigMaps) required by Kloudlite.
func Install(ctx context.Context, opts *k8s.ClientOptions) error {
	kube, err := k8s.NewClient(ctx, opts)
	if err != nil {
		return fmt.Errorf("create Kubernetes client for manifest install: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(kube.Config)
	if err != nil {
		return fmt.Errorf("create kubernetes clientset: %w", err)
	}

	if err := applyStorageClass(ctx, clientset, storageClassYAML); err != nil {
		return fmt.Errorf("apply StorageClass: %w", err)
	}

	if err := applyConfigMap(ctx, clientset, provisionerConfigYAML); err != nil {
		return fmt.Errorf("apply local-path-provisioner ConfigMap: %w", err)
	}

	return nil
}

func applyStorageClass(ctx context.Context, clientset kubernetes.Interface, data []byte) error {
	sc := &storagev1.StorageClass{}
	if err := yaml.Unmarshal(data, sc); err != nil {
		return fmt.Errorf("parse StorageClass: %w", err)
	}

	client := clientset.StorageV1().StorageClasses()
	existing, err := client.Get(ctx, sc.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.Create(ctx, sc, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	sc.ResourceVersion = existing.ResourceVersion
	_, err = client.Update(ctx, sc, metav1.UpdateOptions{})
	return err
}

func applyConfigMap(ctx context.Context, clientset kubernetes.Interface, data []byte) error {
	cm := &corev1.ConfigMap{}
	if err := yaml.Unmarshal(data, cm); err != nil {
		return fmt.Errorf("parse ConfigMap: %w", err)
	}

	client := clientset.CoreV1().ConfigMaps(cm.Namespace)
	existing, err := client.Get(ctx, cm.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.Create(ctx, cm, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	cm.ResourceVersion = existing.ResourceVersion
	_, err = client.Update(ctx, cm, metav1.UpdateOptions{})
	return err
}
