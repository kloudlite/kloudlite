package types

type Action string

const (
	ActionApply   Action = "apply"
	ActionDelete  Action = "delete"
	ActionRestart Action = "restart"
)

type AgentMessage struct {
	AccountName string `json:"accountName"`
	ClusterName string `json:"clusterName"`

	Action Action         `json:"action"`
	Object map[string]any `json:"object"`
}

type AgentErrMessage struct {
	AccountName string `json:"accountName"`
	ClusterName string `json:"clusterName"`

	Error  string         `json:"error"`
	Action Action         `json:"action"`
	Object map[string]any `json:"object"`
}
