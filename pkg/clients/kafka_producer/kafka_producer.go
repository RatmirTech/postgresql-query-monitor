package kafka_producer

type KafkaProducerClient interface {
	Close()
	Produce(topic, message string) error
}
