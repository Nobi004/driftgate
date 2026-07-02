package cmd

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/nobi004/driftgate/internal/provider"
	"github.com/nobi004/driftgate/internal/report"
	"github.com/nobi004/driftgate/internal/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		// Get API key from env
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY not set")
		}

		// Create provider
		p := provider.NewAnthropicClient(apiKey)
		if err := p.ValidateConfig(); err != nil {
			return err
		}

		// Create runner
		concurrency := viper.GetInt("concurrency")
		if concurrency <= 0 {
			concurrency = 5
		}
		r := runner.New(p, concurrency)

		// Build options from flags
		opts := runner.RunOptions{
			SuiteFile: suiteFile,
			TagFilter: viper.GetString("tag"),
			Baseline:  viper.GetBool("baseline"),
			Model:     viper.GetString("model"),
			Provider:  viper.GetString("provider"),
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

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().String("tag", "", "filter tests by tag")
	runCmd.Flags().Bool("baseline", false, "save results as new baseline")
}
