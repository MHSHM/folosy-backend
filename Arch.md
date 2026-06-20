folosy-backend/
├── cmd/
│   └── api/
│       └── main.go                 # The entry point of the project (wires everything together)
├── internal/
│   ├── domain/
│   │   └── data.go                     # shared data structures
│   ├── handler/
│   │   └── user_create_handler.go      # Transpaort layer (HTTP and routing logic)
│   ├── service/
│   │   └── user_create_service.go      # Service Layer (business logic)
│   └── repository/
│       └── postgres_repository.go      # Database LAYER (database-related logic)
├── go.mod
└── go.sum