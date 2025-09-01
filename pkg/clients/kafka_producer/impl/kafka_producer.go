package impl

import (
	"github.com/dreadew/go-common/pkg/clients/kafka_producer"
	"github.com/dreadew/go-common/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

type kafkaProducerClient struct {
	instance *kafka.Producer
}

func New(addr string) (kafka_producer.KafkaProducerClient, error) {
	logger := logger.GetLogger()

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": addr,
	})
	if err != nil {
		logger.Error("error while creating kafka producer client", zap.String("error", err.Error()))
		return nil, err
	}

	return &kafkaProducerClient{
		instance: producer,
	}, nil
}

func (k *kafkaProducerClient) Close() {
	k.instance.Close()
}

func (k *kafkaProducerClient) Produce(topic, message string) error {
	err := k.instance.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(message),
	}, nil)

	if err != nil {
		return err
	}

	k.instance.Flush(15 * 1000)

	return nil
}
