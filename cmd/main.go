package main

import (
	"auth_service/internal/config"
	"auth_service/internal/handlers"
	"auth_service/internal/services"
	"auth_service/internal/storage"
	"auth_service/internal/storage/database"
	"auth_service/internal/storage/memory"
	"time"

	"log"
	"net/http"
)

func main() {
	var store storage.StorageInterface

	go handlers.СleanupVisitors()

	switch config.Mode {
	case "in-memory":
		store = memory.NewMemoryStore()
		log.Println("Using in-memory storage")
	case "postgres":
		db, err := database.NewDatabaseConection(config.PsqlUrl)
		if err != nil {
			log.Printf("failed connection to the database: %v\n", err)
			return
		}
		defer db.Close()

		store = database.NewDatabaseStore(db)
		log.Println("Using PostgreSQL store")
	default:
		log.Fatalf("config.Mode is empty in /internal/config/setting.go")
	}

	authService := services.NewAuthService(store)
	hadler := handlers.RegisterAuthHandler(authService)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/auth/{user_id}", hadler.GenerateTokens())
	mux.HandleFunc("POST /api/auth/refresh", hadler.RefreshTokens())

	serv := &http.Server{
		Addr:         config.ServiceSocket,
		Handler:      handlers.LimiterMiddleware(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("Authentication service is running ...")
	if err := serv.ListenAndServe(); err != nil {
		log.Printf("error when starting the server: %v\n", err)
	}
}
