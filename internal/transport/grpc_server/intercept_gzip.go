package grpc_server

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"internal/adapters/logger"
	pb "internal/service"
	"io"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// unary compression gRPC interceptor
func InterceptGZIPRequests(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		msg := "srv-gzip: failed to get gRPC metadata"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// parse encoding header, gzip encoding and leave if there's no
	token := strings.ToLower("Content-Encoding")
	tokenStr, ok := md[token]
	if !ok || len(tokenStr) == 0 || !strings.Contains(tokenStr[0], "gzip") {
		return handler(ctx, req)
	}

	logger.Info("srv-gzip: handling gzipped request")

	// Type assert the request to access its fields if necessary
	dataReq, ok := req.(*pb.RawData)
	if !ok {
		msg := "srv-gzip: failed to get gRPC body"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	gz, err := gzip.NewReader(bytes.NewReader(dataReq.Message))
	if err != nil {
		logger.Errorf("srv-gzip: error reading compressed msg body: %s", err.Error())
		return nil, err
	}

	defer gz.Close()
	body, err := io.ReadAll(gz)
	if err != nil {
		logger.Errorf("srv-gzip: error extracting msg body: %s", err.Error())
		return nil, err
	}

	dataReq.Message = body

	return handler(ctx, req)
}
