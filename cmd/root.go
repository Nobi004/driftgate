package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "driftgate",
	Short: "Driftgate is a tool for testing infrastructure drift using LLMs",
	Long: `Driftgate runs your LLM prompts against test suites,
detects regressions, catches loops, and integrates with CI/CD.`,
}

// func Execute() {
//     cobra.CheckErr(rootCmd.Execute())
// }

func init() {
	// Config file setup
	viper.SetEnvPrefix("DRIFTGATE")
	viper.AutomaticEnv()

	rootCmd.PersistentFlags().IntP("concurrency", "c", 5, "max parallel test execution")
	rootCmd.PersistentFlags().String("provider", "anthropic", "LLM provider")
	rootCmd.PersistentFlags().String("model", "claude-haiku-4-5-20251001", "model name")

}
