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
		log.Fatalf("Failed to declare and bind queue: %s\n", err)
	}

	fmt.Printf("Queue %s declared and bound\n", queueName)

	// Create new game state
	gameState := gamelogic.NewGameState(username)

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
			id := gameState.CommandSpawn(words[1:])
			fmt.Printf("Unit spawned with ID: %d\n", id)

		case "move":
			if len(words) != 3 {
				fmt.Println("Usage: move <location> <unit_id>")
				continue
			}
			move, err := gameState.CommandMove(words[1:])
			if err != nil {
				fmt.Printf("Error moving unit: %s\n", err)
				continue
			}
			fmt.Printf("Units moved to %s\n", move.ToLocation)

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
