package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/nikimonax/go-metrics/internal/agent"
)

func main() {
	opts, err := ReadOptions()

	if err != nil {
		log.Fatalf("failed read options: %s", err)
	}

	cfg, err := opts.ToAgentConfig()

	if err != nil {
		log.Fatalf("failed build config: %s", err)
	}

	agent, err := agent.New(cfg)

	if err != nil {
		log.Fatalf("failed create agent: %s", err)
	}

	runCtx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	if err := agent.Run(runCtx); err != nil {
		log.Fatalf("failed run agent: %s", err)
	}
}
