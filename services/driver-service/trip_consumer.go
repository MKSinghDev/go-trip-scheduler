package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"

	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type tripEventConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  *Service
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ, service *Service) *tripEventConsumer {
	return &tripEventConsumer{
		rabbitmq,
		service,
	}
}

func (c *tripEventConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.FindAvailableDriversQueue, func(ctx context.Context, msg amqp.Delivery) error {
		var tripEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			return err
		}

		var payload messaging.TripEventData
		if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			return err
		}

		log.Printf("driver received message: %+v", payload)

		switch msg.RoutingKey {
		case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
			return c.handleFindAndNotifyDrivers(ctx, payload)
		}

		log.Printf("unknown trip event: %+v", payload)
		return nil
	})
}

func (c *tripEventConsumer) handleFindAndNotifyDrivers(ctx context.Context, payload messaging.TripEventData) error {
	suitableIDs := c.service.FindAvailableDrivers(payload.Trip.SelectedFare.PackageSlug)

	log.Printf("Found suitable drivers: %v", len(suitableIDs))

	if len(suitableIDs) == 0 {
		// Notify the rider that no drivers are available
		if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventNoDriversFound, contracts.AmqpMessage{
			OwnerID: payload.Trip.UserID,
		}); err != nil {
			log.Printf("Failed to publish message to exchange: %v", err)
			return err
		}
		return nil
	}

	randomIndex := rand.Intn(len(suitableIDs))
	suitableDriverID := suitableIDs[randomIndex]

	marshalledEvent, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Notify the driver about a potential trip
	if err := c.rabbitmq.PublishMessage(ctx, contracts.DriverCmdTripRequest, contracts.AmqpMessage{
		OwnerID: suitableDriverID,
		Data:    marshalledEvent,
	}); err != nil {
		log.Printf("Failed to publish message to exchange: %v", err)
		return err
	}

	return nil
}
