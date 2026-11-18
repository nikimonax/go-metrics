package fxmodule

import (
	"context"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/nikimonax/go-metrics/internal/gen/metric"
	"github.com/nikimonax/go-metrics/internal/server/handler"
)

type grpcServerAdapter struct {
	wrapped *grpc.Server
}

// Serve implements [server].
func (adapter *grpcServerAdapter) Serve(l net.Listener) error {
	return adapter.wrapped.Serve(l)
}

// Shutdown implements [server].
func (adapter *grpcServerAdapter) Shutdown(_ context.Context) error {
	adapter.wrapped.GracefulStop()
	return nil
}

var _ server = (*grpcServerAdapter)(nil)

func GrpcModule() fx.Option {
	return fx.Module(
		"grpc",
		fx.Provide(
			handler.NewUpdateMetricsGrpcHandler,
			provideGrpcServer,
		),
	)
}

func provideGrpcServer(h *handler.UpdateMetricsGrpcHandler) server {
	s := grpc.NewServer()
	metric.RegisterMetricsServer(s, h)
	return &grpcServerAdapter{wrapped: s}
}
