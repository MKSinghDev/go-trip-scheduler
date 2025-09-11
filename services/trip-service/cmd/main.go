package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"ride-sharing/services/trip-service/internal/infrastructure/events"
	"ride-sharing/services/trip-service/internal/infrastructure/grpc"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"

	grpcserver "google.golang.org/grpc"
)

var (
	GrpcAddr    = env.GetString("TRIP_SERVICE_GRPC_ADDR", ":9093")
	rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:    "trip-service",
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

	inmemRepo := repository.NewInmemRepository()
	svc := service.NewService(inmemRepo)

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

	// RabbitMQ Connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()
	log.Println("🐰 Starting RabbitMQ Connection on Trip-service")

	publisher := events.NewTripEventPublisher(rabbitmq)

	// Start driver consumer
	driverConsumer := events.NewDriverConsumer(rabbitmq, svc)
	go driverConsumer.Listen()

	// Start payment consumer
	paymentConsumer := events.NewPaymentConsumer(rabbitmq, svc)
	go paymentConsumer.Listen()

	// Starting the gRPC server
	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	grpc.NewGRPCHandler(grpcServer, svc, publisher)
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
