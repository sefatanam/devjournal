# gRPC and Protocol Buffers

## What is gRPC?

gRPC (Google Remote Procedure Call) is a high-performance, open-source framework for building APIs. It uses Protocol Buffers as its interface definition language and serialization format.

## Why gRPC?

| Feature | REST | gRPC |
|---------|------|------|
| Protocol | HTTP/1.1 | HTTP/2 |
| Payload | JSON (text) | Protobuf (binary) |
| Contract | OpenAPI/Swagger | .proto files |
| Streaming | Limited | Full support |
| Code Generation | Optional | Built-in |
| Type Safety | Manual | Automatic |

## Protocol Buffers (Protobuf)

Protocol Buffers are Google's language-neutral, platform-neutral mechanism for serializing structured data.

### Proto File Structure

```protobuf
// proto/devjournal/v1/journal.proto

syntax = "proto3";

package devjournal.v1;

option go_package = "github.com/devjournal/proto/devjournal/v1;devjournalv1";

// Import timestamp for date handling
import "google/protobuf/timestamp.proto";

// Service definition - defines RPC methods
service JournalService {
  // Unary RPC - single request, single response
  rpc CreateEntry(CreateEntryRequest) returns (CreateEntryResponse);
  rpc GetEntry(GetEntryRequest) returns (GetEntryResponse);
  rpc UpdateEntry(UpdateEntryRequest) returns (UpdateEntryResponse);
  rpc DeleteEntry(DeleteEntryRequest) returns (DeleteEntryResponse);

  // Server streaming - single request, stream of responses
  rpc ListEntries(ListEntriesRequest) returns (ListEntriesResponse);

  // Search with filters
  rpc SearchEntries(SearchEntriesRequest) returns (SearchEntriesResponse);
}

// Message definitions
message JournalEntry {
  string id = 1;
  string user_id = 2;
  string title = 3;
  string content = 4;
  JournalMood mood = 5;
  repeated string tags = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp updated_at = 8;
}

// Enum for mood types
enum JournalMood {
  JOURNAL_MOOD_UNSPECIFIED = 0;
  JOURNAL_MOOD_PRODUCTIVE = 1;
  JOURNAL_MOOD_LEARNING = 2;
  JOURNAL_MOOD_STRUGGLING = 3;
  JOURNAL_MOOD_BREAKTHROUGH = 4;
  JOURNAL_MOOD_TIRED = 5;
}

// Request/Response messages
message CreateEntryRequest {
  string title = 1;
  string content = 2;
  JournalMood mood = 3;
  repeated string tags = 4;
}

message CreateEntryResponse {
  JournalEntry entry = 1;
}

message GetEntryRequest {
  string id = 1;
}

message GetEntryResponse {
  JournalEntry entry = 1;
}

message ListEntriesRequest {
  int32 page = 1;
  int32 page_size = 2;
  JournalMood mood_filter = 3;
}

message ListEntriesResponse {
  repeated JournalEntry entries = 1;
  int32 total_count = 2;
  int32 page = 3;
  int32 page_size = 4;
}
```

## Project Proto Files

Located in: `proto/devjournal/v1/`

| File | Purpose |
|------|---------|
| `journal.proto` | Journal entry CRUD operations |
| `snippet.proto` | Code snippet management |
| `user.proto` | Authentication service |
| `progress.proto` | Progress tracking |

## Code Generation with Buf

### Buf Configuration

```yaml
# proto/buf.yaml
version: v2
modules:
  - path: .
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
```

```yaml
# proto/buf.gen.yaml
version: v2
plugins:
  # Go code generation
  - remote: buf.build/protocolbuffers/go
    out: ../services/go-api/proto
    opt:
      - paths=source_relative

  # Go gRPC generation
  - remote: buf.build/grpc/go
    out: ../services/go-api/proto
    opt:
      - paths=source_relative

  # Connect RPC for Go
  - remote: buf.build/connectrpc/go
    out: ../services/go-api/proto
    opt:
      - paths=source_relative

  # TypeScript generation
  - remote: buf.build/bufbuild/es
    out: ../libs/shared/proto/src/lib

  # Connect RPC for TypeScript
  - remote: buf.build/connectrpc/es
    out: ../libs/shared/proto/src/lib
```

### Generate Code

```bash
cd proto
npx buf generate
```

## Go Server Implementation

### Connect RPC Handler

```go
// services/go-api/internal/handler/grpc/journal_connect.go

package grpc

import (
    "context"
    "connectrpc.com/connect"

    pb "github.com/devjournal/proto/devjournal/v1"
    "github.com/devjournal/proto/devjournal/v1/devjournalv1connect"
)

// JournalConnectHandler implements the JournalService
type JournalConnectHandler struct {
    service *service.JournalService
}

// Ensure interface compliance
var _ devjournalv1connect.JournalServiceHandler = (*JournalConnectHandler)(nil)

func NewJournalConnectHandler(svc *service.JournalService) *JournalConnectHandler {
    return &JournalConnectHandler{service: svc}
}

// CreateEntry - Unary RPC implementation
func (h *JournalConnectHandler) CreateEntry(
    ctx context.Context,
    req *connect.Request[pb.CreateEntryRequest],
) (*connect.Response[pb.CreateEntryResponse], error) {
    // Extract user ID from context (set by auth interceptor)
    userID := ctx.Value("user_id").(string)

    // Convert protobuf to domain model
    entry := &domain.JournalEntry{
        UserID:  userID,
        Title:   req.Msg.Title,
        Content: req.Msg.Content,
        Mood:    convertMood(req.Msg.Mood),
        Tags:    req.Msg.Tags,
    }

    // Call service layer
    created, err := h.service.Create(ctx, entry)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    // Convert domain model to protobuf response
    return connect.NewResponse(&pb.CreateEntryResponse{
        Entry: toProtoEntry(created),
    }), nil
}

// GetEntry - Unary RPC implementation
func (h *JournalConnectHandler) GetEntry(
    ctx context.Context,
    req *connect.Request[pb.GetEntryRequest],
) (*connect.Response[pb.GetEntryResponse], error) {
    userID := ctx.Value("user_id").(string)

    entry, err := h.service.GetByID(ctx, req.Msg.Id, userID)
    if err != nil {
        if errors.Is(err, service.ErrNotFound) {
            return nil, connect.NewError(connect.CodeNotFound, err)
        }
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    return connect.NewResponse(&pb.GetEntryResponse{
        Entry: toProtoEntry(entry),
    }), nil
}

// ListEntries - List with pagination
func (h *JournalConnectHandler) ListEntries(
    ctx context.Context,
    req *connect.Request[pb.ListEntriesRequest],
) (*connect.Response[pb.ListEntriesResponse], error) {
    userID := ctx.Value("user_id").(string)

    filter := domain.JournalFilter{
        Page:     int(req.Msg.Page),
        PageSize: int(req.Msg.PageSize),
    }

    if req.Msg.MoodFilter != pb.JournalMood_JOURNAL_MOOD_UNSPECIFIED {
        filter.Mood = convertMood(req.Msg.MoodFilter)
    }

    entries, total, err := h.service.List(ctx, userID, filter)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    protoEntries := make([]*pb.JournalEntry, len(entries))
    for i, e := range entries {
        protoEntries[i] = toProtoEntry(e)
    }

    return connect.NewResponse(&pb.ListEntriesResponse{
        Entries:    protoEntries,
        TotalCount: int32(total),
        Page:       req.Msg.Page,
        PageSize:   req.Msg.PageSize,
    }), nil
}

// Helper: Convert domain to proto
func toProtoEntry(e *domain.JournalEntry) *pb.JournalEntry {
    return &pb.JournalEntry{
        Id:        e.ID,
        UserId:    e.UserID,
        Title:     e.Title,
        Content:   e.Content,
        Mood:      toProtoMood(e.Mood),
        Tags:      e.Tags,
        CreatedAt: timestamppb.New(e.CreatedAt),
        UpdatedAt: timestamppb.New(e.UpdatedAt),
    }
}
```

### Register Handler

```go
// services/go-api/cmd/api/main.go

func main() {
    // ... initialization ...

    // Create Connect RPC handler
    journalHandler := grpc.NewJournalConnectHandler(journalService)

    // Create path and handler
    path, handler := devjournalv1connect.NewJournalServiceHandler(
        journalHandler,
        connect.WithInterceptors(authInterceptor),
    )

    // Create HTTP mux for gRPC
    grpcMux := http.NewServeMux()
    grpcMux.Handle(path, handler)

    // Start gRPC server
    grpcServer := &http.Server{
        Addr:    ":8081",
        Handler: h2c.NewHandler(grpcMux, &http2.Server{}),
    }

    go grpcServer.ListenAndServe()
}
```

## Angular Client Implementation

### Generated Client Types

```typescript
// libs/shared/proto/src/lib/devjournal/v1/journal_pb.ts

// Auto-generated message classes
export class JournalEntry extends Message<JournalEntry> {
  id = "";
  userId = "";
  title = "";
  content = "";
  mood = JournalMood.UNSPECIFIED;
  tags: string[] = [];
  createdAt?: Timestamp;
  updatedAt?: Timestamp;
}

export class CreateEntryRequest extends Message<CreateEntryRequest> {
  title = "";
  content = "";
  mood = JournalMood.UNSPECIFIED;
  tags: string[] = [];
}
```

```typescript
// libs/shared/proto/src/lib/devjournal/v1/journal_connect.ts

// Auto-generated service client
export const JournalService = {
  typeName: "devjournal.v1.JournalService",
  methods: {
    createEntry: {
      name: "CreateEntry",
      I: CreateEntryRequest,
      O: CreateEntryResponse,
      kind: MethodKind.Unary,
    },
    getEntry: {
      name: "GetEntry",
      I: GetEntryRequest,
      O: GetEntryResponse,
      kind: MethodKind.Unary,
    },
    listEntries: {
      name: "ListEntries",
      I: ListEntriesRequest,
      O: ListEntriesResponse,
      kind: MethodKind.Unary,
    },
    // ... more methods
  },
} as const;
```

### Angular gRPC Service

```typescript
// libs/data-access/api/src/lib/journal-grpc.service.ts

import { Injectable, inject } from '@angular/core';
import { createPromiseClient } from '@connectrpc/connect';
import { createGrpcWebTransport } from '@connectrpc/connect-web';
import { JournalService } from '@devjournal/shared-proto';
import {
  CreateEntryRequest,
  ListEntriesRequest,
  JournalMood
} from '@devjournal/shared-proto';
import { API_CONFIG } from './api.config';
import { AuthStore } from '@devjournal/feature-auth';

@Injectable({ providedIn: 'root' })
export class JournalGrpcService {
  private readonly config = inject(API_CONFIG);
  private readonly authStore = inject(AuthStore);

  private readonly transport = createGrpcWebTransport({
    baseUrl: this.config.grpcUrl,
    // Add auth interceptor
    interceptors: [
      (next) => async (req) => {
        const token = this.authStore.token();
        if (token) {
          req.header.set('Authorization', `Bearer ${token}`);
        }
        return next(req);
      },
    ],
  });

  private readonly client = createPromiseClient(JournalService, this.transport);

  // Create entry
  async createEntry(data: { title: string; content: string; mood: string; tags: string[] }) {
    const request = new CreateEntryRequest({
      title: data.title,
      content: data.content,
      mood: this.toProtoMood(data.mood),
      tags: data.tags,
    });

    const response = await this.client.createEntry(request);
    return this.toDomainEntry(response.entry!);
  }

  // List entries with pagination
  async listEntries(page: number, pageSize: number, mood?: string) {
    const request = new ListEntriesRequest({
      page,
      pageSize,
      moodFilter: mood ? this.toProtoMood(mood) : JournalMood.UNSPECIFIED,
    });

    const response = await this.client.listEntries(request);

    return {
      entries: response.entries.map(e => this.toDomainEntry(e)),
      total: response.totalCount,
      page: response.page,
      pageSize: response.pageSize,
    };
  }

  // Get single entry
  async getEntry(id: string) {
    const response = await this.client.getEntry({ id });
    return this.toDomainEntry(response.entry!);
  }

  // Helper: Convert proto mood to string
  private toDomainMood(mood: JournalMood): string {
    const moodMap: Record<JournalMood, string> = {
      [JournalMood.UNSPECIFIED]: '',
      [JournalMood.PRODUCTIVE]: 'productive',
      [JournalMood.LEARNING]: 'learning',
      [JournalMood.STRUGGLING]: 'struggling',
      [JournalMood.BREAKTHROUGH]: 'breakthrough',
      [JournalMood.TIRED]: 'tired',
    };
    return moodMap[mood];
  }

  // Helper: Convert string to proto mood
  private toProtoMood(mood: string): JournalMood {
    const moodMap: Record<string, JournalMood> = {
      'productive': JournalMood.PRODUCTIVE,
      'learning': JournalMood.LEARNING,
      'struggling': JournalMood.STRUGGLING,
      'breakthrough': JournalMood.BREAKTHROUGH,
      'tired': JournalMood.TIRED,
    };
    return moodMap[mood] || JournalMood.UNSPECIFIED;
  }

  // Helper: Convert proto entry to domain model
  private toDomainEntry(entry: ProtoJournalEntry): JournalEntry {
    return {
      id: entry.id,
      userId: entry.userId,
      title: entry.title,
      content: entry.content,
      mood: this.toDomainMood(entry.mood),
      tags: entry.tags,
      createdAt: entry.createdAt?.toDate() || new Date(),
      updatedAt: entry.updatedAt?.toDate() || new Date(),
    };
  }
}
```

## Auth Interceptor (Go)

```go
// services/go-api/internal/handler/grpc/auth_interceptor.go

package grpc

import (
    "context"
    "strings"

    "connectrpc.com/connect"
)

func NewAuthInterceptor(jwtService *service.AuthService) connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            // Get authorization header
            authHeader := req.Header().Get("Authorization")
            if authHeader == "" {
                return nil, connect.NewError(
                    connect.CodeUnauthenticated,
                    errors.New("missing authorization header"),
                )
            }

            // Extract token
            parts := strings.SplitN(authHeader, " ", 2)
            if len(parts) != 2 || parts[0] != "Bearer" {
                return nil, connect.NewError(
                    connect.CodeUnauthenticated,
                    errors.New("invalid authorization format"),
                )
            }

            // Validate token
            claims, err := jwtService.ValidateToken(parts[1])
            if err != nil {
                return nil, connect.NewError(
                    connect.CodeUnauthenticated,
                    errors.New("invalid token"),
                )
            }

            // Add user ID to context
            ctx = context.WithValue(ctx, "user_id", claims.UserID)

            return next(ctx, req)
        }
    }
}
```

## Connect vs Traditional gRPC

DevJournal uses **Connect RPC** instead of traditional gRPC for several reasons:

| Feature | Traditional gRPC | Connect RPC |
|---------|-----------------|-------------|
| Browser Support | Requires proxy | Native support |
| HTTP Protocol | HTTP/2 only | HTTP/1.1 and HTTP/2 |
| Serialization | Protobuf only | Protobuf + JSON |
| Debugging | Difficult | Easy (JSON mode) |
| Go Integration | Complex | Simple HTTP handlers |

## Key Takeaways

1. **Proto files are the contract** - Define once, generate for all languages
2. **Type safety everywhere** - No manual serialization/deserialization
3. **Efficient binary format** - Smaller payloads than JSON
4. **Code generation** - Less boilerplate, fewer bugs
5. **Connect RPC** - Modern gRPC that works in browsers
6. **Interceptors** - Cross-cutting concerns like auth

## Commands Reference

```bash
# Generate proto code
cd proto && npx buf generate

# Lint proto files
npx buf lint

# Check breaking changes
npx buf breaking --against '.git#branch=main'

# Format proto files
npx buf format -w
```

## Next Steps

- [REST API Implementation](./02-rest-api.md) - Compare with REST approach
- [Angular Signal Store](./05-signal-store.md) - Using gRPC with signals
