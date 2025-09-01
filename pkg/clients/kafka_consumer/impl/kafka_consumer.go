package impl

import (
	"github.com/dreadew/go-common/pkg/clients/kafka_consumer"
	"github.com/dreadew/go-common/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

type kafkaConsumerClient struct {
	instance *kafka.Consumer
}

func New(addr, groupId string) (kafka_consumer.KafkaConsumerClient, error) {
	logger := logger.GetLogger()

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": addr,
		"group.id":          groupId,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		logger.Error("error while creating kafka consumer client", zap.String("error", err.Error()))
		return nil, err
	}

	return &kafkaConsumerClient{
		instance: consumer,
	}, nil
}

func (k *kafkaConsumerClient) Close() {
	k.instance.Close()
}

func (k *kafkaConsumerClient) SubscribeMany(topics ...string) error {
	err := k.instance.SubscribeTopics(topics, nil)
	if err != nil {
		return err
	}
	return nil
}

func (k *kafkaConsumerClient) Subscribe(topic string) error {
	err := k.instance.Subscribe(topic, nil)
	if err != nil {
		return err
	}
	return nil
}

func (k *kafkaConsumerClient) ReadMessage() (*kafka.Message, error) {
	msg, err := k.instance.ReadMessage(-1)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
