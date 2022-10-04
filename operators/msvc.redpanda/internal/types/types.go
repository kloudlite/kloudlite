package types

type ACLUserCreds struct {
	KafkaBrokers string `json:"kafkaBrokers"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

type AdminUserCreds struct {
	AdminEndpoint string `json:"adminEndpoint"`
	KafkaBrokers  string `json:"kafkaBrokers"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}
