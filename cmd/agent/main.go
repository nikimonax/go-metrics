package main

import (
	"github.com/nikimonax/go-metrics/internal/agent"
)

func main() {
	config := ReadOptions().ToAgentConfig()
	agent.New(config).Run()
}
