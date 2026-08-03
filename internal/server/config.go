package server

import "time"

type ServerConfig struct {
	BaseURL      string
	DumpFile     string
	DumpInterval time.Duration
	DumpRestore  bool
}
