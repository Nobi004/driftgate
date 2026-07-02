package report

import (
	"fmt"
	"os"

	"github.com/Nobi004/driftgate/internal/runner"
	"github.com/charmbracelet/lipgloss"
	"github.com/nobi004/driftgate/internal/runner"
)

var (
	passStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	failStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
)

func PrintResults(results []runner.TestResult) {
	passed := 0
	failed := 0

	for _, r := range results {
		if r.Passed {
			fmt.Printf("%s %s (%.2fs)\\n",
				passStyle.Render("✓"), r.Name, r.Duration.Seconds())
			passed++
		} else {
			fmt.Printf("%s %s (%.2fs)\\n",
				failStyle.Render("✗"), r.Name, r.Duration.Seconds())
			if r.Error != "" {
				fmt.Printf("  %s\\n", failStyle.Render(r.Error))
			}
			failed++
		}
	}

	total := passed + failed
	fmt.Printf("\\n%s %d/%d passed\\n",
		infoStyle.Render("→"), passed, total)
}
