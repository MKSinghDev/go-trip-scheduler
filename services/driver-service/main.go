package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"

	grpcserver "google.golang.org/grpc"
)

var (
	GrpcAddr    = env.GetString("TRIP_SERVICE_GRPC_ADDR", ":9094")
	rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:    "driver-service",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	}

	sh, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("Failed to initialize the tracer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)

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

	// RabbitMQ Connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()
	log.Println("🐰 Starting RabbitMQ Connection on Driver-service")

	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	NewGrpcHandler(grpcServer, service)

	consumer := NewTripConsumer(rabbitmq, service)
	go func() {
		if err := consumer.Listen(); err != nil {
			log.Fatalf("failed to listen to the message: %v", err)
		}
	}()

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
