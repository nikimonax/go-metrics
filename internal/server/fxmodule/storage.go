package fxmodule

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/impl/repository"
	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/handler"
)

func StorageModule(cfg *config.ServerConfig) fx.Option {
	if cfg.DatabaseDSN == "" {
		return fx.Module(
			"storage",
			fx.Provide(repository.NewInMemoryMetricRepository),
		)
	}

	return fx.Module(
		"storage",
		fx.Provide(
			provideDatabaseConfig,
			provideDatabase,
			repository.NewPostgresMetricRepository,
			fx.Annotate(
				handler.NewPingDatabaseHandler,
				fx.ResultTags(`name:"handler_database_ping"`),
			),
		),
		fx.Invoke(
			registerDatabaseClose,
			registerDatabaseMigrate,
			fx.Annotate(
				registerDatabasePingHandler,
				fx.ParamTags(
					`name:"router_base"`,
					`name:"middleware_logger"`,
					`name:"handler_database_ping"`,
				),
			),
		),
	)
}

func provideDatabaseConfig(cfg *config.ServerConfig) (*pgx.ConnConfig, error) {
	return pgx.ParseConfig(cfg.DatabaseDSN)
}

func provideDatabase(connCfg *pgx.ConnConfig) *sql.DB {
	return stdlib.OpenDB(*connCfg)
}

func registerDatabaseClose(
	lc fx.Lifecycle,
	db *sql.DB,
) {
	lc.Append(fx.StopHook(db.Close))
}

func registerDatabaseMigrate(
	lc fx.Lifecycle,
	db *sql.DB,
) error {
	m, err := repository.PostgresMigrate(db)

	if err != nil {
		return err
	}

	lc.Append(fx.StartHook(func() error {
		if err := m.Up(); !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		return nil
	}))

	return nil
}

func registerDatabasePingHandler(
	router chi.Router,
	logger Middleware,
	pingHandler http.Handler,
) {
	router.With(logger).Get("/ping", pingHandler.ServeHTTP)
}
