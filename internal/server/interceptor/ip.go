package interceptor

import (
	"context"

	"google.golang.org/grpc/metadata"

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
