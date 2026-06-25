package main

import (
	"context"
	"errors"
	"folosy-backend/internal/auth"
	"folosy-backend/internal/database"
	"folosy-backend/internal/handler"
	"folosy-backend/internal/repository"
	"folosy-backend/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// tokenCleanupInterval is how often the background job prunes expired refresh
// tokens.
const tokenCleanupInterval = 4 * time.Hour

// startTokenCleanup periodically deletes expired refresh tokens until ctx is cancelled.
func startTokenCleanup(ctx context.Context, repo *repository.RefreshTokenRepository) {
	ticker := time.NewTicker(tokenCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := repo.DeleteExpired(ctx)
			if err != nil {
				log.Printf("token cleanup: %v", err)
				continue
			}
			if n > 0 {
				log.Printf("token cleanup: deleted %d expired refresh tokens", n)
			}
		}
	}
}

// durationFromEnv reads a duration (e.g. "15m", "168h") from an env var,
// falling back to a default when the var is unset or unparseable.
func durationFromEnv(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("invalid %s=%q, using default %s: %v", key, raw, fallback, err)
		return fallback
	}
	return d
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")

	dbPool, err := database.Connect(databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	err = database.RunDBMigrations(databaseURL)
	if err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	accessTTL := durationFromEnv("JWT_ACCESS_TTL", 15*time.Minute)
	refreshTTL := durationFromEnv("JWT_REFRESH_TTL", 7*24*time.Hour)
	tokenService := auth.NewTokenService(jwtSecret, accessTTL, refreshTTL)

	userRepository := repository.NewUserRepository(dbPool)
	refreshTokenRepository := repository.NewRefreshTokenRepository(dbPool)
	userService := service.NewUserService(userRepository, refreshTokenRepository, tokenService)
	userHandler := handler.NewUserHandler(userService)
	authMiddleware := auth.NewAuthMiddleware(tokenService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", userHandler.Register)
	mux.HandleFunc("POST /login", userHandler.Login)
	mux.HandleFunc("POST /refresh", userHandler.Refresh)
	
	// Protected: wrapped in RequireAuth
	mux.Handle("GET /me", authMiddleware.RequireAuth(http.HandlerFunc(userHandler.Me)))

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,  // time to send all request headers
		ReadTimeout:       10 * time.Second, // time to send the whole request (headers + body)
		WriteTimeout:      10 * time.Second, // time to receive the whole response
		IdleTimeout:       60 * time.Second, // idle keep-alive connection lifetime
	}

	// signal.NotifyContext gives a ctx that cancels when the OS asks us to stop
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Background job: prune expired refresh tokens every 4h
	go startTokenCleanup(ctx, refreshTokenRepository)

	// Run the server in a goroutine so main is free to wait for the signal.
	go func() {
		log.Println("server listening on :8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Park here until a signal cancels ctx.
	<-ctx.Done()
	log.Println("shutting down...")

	// Shutdown drains in-flight requests; the context bounds that wait to 10s.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
