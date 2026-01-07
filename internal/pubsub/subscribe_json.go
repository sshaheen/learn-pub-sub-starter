package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveryChan, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range deliveryChan {
			var msg T
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				continue
			}
			a := handler(msg)
			switch a {
			case Ack:
				d.Ack(false)
				fmt.Println("Acknowledged")
			case NackRequeue:
				d.Nack(false, true)
				fmt.Println("Negative ack, requeue")
			case NackDiscard:
				d.Nack(false, false)
				fmt.Println("Negative ack, discard")
			}
		}
	}()

	return nil
}
