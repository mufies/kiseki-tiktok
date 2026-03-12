#!/bin/bash

# TikTok Clone - View Logs Script
# This script helps view logs from services

# Check if docker-compose exists
if ! command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

COMPOSE_FILE="docker-compose.complete.yml"

if [ -z "$1" ]; then
    echo "Usage: ./docker-logs.sh [service-name] [optional: number of lines]"
    echo ""
    echo "Available services:"
    echo "  - postgres"
    echo "  - redis"
    echo "  - kafka"
    echo "  - user-service"
    echo "  - video-service"
    echo "  - interaction-service"
    echo "  - event-service"
    echo "  - notification-service"
    echo "  - feed-service"
    echo "  - api-gateway"
    echo "  - frontend"
    echo ""
    echo "Examples:"
    echo "  ./docker-logs.sh api-gateway"
    echo "  ./docker-logs.sh user-service 100"
    echo "  ./docker-logs.sh all  (shows all services)"
    exit 0
fi

SERVICE=$1
LINES=${2:-100}

if [ "$SERVICE" = "all" ]; then
    $COMPOSE_CMD -f $COMPOSE_FILE logs -f --tail=$LINES
else
    $COMPOSE_CMD -f $COMPOSE_FILE logs -f --tail=$LINES $SERVICE
fi
