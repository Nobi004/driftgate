package provider

import (
	"context"
	"fmt"
	"time"
)

// Provider defines the contract for LLM inference backends
type Provider interface {
	// Name returns the provider identifier (e.g., "anthropic", "ollama")
	Name() string
	
	// Generate sends a prompt and returns the generated text
	Generate(ctx context.Context, prompt string) (Response, error)
	
	// Health checks if the provider is reachable and ready
	Health(ctx context.Context) error
}

// Response wraps the LLM output with metadata
type Response struct {
	Text      string
	TokensIn  int
	TokensOut int
	Latency   time.Duration
	Model     string
}

// Config holds provider-agnostic configuration
type Config struct {
	Provider string // "anthropic" | "ollama" | "groq"
	Model    string
	BaseURL  string
	APIKey   string // optional for local providers
	Timeout  time.Duration
}

// Factory creates the appropriate provider based on config
func Factory(cfg Config) (Provider, error) {
	switch cfg.Provider {
	case "anthropic":
		return NewAnthropic(cfg)
	case "ollama":
		return NewOllama(cfg)
	case "groq":
		return NewGroq(cfg)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
}