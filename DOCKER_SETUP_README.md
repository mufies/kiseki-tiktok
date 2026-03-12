# TikTok Clone - Docker Complete Setup 🐳

Complete Docker orchestration for running all TikTok Clone microservices, infrastructure, and frontend in containers.

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Services](#services)
- [Configuration](#configuration)
- [Management Scripts](#management-scripts)
- [Troubleshooting](#troubleshooting)
- [Advanced Usage](#advanced-usage)

---

## 🏗️ Architecture Overview

The complete stack includes:

### Infrastructure Services (4)
- **PostgreSQL 16** - Primary database for all services
- **Redis 7** - Caching and session storage
- **Apache Kafka** - Event streaming and message broker
- **Zookeeper** - Kafka coordination

### Backend Microservices (6)
- **User Service** (Java/Spring Boot) - Port 8081, gRPC 9090
- **Video Service** (Go) - Port 8082, gRPC 9091
- **Interaction Service** (Java/Spring Boot) - Port 8084
- **Event Service** (.NET) - Port 8083, gRPC 9092
- **Notification Service** (.NET) - Port 8085, gRPC 9093
- **Feed Service** (Python/FastAPI) - Port 8086

### Gateway & Frontend (2)
- **API Gateway** (Go) - Port 8080
- **Frontend** (React/Nginx) - Port 3000

**Total: 12 services running in Docker**

```
┌─────────────┐
│   Browser   │
└──────┬──────┘
       │
       ↓ :3000
┌─────────────┐         ┌──────────────┐
│  Frontend   │────────→│ API Gateway  │ :8080
│   (Nginx)   │         │     (Go)     │
└─────────────┘         └──────┬───────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        ↓                      ↓                      ↓
  ┌──────────┐         ┌──────────┐          ┌──────────┐
  │   User   │         │  Video   │          │   Feed   │
  │ Service  │         │ Service  │          │ Service  │
  └────┬─────┘         └────┬─────┘          └────┬─────┘
       │                    │                     │
       │              ┌─────┴─────┐               │
       │              ↓           ↓               │
  ┌────────────┐ ┌──────────┐ ┌──────────┐      │
  │Interaction │ │  Event   │ │Notification│     │
  │  Service   │ │ Service  │ │  Service   │     │
  └────┬───────┘ └────┬─────┘ └────┬───────┘     │
       │              │              │            │
       └──────────────┼──────────────┼────────────┘
                      ↓              ↓
              ┌───────────┐   ┌──────────┐
              │PostgreSQL │   │  Redis   │
              └───────────┘   └──────────┘
                      ↓
              ┌───────────┐
              │   Kafka   │
              │ (Zookeeper)│
              └───────────┘
```

---

## ✅ Prerequisites

### Required
- **Docker** 20.10+ installed and running
- **Docker Compose** 2.0+ or `docker compose` plugin
- **8GB+ RAM** available for Docker
- **20GB+ disk space** for images and volumes

### Check Installation
```bash
docker --version        # Should be 20.10+
docker-compose --version # Should be 2.0+
docker info | grep "Total Memory"  # Check available RAM
```

### Ports Required
Ensure these ports are free:
- **3000** - Frontend
- **8080** - API Gateway
- **8081-8086** - Microservices
- **5433** - PostgreSQL (external, internal: 5432)
- **6379** - Redis
- **9092, 29092** - Kafka
- **2181** - Zookeeper
- **9090-9093** - gRPC ports

---

## 🚀 Quick Start

### Option 1: Using Scripts (Recommended)

```bash
# 1. Navigate to project directory
cd /home/mufies/Code/tiktok-clone

# 2. Start entire stack
./docker-start.sh

# 3. Wait for all services to start (~2-3 minutes)
# The script will show progress

# 4. Access the application
# Open browser: http://localhost:3000
```

### Option 2: Using Docker Compose Directly

```bash
# Start all services
docker-compose -f docker-compose.complete.yml up -d

# View logs
docker-compose -f docker-compose.complete.yml logs -f

# Stop all services
docker-compose -f docker-compose.complete.yml down
```

---

## 🎯 Services

### Infrastructure Services

#### PostgreSQL
```yaml
Port: 5433 (external), 5432 (internal Docker network)
User: admin
Password: admin123
Databases: userdb, videodb, eventdb, interactiondb, notificationdb, feeddb
```

**Connect:**
```bash
docker exec -it tiktok-postgres psql -U admin -d userdb
```

#### Redis
```yaml
Port: 6379
Max Memory: 512MB
Policy: allkeys-lru
```

**Connect:**
```bash
docker exec -it tiktok-redis redis-cli
```

#### Kafka
```yaml
Internal Port: 9092
External Port: 29092
Broker ID: 1
```

**List Topics:**
```bash
docker exec -it tiktok-kafka kafka-topics --list --bootstrap-server localhost:9092
```

### Backend Services

#### User Service (Java/Spring Boot)
- **HTTP:** http://localhost:8081
- **gRPC:** localhost:9090
- **Tech:** Spring Boot 3.2, Spring Data JPA
- **Database:** PostgreSQL (userdb)
- **Events:** Kafka producer

**Health Check:**
```bash
curl http://localhost:8081/actuator/health
```

#### Video Service (Go)
- **HTTP:** http://localhost:8082
- **gRPC:** localhost:9091
- **Tech:** Go 1.21, gRPC, GORM
- **Database:** PostgreSQL (videodb)

**Health Check:**
```bash
curl http://localhost:8082/health
```

#### Interaction Service (Java/Spring Boot)
- **HTTP:** http://localhost:8084
- **Tech:** Spring Boot 3.2, gRPC client
- **Database:** PostgreSQL (interactiondb)
- **Dependencies:** Video Service (gRPC), Kafka

**Health Check:**
```bash
curl http://localhost:8084/actuator/health
```

#### Event Service (.NET)
- **HTTP:** http://localhost:8083
- **gRPC:** localhost:9092
- **Tech:** .NET 8, Entity Framework Core
- **Database:** PostgreSQL (eventdb)
- **Cache:** Redis

**Health Check:**
```bash
curl http://localhost:8083/health
```

#### Notification Service (.NET)
- **HTTP:** http://localhost:8085
- **gRPC:** localhost:9093
- **Tech:** .NET 8, Entity Framework Core
- **Database:** PostgreSQL (notificationdb)
- **Cache:** Redis
- **Events:** Kafka consumer

**Health Check:**
```bash
curl http://localhost:8085/health
```

#### Feed Service (Python/FastAPI)
- **HTTP:** http://localhost:8086
- **Tech:** Python 3.11, FastAPI, SQLAlchemy
- **Database:** PostgreSQL (feeddb)
- **Cache:** Redis
- **Events:** Kafka consumer

**Health Check:**
```bash
curl http://localhost:8086/health
```

### Gateway & Frontend

#### API Gateway (Go)
- **HTTP:** http://localhost:8080
- **Tech:** Go 1.21, Gin framework
- **Routes:** Proxies to all backend services

**Test:**
```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/users/health
```

#### Frontend (React + Nginx)
- **HTTP:** http://localhost:3000
- **Tech:** React 19, TypeScript, Vite
- **Server:** Nginx with SPA routing
- **API Proxy:** Proxies /api to API Gateway

**Access:**
```
Browser: http://localhost:3000
```

---

## ⚙️ Configuration

### Environment Variables

All services use environment variables for configuration. Edit `docker-compose.complete.yml` to change settings.

#### Database Configuration
```yaml
SPRING_DATASOURCE_URL: jdbc:postgresql://postgres:5432/userdb
SPRING_DATASOURCE_USERNAME: admin
SPRING_DATASOURCE_PASSWORD: admin123
```

#### Kafka Configuration
```yaml
SPRING_KAFKA_BOOTSTRAP_SERVERS: kafka:9092
Kafka__BootstrapServers: kafka:9092
```

#### Service URLs (API Gateway)
```yaml
USER_SERVICE_URL: http://user-service:8081
VIDEO_SERVICE_URL: http://video-service:8082
# ... etc
```

### Volumes

Data persists in Docker volumes:
- `postgres-data` - PostgreSQL databases
- `redis-data` - Redis cache
- `kafka-data` - Kafka messages
- `zookeeper-data` - Zookeeper metadata

**List volumes:**
```bash
docker volume ls | grep tiktok
```

**Remove volumes (WARNING: Deletes all data):**
```bash
docker-compose -f docker-compose.complete.yml down -v
```

---

## 🛠️ Management Scripts

### docker-start.sh
Starts all services in the correct order with health checks.

```bash
./docker-start.sh
```

**What it does:**
1. Checks Docker is running
2. Starts infrastructure (Postgres, Redis, Kafka)
3. Waits for infrastructure to be healthy
4. Starts backend microservices
5. Starts API Gateway
6. Starts Frontend
7. Shows service status and access points

### docker-stop.sh
Gracefully stops all services.

```bash
./docker-stop.sh
```

### docker-logs.sh
View logs from specific services.

```bash
# View logs for a specific service
./docker-logs.sh api-gateway

# View last 200 lines
./docker-logs.sh user-service 200

# View all services
./docker-logs.sh all
```

---

## 🔧 Troubleshooting

### Services Won't Start

**Check Docker resources:**
```bash
docker system df
docker system prune  # Clean up unused resources
```

**Check logs:**
```bash
./docker-logs.sh [service-name]
```

**Restart a specific service:**
```bash
docker-compose -f docker-compose.complete.yml restart user-service
```

### Port Already in Use

**Find process using port:**
```bash
lsof -i :8080  # Replace with your port
```

**Kill process:**
```bash
kill -9 <PID>
```

### Database Connection Errors

**Check PostgreSQL is running:**
```bash
docker exec -it tiktok-postgres pg_isready -U admin
```

**View PostgreSQL logs:**
```bash
./docker-logs.sh postgres
```

**Recreate database:**
```bash
docker-compose -f docker-compose.complete.yml down postgres
docker volume rm tiktok-clone_postgres-data
docker-compose -f docker-compose.complete.yml up -d postgres
```

### Kafka Connection Issues

**Check Kafka is healthy:**
```bash
docker exec -it tiktok-kafka kafka-broker-api-versions --bootstrap-server localhost:9092
```

**Recreate Kafka:**
```bash
docker-compose -f docker-compose.complete.yml restart kafka
```

### Frontend Not Loading

**Check Nginx logs:**
```bash
docker exec -it tiktok-frontend tail -f /var/log/nginx/error.log
```

**Rebuild frontend:**
```bash
docker-compose -f docker-compose.complete.yml up -d --build frontend
```

### Service Health Check Failing

**View container status:**
```bash
docker ps -a
```

**Inspect container:**
```bash
docker inspect tiktok-user-service
```

**Shell into container:**
```bash
docker exec -it tiktok-user-service sh
```

---

## 🎓 Advanced Usage

### Building Specific Services

```bash
# Build only one service
docker-compose -f docker-compose.complete.yml build user-service

# Build without cache
docker-compose -f docker-compose.complete.yml build --no-cache video-service

# Build and restart
docker-compose -f docker-compose.complete.yml up -d --build api-gateway
```

### Scaling Services

```bash
# Scale feed service to 3 instances
docker-compose -f docker-compose.complete.yml up -d --scale feed-service=3
```

### View Resource Usage

```bash
# Real-time stats
docker stats

# Specific service
docker stats tiktok-user-service
```

### Execute Commands in Containers

```bash
# PostgreSQL query
docker exec -it tiktok-postgres psql -U admin -d userdb -c "SELECT * FROM users LIMIT 5;"

# Redis command
docker exec -it tiktok-redis redis-cli GET some_key

# Shell access
docker exec -it tiktok-api-gateway sh
```

### Network Inspection

```bash
# List networks
docker network ls | grep tiktok

# Inspect network
docker network inspect tiktok-clone_tiktok-network

# View container IPs
docker inspect -f '{{.Name}} - {{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $(docker ps -aq)
```

### Export/Import Data

```bash
# Backup PostgreSQL
docker exec -it tiktok-postgres pg_dumpall -U admin > backup.sql

# Restore PostgreSQL
docker exec -i tiktok-postgres psql -U admin < backup.sql

# Backup volumes
docker run --rm -v tiktok-clone_postgres-data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup.tar.gz -C /data .

# Restore volumes
docker run --rm -v tiktok-clone_postgres-data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres-backup.tar.gz -C /data
```

### Development Mode

For development with hot reload:

```bash
# Stop production frontend
docker-compose -f docker-compose.complete.yml stop frontend

# Run frontend in dev mode (outside Docker)
cd ../test-app
npm run dev
```

---

## 📊 Monitoring

### View All Container Status

```bash
docker-compose -f docker-compose.complete.yml ps
```

### Check Service Health

```bash
# API Gateway
curl http://localhost:8080/health

# User Service
curl http://localhost:8081/actuator/health

# Video Service
curl http://localhost:8082/health

# All services script
for port in 8080 8081 8082 8083 8084 8085 8086; do
  echo "Port $port: $(curl -s http://localhost:$port/health 2>&1 | head -1)"
done
```

### View Logs in Real-Time

```bash
# All services
docker-compose -f docker-compose.complete.yml logs -f

# Specific service
docker-compose -f docker-compose.complete.yml logs -f user-service

# Multiple services
docker-compose -f docker-compose.complete.yml logs -f user-service video-service
```

---

## 🎯 Production Deployment

For production deployment, consider:

1. **Use secrets management** (Docker Secrets, Vault)
2. **Enable SSL/TLS** (Let's Encrypt, custom certs)
3. **Add load balancer** (nginx, HAProxy, Traefik)
4. **Set up monitoring** (Prometheus, Grafana)
5. **Configure log aggregation** (ELK Stack, Loki)
6. **Implement backup strategy** for volumes
7. **Use container orchestration** (Kubernetes, Docker Swarm)

---

## 📝 Summary

**Start stack:**
```bash
./docker-start.sh
```

**Access app:**
```
http://localhost:3000
```

**View logs:**
```bash
./docker-logs.sh all
```

**Stop stack:**
```bash
./docker-stop.sh
```

---

## 🆘 Support

**Issues:** Check service logs with `./docker-logs.sh [service-name]`

**Resources:** Ensure Docker has 8GB+ RAM allocated

**Clean slate:** `docker-compose -f docker-compose.complete.yml down -v` (WARNING: Deletes all data)

---

**Version:** 1.0.0
**Last Updated:** March 11, 2026
**Status:** ✅ Production Ready
