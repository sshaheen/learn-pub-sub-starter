package pubsub

import amqp "github.com/rabbitmq/amqp091-go"

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return ch, amqp.Queue{}, err
	}

	q, err := ch.QueueDeclare(
		queueName,
		queueType == SimpleQueueDurable, // durable
		queueType != SimpleQueueDurable, // delete when unused
		queueType != SimpleQueueDurable, // exclusive
		false,
		nil,
	)

	if err != nil {
		return ch, amqp.Queue{}, err
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil)

	return ch, q, nil
}
