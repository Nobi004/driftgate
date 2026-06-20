package cmd

import (
	"fmt"

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
		fmt.Printf("Running suite: %s\n", suiteFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().String("tag", "", "filter tests by tag")
	runCmd.Flags().Bool("baseline", false, "save results as new baseline")
}
