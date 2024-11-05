package grpc_server

import (
	"context"
	"errors"
	"strings"

	"internal/adapters/cryptor"
	"internal/adapters/logger"

	pb "internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// unary decryption gRPC interceptor
func InterceptEncryptedRequests(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		msg := "cryptor: failed to get gRPC metadata"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// parse encryption header and leave if there's no
	token := strings.ToLower(cryptor.GetEncryptionToken())
	tokenStr, ok := md[token]
	if !ok || len(tokenStr) == 0 || strings.TrimSpace(tokenStr[0]) != "true" {
		return handler(ctx, req)
	}

	if !cryptor.CanDecrypt() {
		msg := "cryptor: server is not configured to read encrypted messages"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// Type assert the request to access its fields if necessary
	dataReq, ok := req.(*pb.RawData)
	if !ok {
		msg := "cryptor: failed to get gRPC body"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	//reading out body
	if len(dataReq.Message) == 0 {
		logger.Info("cryptor: empty body, no decryption required")
		return handler(ctx, req)
	}

	logger.Info("cryptor: handling encrypted request")

	decrBody, err := cryptor.Decrypt(&dataReq.Message)
	if err != nil {
		logger.Error("cryptor: error decrypting payload: " + err.Error())
		return nil, errors.New("server could not read encrypted message")
	}

	dataReq.Message = decrBody

	logger.Info("cryptor: successfully decrypted")
	return handler(ctx, req)
}
