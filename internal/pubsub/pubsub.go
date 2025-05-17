package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

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
		nil,
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

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, Key string, simpleQueueType SimpleQueueType, handler func(T) (ackType AckType)) error {
	chn, queue, err := DeclareAndBind(conn, exchange, queueName, Key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("failed to declare and bind: %w", err)
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
			var val T
			err := json.Unmarshal(msg.Body, &val)
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
				msg.Nack(false, true)
				fmt.Printf("acktype: NackDiscard!. \n")
			default:
				fmt.Printf("Unknown acktype!. \n")
			}
		}
	}()

	return nil
}
