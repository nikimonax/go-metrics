package main

import (
	"time"

	"github.com/nikimonax/go-metrics/internal/agent"
)

func main() {
	config := agent.AgentConfig{
		BaseUrl:        "http://localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}

	agent.New(config).Run()
}
