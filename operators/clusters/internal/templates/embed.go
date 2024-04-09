package templates

import (
	"embed"
	"path/filepath"

	ct "github.com/kloudlite/operator/apis/common-types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/operator/pkg/templates"
)

//go:embed *
var templatesDir embed.FS

type templateFile string

const (
	ClusterJobTemplate        templateFile = "./cluster-job.yml.tpl"
	S3BucketJobTemplate       templateFile = "./s3-bucket-job.yml.tpl"
	RBACForClusterJobTemplate templateFile = "./rbac-for-cluster-job.yml.tpl"

	AwsVPCJob templateFile = "./aws-vpc-job.yml.tpl"
	GcpVPCJob templateFile = "./gcp-vpc-job.yml.tpl"
)

func Read(t templateFile) ([]byte, error) {
	return templatesDir.ReadFile(filepath.Join(string(t)))
}

func ParseBytes(b []byte, values any) ([]byte, error) {
	return templates.NewTextTemplate("clusters").ParseBytes(b, values)
}

type GcpVPCJobVars struct {
	JobMetadata metav1.ObjectMeta

	JobImage           string
	JobImagePullPolicy string

	TFStateSecretNamespace string
	TFStateSecretName      string

	ValuesJSON string

	CloudProvider ct.CloudProvider

	VPCOutputSecretName      string
	VPCOutputSecretNamespace string
}
