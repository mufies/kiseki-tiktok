# TikTok Clone - Quick Start Guide 🚀

## 30-Second Start

```bash
cd /home/mufies/Code/tiktok-clone
./docker-start.sh
```

Wait 2-3 minutes, then open: **http://localhost:3000**

---

## What Gets Started?

| Service | Port | Type | Purpose |
|---------|------|------|---------|
| **Frontend** | 3000 | React/Nginx | User interface |
| **API Gateway** | 8080 | Go | API entry point |
| **User Service** | 8081 | Java | User management |
| **Video Service** | 8082 | Go | Video operations |
| **Event Service** | 8083 | .NET | Event tracking |
| **Interaction Service** | 8084 | Java | Likes/Comments |
| **Notification Service** | 8085 | .NET | Notifications |
| **Feed Service** | 8086 | Python | Video feed |
| **PostgreSQL** | 5433 (external) | Database | Data storage |
| **Redis** | 6379 | Cache | Session cache |
| **Kafka** | 9092 | Message Queue | Event streaming |

---

## Quick Commands

```bash
# Start everything
./docker-start.sh

# View logs
./docker-logs.sh api-gateway
./docker-logs.sh all

# Stop everything
./docker-stop.sh

# Restart a service
docker-compose -f docker-compose.complete.yml restart user-service

# Check status
docker-compose -f docker-compose.complete.yml ps

# Clean everything (⚠️ deletes data)
docker-compose -f docker-compose.complete.yml down -v
```

---

## Health Checks

```bash
# Check all services
curl http://localhost:3000          # Frontend
curl http://localhost:8080/health   # API Gateway
curl http://localhost:8081/actuator/health  # User Service
curl http://localhost:8082/health   # Video Service
curl http://localhost:8083/health   # Event Service
curl http://localhost:8084/actuator/health  # Interaction Service
curl http://localhost:8085/health   # Notification Service
curl http://localhost:8086/health   # Feed Service
```

---

## Troubleshooting

**Ports in use?**
```bash
lsof -i :8080  # Check who's using port 8080
```

**Service won't start?**
```bash
./docker-logs.sh [service-name]
docker-compose -f docker-compose.complete.yml restart [service-name]
```

**Need to rebuild?**
```bash
docker-compose -f docker-compose.complete.yml up -d --build [service-name]
```

**Clean slate?**
```bash
./docker-stop.sh
docker system prune -a --volumes  # ⚠️ Deletes EVERYTHING
./docker-start.sh
```

---

## Database Access

```bash
# PostgreSQL
docker exec -it tiktok-postgres psql -U admin -d userdb

# Redis
docker exec -it tiktok-redis redis-cli

# Kafka topics
docker exec -it tiktok-kafka kafka-topics --list --bootstrap-server localhost:9092
```

---

## Development Tips

**Watch logs in real-time:**
```bash
./docker-logs.sh all | grep ERROR
```

**Shell into a container:**
```bash
docker exec -it tiktok-api-gateway sh
```

**Check resource usage:**
```bash
docker stats
```

**Network inspection:**
```bash
docker network inspect tiktok-clone_tiktok-network
```

---

## Full Documentation

See `DOCKER_SETUP_README.md` for complete details.

