package grpc_server

import (
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

// main gRPC server routine
func ServeGRPC() *grpc.Server {

	lis, err := net.Listen("tcp", app.Sc.Endpoint)
	if err != nil {
		logger.Fatalf("error listening gRPC: %v", err)
	}

	options := grpc.ChainUnaryInterceptor(
		InterceptLogger,
		InterceptTrustedNetworkRequests,
		InterceptSignedRequests,
		InterceptGZIPRequests,
		InterceptEncryptedRequests,
	)

	grpcSrv := grpc.NewServer(options)
	pb.RegisterMetricStorageServer(grpcSrv, new(grpcServer))

	reflection.Register(grpcSrv)

	logger.Infof("srv: serving gRPC on %s", app.Sc.Endpoint)
	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Fatalf("error running gRPC server: %v", err)
		}
	}()

	return grpcSrv
}
