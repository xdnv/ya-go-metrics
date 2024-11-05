// gRPC part of agent code
package main

import (
	"context"
	"internal/adapters/logger"
	"internal/app"
	pb "internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func PostGRPC(ctx context.Context, ac app.AgentConfig, counterType string, counterName string, value string) (*Response, error) {
	res := NewResponse()

	conn, err := grpc.NewClient(ac.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("Error connecting to gRPC server: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewMetricStorageClient(conn)

	data := new(pb.Metric)
	data.Name = counterName
	data.Type = counterType
	data.Value = value

	resp, err := client.UpdateMetricV1(ctx, data)
	if err != nil {
		logger.Errorf("error querying gRPC: %v", err)
		return nil, err
	}

	res.StatusCode = 200
	res.Status = string(resp.Message)

	return res, nil
}

// simple GRPC post function based on Message input format
func PostMessageGRPC(m *Message, bulkUpdate bool) (*Response, error) {
	conn, err := grpc.NewClient(ac.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewMetricStorageClient(conn)

	// Create metadata and a new context with the metadata
	md := metadata.New(m.Metadata)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	br := new(pb.RawData)
	br.Message = m.Body.Bytes()

	//resp, err := client.UpdateMetricV1(context.Background(), mr)
	var resp *pb.RawData
	if bulkUpdate {
		resp, err = client.UpdateMetrics(ctx, br)
	} else {
		resp, err = client.UpdateMetricV2(ctx, br)
	}

	if err != nil {
		//log.Fatalf("Ошибка gRPC запроса: %v", err)
		return nil, err
	}

	res := NewResponse()
	res.StatusCode = 200
	res.Status = string(resp.Message)

	return res, nil
}
