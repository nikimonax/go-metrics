package server

import "time"

const (
	DefaultLifespanCloseTimeout = 5 * time.Second
	DefaultServerStopTimeout    = 5 * time.Second
)

type ServerConfig struct {
	BaseURL              string
	DumpFile             string
	DumpInterval         time.Duration
	DumpRestore          bool
	LifespanCloseTimeout time.Duration
	ServerStopTimeout    time.Duration
}
