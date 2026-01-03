# Docker Deployment

## Overview

DevJournal uses Docker and Docker Compose for containerized deployment. This ensures consistent environments across development, staging, and production.

## Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                        Docker Network                           │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Angular    │  │   Go API     │  │    Envoy     │         │
│  │   Frontend   │  │   Backend    │  │    Proxy     │         │
│  │   (SSR)      │  │              │  │   (gRPC-Web) │         │
│  │   :4000      │  │  :8080/:8081 │  │    :9090     │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
│         │                 │                  │                  │
│         └────────┬────────┴─────────┬───────┘                  │
│                  │                  │                           │
│         ┌───────┴──────┐   ┌───────┴──────┐                    │
│         │  PostgreSQL  │   │   MongoDB    │                    │
│         │    :5432     │   │    :27017    │                    │
│         └──────────────┘   └──────────────┘                    │
└────────────────────────────────────────────────────────────────┘
                              │
                         Port Mapping
                              │
┌─────────────────────────────┴──────────────────────────────────┐
│  Host Machine                                                   │
│  4200 → Frontend                                               │
│  8080 → REST API                                               │
│  8081 → gRPC API                                               │
│  9090 → Envoy (gRPC-Web)                                       │
└────────────────────────────────────────────────────────────────┘
```

## Docker Compose Configuration

```yaml
# docker/docker-compose.yml

version: '3.8'

services:
  # ============================================
  # PostgreSQL Database
  # ============================================
  postgres:
    image: postgres:16-alpine
    container_name: devjournal-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-devjournal}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-devpass}
      POSTGRES_DB: ${POSTGRES_DB:-devjournal}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init/postgres:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-devjournal}"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - devjournal-network

  # ============================================
  # MongoDB Database
  # ============================================
  mongodb:
    image: mongo:7
    container_name: devjournal-mongodb
    restart: unless-stopped
    environment:
      MONGO_INITDB_ROOT_USERNAME: ${MONGO_USER:-devjournal}
      MONGO_INITDB_ROOT_PASSWORD: ${MONGO_PASSWORD:-devpass}
      MONGO_INITDB_DATABASE: ${MONGO_DB:-devjournal}
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
      - ./init/mongo:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - devjournal-network

  # ============================================
  # Go API Backend
  # ============================================
  api:
    build:
      context: ../services/go-api
      dockerfile: Dockerfile
    container_name: devjournal-api
    restart: unless-stopped
    environment:
      PORT: 8080
      GRPC_PORT: 8081
      DB_URL: postgres://${POSTGRES_USER:-devjournal}:${POSTGRES_PASSWORD:-devpass}@postgres:5432/${POSTGRES_DB:-devjournal}?sslmode=disable
      MONGO_URL: mongodb://${MONGO_USER:-devjournal}:${MONGO_PASSWORD:-devpass}@mongodb:27017/${MONGO_DB:-devjournal}?authSource=admin
      MONGO_DB: ${MONGO_DB:-devjournal}
      JWT_SECRET: ${JWT_SECRET:-your-super-secret-jwt-key-change-in-production}
      ENVIRONMENT: ${ENVIRONMENT:-development}
    ports:
      - "8080:8080"   # REST API
      - "8081:8081"   # gRPC API
    depends_on:
      postgres:
        condition: service_healthy
      mongodb:
        condition: service_healthy
    networks:
      - devjournal-network

  # ============================================
  # Envoy Proxy (gRPC-Web Gateway)
  # ============================================
  envoy:
    image: envoyproxy/envoy:v1.28-latest
    container_name: devjournal-envoy
    restart: unless-stopped
    volumes:
      - ./envoy/envoy.yaml:/etc/envoy/envoy.yaml:ro
    ports:
      - "9090:9090"
    depends_on:
      - api
    networks:
      - devjournal-network

  # ============================================
  # Angular Frontend (SSR)
  # ============================================
  frontend:
    build:
      context: ..
      dockerfile: docker/Dockerfile.frontend
    container_name: devjournal-frontend
    restart: unless-stopped
    environment:
      PORT: 4000
      API_URL: http://api:8080
      GRPC_URL: http://envoy:9090
    ports:
      - "4200:4000"
    depends_on:
      - api
      - envoy
    networks:
      - devjournal-network

# ============================================
# Volumes
# ============================================
volumes:
  postgres_data:
    driver: local
  mongodb_data:
    driver: local

# ============================================
# Networks
# ============================================
networks:
  devjournal-network:
    driver: bridge
```

## Dockerfiles

### Go API Dockerfile

```dockerfile
# docker/Dockerfile.backend (or services/go-api/Dockerfile)

# ============================================
# Build Stage
# ============================================
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /app/api \
    ./cmd/api

# ============================================
# Runtime Stage
# ============================================
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/api .

# Use non-root user
USER appuser

# Expose ports
EXPOSE 8080 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./api"]
```

### Angular Frontend Dockerfile

```dockerfile
# docker/Dockerfile.frontend

# ============================================
# Build Stage
# ============================================
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build the application (SSR)
RUN npm run build:web:production

# ============================================
# Runtime Stage
# ============================================
FROM node:20-alpine

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy built application
COPY --from=builder /app/dist/apps/web ./dist
COPY --from=builder /app/node_modules ./node_modules

# Use non-root user
USER appuser

# Expose port
EXPOSE 4000

# Environment variables
ENV PORT=4000
ENV NODE_ENV=production

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:4000/health || exit 1

# Run the SSR server
CMD ["node", "dist/server/server.mjs"]
```

## Envoy Configuration (gRPC-Web)

```yaml
# docker/envoy/envoy.yaml

admin:
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 9901

static_resources:
  listeners:
    - name: listener_0
      address:
        socket_address:
          address: 0.0.0.0
          port_value: 9090
      filter_chains:
        - filters:
            - name: envoy.filters.network.http_connection_manager
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                codec_type: AUTO
                stat_prefix: ingress_http
                route_config:
                  name: local_route
                  virtual_hosts:
                    - name: local_service
                      domains: ["*"]
                      routes:
                        - match:
                            prefix: "/"
                          route:
                            cluster: grpc_service
                            timeout: 60s
                            max_stream_duration:
                              grpc_timeout_header_max: 60s
                      cors:
                        allow_origin_string_match:
                          - prefix: "*"
                        allow_methods: GET, PUT, DELETE, POST, OPTIONS
                        allow_headers: authorization, keep-alive, user-agent, cache-control, content-type, content-transfer-encoding, x-accept-content-transfer-encoding, x-accept-response-streaming, x-user-agent, x-grpc-web, grpc-timeout
                        expose_headers: grpc-status, grpc-message
                        max_age: "1728000"
                http_filters:
                  - name: envoy.filters.http.grpc_web
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.grpc_web.v3.GrpcWeb
                  - name: envoy.filters.http.cors
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.cors.v3.Cors
                  - name: envoy.filters.http.router
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router

  clusters:
    - name: grpc_service
      connect_timeout: 5s
      type: LOGICAL_DNS
      lb_policy: ROUND_ROBIN
      http2_protocol_options: {}
      load_assignment:
        cluster_name: grpc_service
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: api
                      port_value: 8081
```

## Database Initialization Scripts

### PostgreSQL Init

```sql
-- docker/init/postgres/01-init.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    avatar_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create journal_entries table
CREATE TABLE IF NOT EXISTS journal_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    mood VARCHAR(50),
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_journal_entries_user_id ON journal_entries(user_id);
CREATE INDEX IF NOT EXISTS idx_journal_entries_created_at ON journal_entries(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Full-text search index
CREATE INDEX IF NOT EXISTS idx_journal_entries_search ON journal_entries
    USING GIN (to_tsvector('english', title || ' ' || content));
```

### MongoDB Init

```javascript
// docker/init/mongo/01-init.js

db = db.getSiblingDB('devjournal');

// Create snippets collection with indexes
db.createCollection('snippets');

db.snippets.createIndex({ user_id: 1, created_at: -1 });
db.snippets.createIndex({ language: 1 });
db.snippets.createIndex({ tags: 1 });
db.snippets.createIndex(
  { title: "text", description: "text", code: "text" },
  { weights: { title: 10, description: 5, code: 1 } }
);

print('MongoDB initialization complete');
```

## Environment Variables

### Development (.env.development)

```env
# Database
POSTGRES_USER=devjournal
POSTGRES_PASSWORD=devpass
POSTGRES_DB=devjournal

MONGO_USER=devjournal
MONGO_PASSWORD=devpass
MONGO_DB=devjournal

# API
JWT_SECRET=dev-secret-key-not-for-production
ENVIRONMENT=development

# Frontend
API_URL=http://localhost:8080
GRPC_URL=http://localhost:9090
```

### Production (.env.production)

```env
# Database - USE STRONG PASSWORDS!
POSTGRES_USER=devjournal_prod
POSTGRES_PASSWORD=CHANGE_ME_STRONG_PASSWORD_123!@#
POSTGRES_DB=devjournal

MONGO_USER=devjournal_prod
MONGO_PASSWORD=CHANGE_ME_ANOTHER_STRONG_PASSWORD_456!@#
MONGO_DB=devjournal

# API - USE STRONG SECRET!
JWT_SECRET=CHANGE_ME_GENERATE_A_SECURE_256_BIT_SECRET_KEY
ENVIRONMENT=production

# Frontend
API_URL=https://api.devjournal.com
GRPC_URL=https://grpc.devjournal.com
```

## Docker Commands

### Development

```bash
# Start all services
cd docker
docker compose up -d

# View logs
docker compose logs -f

# View specific service logs
docker compose logs -f api
docker compose logs -f frontend

# Stop all services
docker compose down

# Stop and remove volumes (WARNING: deletes data!)
docker compose down -v

# Rebuild a specific service
docker compose build api
docker compose up -d api

# Rebuild all services
docker compose build
docker compose up -d
```

### Production

```bash
# Start with production config
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Scale services
docker compose up -d --scale api=3

# Update a service with zero downtime
docker compose up -d --no-deps --build api
```

## Health Checks

### API Health Endpoint

```go
// services/go-api/cmd/api/main.go

r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    health := map[string]string{
        "status": "healthy",
        "time":   time.Now().Format(time.RFC3339),
    }

    // Check PostgreSQL
    if err := pgPool.Ping(r.Context()); err != nil {
        health["postgres"] = "unhealthy"
        health["status"] = "degraded"
    } else {
        health["postgres"] = "healthy"
    }

    // Check MongoDB
    if err := mongoClient.Ping(r.Context(), nil); err != nil {
        health["mongodb"] = "unhealthy"
        health["status"] = "degraded"
    } else {
        health["mongodb"] = "healthy"
    }

    w.Header().Set("Content-Type", "application/json")
    if health["status"] != "healthy" {
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    json.NewEncoder(w).Encode(health)
})
```

## Production Checklist

### Security
- [ ] Change all default passwords
- [ ] Use strong JWT secret (256-bit)
- [ ] Enable HTTPS with valid certificates
- [ ] Configure proper CORS origins
- [ ] Run containers as non-root users
- [ ] Use secrets management (Docker secrets, Vault)

### Performance
- [ ] Configure proper resource limits
- [ ] Enable connection pooling
- [ ] Set up caching (Redis)
- [ ] Configure log rotation
- [ ] Set up monitoring (Prometheus, Grafana)

### Reliability
- [ ] Configure health checks
- [ ] Set up automated backups
- [ ] Configure restart policies
- [ ] Set up alerting
- [ ] Document recovery procedures

### Monitoring

```yaml
# Add to docker-compose.yml for monitoring
services:
  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9091:9090"

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
```

## Multi-Stage Builds Benefits

1. **Smaller images** - Only runtime files included
2. **Security** - No build tools in production
3. **Caching** - Faster rebuilds
4. **Consistency** - Same build everywhere

## Key Takeaways

1. **Use multi-stage builds** - Smaller, more secure images
2. **Health checks** - Enable orchestrator self-healing
3. **Environment variables** - Configuration without rebuilding
4. **Named volumes** - Persistent data survives restarts
5. **Networks** - Service isolation and discovery
6. **Non-root users** - Security best practice
7. **Secrets management** - Never commit passwords

## Next Steps

- [Overview](./00-overview.md) - Return to project overview
- [Authentication](./07-authentication.md) - Secure your deployment
