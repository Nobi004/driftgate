package cmd

import "github.com/spf13/cobra"

// Execute is the entry point for the driftgate application
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
