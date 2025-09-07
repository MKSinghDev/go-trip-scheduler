package main

import (
	"context"

	pb "ride-sharing/shared/proto/driver"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedDriverServiceServer

	Service *Service
}

func NewGrpcHandler(s *grpc.Server, service *Service) {
	handler := &gRPCHandler{
		Service: service,
	}

	pb.RegisterDriverServiceServer(s, handler)
}

func (h *gRPCHandler) RegisterDriver(ctx context.Context, r *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	driver, err := h.Service.RegisterDriver(r.GetDriverID(), r.GetPackageSlug())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register driver")
	}
	return &pb.RegisterDriverResponse{
		Driver: driver,
	}, nil
}

func (h *gRPCHandler) UnregisterDriver(ctx context.Context, r *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	h.Service.UnregisterDriver(r.GetDriverID())

	return &pb.RegisterDriverResponse{
		Driver: &pb.Driver{
			Id: r.GetDriverID(),
		},
	}, nil
}
