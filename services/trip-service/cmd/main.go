package main

import (
	"log"
	"net/http"

	h "ride-sharing/services/trip-service/internal/infrastructure/http"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"
)

var httpAddr = env.GetString("TRIP_SERVICE_HTTP_ADDR", ":8083")

func main() {
	// ctx := context.Background()

	inmemRepo := repository.NewInmemRepository()
	svc := service.NewService(inmemRepo)

	handlers := h.HTTPHandler{Service: svc}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /preview", handlers.HandleTripPreview)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	log.Printf("🚀 Trip-service listening at %v", httpAddr)

	if err := server.ListenAndServe(); err != nil {
		log.Printf("💥 Failed to start Trip-service! %v", err)
	}
}
