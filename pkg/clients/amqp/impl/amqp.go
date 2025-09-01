package impl

import (
	amqp_internal "github.com/dreadew/go-common/pkg/clients/amqp"
	"github.com/dreadew/go-common/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type amqpClient struct {
	channel *amqp.Channel
	conn    *amqp.Connection
}

func New(addr string) (amqp_internal.AmqpClient, error) {
	logger := logger.GetLogger()

	conn, err := amqp.Dial(addr)
	if err != nil {
		logger.Error("error while creating amqp client", zap.String("error", err.Error()))
		return nil, err
	}

	chann, err := conn.Channel()
	if err != nil {
		logger.Error("error while creating amqp client", zap.String("error", err.Error()))
		return nil, err
	}

	return &amqpClient{
		conn:    conn,
		channel: chann,
	}, nil
}

func (a *amqpClient) Chann() *amqp.Channel {
	return a.channel
}

func (a *amqpClient) Close() error {
	return a.conn.Close()
}
