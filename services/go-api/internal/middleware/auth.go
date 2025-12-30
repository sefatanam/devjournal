package middleware

import (
	"context"
	"net/http"
	"strings"

	"devjournal/internal/service"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey      contextKey = "userID"
	UserEmailKey   contextKey = "userEmail"
	UserNameKey    contextKey = "userName"
)

// AuthMiddleware validates JWT tokens and adds user info to context
func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			// Try Authorization header first
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				// Parse "Bearer <token>"
				const prefix = "Bearer "
				if len(authHeader) >= len(prefix) && strings.HasPrefix(authHeader, prefix) {
					tokenString = authHeader[len(prefix):]
				}
			}

			// Fall back to query parameter (for WebSocket connections)
			if tokenString == "" {
				tokenString = r.URL.Query().Get("token")
			}

			if tokenString == "" {
				http.Error(w, `{"error":"missing authorization"}`, http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID) // Already a uuid.UUID
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserNameKey, claims.DisplayName)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the user ID from context as string
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return userID.String()
	}
	return ""
}

// GetUserUUID extracts the user ID from context as UUID
func GetUserUUID(ctx context.Context) uuid.UUID {
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return userID
	}
	return uuid.Nil
}

// GetUserEmail extracts the user email from context
func GetUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(UserEmailKey).(string); ok {
		return email
	}
	return ""
}

// GetUserName extracts the user display name from context
func GetUserName(ctx context.Context) string {
	if name, ok := ctx.Value(UserNameKey).(string); ok {
		return name
	}
	return ""
}
