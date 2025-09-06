package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"ride-sharing/services/trip-service/internal/infrastructure/grpc"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = env.GetString("TRIP_SERVICE_GRPC_ADDR", ":9093")

func main() {
	inmemRepo := repository.NewInmemRepository()
	svc := service.NewService(inmemRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	l, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpcserver.NewServer()
	grpc.NewGRPCHandler(grpcServer, svc)
	log.Printf("🚀 Starting gRPC server Trip-service on port %v", l.Addr().String())

	go func() {
		if err := grpcServer.Serve(l); err != nil {
			log.Printf("💥 Failed to start Trip-service! %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down the Trip-service server...")
	grpcServer.GracefulStop()
}
