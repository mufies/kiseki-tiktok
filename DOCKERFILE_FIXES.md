# Dockerfile Fixes Applied 🔧

## Issues Fixed

### 1. Go Version Incompatibility ✅

**Problem:**
- Go services required Go 1.25.7 (which doesn't exist)
- Dockerfiles used Go 1.21

**Solution:**
- Updated Dockerfiles to Go 1.23
- Updated go.mod files to Go 1.23

**Files Changed:**
- `/video-service/Dockerfile` - Go 1.21 → 1.23
- `/video-service/go.mod` - Go 1.25.7 → 1.23
- `/api-gateway/Dockerfile` - Go 1.21 → 1.23
- `/api-gateway/go.mod` - Go 1.25.7 → 1.23

### 2. Port Mismatches ✅

**Problem:**
- Feed service Dockerfile exposed port 8001
- Event service Dockerfile exposed port 5001
- Docker-compose expected ports 8086 and 8083

**Solution:**
- Updated Dockerfiles to match docker-compose ports

**Files Changed:**
- `/feed-service/Dockerfile` - Port 8001 → 8086
- `/event-service/Dockerfile` - Port 5001 → 8083

### 3. .NET NuGet Package Restore Issues ✅

**Problem:**
- Event service failed with "Package Microsoft.CodeAnalysis.Analyzers not found"
- Using `--no-restore` flag incorrectly in multi-stage builds

**Solution:**
- Removed `--no-restore` flag from publish step
- Added proper multi-stage build pattern
- Added `/p:UseAppHost=false` for smaller images
- Standardized both .NET service Dockerfiles

**Files Changed:**
- `/event-service/Dockerfile` - Improved multi-stage build, removed --no-restore
- `/notification-service/NotificationService/Dockerfile` - Standardized pattern

---

## Summary of All Dockerfiles

| Service | Language | Base Image | Exposed Ports | Status |
|---------|----------|------------|---------------|--------|
| **user-service** | Java 17 | maven:3.9-temurin-17 | 8081 | ✅ OK |
| **video-service** | Go 1.23 | golang:1.23-alpine | 8082, 9091 | ✅ Fixed |
| **interaction-service** | Java 21 | maven:3.9-temurin-21 | 8084 | ✅ OK |
| **event-service** | .NET 8 | dotnet/aspnet:8.0 | 8083, 9092 | ✅ Fixed |
| **notification-service** | .NET 8 | dotnet/aspnet:8.0 | 8085, 9093 | ✅ OK |
| **feed-service** | Python 3.11 | python:3.11-slim | 8086 | ✅ Fixed |
| **api-gateway** | Go 1.23 | golang:1.23-alpine | 8080 | ✅ Fixed |
| **frontend** | React + Nginx | nginx:alpine | 80 (→3000) | ✅ OK |

---

## Build Commands

### Build All Services

```bash
cd /home/mufies/Code/tiktok-clone
docker-compose -f docker-compose.complete.yml build
```

### Build Specific Service

```bash
# Video service
docker-compose -f docker-compose.complete.yml build video-service

# API Gateway
docker-compose -f docker-compose.complete.yml build api-gateway

# Feed service
docker-compose -f docker-compose.complete.yml build feed-service
```

### Build with No Cache (if needed)

```bash
docker-compose -f docker-compose.complete.yml build --no-cache
```

---

## Common Build Issues & Solutions

### Issue: Go Module Download Fails

**Error:**
```
go: go.mod requires go >= X.X.X (running go Y.Y.Y)
```

**Solution:**
Update Dockerfile to use correct Go version:
```dockerfile
FROM golang:1.23-alpine AS builder
```

### Issue: Port Already in Use

**Error:**
```
Bind for 0.0.0.0:XXXX failed: port is already allocated
```

**Solution:**
1. Check what's using the port: `lsof -i :XXXX`
2. Stop the process or change the port in docker-compose.yml

### Issue: Dependencies Not Found

**Error:**
```
Cannot find module/package XXX
```

**Solution:**
- For Go: Ensure `go.mod` and `go.sum` are copied before `RUN go mod download`
- For Java: Ensure `pom.xml` is copied before `RUN mvn dependency:go-offline`
- For .NET: Ensure `.csproj` is copied before `RUN dotnet restore`
- For Python: Ensure `requirements.txt` exists and is valid

### Issue: Multi-stage Build Fails

**Error:**
```
failed to compute cache key: not found
```

**Solution:**
Ensure all COPY commands reference files that exist:
```dockerfile
# Check that these files exist before building
COPY go.mod go.sum ./
COPY pom.xml .
COPY requirements.txt .
```

---

## Best Practices Applied

### ✅ Multi-stage Builds
- Smaller final images
- Build dependencies not included in runtime
- Example: Go services ~50MB instead of ~400MB

### ✅ Layer Caching
- Dependencies installed before source code
- Faster rebuilds when only code changes

```dockerfile
# Good: Dependencies cached separately
COPY go.mod go.sum ./
RUN go mod download
COPY . .  # Only invalidates this layer on code change
```

### ✅ Alpine Base Images
- Smaller image sizes
- Security: Fewer packages = smaller attack surface
- Example: `golang:1.23-alpine` vs `golang:1.23`

### ✅ Non-root User (where applicable)
- Security best practice
- Prevents privilege escalation

### ✅ .dockerignore Files
- Exclude unnecessary files from build context
- Faster builds
- Smaller images

**Should include:**
```
node_modules/
.git/
.env
*.log
dist/
target/
bin/
obj/
```

---

## Verification

### Check All Images Built Successfully

```bash
docker images | grep tiktok
```

Expected output:
```
tiktok-clone-frontend              latest   ...
tiktok-clone-api-gateway           latest   ...
tiktok-clone-user-service          latest   ...
tiktok-clone-video-service         latest   ...
tiktok-clone-interaction-service   latest   ...
tiktok-clone-event-service         latest   ...
tiktok-clone-notification-service  latest   ...
tiktok-clone-feed-service          latest   ...
```

### Check Image Sizes

```bash
docker images --format "table {{.Repository}}\t{{.Size}}" | grep tiktok | sort
```

**Expected sizes (approximate):**
- Go services: 30-50 MB
- Java services: 200-300 MB
- .NET services: 200-250 MB
- Python service: 150-200 MB
- Frontend: 50-100 MB

---

## Next Steps

1. **Start Services:**
   ```bash
   ./docker-start.sh
   ```

2. **Monitor Logs:**
   ```bash
   ./docker-logs.sh all
   ```

3. **Health Check:**
   ```bash
   ./docker-health-check.sh
   ```

4. **Test Application:**
   - Frontend: http://localhost:3000
   - API Gateway: http://localhost:8080

---

**Status:** ✅ All Dockerfiles Fixed and Ready
**Last Updated:** March 11, 2026
