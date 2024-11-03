package main

import (
	"fmt"
	"internal/adapters/logger"
	"internal/app"
	pb "internal/service"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// gRPC server proto
type grpcServer struct {
	// Embed the unimplemented server
	pb.UnimplementedMetricStorageServer
}

func serve_grpc() *grpc.Server {

	lis, err := net.Listen("tcp", app.Sc.Endpoint)
	if err != nil {
		logger.Fatal(fmt.Sprintf("error listening gRPC: %v", err))
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterMetricStorageServer(grpcSrv, new(grpcServer))

	reflection.Register(grpcSrv)

	logger.Info(fmt.Sprintf("srv: serving gRPC on %s", app.Sc.Endpoint))
	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Fatal(fmt.Sprintf("error running gRPC server: %v", err))
		}
	}()

	return grpcSrv
}
