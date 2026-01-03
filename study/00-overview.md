# DevJournal - Project Overview

## What is DevJournal?

DevJournal is a full-stack developer learning platform designed to teach modern web development patterns through practical implementation. It's a polyglot monorepo combining Angular frontend with a Go backend.

## Key Features

1. **Journal Entries** - Daily developer logs with mood tracking
2. **Code Snippets Library** - Store and organize code snippets with syntax highlighting
3. **Progress Tracking** - Streaks, stats, and learning analytics
4. **Study Chat** - Real-time WebSocket messaging for study groups

## Technology Stack

### Frontend
- **Angular 19** - Latest Angular with standalone components
- **NgRx Signal Store** - Modern state management with signals
- **SCSS** - CSS preprocessing with design tokens
- **SSR** - Server-side rendering for SEO

### Backend
- **Go (Golang)** - High-performance backend
- **Raw SQL with pgx** - No ORM, explicit database queries
- **Connect RPC** - gRPC-Web compatible API
- **gorilla/websocket** - Real-time communication

### Databases
- **PostgreSQL** - Relational data (users, journals, progress)
- **MongoDB** - Document storage (flexible code snippets)

### Infrastructure
- **Docker Compose** - Container orchestration
- **Envoy Proxy** - gRPC-Web gateway
- **Nx Monorepo** - Build system and workspace management

## Project Structure

```
devjournal/
├── apps/
│   └── web/                    # Angular frontend application
│       └── src/app/
│           ├── pages/          # Route pages
│           └── app.routes.ts   # Application routes
│
├── libs/
│   ├── features/               # Feature modules
│   │   ├── auth/              # Authentication
│   │   ├── journal/           # Journal entries
│   │   ├── snippets/          # Code snippets
│   │   ├── progress/          # Progress tracking
│   │   └── chat/              # Real-time chat
│   │
│   ├── data-access/
│   │   └── api/               # API services (REST + gRPC)
│   │
│   └── shared/
│       ├── models/            # TypeScript interfaces
│       ├── proto/             # Generated proto types
│       └── ui/                # Shared UI components
│
├── services/
│   └── go-api/                # Go backend
│       ├── cmd/api/           # Entry point
│       └── internal/
│           ├── config/        # Configuration
│           ├── database/      # DB connections
│           ├── domain/        # Domain models
│           ├── repository/    # Data access
│           ├── service/       # Business logic
│           ├── handler/       # API handlers
│           └── middleware/    # HTTP middleware
│
├── proto/
│   └── devjournal/v1/         # Protocol Buffer definitions
│
└── docker/
    └── docker-compose.yml     # Container orchestration
```

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                         FRONTEND                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Components  │  │ Signal Store │  │ API Services │          │
│  │  (Angular)   │──│   (NgRx)     │──│ (REST/gRPC)  │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTP / gRPC / WebSocket
┌────────────────────────────┴────────────────────────────────────┐
│                          BACKEND                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Handlers   │──│   Services   │──│ Repositories │          │
│  │  (REST/gRPC) │  │   (Logic)    │  │ (Data Access)│          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────┴────────────────────────────────────┐
│                        DATABASES                                 │
│  ┌────────────────────┐         ┌────────────────────┐          │
│  │    PostgreSQL      │         │      MongoDB       │          │
│  │  (Users, Journals) │         │    (Snippets)      │          │
│  └────────────────────┘         └────────────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

## API Protocols

### REST API (Port 8080)
Traditional HTTP/JSON API for standard CRUD operations.

### gRPC/Connect RPC (Port 8081)
Type-safe, high-performance API using Protocol Buffers.

### WebSocket (Port 8080/ws)
Real-time bidirectional communication for chat.

## Key Learning Concepts

1. **Clean Architecture** - Separation of concerns
2. **Repository Pattern** - Data access abstraction
3. **Signal-based State** - Reactive UI state management
4. **Dual Protocol APIs** - REST and gRPC side-by-side
5. **Polyglot Persistence** - SQL and NoSQL together
6. **Real-time Features** - WebSocket implementation
7. **JWT Authentication** - Stateless auth tokens
8. **Docker Deployment** - Container orchestration

## Next Steps

Continue to the following study guides:
1. [gRPC and Protocol Buffers](./01-grpc-protobuf.md)
2. [REST API Implementation](./02-rest-api.md)
3. [PostgreSQL Database](./03-postgresql.md)
4. [MongoDB Database](./04-mongodb.md)
5. [Angular Signal Store](./05-signal-store.md)
6. [WebSocket Chat](./06-websocket-chat.md)
7. [Authentication & JWT](./07-authentication.md)
8. [Docker Deployment](./08-docker-deployment.md)
