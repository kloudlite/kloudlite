package types

type Action string

const (
	ActionApply  Action = "apply"
	ActionDelete Action = "delete"
)

type AgentMessage struct {
	AccountName string `json:"accountName"`
	ClusterName string `json:"clusterName"`

	Action Action `json:"action"`
	// Yamls  []byte        `json:"yamls,omitempty"`
	Object map[string]any `json:"object"`
}

type AgentErrMessage struct {
	AccountName string `json:"accountName"`
	ClusterName string `json:"clusterName"`

	Error  error          `json:"error"`
	Action Action         `json:"action"`
	Object map[string]any `json:"object"`
	// Yamls  []byte        `json:"yamls"`
}
