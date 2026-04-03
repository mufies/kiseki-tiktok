# TikTok Clone

Microservices-based video sharing platform.

## Tech Stack

### Backend Services
- User Service - Java Spring Boot
- Video Service - Go
- Interaction Service - Java Spring Boot
- Event Service - .NET
- Feed Service - Python FastAPI
- Notification Service - .NET
- Stream Service - Go
- API Gateway - Go

### Frontend
- React + Vite

### Infrastructure
- PostgreSQL
- Redis
- Kafka
- MinIO
- Zookeeper

## Setup

### Prerequisites
- Docker
- Docker Compose

### Installation

1. Clone repository
```bash
git clone <repository-url>
cd tiktok-clone
```

2. Copy environment file
```bash
cp .env.example .env
```

3. Start infrastructure
```bash
docker-compose -f docker-compose.infrastructure.yml up -d
```

4. Start application services
```bash
docker-compose up -d
```

### Access

- Frontend: http://localhost:5173
- API Gateway: http://localhost:8080
- MinIO Console: http://localhost:9001

## Services

### User Service
- Port: 8081 (HTTP), 50053 (gRPC)
- Database: userdb

### Video Service
- Port: 8082 (HTTP), 50052 (gRPC)
- Database: videodb
- Storage: MinIO

### Interaction Service
- Port: 8084 (HTTP), 50054 (gRPC)
- Database: interactiondb

### Event Service
- Port: 5001 (HTTP), 5002 (gRPC)
- Database: eventdb

### Feed Service
- Port: 8001 (HTTP)
- Database: feeddb

### Notification Service
- Port: 8085 (HTTP), 9093 (gRPC)
- Database: notificationdb

### Stream Service
- Port: 8083 (HTTP), 50055 (gRPC), 1935 (RTMP)
- Database: streamdb
- Storage: MinIO

## Development

### Hot Reload
All services support hot reload in development mode.

### Logs
```bash
docker-compose logs -f [service-name]
```

### Stop Services
```bash
docker-compose down
```

### Clean Volumes
```bash
docker-compose down -v
```

## Architecture

```
Frontend (React)
    |
API Gateway (Go)
    |
    +-- User Service (Java)
    +-- Video Service (Go)
    +-- Interaction Service (Java)
    +-- Event Service (.NET)
    +-- Feed Service (Python)
    +-- Notification Service (.NET)
    +-- Stream Service (Go)
```

## License

See LICENSE file.
