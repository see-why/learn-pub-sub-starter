package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	connectionString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s\n", err)
	}
	defer conn.Close()

	fmt.Println("Connected to RabbitMQ")

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s\n", err)
	}
	defer channel.Close()
	fmt.Println("Channel opened")

	routingKey := fmt.Sprintf("%s.*", routing.GameLogSlug)
	queueName := "game_logs"

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		queueName,
		routingKey,
		pubsub.Durable,
	)
	if err != nil {
		log.Fatalf("Failed to declare and bind queue: %s\n", err)
	}
	fmt.Printf("Queue %s declared and bound\n", queueName)

	// Subscribe to game logs
	gameLogsRoutingKey := fmt.Sprintf("%s.*", routing.GameLogSlug)
	err = pubsub.SubscribeGob(conn, routing.ExchangePerilTopic, routing.GameLogSlug, gameLogsRoutingKey, pubsub.Durable, func(log routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(log)
		if err != nil {
			fmt.Printf("Failed to write log: %v\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to game logs: %s\n", err)
	}

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		if words[0] == "quit" {
			fmt.Println("Quitting...")
			os.Exit(0)
		}

		if words[0] == "pause" {
			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})

			if err != nil {
				log.Printf("Failed to publish message: %s\n", err)
				continue
			}
			fmt.Println("Message published")

			fmt.Println("Game paused")
			continue
		}

		if words[0] == "resume" {
			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				log.Printf("Failed to publish message: %s\n", err)
				continue
			}
			fmt.Println("Message published")

			fmt.Println("Game resumed")
			continue
		}

		fmt.Println("I do not understand that command")
	}
}
