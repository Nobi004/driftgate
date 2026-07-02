package config

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Suite struct {
	Provider    string     `yaml:"provider"`
	Model       string     `yaml:"model"`
	Timeout     int        `yaml:"timeout"`
	Concurrency int        `yaml:"concurrency"`
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
	Type   string `yaml:"type"`
	Value  string `yaml:"value,omitempty"`
	Schema any    `yaml:"schema,omitempty"`
	Max    int    `yaml:"max,omitempty"`
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

	return &suite, suite.Validate()
}

func (s *Suite) Validate() error {
	if s.Provider == "" {
		return fmt.Errorf("provider is required")
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
