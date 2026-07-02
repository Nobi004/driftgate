package assertion

import (
	"fmt"
	"strings"
)

type ContainsAssertion struct {
	Value         string
	CaseSensitive bool
	Negate        bool
}

func (a ContainsAssertion) Assert(actual string) AssertionResult {
	haystack := actual
	needle := a.Value

	if !a.CaseSensitive {
		haystack = strings.ToLower(actual)
		needle = strings.ToLower(a.Value)
	}

	found := strings.Contains(haystack, needle)
	passed := found != a.Negate // XDR with negate

	if passed {
		return AssertionResult{
			Passed:  true,
			Message: fmt.Sprintf("contains %q", a.Value),
		}
	}
	return AssertionResult{
		Passed:  false,
		Message: fmt.Sprintf("does not contain %q", a.Value),
	}
}

func (a ContainsAssertion) Name() string {
	if a.Negate {
		return "not_contains"
	}
	return "contains"
}

func (a ContainsAssertion) Describe() string {
	if a.Negate {
		return fmt.Sprintf("output should NOT contain %q", a.Value)
	}
	return fmt.Sprintf("output should contain %q", a.Value)
}
