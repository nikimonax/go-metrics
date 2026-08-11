package zapextra

import (
	"errors"
	"log"

	"go.uber.org/zap"
)

type LogEnv = int

const (
	EnvDev LogEnv = iota
	EnvProd
)

var (
	ErrUnknownEnv = errors.New("unknown env")
)

func NewZapLogger(env LogEnv) *zap.Logger {
	var config zap.Config

	switch env {
	case EnvDev:
		config = zap.NewDevelopmentConfig()
	case EnvProd:
		config = zap.NewProductionConfig()
	default:
		panic(ErrUnknownEnv)
	}

	config.DisableCaller = true

	logger, err := config.Build()

	if err != nil {
		log.Fatalf("failed create zap logger: %s", err)
	}

	return logger
}
