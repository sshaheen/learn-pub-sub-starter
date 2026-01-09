package pubsub

import (
	"bytes"
	"encoding/gob"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGOB[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}

	err = ch.Qos(10, 0, false)
	if err != nil {
		return err
	}

	deliveryChan, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for d := range deliveryChan {
			buf := bytes.NewBuffer(d.Body)
			dec := gob.NewDecoder(buf)
			var msg T
			err := dec.Decode(&msg)
			if err != nil {
				fmt.Println("decode error:", err)
				d.Nack(false, false) // or true, depending on what you want
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
