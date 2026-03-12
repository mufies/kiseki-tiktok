#!/bin/bash

# TikTok Clone - Health Check Script
# This script checks the health status of all services

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║         🏥 TikTok Clone - Service Health Check                      ║"
echo "╚══════════════════════════════════════════════════════════════════════╝"
echo ""

# Function to check HTTP endpoint
check_http() {
    local name=$1
    local url=$2
    local response=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 "$url" 2>/dev/null)

    if [ "$response" = "200" ] || [ "$response" = "404" ]; then
        echo -e "${GREEN}✓${NC} $name - OK (HTTP $response)"
        return 0
    else
        echo -e "${RED}✗${NC} $name - FAILED (HTTP $response)"
        return 1
    fi
}

# Function to check TCP port
check_tcp() {
    local name=$1
    local host=$2
    local port=$3

    if nc -z -w 3 "$host" "$port" 2>/dev/null; then
        echo -e "${GREEN}✓${NC} $name - OK"
        return 0
    else
        echo -e "${RED}✗${NC} $name - FAILED"
        return 1
    fi
}

# Check if services are running
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📦 Container Status"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
docker ps --format "table {{.Names}}\t{{.Status}}" | grep tiktok || echo "No containers running"
echo ""

# Infrastructure health checks
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔧 Infrastructure Services"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
check_tcp "PostgreSQL" localhost 5433
check_tcp "Redis" localhost 6379
check_tcp "Kafka" localhost 29092
check_tcp "Zookeeper" localhost 2181
echo ""

# Backend service health checks
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 Backend Microservices"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
check_http "User Service" "http://localhost:8081/actuator/health"
check_http "Video Service" "http://localhost:8082/health"
check_http "Event Service" "http://localhost:8083/health"
check_http "Interaction Service" "http://localhost:8084/actuator/health"
check_http "Notification Service" "http://localhost:8085/health"
check_http "Feed Service" "http://localhost:8086/health"
echo ""

# Gateway and frontend
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🌐 Gateway & Frontend"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
check_http "API Gateway" "http://localhost:8080/health"
check_http "Frontend" "http://localhost:3000"
echo ""

# Database connectivity test
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🗄️  Database Connectivity"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if docker exec tiktok-postgres pg_isready -U admin > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} PostgreSQL - Accepting connections"
else
    echo -e "${RED}✗${NC} PostgreSQL - Not accepting connections"
fi

if docker exec tiktok-redis redis-cli ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Redis - PONG"
else
    echo -e "${RED}✗${NC} Redis - No response"
fi
echo ""

# Resource usage
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Resource Usage"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep tiktok || echo "No stats available"
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Health check complete"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "💡 Access Points:"
echo "   Frontend:     http://localhost:3000"
echo "   API Gateway:  http://localhost:8080"
echo ""
