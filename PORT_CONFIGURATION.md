# Port Configuration 🔌

## PostgreSQL Port Change

Since you have PostgreSQL running locally on port **5432**, the Docker PostgreSQL has been configured to use port **5433** for external access.

### Port Mapping

```
External (from your computer) → 5433
Internal (Docker containers)   → 5432
```

### How It Works

**Inside Docker Network:**
- All microservices connect to `postgres:5432` (unchanged)
- Services use internal Docker DNS resolution
- No configuration changes needed in services

**Outside Docker (from your computer):**
- Connect using `localhost:5433`
- This avoids conflict with your local PostgreSQL on 5432

### Access Examples

#### From Your Computer

```bash
# psql command
psql -h localhost -p 5433 -U admin -d userdb

# Docker exec (recommended)
docker exec -it tiktok-postgres psql -U admin -d userdb

# Connection string
postgresql://admin:admin123@localhost:5433/userdb
```

#### From Inside Docker Containers

Services use the standard internal port:

```yaml
# Example from docker-compose
SPRING_DATASOURCE_URL: jdbc:postgresql://postgres:5432/userdb
ConnectionStrings__DefaultConnection: "Host=postgres;Port=5432;Database=eventdb;..."
```

**No changes needed** - services automatically use internal Docker networking.

---

## All Port Assignments

### External Access (from localhost)

| Service | External Port | Internal Port | Notes |
|---------|--------------|---------------|-------|
| Frontend | 3000 | 80 | Nginx serves on 80 internally |
| API Gateway | 8080 | 8080 | Same port mapping |
| User Service | 8081 | 8081 | Same port mapping |
| Video Service | 8082 | 8082 | Same port mapping |
| Event Service | 8083 | 8083 | Same port mapping |
| Interaction Service | 8084 | 8084 | Same port mapping |
| Notification Service | 8085 | 8085 | Same port mapping |
| Feed Service | 8086 | 8086 | Same port mapping |
| **PostgreSQL** | **5433** | **5432** | **Changed to avoid conflict** |
| Redis | 6379 | 6379 | Same port mapping |
| Kafka | 29092 | 9092 | Different ports for external/internal |
| Zookeeper | 2181 | 2181 | Same port mapping |

### gRPC Ports (External)

| Service | Port | Protocol |
|---------|------|----------|
| User Service | 9090 | gRPC |
| Video Service | 9091 | gRPC |
| Event Service | 9092 | gRPC |
| Notification Service | 9093 | gRPC |

---

## Checking for Port Conflicts

Before starting Docker services:

```bash
# Check if ports are in use
lsof -i :3000   # Frontend
lsof -i :8080   # API Gateway
lsof -i :5433   # PostgreSQL (Docker)
lsof -i :6379   # Redis
lsof -i :29092  # Kafka

# Check your local PostgreSQL
lsof -i :5432   # Should show your local PostgreSQL
```

---

## If You Need Different Ports

Edit `docker-compose.complete.yml` and change the port mapping:

```yaml
ports:
  - "EXTERNAL_PORT:INTERNAL_PORT"

# Example: Change PostgreSQL to port 5434
ports:
  - "5434:5432"
```

Then update health checks and documentation accordingly.

---

## Database Connection Examples

### From Your Computer

```bash
# Using psql
psql -h localhost -p 5433 -U admin -d userdb
# Password: admin123

# Using Docker exec (easier)
docker exec -it tiktok-postgres psql -U admin -d userdb

# Using DBeaver/pgAdmin
Host: localhost
Port: 5433
User: admin
Password: admin123
Database: userdb (or videodb, eventdb, etc.)
```

### Available Databases

All accessible on `localhost:5433`:

- `userdb` - User service data
- `videodb` - Video service data
- `eventdb` - Event service data
- `interactiondb` - Interaction service data
- `notificationdb` - Notification service data
- `feeddb` - Feed service data

---

## Troubleshooting

### "Port already in use" error

```bash
# Find what's using the port
lsof -i :5433

# Kill the process (if needed)
kill -9 <PID>
```

### Can't connect to PostgreSQL

```bash
# Check if container is running
docker ps | grep postgres

# Check PostgreSQL logs
docker logs tiktok-postgres

# Verify port mapping
docker port tiktok-postgres
# Should show: 5432/tcp -> 0.0.0.0:5433
```

### Services can't connect to database

**This shouldn't happen** - services use internal Docker network (`postgres:5432`), not external ports.

If it does happen:
```bash
# Check Docker network
docker network inspect tiktok-clone_tiktok-network

# Verify PostgreSQL is in the network
docker inspect tiktok-postgres | grep NetworkMode
```

---

## Summary

✅ **Local PostgreSQL**: Runs on port 5432 (unchanged)
✅ **Docker PostgreSQL**: Accessible on port 5433 (external)
✅ **Docker Services**: Connect to `postgres:5432` (internal)
✅ **No conflicts**: Both can run simultaneously

**You can access:**
- Your local databases on `localhost:5432`
- Docker databases on `localhost:5433`

---

**Last Updated:** March 11, 2026
