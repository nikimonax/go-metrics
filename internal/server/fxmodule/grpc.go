package fxmodule

import (
	"context"
	"net"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

	"github.com/nikimonax/go-metrics/internal/gen/metric"
	"github.com/nikimonax/go-metrics/internal/impl/publisher"
	"github.com/nikimonax/go-metrics/internal/lib/grpcextra"
	"github.com/nikimonax/go-metrics/internal/server/handler"
	"github.com/nikimonax/go-metrics/internal/server/interceptor"
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
		// utils
		fx.Provide(
			provideClientIPFromGrpcProvider,
		),
		//server
		fx.Provide(
			handler.NewUpdateMetricsGrpcHandler,
			provideGrpcServer,
		),
	)
}

func provideClientIPFromGrpcProvider() publisher.ClientIPProvider {
	return interceptor.GetClientIP
}

func provideGrpcServer(
	sugar *zap.SugaredLogger,
	h *handler.UpdateMetricsGrpcHandler,
) server {
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(
				grpcextra.InterceptorSugarLogger(sugar),
			),
		),
	)
	metric.RegisterMetricsServer(s, h)
	return &grpcServerAdapter{wrapped: s}
}
