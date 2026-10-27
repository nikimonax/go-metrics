package main

import (
	"log"

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

	server.Run()
}
