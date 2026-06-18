package config

import "time"

// Config holds the configuration for the driftgate application
type Config struct {
	APIKey      string
	Provider    string
	Model       string
	Timeout     time.Duration
	Concurrency int
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Provider:    "anthropic",
		Model:       "claude-haiku-4-5-20251001",
		Timeout:     30 * time.Second,
		Concurrency: 5,
	}
}
