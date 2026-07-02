package report

import (
	"fmt"
	"strings"

	"github.com/nobi004/driftgate/internal/runner"
	"github.com/charmbracelet/lipgloss"
)

var (
	passStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	failStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	tagStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
)

func PrintResults(results []runner.TestResult) {
	passed := 0
	failed := 0

	for _, r := range results {
		tags := ""
		if len(r.Tags) > 0 {
			tags = tagStyle.Render(fmt.Sprintf(" [%s]", strings.Join(r.Tags, ", ")))
		}

		if r.Passed {
			fmt.Printf("%s %s%s (%.2fs)\n",
				passStyle.Render("\u2713"), r.Name, tags, r.Duration)
			passed++
		} else {
			fmt.Printf("%s %s%s (%.2fs)\n",
				failStyle.Render("\u2717"), r.Name, tags, r.Duration)
			if r.Error != "" {
				fmt.Printf("  %s\n", failStyle.Render(r.Error))
			}
			failed++
		}
	}

	total := passed + failed
	fmt.Printf("\n%s %d/%d passed\n",
		infoStyle.Render("\u2192"), passed, total)
}
