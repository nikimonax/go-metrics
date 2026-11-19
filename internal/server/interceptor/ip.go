package interceptor

import (
	"context"
	"net/netip"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/nikimonax/go-metrics/internal/shared/config"
)

func GetClientIP(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return ""
	}

	vals := md.Get(config.SourceIPKey)

	if len(vals) == 0 {
		return ""
	}

	return vals[0]
}

func AllowIPs(prefixes []netip.Prefix) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		raw := GetClientIP(ctx)

		if raw == "" {
			return nil, status.Error(codes.PermissionDenied, "missing client ip")
		}

		addr, err := netip.ParseAddr(raw)

		if err != nil {
			return nil, status.Error(codes.PermissionDenied, "invalid client ip")
		}

		addr = addr.Unmap()

		for _, prefix := range prefixes {
			if prefix.Contains(addr) {
				return handler(ctx, req)
			}
		}

		return nil, status.Error(codes.PermissionDenied, "not allowed")
	}
}
