package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	connectionString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s\n", err)
	}
	defer conn.Close()

	fmt.Println("Connected to RabbitMQ")

	username, err := gamelogic.ClientWelcome()

	if err != nil {
		log.Fatalf("Failed to welcome client: %s\n", err)
	}
	fmt.Printf("Welcome, %s!\n", username)

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Fatalf("Failed to declare and bind queue %s: %s\n", queueName, err)
	}

	fmt.Printf("Queue %s declared and bound\n", queueName)

	armyMovesRoutingKey := fmt.Sprintf("%s.*", routing.ArmyMovesPrefix)
	armyMovesQueue := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username)
	armyQueuesRoutingKey := armyMovesQueue 

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		armyMovesQueue,
		armyMovesRoutingKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Fatalf("Failed to declare and bind queue %s: %s\n", armyMovesQueue, err)
	}

	fmt.Printf("Queue %s declared and bound\n", armyMovesQueue)

	// Create new game state
	gameState := gamelogic.NewGameState(username)
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.Transient, handlerPause(gameState))

	if err != nil {
		log.Fatalf("Failed to subscribe to queue: %s\n", err)
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, armyMovesQueue, armyMovesRoutingKey, pubsub.Transient, handlerMove(gameState))
	if err != nil {
		log.Fatalf("Failed to subscribe to queue: %s\n", err)
	}

	// REPL loop
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		command := strings.ToLower(words[0])

		switch command {
		case "spawn":
			if len(words) != 3 {
				fmt.Println("Usage: spawn <location> <unit_type>")
				continue
			}
			id, err := gameState.CommandSpawn(words)

			if err != nil {
				fmt.Printf("Error spawning unit: %s\n", err)
				continue
			}
			fmt.Printf("Unit spawned with ID: %d\n", id)

		case "move":
			if len(words) != 3 {
				fmt.Println("Usage: move <location> <unit_id>")
				continue
			}

			move, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Printf("Error moving unit: %s\n", err)
				continue
			}
			fmt.Printf("Units moved to %s\n", move.ToLocation)

			channel, err := conn.Channel()
			if err != nil {
				log.Fatalf("Failed to open a channel: %s\n", err)
			}

			err = pubsub.PublishJSON(channel, routing.ExchangePerilTopic, armyQueuesRoutingKey, move)
			if err != nil {
				log.Fatalf("Failed to publish to exchange %s: %s\n", routing.ExchangePerilTopic, err)
			}

			fmt.Printf("Moves published to %s\n", armyMovesQueue)

		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")
		}
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")

		gs.HandlePause(ps)
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(am gamelogic.ArmyMove) {
		defer fmt.Print("> ")

		gs.HandleMove(am)
	}
}
