package main

import (
	"folosy-backend/internal/database"
	"folosy-backend/internal/handler"
	"folosy-backend/internal/repository"
	"folosy-backend/internal/service"
	"log"
	"net/http"
	"os"
)

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

	userRepository := repository.NewUserRepository(dbPool)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	http.HandleFunc("POST /register", userHandler.Register)

	log.Println("server listening on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
