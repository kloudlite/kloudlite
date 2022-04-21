package op_crds

type ManagedResourceSpec struct {
	Type       string `json:"type"`
	ManagedSvc string `json:"managedSvc"`
	Values     string `json:"values,omitempty"`
}

type ManagedResource struct {
	Name      string              `json:"name"`
	NameSpace string              `json:"nameSpace"`
	Spec      ManagedResourceSpec `json:"spec,omitempty"`
	Status    Status              `json:"status,omitempty"`
}
