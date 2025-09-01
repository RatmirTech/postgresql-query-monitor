package amqp_sender

type AmqpSenderClient interface {
	Send(info interface{}, queue string) error
	Close() error
}
