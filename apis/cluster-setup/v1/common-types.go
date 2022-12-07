package v1

import (
	corev1 "k8s.io/api/core/v1"
	ct "operators.kloudlite.io/apis/common-types"
)

type S3 struct {
	AwsAccessKeyId     string `json:"awsAccessKeyId"`
	AwsSecretAccessKey string `json:"awsSecretAccessKey"`
	BucketName         string `json:"bucketName"`
	Endpoint           string `json:"endpoint"`
}

type LokiValues struct {
	S3        S3           `json:"s3"`
	Resources ct.Resources `json:"resources"`
	Url       string       `json:"url,omitempty"`
}

type PrometheusValues struct {
	Resources ct.Resources `json:"resources"`
}

// +kubebuilder:object:generate=true

type Cloudflare struct {
	Email        string          `json:"email"`
	SecretKeyRef ct.SecretKeyRef `json:"secretKeyRef"`
	DnsNames     []string        `json:"dnsNames"`
}

// +kubebuilder:object:generate=true

type ClusterIssuer struct {
	Name         string      `json:"name"`
	AcmeEmail    string      `json:"acmeEmail"`
	Cloudflare   *Cloudflare `json:"cloudflare,omitempty"`
	IngressClass string      `json:"ingressClass"`
}

type CertManagerValues struct {
	Tolerations   []corev1.Toleration `json:"tolerations,omitempty"`
	NodeSelector  map[string]string   `json:"nodeSelector,omitempty"`
	PodLabels     map[string]string   `json:"podLabels,omitempty"`
	ClusterIssuer ClusterIssuer       `json:"clusterIssuer"`
}

type IngressValues struct {
	ClassName    string              `json:"className"`
	PodLabels    map[string]string   `json:"podLabels,omitempty"`
	Resources    ct.Resources        `json:"resources"`
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
}

type GithubReleaseArtifacts struct {
	// Github Repo in form of <owner>/<rep-name>
	Repo string `json:"repo"`
	// Github Release Tag
	Tag string `json:"tag"`
	// list of artifact names that we want to refer
	Artifacts []string `json:"artifacts,omitempty"`

	TokenSecret SecretKeyReference `json:"ghTokenSecret"`
}

type Operators struct {
	Manifests []GithubReleaseArtifacts `json:"manifests"`
}
