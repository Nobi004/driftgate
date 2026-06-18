package assertion

// AssertionResult tells us if a test passed or failed
type AssertionResult struct {
	Passed  bool
	Message string
	Diff    string // optional: expected vs actual
}

// Assertion is the contract every assertion must satisfy
type Assertion interface {
	Assert(actual string) AssertionResult
	Name() string
	Describe() string
}
