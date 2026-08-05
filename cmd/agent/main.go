package main

import (
	"flag"
	"strings"

	"github.com/nikimonax/go-metrics/internal/agent"
)

func main() {
	config := agent.AgentConfig{}

	flag.StringVar(
		&config.BaseURL,
		"a",
		"http://localhost:8080",
		"metrics server base url",
	)
	flag.Int64Var(
		&config.PollIntervalSecs,
		"p",
		2,
		"collect metrics interval",
	)
	flag.Int64Var(
		&config.ReportIntervalSecs,
		"r",
		10,
		"send metrics interval",
	)
	flag.Parse()

	hasScheme := false
	hasScheme = hasScheme || strings.HasPrefix(config.BaseURL, "http://")
	hasScheme = hasScheme || strings.HasPrefix(config.BaseURL, "https://")

	if !hasScheme {
		config.BaseURL = "http://" + config.BaseURL
	}

	agent.New(config).Run()
}
