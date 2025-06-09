package v1

type GCPServiceAccount struct {
	Enabled bool     `json:"enabled"`
	Email   *string  `json:"email,omitempty"`
	Scopes  []string `json:"scopes,omitempty"`
}
