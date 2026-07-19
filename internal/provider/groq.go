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

type groqProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewGroq creates a Groq API provider (OpenAI-compatible)
func NewGroq(cfg Config) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("groq provider requires api_key")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1"
	}
	return &groqProvider{
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		baseURL: baseURL,
		client:  &http.Client{Timeout: cfg.Timeout},
	}, nil
}

func (g *groqProvider) Name() string { return "groq" }

func (g *groqProvider) Generate(ctx context.Context, prompt string) (Response, error) {
	start := time.Now()

	payload := map[string]interface{}{
		"model":       g.model,
		"max_tokens":  1024,
		"temperature": 0.0, // deterministic for testing
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("groq request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return Response{}, fmt.Errorf("groq API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Response{}, err
	}

	if len(result.Choices) == 0 {
		return Response{}, fmt.Errorf("empty response from groq")
	}

	return Response{
		Text:      result.Choices[0].Message.Content,
		TokensIn:  result.Usage.PromptTokens,
		TokensOut: result.Usage.CompletionTokens,
		Latency:   time.Since(start),
		Model:     g.model,
	}, nil
}

func (g *groqProvider) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", g.baseURL+"/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("groq health check failed: %d", resp.StatusCode)
	}
	return nil
}