package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nobi004/driftgate/internal/assertion"
	"github.com/nobi004/driftgate/internal/config"
	"github.com/nobi004/driftgate/internal/provider"
)

// Result holds the outcome of a single test
type Result struct {
	Name      string
	Passed    bool
	Duration  int64 //ms
	Error     string
	Response  string
	TokensIn  int
	TokensOut int
}
type Runner struct {
	provider provider.Provider
	pool     *WorkerPool
}

type RunOptions struct {
	SuiteFile string
	TagFilter string
	Baseline  bool
	Model     string
	Provider  string
}

type TestResult struct {
	Name     string    `json:"name"`
	Passed   bool      `json:"passed"`
	Duration float64   `json:"duration"`
	Error    string    `json:"error,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
	RunAt    time.Time `json:"run_at"`
}

type Baseline struct {
	RunAt   time.Time    `json:"run_at"`
	Results []TestResult `json:"results"`
}

func New(p provider.Provider, concurrency int) *Runner {
	return &Runner{
		provider: p,
		pool:     NewWorkerPool(concurrency),
	}
}

func (r *Runner) Execute(ctx context.Context, opts RunOptions) ([]TestResult, error) {
	suite, err := config.LoadSuite(opts.SuiteFile)
	if err != nil {
		return nil, fmt.Errorf("load suite: %w", err)
	}

	// CLI flags override suite config
	if opts.Model != "" {
		suite.Model = opts.Model
	}
	if opts.Provider != "" {
		suite.Provider = opts.Provider
	}

	// Build tasks with tag filtering
	tasks := make([]func() error, 0, len(suite.Tests))
	results := make([]TestResult, 0, len(suite.Tests))

	for _, tc := range suite.Tests {
		if tc.Skip {
			continue
		}

		// Tag filtering
		if opts.TagFilter != "" && !hasTag(tc.Tags, opts.TagFilter) {
			continue
		}

		idx := len(results)
		results = append(results, TestResult{
			Name: tc.Name,
			Tags: tc.Tags,
		})

		tc := tc
		tasks = append(tasks, func() error {
			start := time.Now()

			prompt, err := tc.RenderPrompt()
			if err != nil {
				results[idx] = TestResult{
					Name:     tc.Name,
					Passed:   false,
					Duration: time.Since(start).Seconds(),
					Error:    err.Error(),
					Tags:     tc.Tags,
					RunAt:    time.Now(),
				}
				return nil
			}

			resp, err := r.provider.Generate(ctx, prompt)
			if err != nil {
				results[idx] = TestResult{
					Name:     tc.Name,
					Passed:   false,
					Duration: time.Since(start).Seconds(),
					Error:    err.Error(),
					Tags:     tc.Tags,
					RunAt:    time.Now(),
				}
				return nil
			}

			// Run assertions
			passed := true
			var failReasons []string
			for _, ac := range tc.Assertions {
				if ac.Type == "contains" {
					a := assertion.ContainsAssertion{
						Value:         ac.Value,
						CaseSensitive: ac.CaseSensitive,
						Negate:        ac.Negate,
					}
					result := a.Assert(resp.Text)
					if !result.Passed {
						passed = false
						failReasons = append(failReasons, result.Message)
					}
				}
			}

			tr := TestResult{
				Name:     tc.Name,
				Passed:   passed,
				Duration: time.Since(start).Seconds(),
				Tags:     tc.Tags,
				RunAt:    time.Now(),
			}
			if !passed {
				tr.Error = strings.Join(failReasons, "; ")
			}
			results[idx] = tr

			return nil
		})
	}

	if err := r.pool.Run(ctx, tasks); err != nil {
		return nil, err
	}

	// Save baseline if requested
	if opts.Baseline {
		if err := saveBaseline(results); err != nil {
			return nil, fmt.Errorf("save baseline: %w", err)
		}
	}

	return results, nil
}

func hasTag(tags []string, target string) bool {
	for _, t := range tags {
		if t == target {
			return true
		}
	}
	return false
}

func saveBaseline(results []TestResult) error {
	baseline := Baseline{
		RunAt:   time.Now(),
		Results: results,
	}

	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(".driftgate/baseline.json", data, 0644)
}

func LoadBaseline() (*Baseline, error) {
	data, err := os.ReadFile(".driftgate/baseline.json")
	if err != nil {
		return nil, err
	}

	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, err
	}

	return &baseline, nil
}
