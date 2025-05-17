#!/bin/bash

case "$1" in
    start)
        echo "Starting RabbitMQ container..."
        docker run -d --rm --name MrRabbit -p 5672:5672 -p 15672:15672 rabbitmq:4.0.2-management
        ;;
    stop)
        echo "Stopping RabbitMQ container..."
        docker stop MrRabbit
        ;;
    logs)
        echo "Fetching logs for RabbitMQ container..."
        docker logs -f MrRabbit
        ;;
    *)
        echo "Usage: $0 {start|stop|logs}"
        exit 1
esac
