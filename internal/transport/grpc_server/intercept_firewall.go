package grpc_server

import (
	"context"
	"errors"
	"internal/adapters/firewall"
	"internal/adapters/logger"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// unary firewall gRPC interceptor
func InterceptTrustedNetworkRequests(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

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
