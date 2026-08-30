package repository_test

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/go-metrics/internal/app"
	"github.com/nikimonax/go-metrics/internal/app/interfaces"
	"github.com/nikimonax/go-metrics/internal/domain"
	"github.com/nikimonax/go-metrics/internal/impl/repository"
)

type DatabaseConfig struct {
	User     string `env:"DATABASE_USER"`
	Password string `env:"DATABASE_PASSWORD"`
	Host     string `env:"DATABASE_HOST"`
	Port     uint16 `env:"DATABASE_PORT"`
	Name     string `env:"DATABASE_NAME"`
}

func (cfg DatabaseConfig) Validate() error {
	var err error

	if cfg.User == "" {
		err = errors.Join(err, errors.New("required 'DATABASE_USER' env"))
	}

	if cfg.Password == "" {
		err = errors.Join(err, errors.New("required 'DATABASE_PASSWORD' env"))
	}

	if cfg.Host == "" {
		err = errors.Join(err, errors.New("required 'DATABASE_HOST' env"))
	}

	if cfg.Port == 0 {
		err = errors.Join(err, errors.New("required 'DATABASE_PORT' env"))
	}

	if cfg.Name == "" {
		err = errors.Join(err, errors.New("required 'DATABASE_NAME' env"))
	}

	return err
}

func (cfg DatabaseConfig) BuildDSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host + fmt.Sprintf(":%d", cfg.Port),
		Path:     cfg.Name,
		RawQuery: "sslmode=disable",
	}

	return u.String()
}

func findProjectRoot() (string, error) {
	wd, err := os.Getwd()

	if err != nil {
		return "", err
	}

	for {
		gomod := filepath.Join(wd, "go.mod")

		if _, err := os.Stat(gomod); err == nil {
			return wd, nil
		}

		// Move up one directory level
		parent := filepath.Dir(wd)

		if parent == wd {
			// Reached the file system root
			break
		}

		wd = parent
	}
	return "", os.ErrNotExist
}

func scenario(t *testing.T, repo interfaces.MetricRepository) {
	metricA := domain.NewCounterMetric("A", 42)
	metricB := domain.NewGaugeMetric("B", 3.14)
	metricsAB := []domain.Metric{metricA, metricB}

	// 1. initial empty repo
	metrics, err := repo.GetAll(t.Context())
	require.NoError(t, err)
	require.Empty(t, metrics)

	//  2. update two metrics
	err = repo.UpdateBatch(t.Context(), metricsAB)
	require.NoError(t, err)

	// 3. check match
	metrics, err = repo.GetAll(t.Context())
	require.NoError(t, err)
	require.ElementsMatch(t, metrics, metricsAB)

	// 4. check not existing metric
	_, err = repo.Get(t.Context(), domain.Counter, "C")
	require.ErrorIs(t, err, app.ErrMetricNotFound)

	// 5. update existing metric
	err = repo.Update(t.Context(), domain.NewCounterMetric("A", 1))
	require.NoError(t, err)

	// 6. check existing metric updated
	metric, err := repo.Get(t.Context(), metricA.Type(), metricA.Name())
	require.NoError(t, err)
	require.Equal(t, "43", metric.Value().String())

	// 7. pop all metrics
	metrics, err = repo.PopAll(t.Context())
	require.NoError(t, err)
	require.Len(t, metrics, 2)

	// 8. check metrics popped
	metricsLeft, err := repo.GetAll(t.Context())
	require.NoError(t, err)
	require.Empty(t, metricsLeft)

	// 9. again write back
	err = repo.UpdateBatch(t.Context(), metrics)
	require.NoError(t, err)

	// 10. check written
	metrics, err = repo.GetAll(t.Context())
	require.NoError(t, err)
	require.Len(t, metrics, 2)

	// 11. clear
	err = repo.Clear(t.Context())
	require.NoError(t, err)

	// 12. check clear
	metrics, err = repo.GetAll(t.Context())
	require.NoError(t, err)
	require.Empty(t, metrics)
}

func TestMetricRepository(t *testing.T) {
	t.Run("inmemory", func(t *testing.T) {
		repo := repository.NewInMemoryMetricRepository()
		scenario(t, repo)
	})

	t.Run("postgres", func(t *testing.T) {
		root, err := findProjectRoot()

		if err != nil {
			t.Skipf("failed find project root: %s", err)
		}

		envFile := filepath.Join(root, ".env")

		if err := godotenv.Load(envFile); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				t.Skipf("failed read .env file: %s", err)
			}

			t.Logf("env file not exists: %s", envFile)
		}

		var dbConfig DatabaseConfig

		if err := env.Parse(&dbConfig); err != nil {
			t.Skipf("failed parse database config: %s", err)
		}

		if err := dbConfig.Validate(); err != nil {
			t.Skipf("failed validate database config: %s", err)
		}

		pgxConfig, err := pgx.ParseConfig(dbConfig.BuildDSN())

		if err != nil {
			t.Skipf("failed parse pgx config: %s", err)
		}

		db := stdlib.OpenDB(*pgxConfig)
		t.Cleanup(func() { assert.NoError(t, db.Close()) })

		if err := db.PingContext(t.Context()); err != nil {
			t.Skipf("database not available: %s", err)
		}

		m, err := repository.PostgresMigrate(db)
		require.NoError(t, err)

		err = m.Up()

		if errors.Is(err, migrate.ErrNoChange) {
			err = nil
		}
		require.NoError(t, err)

		t.Cleanup(func() { assert.NoError(t, m.Down()) })

		repo := repository.NewPostgresMetricRepository(db)
		scenario(t, repo)
	})

}
