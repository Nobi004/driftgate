package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSuite_ValidYAML(t *testing.T) {
	// Set dummy API key for testing
	os.Setenv("GROQ_API_KEY", "test-key")
	defer os.Unsetenv("GROQ_API_KEY")

	content := `
provider: groq
model: llama3-8b-8192
timeout: 30s
concurrency: 5
tests:
  - name: "test one"
    prompt: "Say hello"
    assertions:
      - type: contains
        value: "hello"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "suite.yaml")
	os.WriteFile(path, []byte(content), 0644)

	suite, err := LoadSuite(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if suite.Provider != "groq" {
		t.Errorf("expected provider 'groq', got %s", suite.Provider)
	}
	if suite.Model != "llama3-8b-8192" {
		t.Errorf("expected model 'llama3-8b-8192', got %s", suite.Model)
	}
	if len(suite.Tests) != 1 {
		t.Errorf("expected 1 test, got %d", len(suite.Tests))
	}
}

func TestLoadSuite_WithTags(t *testing.T) {
	os.Setenv("GROQ_API_KEY", "test-key")
	defer os.Unsetenv("GROQ_API_KEY")

	content := `
provider: groq
model: llama3-8b-8192
tests:
  - name: "tagged test"
    tags: [smoke, regression]
    prompt: "test"
    assertions:
      - type: contains
        value: "test"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "suite.yaml")
	os.WriteFile(path, []byte(content), 0644)

	suite, err := LoadSuite(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(suite.Tests[0].Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(suite.Tests[0].Tags))
	}
}

func TestLoadSuite_WithSkip(t *testing.T) {
	os.Setenv("GROQ_API_KEY", "test-key")
	defer os.Unsetenv("GROQ_API_KEY")

	content := `
provider: groq
model: llama3-8b-8192
tests:
  - name: "skipped test"
    skip: true
    prompt: "test"
    assertions:
      - type: contains
        value: "test"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "suite.yaml")
	os.WriteFile(path, []byte(content), 0644)

	suite, err := LoadSuite(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !suite.Tests[0].Skip {
		t.Error("expected test to be skipped")
	}
}

func TestLoadSuite_WithVariables(t *testing.T) {
	os.Setenv("GROQ_API_KEY", "test-key")
	defer os.Unsetenv("GROQ_API_KEY")

	content := `
provider: groq
model: llama3-8b-8192
tests:
  - name: "template test"
    prompt: "Hello {{.Name}}"
    variables:
      Name: "World"
    assertions:
      - type: contains
        value: "World"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "suite.yaml")
	os.WriteFile(path, []byte(content), 0644)

	suite, err := LoadSuite(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rendered, err := suite.Tests[0].RenderPrompt()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rendered != "Hello World" {
		t.Errorf("expected 'Hello World', got %s", rendered)
	}
}

func TestLoadSuite_MissingProvider(t *testing.T) {
	content := `
model: llama3-8b-8192
tests:
  - name: "test"
    prompt: "test"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "suite.yaml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := LoadSuite(path)
	if err == nil {
		t.Error("expected error for missing provider")
	}
}

func TestLoadSuite_NoTests(t *testing.T) {
	content := `
provider: groq
model: llama3-8b-8192
tests: []
`
	dir := t.TempDir()
	path := filepath.Join(dir, "suite.yaml")
	os.WriteFile(path, []byte(content), 0644)

	_, err := LoadSuite(path)
	if err == nil {
		t.Error("expected error for empty tests")
	}
}

func TestLoadSuite_FileNotFound(t *testing.T) {
	_, err := LoadSuite("/nonexistent/suite.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestGetTimeout_Default(t *testing.T) {
	suite := &Suite{}
	if suite.GetTimeout().Seconds() != 30 {
		t.Errorf("expected 30s default timeout")
	}
}

func TestGetTimeout_DurationString(t *testing.T) {
	suite := &Suite{Timeout: "60s"}
	if suite.GetTimeout().Seconds() != 60 {
		t.Errorf("expected 60s timeout, got %v", suite.GetTimeout())
	}
}

func TestGetTimeout_PlainNumber(t *testing.T) {
	suite := &Suite{Timeout: "45"}
	if suite.GetTimeout().Seconds() != 45 {
		t.Errorf("expected 45s timeout, got %v", suite.GetTimeout())
	}
}

func TestRenderPrompt_NoVariables(t *testing.T) {
	tc := &TestCase{Prompt: "Hello World"}
	rendered, err := tc.RenderPrompt()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rendered != "Hello World" {
		t.Errorf("expected 'Hello World', got %s", rendered)
	}
}

func TestRenderPrompt_WithVariables(t *testing.T) {
	tc := &TestCase{
		Prompt:    "Hello {{.Name}}, you are {{.Age}}",
		Variables: map[string]string{"Name": "Alice", "Age": "30"},
	}
	rendered, err := tc.RenderPrompt()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rendered != "Hello Alice, you are 30" {
		t.Errorf("expected 'Hello Alice, you are 30', got %s", rendered)
	}
}

func TestRenderPrompt_EmptyTemplate(t *testing.T) {
	tc := &TestCase{Prompt: ""}
	rendered, err := tc.RenderPrompt()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rendered != "" {
		t.Errorf("expected empty string, got %s", rendered)
	}
}