package env

type Env struct {
	//KafkaBrokers            string `env:"KAFKA_BROKERS" required:"true"`
	//KafkaUsername           string `env:"KAFKA_USERNAME" required:"true"`
	//KafkaPassword           string `env:"KAFKA_PASSWORD" required:"true"`
	//KafkaSubscriptionTopics string `env:"KAFKA_SUBSCRIPTION_TOPICS" required:"true"`
	//KafkaConsumerGroupId    string `env:"KAFKA_CONSUMER_GROUP_ID" required:"true"`
	EventsDbUri  string `env:"EVENTS_DB_URI" required:"true"`
	EventsDbName string `env:"EVENTS_DB_NAME" required:"true"`
	NatsURL      string `env:"NATS_URL" required:"true"`
	EventLogNatsStream   string `env:"EVENT_LOG_STREAM" required:"true"`
}
