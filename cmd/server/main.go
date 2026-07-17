package main

import (
	"github.com/nikimonax/go-metrics/internal/server"
)

func main() {
	config := ReadOptions().ToServerConfig()
	server.New(config).Run()
}
