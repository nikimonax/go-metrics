package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/nikimonax/go-metrics/internal/server"
)

// Server entrypoint
//
//	@title			GO Metrics API
//	@version		1.0.0
//	@description	API сервера метрик.
//	@BasePath		/
func main() {
	opts, err := ReadOptions()

	if err != nil {
		log.Fatalf("failed read options: %s", err)
	}

	cfg := opts.ToServerConfig()
	server, err := server.New(cfg)

	if err != nil {
		log.Fatalf("failed create server: %s", err)
	}

	runCtx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	if err := server.Run(runCtx); err != nil {
		log.Fatalf("failed run server: %s", err)
	}
}
