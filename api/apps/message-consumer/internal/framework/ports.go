package framework

type Config struct {
	IsDev bool
	KafkaBrokers    string
	ConsumerGroupId string
	TopicPrefix     string
}
