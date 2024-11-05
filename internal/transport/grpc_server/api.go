package grpc_server

import (
	"bytes"
	"context"
	"fmt"

	"internal/adapters/logger"
	"internal/app"
	"internal/domain"
	"internal/ports/storage"
	pb "internal/service"
)

// grpc index page handler
func (s *grpcServer) GetMetrics(ctx context.Context, req *pb.EmptyRequest) (*pb.Metrics, error) {
	ms := new(pb.Metrics)

	metrics := app.Stor.GetMetrics()
	for _, key := range storage.SortKeys(metrics) {
		metric, _ := app.Stor.GetMetric(key)

		m := new(pb.Metric)
		m.Type = metric.GetType()
		m.Name = key
		m.Value = fmt.Sprintf("%v", metric.GetValue())

		ms.Data = append(ms.Data, m)
	}

	return ms, nil
}

// grpc PingDB page handler
func (s *grpcServer) PingDB(ctx context.Context, req *pb.EmptyRequest) (*pb.DataResponse, error) {
	dr := new(pb.DataResponse)

	hs := app.PingDBServer()
	if hs.Err != nil {
		logger.Error(hs.Message)
		dr.Message = hs.Message
		return dr, hs.Err
	}

	dr.Message = hs.Message
	return dr, nil
}

// grpc RequestMetricV1 processor
func (s *grpcServer) RequestMetricV1(ctx context.Context, req *pb.Metric) (*pb.DataResponse, error) {
	dr := new(pb.DataResponse)
	mr := new(domain.MetricRequest)
	mr.Type = req.Type
	mr.Name = req.Name

	hs := app.RequestMetricV1(mr)
	if hs.Err != nil {
		logger.Error("RequestMetricV1: " + hs.Message)
		dr.Message = hs.Message
		return dr, hs.Err
	}

	dr.Message = hs.Message
	return dr, nil
}

// grpc RequestMetricV2 processor
func (s *grpcServer) RequestMetricV2(ctx context.Context, req *pb.RawData) (*pb.RawData, error) {
	dr := new(pb.RawData)

	//r := strings.NewReader(req.Message)
	r := bytes.NewBuffer(req.Message)
	//logger.Info("RequestMetricV2: msg " + req.Message)

	data, hs := app.RequestMetricV2(r)
	if hs.Err != nil {
		logger.Error("RequestMetricV2: " + hs.Message)
		dr.Message = []byte(hs.Message)
		return dr, hs.Err
	}

	dr.Message = *data
	return dr, nil
}

// grpc UpdateMetricV1 processor
func (s *grpcServer) UpdateMetricV1(ctx context.Context, req *pb.Metric) (*pb.DataResponse, error) {
	dr := new(pb.DataResponse)
	mr := new(domain.MetricRequest)
	mr.Type = req.Type
	mr.Name = req.Name
	mr.Value = req.Value

	hs := app.UpdateMetricV1(mr)
	if hs.Err != nil {
		logger.Error("UpdateMetricV1: " + hs.Message)
		dr.Message = hs.Message
		return dr, hs.Err
	}

	dr.Message = hs.Message
	return dr, nil
}

// grpc UpdateMetricV2 processor
func (s *grpcServer) UpdateMetricV2(ctx context.Context, req *pb.RawData) (*pb.RawData, error) {
	dr := new(pb.RawData)

	//r := strings.NewReader(req.Message)
	r := bytes.NewBuffer(req.Message)
	//logger.Info("UpdateMetricV2: msg " + req.Message)

	data, hs := app.UpdateMetricV2(r)
	if hs.Err != nil {
		logger.Error("UpdateMetricV2: " + hs.Message)
		dr.Message = []byte(hs.Message)
		return dr, hs.Err
	}

	dr.Message = *data
	return dr, nil
}

// grpc UpdateMetrics processor
func (s *grpcServer) UpdateMetrics(ctx context.Context, req *pb.RawData) (*pb.RawData, error) {
	dr := new(pb.RawData)

	//r := strings.NewReader(req.Message)
	r := bytes.NewBuffer(req.Message)
	//logger.Info("UpdateMetricV2: msg " + req.Message)

	data, hs := app.UpdateMetrics(r)
	if hs.Err != nil {
		logger.Error("UpdateMetrics: " + hs.Message)
		dr.Message = []byte(hs.Message)
		return dr, hs.Err
	}

	//dr.Message = string(*data)
	dr.Message = *data
	return dr, nil
}
