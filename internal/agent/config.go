package agent

import "time"

type AgentConfig struct {
	BaseURL        string
	PollInterval   time.Duration
	ReportInterval time.Duration
}
