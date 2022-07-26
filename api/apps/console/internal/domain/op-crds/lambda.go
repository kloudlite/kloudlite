package op_crds

type LambdaSpec struct {
	Containers   []Container       `json:"containers,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type LambdaMetadata struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

const LambdaAPIVersion = "serverless.kloudlite.io/v1"
const LambdaKind = "Lambda"

type Lambda struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`

	Metadata LambdaMetadata `json:"metadata"`
	Spec     LambdaSpec     `json:"spec,omitempty"`
}
