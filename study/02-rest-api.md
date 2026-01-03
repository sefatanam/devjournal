# REST API Implementation

## Overview

DevJournal implements a RESTful API using Go's standard `net/http` package with the `chi` router. The API follows REST conventions and returns JSON responses.

## API Structure

```
/api
├── /auth
│   ├── POST /register     # Create new user
│   ├── POST /login        # Authenticate user
│   ├── GET  /me           # Get current user
│   └── POST /refresh      # Refresh JWT token
│
├── /entries               # Journal entries
│   ├── GET    /           # List entries (paginated)
│   ├── POST   /           # Create entry
│   ├── GET    /{id}       # Get single entry
│   ├── PUT    /{id}       # Update entry
│   └── DELETE /{id}       # Delete entry
│
├── /snippets              # Code snippets
│   ├── GET    /           # List snippets
│   ├── POST   /           # Create snippet
│   ├── GET    /{id}       # Get single snippet
│   ├── PUT    /{id}       # Update snippet
│   ├── DELETE /{id}       # Delete snippet
│   └── GET    /stats/languages  # Language statistics
│
├── /groups                # Study groups
│   ├── GET    /           # List user's groups
│   ├── GET    /discover   # Public groups
│   ├── POST   /           # Create group
│   ├── GET    /{id}       # Get group details
│   ├── DELETE /{id}       # Delete group
│   ├── POST   /{id}/join  # Join group
│   ├── POST   /{id}/leave # Leave group
│   └── GET    /{id}/members # List members
│
└── /progress              # Progress tracking
    ├── GET /summary       # Overall stats
    ├── GET /today         # Today's activity
    ├── GET /weekly        # Weekly breakdown
    └── GET /streak        # Current streak
```

## Router Setup with Chi

```go
// services/go-api/cmd/api/main.go

package main

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
)

func main() {
    // Initialize dependencies (DB, services, etc.)
    // ...

    // Create router
    r := chi.NewRouter()

    // Global middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.RequestID)
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:4200", "http://localhost:4000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Authorization", "Content-Type"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    // Public routes
    r.Route("/api/auth", func(r chi.Router) {
        r.Post("/register", authHandler.Register)
        r.Post("/login", authHandler.Login)
        r.Post("/refresh", authHandler.RefreshToken)
    })

    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(authMiddleware.Authenticate)

        // Auth
        r.Get("/api/auth/me", authHandler.GetProfile)

        // Journal entries
        r.Route("/api/entries", func(r chi.Router) {
            r.Get("/", journalHandler.List)
            r.Post("/", journalHandler.Create)
            r.Get("/{id}", journalHandler.GetByID)
            r.Put("/{id}", journalHandler.Update)
            r.Delete("/{id}", journalHandler.Delete)
        })

        // Snippets
        r.Route("/api/snippets", func(r chi.Router) {
            r.Get("/", snippetHandler.List)
            r.Post("/", snippetHandler.Create)
            r.Get("/stats/languages", snippetHandler.GetLanguageStats)
            r.Get("/{id}", snippetHandler.GetByID)
            r.Put("/{id}", snippetHandler.Update)
            r.Delete("/{id}", snippetHandler.Delete)
        })

        // Study groups
        r.Route("/api/groups", func(r chi.Router) {
            r.Get("/", groupHandler.ListMyGroups)
            r.Get("/discover", groupHandler.ListPublicGroups)
            r.Post("/", groupHandler.Create)
            r.Get("/{id}", groupHandler.GetByID)
            r.Delete("/{id}", groupHandler.Delete)
            r.Post("/{id}/join", groupHandler.Join)
            r.Post("/{id}/leave", groupHandler.Leave)
            r.Get("/{id}/members", groupHandler.ListMembers)
        })

        // Progress
        r.Route("/api/progress", func(r chi.Router) {
            r.Get("/summary", progressHandler.GetSummary)
            r.Get("/today", progressHandler.GetToday)
            r.Get("/weekly", progressHandler.GetWeekly)
            r.Get("/streak", progressHandler.GetStreak)
        })
    })

    // WebSocket (separate auth check)
    r.Get("/ws/chat/{roomId}", chatHandler.HandleWebSocket)

    // Start server
    log.Println("Starting server on :8080")
    http.ListenAndServe(":8080", r)
}
```

## REST Handler Implementation

### Journal Handler

```go
// services/go-api/internal/handler/rest/journal_handler.go

package rest

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"
)

type JournalHandler struct {
    service *service.JournalService
}

func NewJournalHandler(svc *service.JournalService) *JournalHandler {
    return &JournalHandler{service: svc}
}

// List - GET /api/entries
func (h *JournalHandler) List(w http.ResponseWriter, r *http.Request) {
    // Get user ID from context (set by auth middleware)
    userID := r.Context().Value("user_id").(string)

    // Parse query parameters
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }

    pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
    if pageSize < 1 || pageSize > 50 {
        pageSize = 10
    }

    mood := r.URL.Query().Get("mood")
    search := r.URL.Query().Get("search")

    // Build filter
    filter := domain.JournalFilter{
        Page:     page,
        PageSize: pageSize,
        Mood:     mood,
        Search:   search,
    }

    // Call service
    entries, total, err := h.service.List(r.Context(), userID, filter)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to fetch entries")
        return
    }

    // Build response
    response := map[string]interface{}{
        "data":       entries,
        "total":      total,
        "page":       page,
        "pageSize":   pageSize,
        "totalPages": (total + pageSize - 1) / pageSize,
    }

    respondJSON(w, http.StatusOK, response)
}

// Create - POST /api/entries
func (h *JournalHandler) Create(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)

    // Parse request body
    var req CreateEntryRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Validate
    if req.Title == "" {
        respondError(w, http.StatusBadRequest, "Title is required")
        return
    }
    if req.Content == "" {
        respondError(w, http.StatusBadRequest, "Content is required")
        return
    }

    // Create entry
    entry := &domain.JournalEntry{
        UserID:  userID,
        Title:   req.Title,
        Content: req.Content,
        Mood:    req.Mood,
        Tags:    req.Tags,
    }

    created, err := h.service.Create(r.Context(), entry)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to create entry")
        return
    }

    respondJSON(w, http.StatusCreated, created)
}

// GetByID - GET /api/entries/{id}
func (h *JournalHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    entryID := chi.URLParam(r, "id")

    entry, err := h.service.GetByID(r.Context(), entryID, userID)
    if err != nil {
        if errors.Is(err, service.ErrNotFound) {
            respondError(w, http.StatusNotFound, "Entry not found")
            return
        }
        if errors.Is(err, service.ErrForbidden) {
            respondError(w, http.StatusForbidden, "Access denied")
            return
        }
        respondError(w, http.StatusInternalServerError, "Failed to fetch entry")
        return
    }

    respondJSON(w, http.StatusOK, entry)
}

// Update - PUT /api/entries/{id}
func (h *JournalHandler) Update(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    entryID := chi.URLParam(r, "id")

    var req UpdateEntryRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    entry := &domain.JournalEntry{
        ID:      entryID,
        UserID:  userID,
        Title:   req.Title,
        Content: req.Content,
        Mood:    req.Mood,
        Tags:    req.Tags,
    }

    updated, err := h.service.Update(r.Context(), entry)
    if err != nil {
        if errors.Is(err, service.ErrNotFound) {
            respondError(w, http.StatusNotFound, "Entry not found")
            return
        }
        respondError(w, http.StatusInternalServerError, "Failed to update entry")
        return
    }

    respondJSON(w, http.StatusOK, updated)
}

// Delete - DELETE /api/entries/{id}
func (h *JournalHandler) Delete(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    entryID := chi.URLParam(r, "id")

    err := h.service.Delete(r.Context(), entryID, userID)
    if err != nil {
        if errors.Is(err, service.ErrNotFound) {
            respondError(w, http.StatusNotFound, "Entry not found")
            return
        }
        respondError(w, http.StatusInternalServerError, "Failed to delete entry")
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// Request/Response types
type CreateEntryRequest struct {
    Title   string   `json:"title"`
    Content string   `json:"content"`
    Mood    string   `json:"mood"`
    Tags    []string `json:"tags"`
}

type UpdateEntryRequest struct {
    Title   string   `json:"title"`
    Content string   `json:"content"`
    Mood    string   `json:"mood"`
    Tags    []string `json:"tags"`
}
```

### Helper Functions

```go
// services/go-api/internal/handler/rest/response.go

package rest

import (
    "encoding/json"
    "net/http"
)

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)

    if data != nil {
        json.NewEncoder(w).Encode(data)
    }
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
    respondJSON(w, status, map[string]string{
        "error": message,
    })
}

// ErrorResponse represents an API error
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code,omitempty"`
    Details string `json:"details,omitempty"`
}
```

## Auth Middleware

```go
// services/go-api/internal/middleware/auth.go

package middleware

import (
    "context"
    "net/http"
    "strings"
)

type AuthMiddleware struct {
    authService *service.AuthService
}

func NewAuthMiddleware(authSvc *service.AuthService) *AuthMiddleware {
    return &AuthMiddleware{authService: authSvc}
}

// Authenticate validates JWT and adds user to context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, `{"error":"Authorization header required"}`, http.StatusUnauthorized)
            return
        }

        // Extract token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, `{"error":"Invalid authorization format"}`, http.StatusUnauthorized)
            return
        }

        token := parts[1]

        // Validate token
        claims, err := m.authService.ValidateToken(token)
        if err != nil {
            http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
            return
        }

        // Add user ID to context
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "user_email", claims.Email)

        // Continue with updated context
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Angular HTTP Client

### API Service

```typescript
// libs/data-access/api/src/lib/journal-api.service.ts

import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import {
  JournalEntry,
  CreateJournalRequest,
  UpdateJournalRequest,
  JournalFilter,
  PaginatedResponse,
} from '@devjournal/shared-models';
import { API_CONFIG } from './api.config';

@Injectable({ providedIn: 'root' })
export class JournalApiService {
  private readonly http = inject(HttpClient);
  private readonly config = inject(API_CONFIG);

  private get baseUrl(): string {
    return `${this.config.baseUrl}/api/entries`;
  }

  // List entries with pagination and filters
  list(filter: JournalFilter): Observable<PaginatedResponse<JournalEntry>> {
    let params = new HttpParams()
      .set('page', filter.page.toString())
      .set('pageSize', filter.pageSize.toString());

    if (filter.mood) {
      params = params.set('mood', filter.mood);
    }

    if (filter.search) {
      params = params.set('search', filter.search);
    }

    return this.http.get<PaginatedResponse<JournalEntry>>(this.baseUrl, { params });
  }

  // Get single entry
  getById(id: string): Observable<JournalEntry> {
    return this.http.get<JournalEntry>(`${this.baseUrl}/${id}`);
  }

  // Create new entry
  create(data: CreateJournalRequest): Observable<JournalEntry> {
    return this.http.post<JournalEntry>(this.baseUrl, data);
  }

  // Update entry
  update(id: string, data: UpdateJournalRequest): Observable<JournalEntry> {
    return this.http.put<JournalEntry>(`${this.baseUrl}/${id}`, data);
  }

  // Delete entry
  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }

  // Search entries
  search(query: string): Observable<JournalEntry[]> {
    const params = new HttpParams().set('search', query);
    return this.http
      .get<PaginatedResponse<JournalEntry>>(this.baseUrl, { params })
      .pipe(map((response) => response.data));
  }
}
```

### Auth Interceptor

```typescript
// libs/features/auth/src/lib/auth.interceptor.ts

import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { AuthStore } from './auth.store';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authStore = inject(AuthStore);
  const token = authStore.token();

  // Skip auth for login/register
  if (req.url.includes('/auth/login') || req.url.includes('/auth/register')) {
    return next(req);
  }

  // Add token if available
  if (token) {
    const authReq = req.clone({
      headers: req.headers.set('Authorization', `Bearer ${token}`),
    });
    return next(authReq);
  }

  return next(req);
};
```

### App Config

```typescript
// apps/web/src/app/app.config.ts

import { ApplicationConfig, provideZoneChangeDetection } from '@angular/core';
import { provideRouter, withComponentInputBinding, withViewTransitions } from '@angular/router';
import { provideHttpClient, withInterceptors, withFetch } from '@angular/common/http';
import { provideClientHydration, withEventReplay } from '@angular/platform-browser';
import { authInterceptor } from '@devjournal/feature-auth';
import { API_CONFIG, ApiConfig } from '@devjournal/data-access-api';
import { routes } from './app.routes';

export const appConfig: ApplicationConfig = {
  providers: [
    provideZoneChangeDetection({ eventCoalescing: true }),
    provideRouter(routes, withComponentInputBinding(), withViewTransitions()),
    provideClientHydration(withEventReplay()),
    provideHttpClient(
      withFetch(),
      withInterceptors([authInterceptor])
    ),
    {
      provide: API_CONFIG,
      useValue: {
        baseUrl: '',  // Same origin
        grpcUrl: 'http://localhost:8081',
      } satisfies ApiConfig,
    },
  ],
};
```

## Request/Response Examples

### Create Entry

**Request:**
```http
POST /api/entries HTTP/1.1
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

{
  "title": "Learning Angular Signals",
  "content": "Today I explored Angular signals...",
  "mood": "learning",
  "tags": ["angular", "signals", "frontend"]
}
```

**Response:**
```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "userId": "user-123",
  "title": "Learning Angular Signals",
  "content": "Today I explored Angular signals...",
  "mood": "learning",
  "tags": ["angular", "signals", "frontend"],
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z"
}
```

### List Entries (Paginated)

**Request:**
```http
GET /api/entries?page=1&pageSize=10&mood=learning HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Learning Angular Signals",
      "mood": "learning",
      "tags": ["angular"],
      "createdAt": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 25,
  "page": 1,
  "pageSize": 10,
  "totalPages": 3
}
```

### Error Response

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "error": "Invalid or expired token"
}
```

## REST vs gRPC Comparison in DevJournal

| Aspect | REST | gRPC |
|--------|------|------|
| URL | `/api/entries` | `/devjournal.v1.JournalService/ListEntries` |
| Content-Type | `application/json` | `application/grpc-web+proto` |
| Request | JSON object | Protobuf binary |
| Response | JSON object | Protobuf binary |
| Type Safety | Manual validation | Generated types |
| Browser Support | Native | Via Connect RPC |
| Debugging | Easy (JSON readable) | Needs tools |
| Performance | Good | Better (smaller payloads) |

## Best Practices

1. **Use proper HTTP methods** - GET for read, POST for create, PUT for update, DELETE for remove
2. **Return appropriate status codes** - 200 OK, 201 Created, 204 No Content, 400 Bad Request, etc.
3. **Validate input** - Check required fields before processing
4. **Handle errors gracefully** - Return meaningful error messages
5. **Use pagination** - Don't return unbounded lists
6. **Add filtering/sorting** - Query params for flexibility
7. **Secure endpoints** - Auth middleware for protected routes
8. **CORS configuration** - Allow frontend origins

## Next Steps

- [PostgreSQL Database](./03-postgresql.md) - Data persistence
- [Authentication & JWT](./07-authentication.md) - Security details
