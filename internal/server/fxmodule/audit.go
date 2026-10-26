package fxmodule

import (
	"errors"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nikimonax/go-metrics/internal/app/event"
	"github.com/nikimonax/go-metrics/internal/impl/audit"
	"github.com/nikimonax/go-metrics/internal/impl/publisher"
	"github.com/nikimonax/go-metrics/internal/lib/work"
	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/middleware"
	"github.com/nikimonax/go-metrics/internal/shared/retry"
)

func AuditModule(cfg *config.ServerConfig) fx.Option {
	if cfg.Audit.File == "" && cfg.Audit.URL == "" {
		return fx.Options()
	}

	provide := []any{
		fx.Annotate(
			provideAuditPoolConfig,
			fx.ResultTags(`name:"pool_config_audit"`),
		),
		fx.Annotate(
			work.NewPool,
			fx.ParamTags(`name:"pool_config_audit"`),
			fx.ResultTags(`name:"pool_audit"`),
		),
		fx.Annotate(
			work.PoolAsSubmitter,
			fx.ParamTags(`name:"pool_audit"`),
			fx.ResultTags(`name:"submitter_audit"`),
		),
		fx.Annotate(
			provideAuditEventHandler,
			fx.ParamTags("", `name:"submitter_audit"`),
			fx.ResultTags(`name:"event_handler_audit"`),
		),
		fx.Annotate(
			provideAuditClient,
			fx.ParamTags(`group:"audit_clients"`),
		),
	}

	invoke := []any{
		fx.Annotate(
			registerPoolLifecycle,
			fx.ParamTags("", `name:"pool_audit"`),
		),
		fx.Annotate(
			registerAuditEventHandler,
			fx.ParamTags("", `name:"event_handler_audit"`),
		),
	}

	if cfg.Audit.File != "" {
		provide = append(
			provide,
			provideFileAuditClient,
			fx.Annotate(
				provideFileAuditClientAsInterface,
				fx.ResultTags(`group:"audit_clients"`),
			),
		)
		invoke = append(
			invoke,
			registerFileAuditClientClose,
		)
	}

	if cfg.Audit.URL != "" {
		provide = append(
			provide,
			fx.Annotate(
				provideHTTPAuditClient,
				fx.ResultTags(`group:"audit_clients"`),
			),
		)
	}

	return fx.Module(
		"audit",
		fx.Provide(provide...),
		fx.Invoke(invoke...),
	)
}

func provideAuditPoolConfig(
	sugar *zap.SugaredLogger,
) work.PoolConfig {
	return work.PoolConfig{
		WorkerConfig:    config.NewWorkerConfig(sugar),
		LifecycleConfig: config.NewLifecycleConfig(),
		WorkerCount:     1,
		QueueSize:       20,
	}
}

func provideFileAuditClient(cfg *config.ServerConfig) (*audit.FileClient, error) {
	return audit.NewFileClient(cfg.Audit.File)
}

func provideFileAuditClientAsInterface(client *audit.FileClient) audit.Client {
	return client
}

func provideHTTPAuditClient(cfg *config.ServerConfig) audit.Client {
	var client audit.Client = audit.NewHTTPClient(cfg.Audit.URL)

	if cfg.Audit.Backoff.Retry == 0 {
		return client
	}

	return audit.NewRetryClient(
		client,
		cfg.Audit.Backoff.Build,
		retry.HTTPErrorIsRetryable,
	)
}

func provideAuditClient(clients []audit.Client) (audit.Client, error) {
	switch len(clients) {
	case 0:
		return nil, errors.New("no audit clients provided")
	case 1:
		return clients[0], nil
	default:
		return audit.NewClientGroup(clients), nil
	}
}

func provideAuditEventHandler(
	client audit.Client,
	submitter work.Submitter,
) publisher.EventHandler {
	return publisher.NewAuditEventHandler(
		client,
		submitter,
		middleware.GetClientIP,
	)
}

func registerFileAuditClientClose(
	lc fx.Lifecycle,
	client *audit.FileClient,
) {
	lc.Append(fx.StopHook(client.Close))
}

func registerAuditEventHandler(
	dp *publisher.EventDispatcher,
	handler publisher.EventHandler,
) {
	dp.Register(event.MetricsUpdatedEventName, handler)
}
