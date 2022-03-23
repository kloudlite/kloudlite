package app

type App interface {
	Handle(msg *Message) error
}

// type Message struct {
// 	Action       string            `json:"action"`
// 	ResourceType string            `json:"resourceType"`
// 	ResourceId   string            `json:"resourceId"`
// 	Metadata     map[string]string `json:"metadata"`
// }

type Message struct {
	JobId        string            `json:"jobId"`
}

type M map[string]interface{}
