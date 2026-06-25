package main

import (
	"folosy-backend/internal/auth"
	"folosy-backend/internal/database"
	"folosy-backend/internal/handler"
	"folosy-backend/internal/repository"
	"folosy-backend/internal/service"
	"log"
	"net/http"
	"os"
	"time"
)

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

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", userHandler.Register)
	mux.HandleFunc("POST /login", userHandler.Login)
	mux.HandleFunc("POST /refresh", userHandler.Refresh)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,  // time to send all request headers
		ReadTimeout:       10 * time.Second, // time to send the whole request (headers + body)
		WriteTimeout:      10 * time.Second, // time to receive the whole response
		IdleTimeout:       60 * time.Second, // idle keep-alive connection lifetime
	}

	log.Println("server listening on :8080")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
