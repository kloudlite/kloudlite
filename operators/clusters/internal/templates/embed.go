package templates

import (
	"embed"
	"path/filepath"

	"github.com/kloudlite/operator/pkg/templates"
)

//go:embed *
var templatesDir embed.FS

type templateFile string

const (
	ClusterJobTemplate        templateFile = "./cluster-job.yml.tpl"
	S3BucketJobTemplate       templateFile = "./s3-bucket-job.yml.tpl"
	RBACForClusterJobTemplate templateFile = "./rbac-for-cluster-job.yml.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

var ParseBytes = templates.ParseBytes
