# DevJournal - Polyglot Learning Platform

A full-stack learning journal application designed to teach Angular 21 + Go backend development patterns. Built for studying modern web development practices and preparing for intermediate/senior developer interviews.

[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/deploy/devjournal?referralCode=rso6FB&utm_medium=integration&utm_source=template&utm_campaign=generic)
## Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Frontend** | Angular 21 + Signal Store | Modern reactive UI with signals |
| **Backend** | Go (Golang) | High-performance API server |
| **API Protocols** | REST/JSON + gRPC + Protobuf | Multiple API patterns |
| **Real-time** | WebSocket (gorilla/websocket) | Live chat functionality |
| **Primary DB** | PostgreSQL (raw SQL with pgx) | Relational data storage |
| **Secondary DB** | MongoDB | Flexible document storage |
| **Monorepo** | Nx Workspace | Project management |
| **Deployment** | Docker + Docker Compose | Containerization |

## Project Structure

```
devjournal/
├── apps/
│   └── web/                    # Angular 21 frontend
├── libs/
│   ├── shared/
│   │   ├── ui/                 # Reusable UI components
│   │   ├── util/               # Utilities
│   │   └── models/             # TypeScript interfaces
│   ├── features/
│   │   ├── auth/               # Authentication feature
│   │   ├── journal/            # Journal entries feature
│   │   ├── snippets/           # Code snippets feature
│   │   ├── chat/               # Real-time chat feature
│   │   └── progress/           # Progress tracking feature
│   └── data-access/
│       └── api/                # HTTP/gRPC clients
├── services/
│   └── go-api/                 # Go backend service
│       ├── cmd/api/            # Entry point
│       ├── internal/           # Private packages
│       │   ├── config/         # Configuration
│       │   ├── handler/        # HTTP/gRPC/WebSocket handlers
│       │   ├── middleware/     # Auth, CORS, logging
│       │   ├── repository/     # PostgreSQL & MongoDB repos
│       │   ├── service/        # Business logic
│       │   └── domain/         # Domain models
│       └── pkg/                # Public utilities
├── proto/                      # Protobuf definitions
│   └── devjournal/v1/
├── docker/                     # Docker configuration
└── docs/                       # Documentation
```

## Quick Start

### Prerequisites

- **Node.js** >= 20.x
- **Go** >= 1.21
- **Docker** and Docker Compose
- **protoc** (Protocol Buffers compiler)
- **buf** CLI (for proto generation)

### 1. Clone and Install

```bash
# Clone the repository
git clone <repo-url>
cd devjournal

# Install Node.js dependencies
npm install

# Install Go dependencies
cd services/go-api
go mod download
cd ../..
```

### 2. Start Databases (Docker)

```bash
# Start PostgreSQL and MongoDB
docker-compose -f docker/docker-compose.dev.yml up -d

# Verify containers are running
docker ps
```

### 3. Run Database Migrations

The migrations run automatically when PostgreSQL starts. To run manually:

```bash
# Connect to PostgreSQL
docker exec -it devjournal-postgres psql -U devjournal -d devjournal

# Run migration files (if needed)
\i /docker-entrypoint-initdb.d/001_create_users.sql
```

### 4. Start the Backend

```bash
# From project root
cd services/go-api
go run ./cmd/api/main.go

# Or using Nx
npx nx serve go-api
```

The API will be available at:
- HTTP: http://localhost:8080
- gRPC: http://localhost:8081
- Health check: http://localhost:8080/health

### 5. Start the Frontend

```bash
# From project root
npx nx serve web

# Or with specific port
npx ng serve --port 4200
```

Open http://localhost:4200 in your browser.

## Development Commands

### Nx Commands

```bash
# Serve Angular frontend
npx nx serve web

# Build Angular for production
npx nx build web --configuration=production

# Run tests
npx nx test web
npx nx test go-api

# Lint code
npx nx lint web
npx nx lint go-api

# Generate Angular library
npx nx g @nx/angular:library my-lib --directory=libs/shared/my-lib

# Generate Angular component
npx nx g @nx/angular:component my-component --project=feature-auth
```

### Go Commands

```bash
# Run Go server
cd services/go-api
go run ./cmd/api/main.go

# Build Go binary
go build -o dist/server ./cmd/api

# Run tests
go test -v ./...

# Run with race detector
go run -race ./cmd/api/main.go
```

### Docker Commands

```bash
# Development (databases only)
docker-compose -f docker/docker-compose.dev.yml up -d

# Full stack
docker-compose -f docker/docker-compose.yml up -d

# View logs
docker-compose -f docker/docker-compose.dev.yml logs -f

# Stop containers
docker-compose -f docker/docker-compose.dev.yml down

# Stop and remove volumes
docker-compose -f docker/docker-compose.dev.yml down -v
```

### Protobuf Generation

```bash
# Install buf CLI (macOS)
brew install bufbuild/buf/buf

# Generate code
cd proto
buf generate

# Lint proto files
buf lint

# Check for breaking changes
buf breaking --against '.git#branch=main'
```

## API Endpoints

### REST API (HTTP)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/auth/register | Register new user |
| POST | /api/auth/login | Login user |
| GET | /api/entries | List journal entries |
| POST | /api/entries | Create journal entry |
| GET | /api/entries/:id | Get journal entry |
| PUT | /api/entries/:id | Update journal entry |
| DELETE | /api/entries/:id | Delete journal entry |
| GET | /api/snippets | List code snippets |
| POST | /api/snippets | Create code snippet |
| GET | /api/snippets/:id | Get code snippet |
| PUT | /api/snippets/:id | Update code snippet |
| DELETE | /api/snippets/:id | Delete code snippet |

### WebSocket

```
ws://localhost:8080/ws/chat/{room}
```

### gRPC Services

- `JournalService` - CRUD operations for journal entries
- `SnippetService` - CRUD operations for code snippets
- `AuthService` - User authentication
- `ProgressService` - Learning progress tracking

## Database Schemas

### PostgreSQL Tables

- `users` - User accounts
- `journal_entries` - Learning journal entries
- `learning_progress` - Daily progress tracking
- `study_groups` - Chat room groups
- `study_group_members` - Group membership

### MongoDB Collections

- `snippets` - Code snippets with flexible metadata

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| HTTP_PORT | 8080 | HTTP server port |
| GRPC_PORT | 8081 | gRPC server port |
| POSTGRES_URL | - | PostgreSQL connection string |
| MONGO_URL | - | MongoDB connection string |
| MONGO_DB | devjournal | MongoDB database name |
| JWT_SECRET | - | JWT signing secret |
| ENVIRONMENT | development | Runtime environment |

## Key Learning Patterns

### Go Backend Patterns

1. **Clean Architecture** - cmd/internal/pkg structure
2. **Repository Pattern** - Raw SQL with pgx (no ORM)
3. **Service Layer** - Business logic separation
4. **Middleware Chain** - Auth, CORS, logging, recovery
5. **WebSocket Hub** - Concurrent connection management
6. **gRPC Server** - Protobuf-based RPC

### Angular Frontend Patterns

1. **Signal Store** - @ngrx/signals for state management
2. **Standalone Components** - Modern Angular architecture
3. **Modern Control Flow** - @if, @for, @switch
4. **Functional Guards** - Route protection
5. **HTTP Interceptors** - Token injection
6. **WebSocket Integration** - RxJS webSocket

### Database Patterns

1. **Raw SQL** - No ORM, explicit queries
2. **Migrations** - Version-controlled schema
3. **Connection Pooling** - pgxpool for PostgreSQL
4. **Flexible Schema** - MongoDB for varied content

## Implementation Phases

1. **Phase 1** - Project Foundation (Nx, Go skeleton, Docker) ✅
2. **Phase 2** - Authentication System (JWT, guards)
3. **Phase 3** - Journal Entries (REST CRUD)
4. **Phase 4** - Code Snippets (MongoDB)
5. **Phase 5** - gRPC Integration
6. **Phase 6** - Real-time Chat (WebSocket)
7. **Phase 7** - Progress Tracking
8. **Phase 8** - Polish & Documentation

## Nx Workspace

This project uses [Nx](https://nx.dev) for monorepo management.

```bash
# Show project graph
npx nx graph

# Show available targets for a project
npx nx show project web

# Run affected tests only
npx nx affected -t test
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - feel free to use this project for learning purposes.

---

Built with care for developers learning full-stack development.
