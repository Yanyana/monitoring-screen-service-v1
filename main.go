package main

import (
	"fmt"
	"go-service/config"
	"go-service/database"
	"go-service/handlers"
	"go-service/sse"
	"log"
	"net/http"

	"github.com/rs/cors"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize PostgreSQL
	pgDB := database.InitPostgres(cfg)
	defer pgDB.Close()

	// Initialize Redis
	redisClient := database.InitRedis(cfg)
	defer redisClient.Close()

	// CORS Configuration: Allow all origins
	corsOptions := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Inisialisasi SSE hub
	hub := sse.NewSSEHub()

	// Jalankan listener Redis di goroutine
	go sse.ListenRedisChannel(redisClient, hub)

	// Endpoint SSE
	http.HandleFunc("/sse/registration", sse.SSEHandler(hub))
	http.HandleFunc("/sse/patient-result", sse.SSEHandler(hub))

	http.HandleFunc("/example", handlers.ExampleHandler(pgDB, redisClient))
	http.HandleFunc("/worklist/monitoring/patients", handlers.GetPatientRegistrations(pgDB))

	// Wrap the handler with CORS middleware
	handler := corsOptions.Handler(http.DefaultServeMux)

	// Start HTTP server
	log.Printf("Server started at :%s", cfg.ServerPort)
	err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.ServerPort), handler)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
