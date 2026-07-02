package assertion

import (
	"testing"
)

func TestContainsAssertion_BasicMatch(t *testing.T) {
	a := ContainsAssertion{Value: "hello", CaseSensitive: false}
	result := a.Assert("hello world")
	if !result.Passed {
		t.Errorf("expected pass, got fail: %s", result.Message)
	}
}

func TestContainsAssertion_NoMatch(t *testing.T) {
	a := ContainsAssertion{Value: "goodbye", CaseSensitive: false}
	result := a.Assert("hello world")
	if result.Passed {
		t.Errorf("expected fail, got pass")
	}
}

func TestContainsAssertion_CaseSensitive(t *testing.T) {
	a := ContainsAssertion{Value: "Hello", CaseSensitive: true}
	result := a.Assert("hello world")
	if result.Passed {
		t.Errorf("expected fail for case sensitive match")
	}
}

func TestContainsAssertion_CaseInsensitive(t *testing.T) {
	a := ContainsAssertion{Value: "Hello", CaseSensitive: false}
	result := a.Assert("hello world")
	if !result.Passed {
		t.Errorf("expected pass for case insensitive match")
	}
}

func TestContainsAssertion_Negate(t *testing.T) {
	a := ContainsAssertion{Value: "hello", CaseSensitive: false, Negate: true}
	result := a.Assert("hello world")
	if result.Passed {
		t.Errorf("expected fail for negated match")
	}
}

func TestContainsAssertion_NegateNoMatch(t *testing.T) {
	a := ContainsAssertion{Value: "goodbye", CaseSensitive: false, Negate: true}
	result := a.Assert("hello world")
	if !result.Passed {
		t.Errorf("expected pass for negated non-match")
	}
}

func TestContainsAssertion_EmptyValue(t *testing.T) {
	a := ContainsAssertion{Value: "", CaseSensitive: false}
	result := a.Assert("hello world")
	if !result.Passed {
		t.Errorf("expected pass for empty value")
	}
}

func TestContainsAssertion_EmptyActual(t *testing.T) {
	a := ContainsAssertion{Value: "hello", CaseSensitive: false}
	result := a.Assert("")
	if result.Passed {
		t.Errorf("expected fail for empty actual")
	}
}

func TestContainsAssertion_Name(t *testing.T) {
	a := ContainsAssertion{Value: "test", Negate: false}
	if a.Name() != "contains" {
		t.Errorf("expected 'contains', got %s", a.Name())
	}

	a.Negate = true
	if a.Name() != "not_contains" {
		t.Errorf("expected 'not_contains', got %s", a.Name())
	}
}

func TestContainsAssertion_Describe(t *testing.T) {
	a := ContainsAssertion{Value: "test", Negate: false}
	desc := a.Describe()
	if desc == "" {
		t.Error("expected non-empty description")
	}

	a.Negate = true
	desc = a.Describe()
	if desc == "" {
		t.Error("expected non-empty description")
	}
}
