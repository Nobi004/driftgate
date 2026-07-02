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

func (c *AnthropicClient) CallLLM(ctx context.Context, req Request) (Response, error) {
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
			backoff := time.Duration(1<<attempt) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return Response{}, ctx.Err()
			}
		}

		resp, err := c.doRequest(ctx, body)
		if err != nil {
			if attempt < maxRetries-1 {
				continue
			}
			return Response{}, fmt.Errorf("all retries exhausted: %w", err)
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()
			continue
		}
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			continue
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			return Response{}, fmt.Errorf("API error %d", resp.StatusCode)
		}

		result, err := decodeResponse(resp)
		resp.Body.Close()
		if err != nil {
			return Response{}, err
		}
		return result, nil
	}

	return Response{}, fmt.Errorf("all retries exhausted")
}

func (c *AnthropicClient) doRequest(ctx context.Context, body []byte) (*http.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("content-type", "application/json")

	return c.httpClient.Do(httpReq)
}

func decodeResponse(resp *http.Response) (Response, error) {
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
