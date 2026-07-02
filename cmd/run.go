package cmd

import (
	"fmt"
	"github.com/Nobi004/driftgate/internal/provider"
	"github.com/Nobi004/driftgate/internal/report"
	"github.com/Nobi004/driftgate/internal/runner"
	"github.com/spf13/viper"
	"os"

	"github.com/spf13/cobra"
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
		r := runner.New(p, concurrency)

		// Execute
		ctx := cmd.Context()
		results, err := r.Execute(ctx, suiteFile)
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
