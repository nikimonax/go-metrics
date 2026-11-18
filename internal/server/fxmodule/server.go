package fxmodule

import (
	"context"
	"errors"
	"net"
	"net/http"

	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/server/config"
)

func ServerModule() fx.Option {
	return fx.Module(
		"server",
		fx.Invoke(registerServerLifecycle),
	)
}

type server interface {
	Serve(net.Listener) error
	Shutdown(context.Context) error
}

func registerServerLifecycle(
	lc fx.Lifecycle,
	cfg *config.ServerConfig,
	srv server,
) {
	errs := make(chan error, 1)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lc := net.ListenConfig{}

			ln, err := lc.Listen(ctx, "tcp", cfg.Listen)
			if err != nil {
				return err
			}

			go func() {
				err := srv.Serve(ln)

				if err != nil && err != http.ErrServerClosed {
					errs <- err
				}

				close(errs)
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownErr := srv.Shutdown(ctx)

			select {
			case <-ctx.Done():
				return shutdownErr
			case serveErr := <-errs:
				return errors.Join(shutdownErr, serveErr)
			}
		},
	})
}
