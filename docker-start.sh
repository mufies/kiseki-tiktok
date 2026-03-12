#!/bin/bash

# TikTok Clone - Docker Stack Startup Script
# This script starts all services in the correct order

set -e

echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║         🚀 Starting TikTok Clone - Complete Stack                   ║"
echo "╚══════════════════════════════════════════════════════════════════════╝"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Docker is running${NC}"
echo ""

# Check if docker-compose exists
if ! command -v docker-compose &> /dev/null; then
    echo -e "${YELLOW}⚠ docker-compose not found, using 'docker compose' instead${NC}"
    COMPOSE_CMD="docker compose"
else
    COMPOSE_CMD="docker-compose"
fi

# Use the complete docker-compose file
COMPOSE_FILE="docker-compose.complete.yml"

if [ ! -f "$COMPOSE_FILE" ]; then
    echo -e "${RED}✗ $COMPOSE_FILE not found!${NC}"
    exit 1
fi

echo "Using compose file: $COMPOSE_FILE"
echo ""

# Function to wait for service
wait_for_service() {
    local service=$1
    local max_wait=60
    local waited=0

    echo -n "Waiting for $service to be healthy..."

    while [ $waited -lt $max_wait ]; do
        if $COMPOSE_CMD -f $COMPOSE_FILE ps | grep $service | grep -q "healthy\|running"; then
            echo -e " ${GREEN}✓${NC}"
            return 0
        fi
        echo -n "."
        sleep 2
        waited=$((waited + 2))
    done

    echo -e " ${YELLOW}⚠ (timeout but continuing)${NC}"
    return 1
}

# Step 1: Start infrastructure services
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📦 STEP 1: Starting Infrastructure Services"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

$COMPOSE_CMD -f $COMPOSE_FILE up -d postgres redis zookeeper kafka

wait_for_service "postgres"
wait_for_service "redis"
wait_for_service "kafka"

echo ""

# Step 2: Start backend microservices
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔧 STEP 2: Starting Backend Microservices"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

$COMPOSE_CMD -f $COMPOSE_FILE up -d \
    user-service \
    video-service \
    interaction-service \
    event-service \
    notification-service \
    feed-service

echo "⏳ Waiting for services to start (30s)..."
sleep 30

echo ""

# Step 3: Start API Gateway
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🌐 STEP 3: Starting API Gateway"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

$COMPOSE_CMD -f $COMPOSE_FILE up -d api-gateway

wait_for_service "api-gateway"

echo ""

# Step 4: Start Frontend
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "💻 STEP 4: Starting Frontend"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

$COMPOSE_CMD -f $COMPOSE_FILE up -d frontend

wait_for_service "frontend"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✅ All Services Started Successfully!${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Service Status:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
$COMPOSE_CMD -f $COMPOSE_FILE ps
echo ""
echo "🌐 Access Points:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Frontend:              http://localhost:3000"
echo "  API Gateway:           http://localhost:8080"
echo "  User Service:          http://localhost:8081"
echo "  Video Service:         http://localhost:8082"
echo "  Event Service:         http://localhost:8083"
echo "  Interaction Service:   http://localhost:8084"
echo "  Notification Service:  http://localhost:8085"
echo "  Feed Service:          http://localhost:8086"
echo ""
echo "  PostgreSQL:            localhost:5432 (admin/admin123)"
echo "  Redis:                 localhost:6379"
echo "  Kafka:                 localhost:29092"
echo ""
echo "📝 Useful Commands:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  View logs:       $COMPOSE_CMD -f $COMPOSE_FILE logs -f [service-name]"
echo "  Stop all:        ./docker-stop.sh"
echo "  Restart service: $COMPOSE_CMD -f $COMPOSE_FILE restart [service-name]"
echo "  Shell access:    $COMPOSE_CMD -f $COMPOSE_FILE exec [service-name] sh"
echo ""
