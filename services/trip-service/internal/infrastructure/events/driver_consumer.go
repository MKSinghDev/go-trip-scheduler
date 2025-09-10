package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	pbd "ride-sharing/shared/proto/driver"

	amqp "github.com/rabbitmq/amqp091-go"
)

type driverEventConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.TripService
}

func NewDriverConsumer(rabbitmq *messaging.RabbitMQ, service domain.TripService) *driverEventConsumer {
	return &driverEventConsumer{
		rabbitmq,
		service,
	}
}

func (c *driverEventConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.DriverCmdTripResponseQueue, func(ctx context.Context, msg amqp.Delivery) error {
		var driverEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &driverEvent); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			return err
		}

		var payload messaging.DriverTripResponseData
		if err := json.Unmarshal(driverEvent.Data, &payload); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			return err
		}

		log.Printf("driver received message: %+v", payload)

		switch msg.RoutingKey {
		case contracts.DriverCmdTripAccept:
			if err := c.handleTripAccepted(ctx, payload.TripID, payload.Driver); err != nil {
				log.Printf("Failed to handle the trip accept: %v", err)
				return err
			}
		case contracts.DriverCmdTripDecline:
			if err := c.handleTripDeclined(ctx, payload.TripID, payload.RiderID); err != nil {
				log.Printf("failed to handle the trip decline; %v", err)
				return err
			}
			return nil
		}

		log.Printf("unknown driver event: %+v", payload)
		return nil
	})
}

func (c *driverEventConsumer) handleTripAccepted(ctx context.Context, tripID string, driver *pbd.Driver) error {
	// 1. Fetch the first
	trip, err := c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	if trip == nil {
		return fmt.Errorf("trip was not found %s", tripID)
	}

	// 2. Update the trip
	if err := c.service.UpdateTrip(ctx, tripID, "accepted", driver); err != nil {
		log.Printf("Failed to update the trip: %v", err)
		return err
	}

	trip, err = c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	// 3. Driver has been assigned -> publish this event to RB
	marshalledTrip, err := json.Marshal(trip)
	if err != nil {
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverAssigned, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    marshalledTrip,
	}); err != nil {
		return err
	}

	// TODO: Notify the payment service to start a payment link
	return nil
}

func (c *driverEventConsumer) handleTripDeclined(ctx context.Context, tripID, riderID string) error {
	// When a driver declines, we should try to find another driver
	trip, err := c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	newPayload := messaging.TripEventData{
		Trip: trip.ToProto(),
	}

	marshalledPayload, err := json.Marshal(newPayload)
	if err != nil {
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverNotInterested, contracts.AmqpMessage{
		OwnerID: riderID,
		Data:    marshalledPayload,
	}); err != nil {
		return err
	}
	return nil
}
