package provider

import "context"

// Request represents a single LLM API Call
type Request struct {
	Model     string
	Prompt    string
	MaxTokens int
}

// Response represents the LLM output
type Response struct {
	Content string
	Tokens  int
}

// Provider is the contract every LLM Client must satisfy
type Provider interface {
	CallLLM(ctx context.Context, req Request) (Response, error)
	Name() string
	ValidateConfig() error
}
