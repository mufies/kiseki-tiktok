#!/bin/bash

# TikTok Clone - Docker Stack Shutdown Script
# This script stops all services gracefully

set -e

echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║         🛑 Stopping TikTok Clone - Complete Stack                   ║"
echo "╚══════════════════════════════════════════════════════════════════════╝"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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

# Stop services in reverse order
echo "🛑 Stopping services..."
$COMPOSE_CMD -f $COMPOSE_FILE down

echo ""
echo -e "${GREEN}✅ All services stopped successfully!${NC}"
echo ""
echo "💡 To remove volumes as well, run:"
echo "   $COMPOSE_CMD -f $COMPOSE_FILE down -v"
echo ""
