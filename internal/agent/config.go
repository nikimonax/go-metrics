package agent

import (
	"net/url"
	"time"
)

type AgentConfig struct {
	BaseURL        *url.URL
	PollInterval   time.Duration
	ReportInterval time.Duration
}
