package op_crds

type routes struct {
	Path string `json:"path"`
	App  string `json:"app"`
	Port uint16 `json:"port"`
}

type RouterSpec struct {
	Domains []string `json:"domains"`
	Routes  []routes `json:"routes"`
}

type Router struct {
	Name      string     `json:"name"`
	NameSpace string     `json:"nameSpace"`
	Spec      RouterSpec `json:"spec,omitempty"`
	Status    Status     `json:"status,omitempty"`
}
