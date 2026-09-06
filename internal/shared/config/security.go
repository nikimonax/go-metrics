package config

import "hash"

const DefaultHashHeaderKey = "HashSHA256"

type SecurityConfig struct {
	Header      string
	HashingKey  string
	HashingFunc func() hash.Hash
}
