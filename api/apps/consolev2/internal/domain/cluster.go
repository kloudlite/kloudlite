package domain

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"text/template"

	"kloudlite.io/apps/consolev2/internal/domain/entities"
	text_templates "kloudlite.io/pkg/text-templates"
)

var (
	//go:embed templates
	dirTemplates embed.FS
)

func fxClusterTemplate() (*template.Template, error) {
	t := template.New("cluster")
	t = text_templates.WithFunctions(t)
	if _, err := t.ParseFS(dirTemplates, "templates/secret.yml.tpl", "templates/cluster.yml.tpl"); err != nil {
		return nil, err
	}
	return t, nil
}

func (d *domain) AddNewCluster(ctx context.Context, name string, subDomain string, kubeConfig string) (*entities.Cluster, error) {
	cluster, err := d.clusterRepo.Create(
		ctx, &entities.Cluster{
			Name:       name,
			SubDomain:  subDomain,
			KubeConfig: base64.RawStdEncoding.EncodeToString([]byte(kubeConfig)),
		},
	)
	if err != nil {
		return nil, err
	}

	b := new(bytes.Buffer)
	if err := d.consoleTemplate.ExecuteTemplate(
		b, "secret.yml.tpl", map[string]any{
			"name":       getClusterKubeConfig(string(cluster.Id)),
			"namespace":  fmt.Sprintf("kl-core"),
			"kubeconfig": kubeConfig,
		},
	); err != nil {
		return nil, err
	}

	b2 := new(bytes.Buffer)
	if err := d.consoleTemplate.ExecuteTemplate(
		b2, "cluster.yml.tpl", map[string]any{
			"name":                 cluster.Id,
			"kubeconfig-name":      getClusterKubeConfig(string(cluster.Id)),
			"kubeconfig-namespace": "kl-core",
		},
	); err != nil {
		return nil, err
	}

	if err := d.k8sYamlClient.ApplyYAML(ctx, b.Bytes(), b2.Bytes()); err != nil {
		return nil, err
	}

	return cluster, nil
}
