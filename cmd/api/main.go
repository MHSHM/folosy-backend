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

	http.HandleFunc("POST /register", userHandler.Register)
	http.HandleFunc("POST /login", userHandler.Login)

	log.Println("server listening on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
