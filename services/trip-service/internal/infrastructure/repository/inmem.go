// Package repository
package repository

import (
	"context"

	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/types"
)

type inmemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewInmemRepository() *inmemRepository {
	return &inmemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inmemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.trips[trip.ID.Hex()] = trip
	return trip, nil
}

func (r *inmemRepository) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*types.OsrmAPIResponse, error) {
	return nil, nil
}
