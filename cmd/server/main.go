package main

import (
	"github.com/nikimonax/go-metrics/internal/server"
)

func main() {
	config := server.ServerConfig{
		BaseURL: "localhost:8080",
	}

	server.New(config).Run()
}
