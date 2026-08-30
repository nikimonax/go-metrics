package repository

import (
	"context"
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/model"
)

const (
	stmtUpdateCounter = `
INSERT INTO metrics (name, type, delta)
VALUES ($1, $2, $3)
ON CONFLICT (name, type) DO UPDATE
SET delta = metrics.delta + EXCLUDED.delta;`

	stmtUpdateGauge = `
INSERT INTO metrics (name, type, value)
VALUES ($1, $2, $3)
ON CONFLICT (name, type) DO UPDATE
SET value = EXCLUDED.value;`

	stmtGet = `
SELECT name, type, delta, value
FROM metrics
WHERE name = $1 AND type = $2;`

	stmtGetAll = `
SELECT name, type, delta, value
FROM metrics
ORDER BY name;`

	stmtClear = `
TRUNCATE TABLE metrics;`
)

func chooseUpdateStmtByType(metricType domain.MetricType) (string, error) {
	switch metricType {
	case domain.Counter:
		return stmtUpdateCounter, nil
	case domain.Gauge:
		return stmtUpdateGauge, nil
	default:
		return "", app.NewErrUnsupportedMetricType(metricType)
	}
}

func updateMetric(
	ctx context.Context,
	executor interface {
		ExecContext(
			ctx context.Context,
			query string,
			args ...any,
		) (sql.Result, error)
	},
	metric domain.Metric,
) error {
	stmt, err := chooseUpdateStmtByType(metric.Type())

	if err != nil {
		return err
	}

	_, err = executor.ExecContext(
		ctx,
		stmt,
		metric.Name(),
		metric.Type(),
		metric.Value().Get(),
	)

	return err
}

func scanMetric(
	scannable interface {
		Scan(dest ...any) error
	},
) (domain.Metric, error) {
	var m model.Metric
	err := scannable.Scan(&m.Name, &m.Type, &m.Delta, &m.Value)

	if err != nil {
		return nil, err
	}

	return m.ToDomain(), nil
}

func readMetric(
	ctx context.Context,
	querier interface {
		QueryRowContext(
			ctx context.Context,
			query string,
			args ...any,
		) *sql.Row
	},
	metricName domain.MetricName,
	metricType domain.MetricType,
) (domain.Metric, error) {
	row := querier.QueryRowContext(ctx, stmtGet, metricName, metricType)

	metric, err := scanMetric(row)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = app.NewErrMetricNotFound(metricType, metricName)
		}

		return nil, err
	}

	return metric, nil
}

func readAllMetrics(
	ctx context.Context,
	querier interface {
		QueryContext(
			ctx context.Context,
			query string,
			args ...any,
		) (*sql.Rows, error)
	},
) (metrics []domain.Metric, err error) {
	rows, err := querier.QueryContext(ctx, stmtGetAll)

	if err != nil {
		return nil, err
	}

	defer func() { err = errors.Join(rows.Close(), err) }()

	for rows.Next() {
		metric, err := scanMetric(rows)

		if err != nil {
			return nil, err
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

func deleteAllMetrics(
	ctx context.Context,
	executor interface {
		ExecContext(
			ctx context.Context,
			query string,
			args ...any,
		) (sql.Result, error)
	},
) error {
	_, err := executor.ExecContext(ctx, stmtClear)
	return err
}

type PostgresMetricRepository struct {
	db *sql.DB
}

// Update implements [interfaces.MetricRepository].
func (repo *PostgresMetricRepository) Update(
	ctx context.Context,
	metric domain.Metric,
) error {
	return updateMetric(ctx, repo.db, metric)
}

// UpdateBatch implements [interfaces.MetricRepository].
func (repo *PostgresMetricRepository) UpdateBatch(
	ctx context.Context,
	metrics []domain.Metric,
) (err error) {
	switch len(metrics) {
	case 0:
		return nil
	case 1:
		return updateMetric(ctx, repo.db, metrics[0])
	}

	tx, err := repo.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			return
		}

		if rbErr := tx.Rollback(); !errors.Is(rbErr, sql.ErrTxDone) {
			err = errors.Join(rbErr, err)
		}
	}()

	for _, metric := range metrics {
		if err := updateMetric(ctx, tx, metric); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Get implements [interfaces.MetricRepository].
func (repo *PostgresMetricRepository) Get(
	ctx context.Context,
	metricType domain.MetricType,
	metricName domain.MetricName,
) (domain.Metric, error) {
	return readMetric(ctx, repo.db, metricName, metricType)
}

// GetAll implements [interfaces.MetricRepository].
func (repo *PostgresMetricRepository) GetAll(
	ctx context.Context,
) ([]domain.Metric, error) {
	return readAllMetrics(ctx, repo.db)
}

// PopAll implements [interfaces.MetricRepository].
func (repo *PostgresMetricRepository) PopAll(
	ctx context.Context,
) (metrics []domain.Metric, err error) {
	tx, err := repo.db.BeginTx(ctx, nil)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err == nil {
			return
		}

		if rbErr := tx.Rollback(); !errors.Is(rbErr, sql.ErrTxDone) {
			err = errors.Join(rbErr, err)
		}
	}()

	metrics, err = readAllMetrics(ctx, tx)

	if err != nil {
		return nil, err
	}

	err = deleteAllMetrics(ctx, tx)

	if err != nil {
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	return metrics, err
}

// Clear implements [interfaces.MetricRepository].
func (repo *PostgresMetricRepository) Clear(ctx context.Context) error {
	return deleteAllMetrics(ctx, repo.db)
}

func NewPostgresMetricRepository(db *sql.DB) interfaces.MetricRepository {
	return &PostgresMetricRepository{db: db}
}

//go:embed migrations/*.sql
var migrationsDirectory embed.FS

func PostgresMigrate(db *sql.DB) (*migrate.Migrate, error) {
	sourceDriver, err := iofs.New(migrationsDirectory, "migrations")

	if err != nil {
		return nil, err
	}

	databaseDriver, err := postgres.WithInstance(db, new(postgres.Config))

	if err != nil {
		return nil, err
	}

	return migrate.NewWithInstance(
		"iofs", sourceDriver,
		"postgres", databaseDriver,
	)
}
