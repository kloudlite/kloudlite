package app

type App interface {
	Handle(msg *Message) error
}

type Message struct {
	Action       string            `json:"action"`
	ProjectId    string            `json:"projectId"`
	ResourceType string            `json:"resourceType"`
	Metadata     map[string]string `json:"metadata"`
}
