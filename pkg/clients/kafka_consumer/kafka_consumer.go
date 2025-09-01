package kafka_consumer

import "github.com/confluentinc/confluent-kafka-go/kafka"

type KafkaConsumerClient interface {
	Close()
	Subscribe(topic string) error
	SubscribeMany(topics ...string) error
	ReadMessage() (*kafka.Message, error)
}
