package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"errors"
	"internal/adapters/cryptor"
	"internal/adapters/firewall"
	"internal/adapters/logger"
	"internal/adapters/signer"
	"internal/app"
	pb "internal/service"
	"io"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

// gRPC server proto
type grpcServer struct {
	// Embed the unimplemented server
	pb.UnimplementedMetricStorageServer
}

// unary logging gRPC interceptor
func interceptLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()

	// Call next handler
	resp, err = handler(ctx, req)

	status := 0
	if err != nil {
		status = 1
		logger.Errorf("error executiong command: %v", err.Error())
	}

	duration := time.Since(start)
	logger.CommandTrace("gRPC", info.FullMethod, status, duration)

	return resp, err
}

// unary firewall gRPC interceptor
func interceptTrustedNetworkRequests(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

	if !firewall.IsFirewallEnabled() {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		msg := "firewall: failed to get gRPC metadata"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	//logger.Infof("Received metadata: %v", md)

	// parse X-Real-IP header
	token := strings.ToLower(firewall.GetFirewallToken())
	tokenStr, ok := md[token]
	if !ok || len(tokenStr) == 0 {
		msg := "firewall: ip not found in grpc metadata"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	ip := net.ParseIP(tokenStr[0])
	if ip == nil {
		msg := "firewall: failed to parse ip from grpc metadata"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	if !firewall.IsTrustedIP(ip) {
		logger.Errorf("firewall: agent belongs to untrusted network, ip=%s", tokenStr)
		return nil, errors.New("access denied")
	}

	logger.Infof("firewall: agent is within trusted network, ip=%s", tokenStr)

	return handler(ctx, req)
}

// unary signer gRPC interceptor
func interceptSignedRequests(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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

// unary compression gRPC interceptor
func interceptGZIPRequests(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

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

// unary decryption gRPC interceptor
func interceptEncryptedRequests(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

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

// main gRPC server routine
func serve_grpc() *grpc.Server {

	lis, err := net.Listen("tcp", app.Sc.Endpoint)
	if err != nil {
		logger.Fatalf("error listening gRPC: %v", err)
	}

	options := grpc.ChainUnaryInterceptor(
		interceptLogger,
		interceptTrustedNetworkRequests,
		interceptSignedRequests,
		interceptGZIPRequests,
		interceptEncryptedRequests,
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
