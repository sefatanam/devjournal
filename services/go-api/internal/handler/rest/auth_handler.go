package rest

import (
	"encoding/json"
	"net/http"

	"devjournal/internal/service"
	"devjournal/pkg/httputil"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

// UserProfile represents the user data in responses
type UserProfile struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		httputil.Error(w, http.StatusBadRequest, "email, password, and displayName are required")
		return
	}

	if len(req.Password) < 6 {
		httputil.Error(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	// Register user
	user, token, err := h.authService.Register(r.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		if err == service.ErrEmailAlreadyExists {
			httputil.Error(w, http.StatusConflict, "email already exists")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	// Return response
	response := AuthResponse{
		Token: token,
		User: UserProfile{
			ID:          user.ID.String(),
			Email:       user.Email,
			DisplayName: user.DisplayName,
		},
	}

	httputil.JSON(w, http.StatusCreated, response)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		httputil.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	// Login user
	user, token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			httputil.Error(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to login")
		return
	}

	// Return response
	response := AuthResponse{
		Token: token,
		User: UserProfile{
			ID:          user.ID.String(),
			Email:       user.Email,
			DisplayName: user.DisplayName,
		},
	}

	httputil.JSON(w, http.StatusOK, response)
}
