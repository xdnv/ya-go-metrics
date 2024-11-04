package grpc_server

import (
	"context"
	"encoding/base64"
	"errors"
	"internal/adapters/logger"
	"internal/adapters/signer"
	pb "internal/service"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// unary signer gRPC interceptor
func InterceptSignedRequests(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if !signer.IsSignedMessagingEnabled() {
		return handler(ctx, req)
	}

	// Type assert the request to access its fields if necessary
	dataReq, ok := req.(*pb.RawData)
	if !ok {
		msg := "signer: failed to get gRPC body"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	body := dataReq.Message

	if len(body) == 0 {
		logger.Info("signer: empty body, no security check needed")
		return handler(ctx, req)
	}

	logger.Info("signer: handling signed request")

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		msg := "signer: failed to get gRPC metadata"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// parse signature header
	token := strings.ToLower(signer.GetSignatureToken())
	tokenStr, ok := md[token]
	if !ok || len(tokenStr) == 0 {
		msg := "signer: ip not found in grpc metadata"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	sigRaw := tokenStr[0]

	sig, err := base64.URLEncoding.DecodeString(sigRaw)
	if err != nil {
		msg := "signer: incorrect message signature format"
		logger.Error(msg)
		return nil, err
	}

	//calculate body signature
	ok = signer.Compare(&sig, &body)
	if !ok {
		// gRPC branch gives no excuses
		//if signer.IsStrictSignedMessagingEnabled() {
		msg := "signer: message security check failed"
		logger.Error(msg)
		return nil, errors.New(msg)
		//}

		//non-strict mode passes yandex iter14 test: yandex gives no actual signature, just a key on startup
		//logger.Error("signer: non-strict message security check FAILED")
	}

	logger.Infof("signer: signature OK, id=%v", sig)
	return handler(ctx, req)
}
