package main

import (
	"flag"

	"github.com/nikimonax/go-metrics/internal/server"
)

func main() {
	config := server.ServerConfig{}

	flag.StringVar(
		&config.BaseURL,
		"a",
		"localhost:8080",
		"host and port to listen",
	)
	flag.Parse()

	server.New(config).Run()
}
