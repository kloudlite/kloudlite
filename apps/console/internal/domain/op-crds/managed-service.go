package op_crds

type ManagedServiceSpec struct {
	Values       string `json:"values"`
	TemplateName string `json:"templateName"`
}

type ManagedService struct {
	Name      string             `json:"name"`
	NameSpace string             `json:"nameSpace"`
	Spec      ManagedServiceSpec `json:"spec,omitempty"`
	Status    Status             `json:"status,omitempty"`
}
