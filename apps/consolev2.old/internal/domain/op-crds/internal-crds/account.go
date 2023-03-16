package internal_crds

type AccountSpec struct {
	AccountId    string   `json:"accountId,omitempty"`
	OwnedDomains []string `json:"ownedDomains,omitempty"`
}

type AccountMetadata struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

const AccountAPIVersion = "management.kloudlite.io/v1"
const AccountKind = "Account"

type Account struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`

	Metadata AccountMetadata `json:"metadata"`
	Spec     AccountSpec     `json:"spec,omitempty"`
}
