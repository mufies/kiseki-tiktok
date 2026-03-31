#!/bin/bash

# TikTok Clone - Docker Helper Script
# Quick commands for managing the application

set -e

BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}============================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

check_infrastructure() {
    print_header "Checking Infrastructure Services"

    echo -n "PostgreSQL (host): "
    if systemctl is-active postgresql >/dev/null 2>&1; then
        print_success "Running"
    else
        print_error "Not Running"
        echo "  Start with: sudo systemctl start postgresql"
        return 1
    fi

    echo -n "Redis (host):      "
    if systemctl is-active redis >/dev/null 2>&1; then
        print_success "Running"
    else
        print_error "Not Running"
        echo "  Start with: sudo systemctl start redis"
        return 1
    fi

    echo -n "MinIO (container): "
    if docker ps --format '{{.Names}}' | grep -q tiktok-minio; then
        print_success "Running"
    else
        print_error "Not Running"
        echo "  Check: docker ps -a | grep minio"
        return 1
    fi

    echo -n "Kafka (container): "
    if docker ps --format '{{.Names}}' | grep -q tiktok-kafka; then
        print_success "Running"
    else
        print_error "Not Running"
        echo "  Start with: docker start tiktok-kafka"
        return 1
    fi

    echo -n "Zookeeper:         "
    if docker ps --format '{{.Names}}' | grep -q tiktok-zookeeper; then
        print_success "Running"
    else
        print_error "Not Running"
        echo "  Start with: docker start tiktok-zookeeper"
        return 1
    fi

    echo ""
    print_success "All infrastructure services are running!"
    return 0
}

check_databases() {
    print_header "Checking Databases"

    # Check if databases exist
    DATABASES=("userdb" "videodb" "eventdb" "interactiondb" "notificationdb" "feeddb")

    for db in "${DATABASES[@]}"; do
        echo -n "Database $db: "
        if psql -U postgres -h localhost -lqt | cut -d \| -f 1 | grep -qw $db; then
            print_success "Exists"
        else
            print_warning "Not found"
            echo "  Create with: psql -U postgres -h localhost -f init-databases.sql"
        fi
    done
}

start_services() {
    print_header "Starting Application Services"

    if ! check_infrastructure >/dev/null 2>&1; then
        print_error "Infrastructure is not ready!"
        echo ""
        check_infrastructure
        exit 1
    fi

    echo "Starting all services..."
    docker-compose up -d

    echo ""
    print_success "Services started!"
    echo ""
    echo "URLs:"
    echo "  Frontend:    http://localhost:5173"
    echo "  API Gateway: http://localhost:8080"
    echo "  MinIO:       http://localhost:9011"
}

start_dev() {
    print_header "Starting Development Mode"

    if ! check_infrastructure >/dev/null 2>&1; then
        print_error "Infrastructure is not ready!"
        echo ""
        check_infrastructure
        exit 1
    fi

    echo "Starting services with dev tools..."
    docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

    echo ""
    print_success "Dev mode started!"
    echo ""
    echo "URLs:"
    echo "  Frontend:        http://localhost:5173"
    echo "  API Gateway:     http://localhost:8080"
    echo "  MinIO Console:   http://localhost:9011 (minioadmin/minioadmin)"
    echo "  pgAdmin:         http://localhost:5050 (admin@tiktok.local/admin)"
    echo "  Kafka UI:        http://localhost:8090"
    echo "  Redis Commander: http://localhost:8091"
}

stop_services() {
    print_header "Stopping Application Services"
    docker-compose down
    print_success "Services stopped!"
}

show_status() {
    print_header "Service Status"
    docker-compose ps
}

show_logs() {
    if [ -z "$1" ]; then
        docker-compose logs -f --tail=100
    else
        docker-compose logs -f --tail=100 "$1"
    fi
}

show_health() {
    print_header "Health Check"

    echo "API Gateway:"
    curl -s http://localhost:8080/health | jq . 2>/dev/null || echo "Not responding"

    echo ""
    echo "Individual Services:"

    echo -n "  Video Service: "
    curl -s http://localhost:8081/health >/dev/null 2>&1 && print_success "OK" || print_error "Down"

    echo -n "  User Service: "
    curl -s http://localhost:8083/actuator/health >/dev/null 2>&1 && print_success "OK" || print_error "Down"

    echo -n "  Interaction Service: "
    curl -s http://localhost:8084/actuator/health >/dev/null 2>&1 && print_success "OK" || print_error "Down"

    echo -n "  Event Service: "
    curl -s http://localhost:5001/health >/dev/null 2>&1 && print_success "OK" || print_error "Down"

    echo -n "  Feed Service: "
    curl -s http://localhost:8001/health >/dev/null 2>&1 && print_success "OK" || print_error "Down"

    echo -n "  Notification Service: "
    curl -s http://localhost:8085/health >/dev/null 2>&1 && print_success "OK" || print_error "Down"
}

init_databases() {
    print_header "Initializing Databases"

    if [ -f "init-databases.sql" ]; then
        psql -U postgres -h localhost -f init-databases.sql
        print_success "Databases initialized!"
    else
        print_error "init-databases.sql not found!"
        exit 1
    fi
}

show_help() {
    cat << EOF
TikTok Clone - Docker Helper Script

Usage: ./docker-helper.sh [command]

Commands:
    check          Check infrastructure status
    check-db       Check if databases exist
    init-db        Initialize databases from init-databases.sql
    start          Start application services
    dev            Start with development tools
    stop           Stop all services
    restart        Restart all services
    status         Show service status
    logs [service] Show logs (optionally for specific service)
    health         Check health of all services
    help           Show this help message

Examples:
    ./docker-helper.sh check
    ./docker-helper.sh start
    ./docker-helper.sh logs video-service
    ./docker-helper.sh health

EOF
}

# Main script
case "$1" in
    check)
        check_infrastructure
        ;;
    check-db)
        check_databases
        ;;
    init-db)
        init_databases
        ;;
    start)
        start_services
        ;;
    dev)
        start_dev
        ;;
    stop)
        stop_services
        ;;
    restart)
        stop_services
        echo ""
        start_services
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs "$2"
        ;;
    health)
        show_health
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
