package runner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nobi004/driftgate/internal/assertion"
	"github.com/nobi004/driftgate/internal/config"
	"github.com/nobi004/driftgate/internal/provider"
)

type Runner struct {
	provider provider.Provider
	pool     *WorkerPool
}

type TestResult struct {
	Name     string
	Passed   bool
	Duration time.Duration
	Error    string
}

func New(p provider.Provider, concurrency int) *Runner {
	return &Runner{
		provider: p,
		pool:     NewWorkerPool(concurrency),
	}
}

func (r *Runner) Execute(ctx context.Context, suiteFile string) ([]TestResult, error) {
	suite, err := config.LoadSuite(suiteFile)
	if err != nil {
		return nil, fmt.Errorf("load suite: %w", err)
	}

	// Build tasks
	tasks := make([]func() error, 0, len(suite.Tests))
	results := make([]TestResult, len(suite.Tests))

	for i, tc := range suite.Tests {
		if tc.Skip {
			results[i] = TestResult{Name: tc.Name, Passed: true}
			continue
		}

		i, tc := i, tc // capture loop variables
		tasks = append(tasks, func() error {
			start := time.Now()

			prompt, err := tc.RenderPrompt()
			if err != nil {
				results[i] = TestResult{
					Name:     tc.Name,
					Passed:   false,
					Duration: time.Since(start),
					Error:    err.Error(),
				}
				return nil // test failed but don't stop other tests
			}

			resp, err := r.provider.CallLLM(ctx, provider.Request{
				Model:     suite.Model,
				Prompt:    prompt,
				MaxTokens: 1024,
			})
			if err != nil {
				results[i] = TestResult{
					Name:     tc.Name,
					Passed:   false,
					Duration: time.Since(start),
					Error:    err.Error(),
				}
				return nil
			}

			// Run assertions
			passed := true
			var failReasons []string
			for _, ac := range tc.Assertions {
				// TODO: Map assertion config to actual assertion
				// For now, just check contains
				if ac.Type == "contains" {
					a := assertion.ContainsAssertion{
						Value:         ac.Value,
						CaseSensitive: false,
					}
					result := a.Assert(resp.Content)
					if !result.Passed {
						passed = false
						failReasons = append(failReasons, result.Message)
					}
				}
			}

			results[i] = TestResult{
				Name:     tc.Name,
				Passed:   passed,
				Duration: time.Since(start),
			}
			if !passed {
				results[i].Error = strings.Join(failReasons, "; ")
			}

			return nil
		})
	}

	if err := r.pool.Run(ctx, tasks); err != nil {
		return nil, err
	}

	return results, nil
}
