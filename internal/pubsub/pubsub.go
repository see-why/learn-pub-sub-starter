package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

// SimpleQueueType represents different types of queues
type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	// Marshal the value to JSON
	body, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Publish the message
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(val); err != nil {
		return fmt.Errorf("failed to gob encode value: %w", err)
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	})
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	chn, err := conn.Channel()

	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to open channel: %w", err)
	}

	isTransient := simpleQueueType == Transient

	queue, err := chn.QueueDeclare(
		queueName,
		simpleQueueType == Durable,
		isTransient,
		isTransient,
		false,
		amqp.Table{
			"x-dead-letter-exchange": routing.ExchangePerilDeadLetter,
		},
	)

	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = chn.QueueBind(
		queue.Name,
		key,
		exchange,
		false,
		nil,
	)

	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to bind queue: %w", err)
	}

	return chn, queue, nil
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	chn, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("failed to declare and bind: %w", err)
	}

	err = chn.Qos(10, 10, true)
	if err != nil {
		return fmt.Errorf("failed to set prefetch count: %w", err)
	}

	msgs, err := chn.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			val, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("failed to parse message: %v\n", err)
				continue
			}
			ackType := handler(val)

			switch ackType {
			case Ack:
				msg.Ack(false)
				fmt.Printf("acktype: Ack!. \n")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Printf("acktype: NackRequeue!. \n")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Printf("acktype: NackDiscard!. \n")
			default:
				fmt.Printf("Unknown acktype!. \n")
			}
		}
	}()

	return nil
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType, handler func(T) AckType) error {
	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, func(data []byte) (T, error) {
		var val T
		err := json.Unmarshal(data, &val)
		return val, err
	})
}

func SubscribeGob[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType, handler func(T) AckType) error {
	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, func(data []byte) (T, error) {
		var val T
		dec := gob.NewDecoder(bytes.NewReader(data))
		err := dec.Decode(&val)
		return val, err
	})
}
