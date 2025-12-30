package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"devjournal/internal/config"
	"devjournal/internal/database"
	grpcHandler "devjournal/internal/handler/grpc"
	"devjournal/internal/handler/rest"
	"devjournal/internal/handler/websocket"
	"devjournal/internal/middleware"
	"devjournal/internal/repository/mongodb"
	"devjournal/internal/repository/postgres"
	"devjournal/internal/service"
	"devjournal/proto/devjournal/v1/devjournalv1connect"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connections
	ctx := context.Background()

	pgPool, err := database.NewPostgresPool(ctx, cfg.DbURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgPool.Close()

	mongoClient, err := database.NewMongoClient(ctx, cfg.MongoURL)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	// Initialize repositories
	userRepo := postgres.NewUserRepository(pgPool)
	journalRepo := postgres.NewJournalRepository(pgPool)
	progressRepo := postgres.NewProgressRepository(pgPool)
	studyGroupRepo := postgres.NewStudyGroupRepository(pgPool)
	snippetRepo := mongodb.NewSnippetRepository(mongoClient, cfg.MongoDB)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	journalService := service.NewJournalService(journalRepo)
	snippetService := service.NewSnippetService(snippetRepo)
	progressService := service.NewProgressService(progressRepo)
	studyGroupService := service.NewStudyGroupService(studyGroupRepo)
	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Setup HTTP router
	router := setupHTTPRouter(cfg, authService, journalService, snippetService, studyGroupService, progressService, hub)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server
	go func() {
		log.Printf("Starting HTTP server on port %d", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start Connect RPC server (gRPC-Web compatible)
	go func() {
		// Create Connect RPC handlers
		journalConnectHandler := grpcHandler.NewJournalConnectHandler(journalService)
		snippetConnectHandler := grpcHandler.NewSnippetConnectHandler(snippetService)

		// Create auth interceptor
		authInterceptor := grpcHandler.AuthInterceptor(authService)
		interceptors := connect.WithInterceptors(authInterceptor)

		// Create mux for Connect RPC
		mux := http.NewServeMux()

		// Register Journal service
		journalPath, journalHandler := devjournalv1connect.NewJournalServiceHandler(
			journalConnectHandler,
			interceptors,
		)
		mux.Handle(journalPath, journalHandler)

		// Register Snippet service
		snippetPath, snippetHandler := devjournalv1connect.NewSnippetServiceHandler(
			snippetConnectHandler,
			interceptors,
		)
		mux.Handle(snippetPath, snippetHandler)

		// Apply CORS for gRPC-Web
		handler := middleware.CORS(mux)

		// Create server with h2c for HTTP/2 without TLS (for development)
		connectServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.GRPCPort),
			Handler: h2c.NewHandler(handler, &http2.Server{}),
		}

		log.Printf("Starting Connect RPC server on port %d", cfg.GRPCPort)
		if err := connectServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Connect RPC server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Servers stopped gracefully")
}

func setupHTTPRouter(
	cfg *config.Config,
	authService *service.AuthService,
	journalService *service.JournalService,
	snippetService *service.SnippetService,
	studyGroupService *service.StudyGroupService,
	progressService *service.ProgressService,
	hub *websocket.Hub,
) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth handlers (public routes)
	authHandler := rest.NewAuthHandler(authService)
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)

	// Protected routes with auth middleware
	authMiddleware := middleware.AuthMiddleware(authService)

	// Journal handlers
	journalHandler := rest.NewJournalHandler(journalService, progressService)
	mux.Handle("GET /api/entries", authMiddleware(http.HandlerFunc(journalHandler.List)))
	mux.Handle("GET /api/entries/{id}", authMiddleware(http.HandlerFunc(journalHandler.Get)))
	mux.Handle("POST /api/entries", authMiddleware(http.HandlerFunc(journalHandler.Create)))
	mux.Handle("PUT /api/entries/{id}", authMiddleware(http.HandlerFunc(journalHandler.Update)))
	mux.Handle("DELETE /api/entries/{id}", authMiddleware(http.HandlerFunc(journalHandler.Delete)))

	// Snippet handlers
	snippetHandler := rest.NewSnippetHandler(snippetService, progressService)
	mux.Handle("GET /api/snippets", authMiddleware(http.HandlerFunc(snippetHandler.List)))
	mux.Handle("GET /api/snippets/{id}", authMiddleware(http.HandlerFunc(snippetHandler.Get)))
	mux.Handle("POST /api/snippets", authMiddleware(http.HandlerFunc(snippetHandler.Create)))
	mux.Handle("PUT /api/snippets/{id}", authMiddleware(http.HandlerFunc(snippetHandler.Update)))
	mux.Handle("DELETE /api/snippets/{id}", authMiddleware(http.HandlerFunc(snippetHandler.Delete)))

	// Study group handlers
	studyGroupHandler := rest.NewStudyGroupHandler(studyGroupService)
	mux.Handle("GET /api/groups", authMiddleware(http.HandlerFunc(studyGroupHandler.List)))
	mux.Handle("GET /api/groups/discover", authMiddleware(http.HandlerFunc(studyGroupHandler.ListPublic)))
	mux.Handle("GET /api/groups/{id}", authMiddleware(http.HandlerFunc(studyGroupHandler.Get)))
	mux.Handle("POST /api/groups", authMiddleware(http.HandlerFunc(studyGroupHandler.Create)))
	mux.Handle("POST /api/groups/{id}/join", authMiddleware(http.HandlerFunc(studyGroupHandler.Join)))
	mux.Handle("POST /api/groups/{id}/leave", authMiddleware(http.HandlerFunc(studyGroupHandler.Leave)))
	mux.Handle("GET /api/groups/{id}/members", authMiddleware(http.HandlerFunc(studyGroupHandler.GetMembers)))
	mux.Handle("DELETE /api/groups/{id}", authMiddleware(http.HandlerFunc(studyGroupHandler.Delete)))

	// Progress handlers
	progressHandler := rest.NewProgressHandler(progressService)
	mux.Handle("GET /api/progress/summary", authMiddleware(http.HandlerFunc(progressHandler.GetSummary)))
	mux.Handle("GET /api/progress/today", authMiddleware(http.HandlerFunc(progressHandler.GetToday)))
	mux.Handle("GET /api/progress/weekly", authMiddleware(http.HandlerFunc(progressHandler.GetWeekly)))
	mux.Handle("GET /api/progress/monthly", authMiddleware(http.HandlerFunc(progressHandler.GetMonthly)))
	mux.Handle("GET /api/progress/streak", authMiddleware(http.HandlerFunc(progressHandler.GetStreak)))

	// WebSocket handler for chat
	wsHandler := websocket.NewChatHandler(hub, authService)
	mux.Handle("GET /ws/chat/{room}", authMiddleware(http.HandlerFunc(wsHandler.HandleWebSocket)))

	// Apply global middleware
	handler := middleware.CORS(mux)
	handler = middleware.Logging(handler)
	handler = middleware.Recovery(handler)

	return handler
}
