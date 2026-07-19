package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/nobi004/driftgate/internal/provider"
	"github.com/nobi004/driftgate/internal/report"
	"github.com/nobi004/driftgate/internal/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var runCmd = &cobra.Command{
	Use:   "run [suite-files]",
	Short: "Run prompt regression tests from a suite file",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		suiteFile := ".driftgate/suite.yaml"
		if len(args) == 1 {
			suiteFile = args[0]
		}

		// Load .env file if present
		godotenv.Load()

		// Load suite to get provider config
		suite, err := loadSuiteConfig(suiteFile)
		if err != nil {
			return fmt.Errorf("load suite: %w", err)
		}

		// Determine provider from flags -> suite -> default
		providerName := ""
		if cmd.Flags().Changed("provider") {
			providerName = viper.GetString("provider")
		} else if suite.Provider != "" {
			providerName = suite.Provider
		} else {
			providerName = "anthropic"
		}

		// Get API key based on provider
		apiKey := getAPIKeyForProvider(providerName)
		if apiKey == "" && providerName != "ollama" {
			return fmt.Errorf("%s_API_KEY not set", providerEnvVar(providerName))
		}

		// Determine model
		model := ""
		if cmd.Flags().Changed("model") {
			model = viper.GetString("model")
		} else if suite.Model != "" {
			model = suite.Model
		} else {
			model = defaultModelForProvider(providerName)
		}

		// Create provider via factory
		cfg := provider.Config{
			Provider: providerName,
			Model:    model,
			APIKey:   apiKey,
			Timeout:  30 * time.Second,
		}
		p, err := provider.Factory(cfg)
		if err != nil {
			return fmt.Errorf("create provider: %w", err)
		}

		// Create runner
		concurrency := viper.GetInt("concurrency")
		if concurrency <= 0 {
			concurrency = 5
		}
		r := runner.New(p, concurrency)

		// Build options from flags
		tagFilter, _ := cmd.Flags().GetString("tag")
		baseline, _ := cmd.Flags().GetBool("baseline")
		opts := runner.RunOptions{
			SuiteFile: suiteFile,
			TagFilter: tagFilter,
			Baseline:  baseline,
			Model:     model,
			Provider:  providerName,
		}

		// Execute
		ctx := cmd.Context()
		results, err := r.Execute(ctx, opts)
		if err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}

		// Report
		report.PrintResults(results)

		// Exit with error if any failed
		for _, res := range results {
			if !res.Passed {
				return fmt.Errorf("some tests failed")
			}
		}
		return nil
	},
}

func loadSuiteConfig(path string) (*SuiteConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read suite file: %w", err)
	}
	var cfg SuiteConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	return &cfg, nil
}

type SuiteConfig struct {
	Provider    string `yaml:"provider"`
	Model       string `yaml:"model"`
	Timeout     string `yaml:"timeout"`
	Concurrency int    `yaml:"concurrency"`
	Tests       []struct{} `yaml:"tests"`
}

func getAPIKeyForProvider(provider string) string {
	switch provider {
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	case "groq":
		return os.Getenv("GROQ_API_KEY")
	case "ollama":
		return os.Getenv("OLLAMA_API_KEY") // optional
	default:
		return ""
	}
}

func providerEnvVar(provider string) string {
	switch provider {
	case "anthropic":
		return "ANTHROPIC"
	case "groq":
		return "GROQ"
	default:
		return provider
	}
}

func defaultModelForProvider(provider string) string {
	switch provider {
	case "anthropic":
		return "claude-haiku-4-5-20251001"
	case "groq":
		return "llama-3.1-8b-instant"
	case "ollama":
		return "llama3.2"
	default:
		return "claude-haiku-4-5-20251001"
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().String("tag", "", "filter tests by tag")
	runCmd.Flags().Bool("baseline", false, "save results as new baseline")
}
