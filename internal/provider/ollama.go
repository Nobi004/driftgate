package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ollamaProvider struct {
	model   string
	baseURL string
	client  *http.Client
}

// NewOllama creates an Ollama local inference provider
func NewOllama(cfg Config) (Provider, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &ollamaProvider{
		model:   cfg.Model,
		baseURL: baseURL,
		client:  &http.Client{Timeout: cfg.Timeout},
	}, nil
}

func (o *ollamaProvider) Name() string { return "ollama" }

func (o *ollamaProvider) Generate(ctx context.Context, prompt string) (Response, error) {
	start := time.Now()

// Ollama's Generate endpoint
	payload := map[string]interface{}{
		"model":  o.model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.0, // deterministic for testing
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("ollama request failed: (is ollama running?) %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return Response{}, fmt.Errorf("ollama error %d: %s",resp.StatusCode,string(bodyBytes))

	}
	var result struct {
		Response     string   `json:"response"`
		Done         bool     `json:"done"`
		EvalCount	 int      `json:"evalCount"` // Tokens Generated 
		PromptCount	 int      `json:"promptCount"` // Tokens in prompt
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Response{}, err
	}
	return Response{
		Text:	  result.Response,
		TokensIn: result.PromptCount,
		TokensOut: result.EvalCount,
		Latency:   time.Since(start),
		Model:     o.model,
	}, nil
}

func (o *ollamaProvider) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}
	
	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama not reachable at %s (run 'ollama serve'): %w", o.baseURL, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama health check failed: %d", resp.StatusCode)
	}
	return nil
}

// ListModels returns available models from the local Ollama instance
func (o *ollamaProvider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}
	return models, nil
}
