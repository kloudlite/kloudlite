package messaging

type KafkaClient interface {
	GetBrokers() string
}

type kafkaClient struct {
	Brokers string
}

func (k *kafkaClient) GetBrokers() string {
	return k.Brokers
}

func NewKafkaClient(brokers string) KafkaClient {
	return &kafkaClient{
		Brokers: brokers,
	}
}
