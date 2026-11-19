package fxmodule

import (
	"context"
	"errors"
	"net"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

	"github.com/nikimonax/go-metrics/internal/gen/metric"
	"github.com/nikimonax/go-metrics/internal/impl/publisher"
	"github.com/nikimonax/go-metrics/internal/lib/grpcextra"
	"github.com/nikimonax/go-metrics/internal/lib/netutil"
	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/handler"
	"github.com/nikimonax/go-metrics/internal/server/interceptor"
)

type grpcServerAdapter struct {
	wrapped *grpc.Server
}

// Serve implements [server].
func (adapter *grpcServerAdapter) Serve(l net.Listener) error {
	err := adapter.wrapped.Serve(l)

	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}

	return err
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
	cfg *config.ServerConfig,
	sugar *zap.SugaredLogger,
	h *handler.UpdateMetricsGrpcHandler,
) (server, error) {
	interceptors := make([]grpc.UnaryServerInterceptor, 0)

	if cfg.TrustedSubnet != "" {
		prefixes, err := netutil.ParsePrefixes([]string{cfg.TrustedSubnet})

		if err != nil {
			return nil, err
		}

		interceptors = append(
			interceptors,
			interceptor.AllowIPs(prefixes),
		)
	}

	interceptors = append(
		interceptors,
		logging.UnaryServerInterceptor(
			grpcextra.InterceptorSugarLogger(sugar),
		),
	)

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))

	metric.RegisterMetricsServer(s, h)
	return &grpcServerAdapter{wrapped: s}, nil
}
