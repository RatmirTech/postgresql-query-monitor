package impl

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dreadew/go-common/pkg/clients/amqp_sender"
	"github.com/dreadew/go-common/pkg/logger"

	amqp_internal "github.com/dreadew/go-common/pkg/clients/amqp"
	"github.com/dreadew/go-common/pkg/clients/amqp/impl"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type amqpSenderClient struct {
	amqp_internal.AmqpClient
}

func New(addr string) (amqp_sender.AmqpSenderClient, error) {
	logger := logger.GetLogger()

	client, err := impl.New(addr)
	if err != nil {
		logger.Error("error while creating amqp sender client", zap.String("error", err.Error()))
		return nil, err
	}

	return &amqpSenderClient{
		AmqpClient: client,
	}, nil
}

func (a *amqpSenderClient) Send(info interface{}, queue string) error {
	req, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	err = a.AmqpClient.Chann().PublishWithContext(context.Background(), "", queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        req,
	})
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}

func (a *amqpSenderClient) SendToExchange(info interface{}, exchange, queue string) error {
	req, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	err = a.AmqpClient.Chann().PublishWithContext(context.Background(), exchange, queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        req,
	})
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}
