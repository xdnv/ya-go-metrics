package grpc_server

import (
	"context"
	"internal/adapters/logger"
	"time"

	"google.golang.org/grpc"
)

// unary logging gRPC interceptor
func InterceptLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()

	// Call next handler
	resp, err = handler(ctx, req)

	status := 0
	if err != nil {
		status = 1
		logger.Errorf("error executing command: %v", err.Error())
	}

	duration := time.Since(start)
	logger.CommandTrace("gRPC", info.FullMethod, status, duration)

	return resp, err
}
