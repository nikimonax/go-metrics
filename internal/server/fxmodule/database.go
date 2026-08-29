package fxmodule

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"

	"github.com/nikimonax/go-metrics/internal/server/config"
	"github.com/nikimonax/go-metrics/internal/server/handler"
)

func DatabaseModule(cfg *config.ServerConfig) fx.Option {
	if cfg.DatabaseDSN == "" {
		return fx.Options()
	}

	return fx.Module(
		"database",
		fx.Provide(
			provideDatabaseConfig,
			provideDatabase,
			fx.Annotate(
				handler.NewPingDatabaseHandler,
				fx.ResultTags(`name:"handler_database_ping"`),
			),
		),
		fx.Invoke(
			registerDatabaseClose,
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

func registerDatabasePingHandler(
	router chi.Router,
	logger Middleware,
	pingHandler http.Handler,
) {
	router.With(logger).Get("/ping", pingHandler.ServeHTTP)
}
