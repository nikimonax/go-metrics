package agent

import (
	"net/url"
	"time"
)

type AgentConfig struct {
	BaseURL        *url.URL
	APIVersion     uint
	PollInterval   time.Duration
	ReportInterval time.Duration
}
