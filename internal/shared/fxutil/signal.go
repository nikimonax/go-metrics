package fxutil

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func RegisterQuitSignalHandler(
	lc fx.Lifecycle,
	shutdowner fx.Shutdowner,
	sugar *zap.SugaredLogger,
) {
	sigChan := make(chan os.Signal, 1)

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			signal.Notify(sigChan, syscall.SIGQUIT)

			go func() {
				for sig := range sigChan {
					switch sig {
					case syscall.SIGQUIT:
						if err := shutdowner.Shutdown(fx.ExitCode(0)); err != nil {
							sugar.Infow("failed shutdown", "err", err)
						}
						return
					}
				}
			}()

			return nil
		},
		OnStop: func(_ context.Context) error {
			signal.Stop(sigChan)
			close(sigChan)
			return nil
		},
	})
}
