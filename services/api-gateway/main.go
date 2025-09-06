package main

import (
	"log"
	"net/http"

	"ride-sharing/shared/env"
)

var httpAddr = env.GetString("HTTP_ADDR", ":8081")

func main() {
	log.Println("🎬 Starting API Gateway")

	mux := http.NewServeMux()
	mux.HandleFunc("POST /trip/preview", handleTripPreview)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	log.Printf("🚀 API Gateway listening at %v", httpAddr)

	if err := server.ListenAndServe(); err != nil {
		log.Printf("💥 Failed to start the API Gateway, error: %v", err)
	}
}
