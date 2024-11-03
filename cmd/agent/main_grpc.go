// gRPC part of agent code
package main

import (
	"context"
	"fmt"
	"internal/adapters/logger"
	"internal/app"
	pb "internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func PostGRPC(ctx context.Context, ac app.AgentConfig, counterType string, counterName string, value string) {
	conn, err := grpc.NewClient(ac.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error(fmt.Sprintf("Error connecting to gRPC server: %v", err))
		return
	}
	defer conn.Close()

	client := pb.NewMetricStorageClient(conn)

	data := new(pb.Metric)
	data.Name = counterName
	data.Type = counterType
	data.Value = value

	resp, err := client.UpdateMetricV1(ctx, data)
	if err != nil {
		logger.Error(fmt.Sprintf("error querying gRPC: %v", err))
		return
	}

	logger.Info(fmt.Sprintf("gRPC reply: %v", resp.Message))
}
