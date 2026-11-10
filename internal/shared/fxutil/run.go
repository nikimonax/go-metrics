package fxutil

import (
	"context"

	"go.uber.org/fx"
)

func RunWithContext(ctx context.Context, app *fx.App) error {
	startCtx, cancel := context.WithTimeout(context.Background(), app.StartTimeout())
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		return err
	}

	select {
	case <-app.Done():
	case <-ctx.Done():
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		return err
	}

	return nil
}
