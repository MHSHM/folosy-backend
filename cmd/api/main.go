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
	dbPool, err := database.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	userRepository := repository.NewUserRepository(dbPool)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	http.HandleFunc("POST /Register", userHandler.CreateUserHandler)

	log.Println("server listening on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
