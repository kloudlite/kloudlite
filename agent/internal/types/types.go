package types

type AgentMessage struct {
	Action  string         `json:"action"`
	Payload map[string]any `json:"payload,omitempty"`
	Yamls   []byte         `json:"yamls,omitempty"`
}

type ErrMessage struct {
	Error  string `json:"error"`
	Action string `json:"action"`
	// Payload map[string]any `json:"payload"`
	Payload []byte `json:"payload"`
}
