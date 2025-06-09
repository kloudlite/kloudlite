package types

type ACLUserCreds struct {
	KafkaBrokers string `json:"KAFKA_BROKERS"`
	Username     string `json:"USERNAME"`
	Password     string `json:"PASSWORD"`
}

type AdminUserCreds struct {
	AdminEndpoint string `json:"ADMIN_ENDPOINT"`
	KafkaBrokers  string `json:"KAFKA_BROKERS"`
	Username      string `json:"USERNAME"`
	Password      string `json:"PASSWORD"`
	RpkAdminFlags string `json:"RPK_ADMIN_FLAGS"`
	RpkSASLFlags  string `json:"RPK_SASL_FLAGS"`
}
