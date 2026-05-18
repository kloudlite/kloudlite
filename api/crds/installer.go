package crds

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kloudlite/kloudlite/api/crds/assets"
	"github.com/kloudlite/kloudlite/api/k8s"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apixclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func Install(ctx context.Context, opts *k8s.ClientOptions, dir string) error {
	kube, err := k8s.NewClient(ctx, opts)
	if err != nil {
		return fmt.Errorf("create Kubernetes client for CRD install: %w", err)
	}
	clientset, err := apixclient.NewForConfig(kube.Config)
	if err != nil {
		return fmt.Errorf("create apiextensions client: %w", err)
	}
	crds, err := Load(dir)
	if err != nil {
		return err
	}

	for i := range crds {
		if err := applyCRD(ctx, clientset, &crds[i]); err != nil {
			return fmt.Errorf("apply CRD %q: %w", crds[i].Name, err)
		}
	}
	return nil
}

func Load(dir string) ([]apixv1.CustomResourceDefinition, error) {
	if dir == "" {
		return assets.All()
	}
	return loadFromDir(dir)
}

func loadFromDir(dir string) ([]apixv1.CustomResourceDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read CRD manifest dir %q: %w", dir, err)
	}

	crds := make([]apixv1.CustomResourceDefinition, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !(strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml")) {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read CRD manifest %q: %w", path, err)
		}
		crd := &apixv1.CustomResourceDefinition{}
		if err := yaml.Unmarshal(data, crd); err != nil {
			return nil, fmt.Errorf("parse CRD manifest %q: %w", path, err)
		}
		if crd.Kind != "CustomResourceDefinition" || crd.Name == "" {
			continue
		}
		crds = append(crds, *crd)
	}
	if len(crds) == 0 {
		return nil, fmt.Errorf("no CRD manifests found in %q", dir)
	}
	return crds, nil
}

func applyCRD(ctx context.Context, clientset *apixclient.Clientset, crd *apixv1.CustomResourceDefinition) error {
	client := clientset.ApiextensionsV1().CustomResourceDefinitions()
	existing, err := client.Get(ctx, crd.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.Create(ctx, crd, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	crd.ResourceVersion = existing.ResourceVersion
	_, err = client.Update(ctx, crd, metav1.UpdateOptions{})
	return err
}
