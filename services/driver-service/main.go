package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"ride-sharing/shared/env"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = env.GetString("TRIP_SERVICE_GRPC_ADDR", ":9094")

func main() {
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

	service := NewService()

	grpcServer := grpcserver.NewServer()
	NewGrpcHandler(grpcServer, service)
	log.Printf("🚀 Starting gRPC server Driver-service on port %v", l.Addr().String())

	go func() {
		if err := grpcServer.Serve(l); err != nil {
			log.Printf("💥 Failed to start Driver-service! %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down the Driver-service server...")
	grpcServer.GracefulStop()
}
