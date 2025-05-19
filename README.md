# Peril - A Pub/Sub Based Strategy Game

Peril is a multiplayer strategy game that demonstrates the power of pub/sub architecture using RabbitMQ. Players can control armies, move units, and engage in wars with other players in real-time.

## Features

- Real-time multiplayer gameplay
- Pub/Sub architecture using RabbitMQ
- Army movement and unit management
- War system with battle outcomes
- Game state persistence
- Event logging system

## Prerequisites

- Go 1.x or higher
- RabbitMQ server running locally (or accessible via network)
- Basic understanding of pub/sub patterns

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/learn-pub-sub-starter.git
cd learn-pub-sub-starter
```

2. Install dependencies:
```bash
go mod download
```

3. Ensure RabbitMQ is running:
```bash
# Default connection string: amqp://guest:guest@localhost:5672/
```

## Running the Game

1. Start the game client:
```bash
go run cmd/client/main.go
```

2. Follow the on-screen prompts to:
   - Enter your username
   - Spawn units
   - Move armies
   - Engage in wars

## Game Commands

- `spawn <location> <unit_type>` - Spawn a new unit at the specified location
- `move <location> <unit_id>` - Move a unit to a new location
- `status` - View your current game state
- `help` - Display available commands
- `quit` - Exit the game

## Architecture

The game uses a pub/sub architecture with the following components:

- **Direct Exchange**: Handles pause/resume game state
- **Topic Exchange**: Manages army movements and war events
- **Durable Queues**: Ensures war events persist across game sessions
- **Transient Queues**: Handles temporary game state updates

## Event Types

1. **Army Moves**: Published when units are moved
2. **War Events**: Triggered when armies from different players meet
3. **Game Logs**: Record important game events and outcomes

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is part of Boot.dev's [Learn Pub/Sub](https://learn.boot.dev/learn-pub-sub) course.
