package config

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

// Suite represents a test suite configuration
type Suite struct {
	Provider    string     `yaml:"provider"`
	Model       string     `yaml:"model"`
	Timeout     string     `yaml:"timeout"`
	Concurrency int        `yaml:"concurrency"`
	APIKey      string     `yaml:"api_key"`
	Tests       []TestCase `yaml:"tests"`
}

type TestCase struct {
	Name       string            `yaml:"name"`
	Tags       []string          `yaml:"tags"`
	Prompt     string            `yaml:"prompt"`
	Variables  map[string]string `yaml:"variables"`
	Assertions []AssertionConfig `yaml:"assertions"`
	Skip       bool              `yaml:"skip"`
}

type AssertionConfig struct {
	Type          string `yaml:"type"`
	Value         string `yaml:"value,omitempty"`
	CaseSensitive bool   `yaml:"case_sensitive,omitempty"`
	Negate        bool   `yaml:"negate,omitempty"`
	Schema        any    `yaml:"schema,omitempty"`
	Max           int    `yaml:"max,omitempty"`
}

func (s *Suite) GetTimeout() time.Duration {
	if s.Timeout == "" {
		return 30 * time.Second
	}

	// Try parsing as duration string (e.g. "30s", "5m")
	if d, err := time.ParseDuration(s.Timeout); err == nil {
		return d
	}

	// Try parsing as plain number (seconds)
	if secs, err := strconv.Atoi(strings.TrimSuffix(s.Timeout, "s")); err == nil {
		return time.Duration(secs) * time.Second
	}

	return 30 * time.Second
}

func (s *Suite) Validate() error {
	validProviders := map[string]bool{
		"anthropic": true,
		"ollama":    true,
		"groq":      true,
	}
	if !validProviders[s.Provider] {
		return fmt.Errorf("provider must be one of: anthropic, ollama, groq (got %q)", s.Provider)
	}
	if len(s.Tests) == 0 {
		return fmt.Errorf("suite must have at least one test")
	}
	return nil
}

func (tc *TestCase) RenderPrompt() (string, error) {
	if len(tc.Variables) == 0 {
		return tc.Prompt, nil
	}

	tmpl, err := template.New("prompt").Parse(tc.Prompt)
	if err != nil {
		return "", fmt.Errorf("parse prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tc.Variables); err != nil {
		return "", fmt.Errorf("render prompt: %w", err)
	}
	return buf.String(), nil
}

func LoadSuite(path string) (*Suite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read suite file %q: %w", path, err)
	}

	var suite Suite
	if err := yaml.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}

	// Set defaults
	if suite.Provider == "" {
		suite.Provider = "anthropic"
	}

	// Resolve API key from env if not set in config
	if suite.APIKey == "" {
		switch suite.Provider {
		case "groq":
			suite.APIKey = os.Getenv("GROQ_API_KEY")
		case "anthropic":
			suite.APIKey = os.Getenv("ANTHROPIC_API_KEY")
		}
	}

	// Validate
	if suite.Model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if suite.Provider == "anthropic" && suite.APIKey == "" {
		return nil, fmt.Errorf("anthropic provider requires api_key or ANTHROPIC_API_KEY env var")
	}
	if suite.Provider == "groq" && suite.APIKey == "" {
		return nil, fmt.Errorf("groq provider requires api_key or GROQ_API_KEY env var")
	}

	return &suite, suite.Validate()
}