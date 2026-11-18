package fxmodule

import (
	"context"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func GrpcModule() fx.Option {
	return fx.Module(
		"grpc",
		fx.Provide(provideGrpcServer),
	)
}

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

func provideGrpcServer() server {
	return &grpcServerAdapter{
		wrapped: grpc.NewServer(),
	}
}
