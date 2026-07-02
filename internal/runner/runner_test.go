package runner

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nobi004/driftgate/internal/provider"
)

type mockProvider struct {
	response string
	err      error
}

func (m *mockProvider) CallLLM(ctx context.Context, req provider.Request) (provider.Response, error) {
	if m.err != nil {
		return provider.Response{}, m.err
	}
	return provider.Response{Content: m.response, Tokens: 10}, nil
}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) ValidateConfig() error {
	return nil
}

func TestHasTag(t *testing.T) {
	tests := []struct {
		tags   []string
		target string
		want   bool
	}{
		{[]string{"smoke", "regression"}, "smoke", true},
		{[]string{"smoke", "regression"}, "regression", true},
		{[]string{"smoke", "regression"}, "integration", false},
		{[]string{}, "smoke", false},
		{nil, "smoke", false},
	}

	for _, tt := range tests {
		result := hasTag(tt.tags, tt.target)
		if result != tt.want {
			t.Errorf("hasTag(%v, %q) = %v, want %v", tt.tags, tt.target, result, tt.want)
		}
	}
}

func TestSaveBaseline(t *testing.T) {
	// Create .driftgate directory
	os.MkdirAll(".driftgate", 0755)
	defer os.RemoveAll(".driftgate")

	results := []TestResult{
		{
			Name:     "test1",
			Passed:   true,
			Duration: 1.5,
			Tags:     []string{"smoke"},
			RunAt:    time.Now(),
		},
	}

	err := saveBaseline(results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	baseline, err := LoadBaseline()
	if err != nil {
		t.Fatalf("failed to load baseline: %v", err)
	}

	if len(baseline.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(baseline.Results))
	}

	if baseline.Results[0].Name != "test1" {
		t.Errorf("expected name 'test1', got %s", baseline.Results[0].Name)
	}
}

func TestLoadBaseline_NotFound(t *testing.T) {
	_, err := LoadBaseline()
	if err == nil {
		t.Error("expected error for missing baseline file")
	}
}

func TestNewWorkerPool(t *testing.T) {
	wp := NewWorkerPool(5)
	if wp.concurrency != 5 {
		t.Errorf("expected concurrency 5, got %d", wp.concurrency)
	}
}

func TestNew(t *testing.T) {
	p := &mockProvider{}
	r := New(p, 3)
	if r.provider == nil {
		t.Error("expected non-nil provider")
	}
	if r.pool == nil {
		t.Error("expected non-nil pool")
	}
}
