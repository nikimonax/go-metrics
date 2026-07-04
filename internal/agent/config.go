package agent

import "time"

type AgentConfig struct {
	BaseUrl        string
	PollInterval   time.Duration
	ReportInterval time.Duration
}
