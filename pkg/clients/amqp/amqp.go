package amqp

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpClient interface {
	Chann() *amqp.Channel
	Close() error
}
