package calculator

import "testing"

// TestAdd tests the Add function.
func TestAdd(t *testing.T) {
        c := New(2)
        result := c.Add(2, 3)
        if result != 5 {
                t.Errorf("Add(2, 3) = %v; want 5", result)
        }
}

// TestSubtract tests the Subtract function.
func TestSubtract(t *testing.T) {
        c := New(2)
        result := c.Subtract(5, 3)
        if result != 2 {
                t.Errorf("Subtract(5, 3) = %v; want 2", result)
        }
}

// NOTE: Tests for Multiply, Divide, and Power are intentionally missing
// to demonstrate incomplete test coverage for the audit.
