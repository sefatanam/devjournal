# Authentication & JWT

## Overview

DevJournal uses JWT (JSON Web Tokens) for stateless authentication. This approach allows:
- No server-side session storage
- Scalable across multiple servers
- Self-contained tokens with user info
- Expiration and refresh mechanism

## JWT Structure

A JWT consists of three parts separated by dots:

```
header.payload.signature

eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.
eyJzdWIiOiIxMjM0NTY3ODkwIiwidXNlcklkIjoiYWJjMTIzIn0.
SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

### Header
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

### Payload (Claims)
```json
{
  "userId": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "displayName": "John Doe",
  "exp": 1704067200,
  "iat": 1703980800
}
```

### Signature
```
HMACSHA256(
  base64UrlEncode(header) + "." + base64UrlEncode(payload),
  secret
)
```

## Go Authentication Service

### JWT Claims

```go
// services/go-api/internal/domain/user.go

package domain

import (
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// User domain model
type User struct {
    ID           string    `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"` // Never expose in JSON
    DisplayName  string    `json:"displayName"`
    AvatarURL    string    `json:"avatarUrl,omitempty"`
    CreatedAt    time.Time `json:"createdAt"`
    UpdatedAt    time.Time `json:"updatedAt"`
}

// JWTClaims extends standard claims with custom fields
type JWTClaims struct {
    UserID      string `json:"userId"`
    Email       string `json:"email"`
    DisplayName string `json:"displayName"`
    jwt.RegisteredClaims
}
```

### Auth Service

```go
// services/go-api/internal/service/auth_service.go

package service

import (
    "context"
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

var (
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrEmailExists        = errors.New("email already registered")
    ErrInvalidToken       = errors.New("invalid token")
    ErrTokenExpired       = errors.New("token expired")
)

type AuthService struct {
    userRepo  repository.UserRepository
    jwtSecret []byte
    tokenTTL  time.Duration
}

func NewAuthService(
    userRepo repository.UserRepository,
    jwtSecret string,
) *AuthService {
    return &AuthService{
        userRepo:  userRepo,
        jwtSecret: []byte(jwtSecret),
        tokenTTL:  24 * time.Hour, // Token valid for 24 hours
    }
}

// Register creates a new user account
func (s *AuthService) Register(
    ctx context.Context,
    email, password, displayName string,
) (*domain.User, string, error) {
    // Check if email exists
    existing, _ := s.userRepo.GetByEmail(ctx, email)
    if existing != nil {
        return nil, "", ErrEmailExists
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(password),
        bcrypt.DefaultCost,
    )
    if err != nil {
        return nil, "", fmt.Errorf("failed to hash password: %w", err)
    }

    // Create user
    user := &domain.User{
        Email:        email,
        PasswordHash: string(hashedPassword),
        DisplayName:  displayName,
    }

    createdUser, err := s.userRepo.Create(ctx, user)
    if err != nil {
        return nil, "", fmt.Errorf("failed to create user: %w", err)
    }

    // Generate token
    token, err := s.generateToken(createdUser)
    if err != nil {
        return nil, "", fmt.Errorf("failed to generate token: %w", err)
    }

    return createdUser, token, nil
}

// Login authenticates a user and returns a token
func (s *AuthService) Login(
    ctx context.Context,
    email, password string,
) (*domain.User, string, error) {
    // Find user by email
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return nil, "", ErrInvalidCredentials
    }

    // Compare password
    err = bcrypt.CompareHashAndPassword(
        []byte(user.PasswordHash),
        []byte(password),
    )
    if err != nil {
        return nil, "", ErrInvalidCredentials
    }

    // Generate token
    token, err := s.generateToken(user)
    if err != nil {
        return nil, "", fmt.Errorf("failed to generate token: %w", err)
    }

    return user, token, nil
}

// ValidateToken validates a JWT and returns claims
func (s *AuthService) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
    token, err := jwt.ParseWithClaims(
        tokenString,
        &domain.JWTClaims{},
        func(token *jwt.Token) (interface{}, error) {
            // Validate signing method
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return s.jwtSecret, nil
        },
    )

    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, ErrTokenExpired
        }
        return nil, ErrInvalidToken
    }

    claims, ok := token.Claims.(*domain.JWTClaims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }

    return claims, nil
}

// RefreshToken generates a new token if the old one is still valid
func (s *AuthService) RefreshToken(
    ctx context.Context,
    tokenString string,
) (string, error) {
    claims, err := s.ValidateToken(tokenString)
    if err != nil && !errors.Is(err, ErrTokenExpired) {
        return "", err
    }

    // Get fresh user data
    user, err := s.userRepo.GetByID(ctx, claims.UserID)
    if err != nil {
        return "", fmt.Errorf("user not found: %w", err)
    }

    return s.generateToken(user)
}

// generateToken creates a new JWT for a user
func (s *AuthService) generateToken(user *domain.User) (string, error) {
    now := time.Now()

    claims := &domain.JWTClaims{
        UserID:      user.ID,
        Email:       user.Email,
        DisplayName: user.DisplayName,
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   user.ID,
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
            Issuer:    "devjournal",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.jwtSecret)
}

// GetProfile returns the user's profile
func (s *AuthService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
    return s.userRepo.GetByID(ctx, userID)
}
```

### Password Hashing

```go
// Using bcrypt for password hashing

import "golang.org/x/crypto/bcrypt"

// Hash password (during registration)
hashedPassword, err := bcrypt.GenerateFromPassword(
    []byte(plainPassword),
    bcrypt.DefaultCost, // Cost factor (10-14 is reasonable)
)

// Compare password (during login)
err := bcrypt.CompareHashAndPassword(
    []byte(storedHash),
    []byte(attemptedPassword),
)
// err == nil means password matches
```

## Auth Handlers

### REST Handler

```go
// services/go-api/internal/handler/rest/auth_handler.go

package rest

import (
    "encoding/json"
    "net/http"
)

type AuthHandler struct {
    authService *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
    return &AuthHandler{authService: authSvc}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Validate
    if req.Email == "" || req.Password == "" || req.DisplayName == "" {
        respondError(w, http.StatusBadRequest, "Email, password, and display name are required")
        return
    }

    if len(req.Password) < 8 {
        respondError(w, http.StatusBadRequest, "Password must be at least 8 characters")
        return
    }

    user, token, err := h.authService.Register(r.Context(), req.Email, req.Password, req.DisplayName)
    if err != nil {
        if errors.Is(err, service.ErrEmailExists) {
            respondError(w, http.StatusConflict, "Email already registered")
            return
        }
        respondError(w, http.StatusInternalServerError, "Failed to register")
        return
    }

    respondJSON(w, http.StatusCreated, AuthResponse{
        User:  user,
        Token: token,
    })
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    user, token, err := h.authService.Login(r.Context(), req.Email, req.Password)
    if err != nil {
        if errors.Is(err, service.ErrInvalidCredentials) {
            respondError(w, http.StatusUnauthorized, "Invalid email or password")
            return
        }
        respondError(w, http.StatusInternalServerError, "Failed to login")
        return
    }

    respondJSON(w, http.StatusOK, AuthResponse{
        User:  user,
        Token: token,
    })
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)

    user, err := h.authService.GetProfile(r.Context(), userID)
    if err != nil {
        respondError(w, http.StatusNotFound, "User not found")
        return
    }

    respondJSON(w, http.StatusOK, user)
}

// RefreshToken generates a new token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
    var req RefreshRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    token, err := h.authService.RefreshToken(r.Context(), req.Token)
    if err != nil {
        respondError(w, http.StatusUnauthorized, "Invalid or expired token")
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{"token": token})
}

// Request/Response types
type RegisterRequest struct {
    Email       string `json:"email"`
    Password    string `json:"password"`
    DisplayName string `json:"displayName"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type RefreshRequest struct {
    Token string `json:"token"`
}

type AuthResponse struct {
    User  *domain.User `json:"user"`
    Token string       `json:"token"`
}
```

### Auth Middleware

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

// Authenticate validates JWT and adds user info to context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            respondUnauthorized(w, "Authorization header required")
            return
        }

        // Extract Bearer token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            respondUnauthorized(w, "Invalid authorization format")
            return
        }

        token := parts[1]

        // Validate token
        claims, err := m.authService.ValidateToken(token)
        if err != nil {
            if errors.Is(err, service.ErrTokenExpired) {
                respondUnauthorized(w, "Token expired")
                return
            }
            respondUnauthorized(w, "Invalid token")
            return
        }

        // Add user info to context
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "user_email", claims.Email)
        ctx = context.WithValue(ctx, "user_display_name", claims.DisplayName)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func respondUnauthorized(w http.ResponseWriter, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusUnauthorized)
    json.NewEncoder(w).Encode(map[string]string{"error": message})
}
```

## Angular Auth Implementation

### Auth API Service

```typescript
// libs/data-access/api/src/lib/auth-api.service.ts

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { User, AuthResponse, RegisterRequest, LoginRequest } from '@devjournal/shared-models';
import { API_CONFIG } from './api.config';

@Injectable({ providedIn: 'root' })
export class AuthApiService {
  private readonly http = inject(HttpClient);
  private readonly config = inject(API_CONFIG);

  private get baseUrl(): string {
    return `${this.config.baseUrl}/api/auth`;
  }

  register(data: RegisterRequest): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.baseUrl}/register`, data);
  }

  login(email: string, password: string): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.baseUrl}/login`, { email, password });
  }

  getProfile(): Observable<User> {
    return this.http.get<User>(`${this.baseUrl}/me`);
  }

  refreshToken(token: string): Observable<{ token: string }> {
    return this.http.post<{ token: string }>(`${this.baseUrl}/refresh`, { token });
  }
}
```

### Auth Interceptor

```typescript
// libs/features/auth/src/lib/auth.interceptor.ts

import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { AuthStore } from './auth.store';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const authStore = inject(AuthStore);
  const router = inject(Router);

  // Skip auth header for public endpoints
  const publicUrls = ['/api/auth/login', '/api/auth/register'];
  if (publicUrls.some((url) => req.url.includes(url))) {
    return next(req);
  }

  // Add token to request
  const token = authStore.token();
  if (token) {
    req = req.clone({
      headers: req.headers.set('Authorization', `Bearer ${token}`),
    });
  }

  // Handle response errors
  return next(req).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status === 401) {
        // Token expired or invalid
        authStore.logout();
        router.navigate(['/login']);
      }
      return throwError(() => error);
    })
  );
};
```

### Auth Store

```typescript
// libs/features/auth/src/lib/auth.store.ts

import { computed, inject } from '@angular/core';
import { signalStore, withState, withComputed, withMethods, patchState } from '@ngrx/signals';
import { rxMethod } from '@ngrx/signals/rxjs-interop';
import { pipe, switchMap, tap } from 'rxjs';
import { tapResponse } from '@ngrx/operators';
import { Router } from '@angular/router';
import { AuthApiService } from '@devjournal/data-access-api';
import { User } from '@devjournal/shared-models';

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  error: string | null;
}

const TOKEN_KEY = 'devjournal_token';
const USER_KEY = 'devjournal_user';

export const AuthStore = signalStore(
  { providedIn: 'root' },

  withState<AuthState>({
    user: null,
    token: null,
    isLoading: false,
    error: null,
  }),

  withComputed((store) => ({
    isAuthenticated: computed(() => !!store.token()),
    displayName: computed(() => store.user()?.displayName || 'User'),
  })),

  withMethods((store, authApi = inject(AuthApiService), router = inject(Router)) => ({
    // Initialize from localStorage
    init() {
      if (typeof window === 'undefined') return; // SSR guard

      const token = localStorage.getItem(TOKEN_KEY);
      const userJson = localStorage.getItem(USER_KEY);

      if (token && userJson) {
        try {
          const user = JSON.parse(userJson);
          patchState(store, { token, user });
        } catch {
          this.logout();
        }
      }
    },

    // Login
    login: rxMethod<{ email: string; password: string }>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap(({ email, password }) =>
          authApi.login(email, password).pipe(
            tapResponse({
              next: (response) => {
                // Save to localStorage
                localStorage.setItem(TOKEN_KEY, response.token);
                localStorage.setItem(USER_KEY, JSON.stringify(response.user));

                patchState(store, {
                  token: response.token,
                  user: response.user,
                  isLoading: false,
                });

                router.navigate(['/dashboard']);
              },
              error: (err: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: err.message || 'Login failed',
                });
              },
            })
          )
        )
      )
    ),

    // Register
    register: rxMethod<{ email: string; password: string; displayName: string }>(
      pipe(
        tap(() => patchState(store, { isLoading: true, error: null })),
        switchMap((data) =>
          authApi.register(data).pipe(
            tapResponse({
              next: (response) => {
                localStorage.setItem(TOKEN_KEY, response.token);
                localStorage.setItem(USER_KEY, JSON.stringify(response.user));

                patchState(store, {
                  token: response.token,
                  user: response.user,
                  isLoading: false,
                });

                router.navigate(['/dashboard']);
              },
              error: (err: Error) => {
                patchState(store, {
                  isLoading: false,
                  error: err.message || 'Registration failed',
                });
              },
            })
          )
        )
      )
    ),

    // Logout
    logout() {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
      patchState(store, { user: null, token: null, error: null });
      router.navigate(['/login']);
    },

    clearError() {
      patchState(store, { error: null });
    },
  }))
);
```

### Auth Guards

```typescript
// libs/features/auth/src/lib/auth.guard.ts

import { inject } from '@angular/core';
import { Router, CanActivateFn } from '@angular/router';
import { AuthStore } from './auth.store';

// Guard for protected routes (requires authentication)
export const authGuard: CanActivateFn = () => {
  const authStore = inject(AuthStore);
  const router = inject(Router);

  if (authStore.isAuthenticated()) {
    return true;
  }

  router.navigate(['/login']);
  return false;
};

// Guard for guest routes (login/register - redirect if already logged in)
export const guestGuard: CanActivateFn = () => {
  const authStore = inject(AuthStore);
  const router = inject(Router);

  if (!authStore.isAuthenticated()) {
    return true;
  }

  router.navigate(['/dashboard']);
  return false;
};
```

### Using Guards in Routes

```typescript
// apps/web/src/app/app.routes.ts

import { Routes } from '@angular/router';
import { authGuard, guestGuard } from '@devjournal/feature-auth';

export const routes: Routes = [
  // Public routes (guest only)
  {
    path: '',
    loadComponent: () => import('./pages/landing/landing.component'),
    canActivate: [guestGuard],
  },
  {
    path: 'login',
    loadComponent: () => import('@devjournal/feature-auth').then(m => m.LoginComponent),
    canActivate: [guestGuard],
  },
  {
    path: 'register',
    loadComponent: () => import('@devjournal/feature-auth').then(m => m.RegisterComponent),
    canActivate: [guestGuard],
  },

  // Protected routes
  {
    path: 'dashboard',
    loadComponent: () => import('./pages/dashboard/dashboard.component'),
    canActivate: [authGuard],
  },
  {
    path: 'journal',
    loadChildren: () => import('@devjournal/feature-journal').then(m => m.journalRoutes),
    canActivate: [authGuard],
  },
  // ... more protected routes
];
```

## Security Best Practices

### 1. Password Requirements
```typescript
// Minimum requirements
const passwordRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/;
// At least 8 chars, 1 uppercase, 1 lowercase, 1 number, 1 special char
```

### 2. Token Storage
```typescript
// Use localStorage for web apps (acceptable for most cases)
// For higher security, consider:
// - HttpOnly cookies (requires backend changes)
// - In-memory storage (lost on refresh)
```

### 3. Token Expiration
```go
// Short-lived access tokens (15min - 24h)
tokenTTL: 24 * time.Hour

// Longer-lived refresh tokens (7-30 days)
refreshTokenTTL: 7 * 24 * time.Hour
```

### 4. HTTPS Only
Always use HTTPS in production to prevent token interception.

### 5. CORS Configuration
```go
cors.Options{
    AllowedOrigins:   []string{"https://your-domain.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Authorization", "Content-Type"},
    AllowCredentials: true,
}
```

## Key Takeaways

1. **JWT is stateless** - Server doesn't store sessions
2. **bcrypt for passwords** - Never store plain text
3. **Middleware for auth** - Validate on every protected request
4. **Guards for routes** - Protect frontend routes
5. **Interceptors for tokens** - Auto-attach to requests
6. **Handle expiration** - Redirect on 401 responses
7. **Secure storage** - localStorage is acceptable for most apps

## Next Steps

- [Docker Deployment](./08-docker-deployment.md) - Deploy with secrets
- [REST API Implementation](./02-rest-api.md) - Protected endpoints
