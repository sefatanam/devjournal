# DevJournal Study Guide

Comprehensive documentation for learning full-stack development through the DevJournal project.

## Table of Contents

| # | Topic | Description |
|---|-------|-------------|
| 00 | [Overview](./00-overview.md) | Project architecture and technology stack |
| 01 | [gRPC & Protocol Buffers](./01-grpc-protobuf.md) | Type-safe APIs with Protobuf and Connect RPC |
| 02 | [REST API](./02-rest-api.md) | Traditional HTTP/JSON API implementation |
| 03 | [PostgreSQL](./03-postgresql.md) | Relational database with pgx and raw SQL |
| 04 | [MongoDB](./04-mongodb.md) | Document storage for flexible data |
| 05 | [Angular Signal Store](./05-signal-store.md) | Modern state management with NgRx signals |
| 06 | [WebSocket Chat](./06-websocket-chat.md) | Real-time communication with gorilla/websocket |
| 07 | [Authentication](./07-authentication.md) | JWT-based auth with bcrypt password hashing |
| 08 | [Docker Deployment](./08-docker-deployment.md) | Container orchestration with Docker Compose |

## Quick Start

```bash
# Clone the repository
git clone https://github.com/yourusername/devjournal.git
cd devjournal

# Start all services with Docker
cd docker
docker compose up -d

# Access the application
# Frontend: http://localhost:4200
# REST API: http://localhost:8080
# gRPC API: http://localhost:8081
# gRPC-Web: http://localhost:9090
```

## Learning Path

### Beginner Path
1. Start with [Overview](./00-overview.md) to understand the architecture
2. Learn [REST API](./02-rest-api.md) - most familiar pattern
3. Explore [PostgreSQL](./03-postgresql.md) for database basics
4. Study [Authentication](./07-authentication.md) for security concepts

### Intermediate Path
1. Deep dive into [Angular Signal Store](./05-signal-store.md)
2. Learn [MongoDB](./04-mongodb.md) for NoSQL patterns
3. Explore [WebSocket Chat](./06-websocket-chat.md) for real-time features
4. Master [Docker Deployment](./08-docker-deployment.md)

### Advanced Path
1. Study [gRPC & Protocol Buffers](./01-grpc-protobuf.md) for type-safe APIs
2. Compare REST vs gRPC implementations
3. Explore polyglot persistence patterns
4. Learn production deployment strategies

## Key Technologies

### Frontend
- Angular 19 with standalone components
- NgRx Signal Store for state management
- Modern control flow (@if, @for, @switch)
- SCSS with design tokens

### Backend
- Go with Chi router
- Connect RPC (gRPC-Web compatible)
- gorilla/websocket for real-time
- Raw SQL with pgx (no ORM)

### Databases
- PostgreSQL 16 for relational data
- MongoDB 7 for document storage

### Infrastructure
- Docker & Docker Compose
- Envoy Proxy for gRPC-Web
- Multi-stage Docker builds

## Code Structure

```
devjournal/
├── apps/web/              # Angular frontend
├── libs/
│   ├── features/          # Feature modules (auth, journal, etc.)
│   ├── data-access/api/   # API services
│   └── shared/            # Models, proto, UI components
├── services/go-api/       # Go backend
├── proto/                 # Protocol Buffer definitions
├── docker/                # Docker configuration
└── study/                 # This documentation
```

## Contributing

Feel free to improve these study materials by:
1. Fixing errors or typos
2. Adding more examples
3. Clarifying complex concepts
4. Adding new topics

## License

MIT License - Learn, modify, and share freely.
