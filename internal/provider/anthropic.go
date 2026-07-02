package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AnthropicClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

func NewAnthropicClient(apiKey string) *AnthropicClient {
	return &AnthropicClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: "https://api.anthropic.com/v1",
	}
}

func (c *AnthropicClient) Name() string {
	return "anthropic"
}

func (c *AnthropicClient) ValidateConfig() error {
	if c.apiKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY not set")
	}
	return nil
}

func (c *AnthropicClient) callLLM(ctx context.Context, req Request) (Response, error) {
	body, err := json.Marshal(map[string]interface{}{
		"model":      req.Model,
		"max_tokens": req.MaxTokens,
		"messages": []map[string]string{
			{"role": "user", "content": req.Prompt},
		},
	})
	if err != nil {
		return Response{}, fmt.Errorf("marshal request: %w", err)
	}

	return c.doWithRetry(ctx, body)
}

func (c *AnthropicClient) doWithRetry(ctx context.Context, body []byte) (Response, error) {
	const maxRetries = 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<attempt) * time.Second // 1s, 2s, 4s
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return Response{}, ctx.Err()
			}
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST",
			c.baseURL+"/messages", bytes.NewReader(body))
		if err != nil {
			return Response{}, fmt.Errorf("create request: %w", err)
		}

		httpReq.Header.Set("x-api-key", c.apiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
		httpReq.Header.Set("content-type", "application/json")

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			if attempt < maxRetries-1 {
				continue // retry on network error
			}
			return Response{}, fmt.Errorf("all retries exhausted: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			continue // rate limit — retry
		}
		if resp.StatusCode >= 500 {
			continue // server error — retry
		}
		if resp.StatusCode != 200 {
			return Response{}, fmt.Errorf("API error %d", resp.StatusCode)
		}

		var result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return Response{}, fmt.Errorf("decode response: %w", err)
		}

		content := ""
		if len(result.Content) > 0 {
			content = result.Content[0].Text
		}

		return Response{
			Content: content,
			Tokens:  result.Usage.InputTokens + result.Usage.OutputTokens,
		}, nil
	}

	return Response{}, fmt.Errorf("all retries exhausted")
}
