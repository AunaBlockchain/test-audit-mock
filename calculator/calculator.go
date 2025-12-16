// Package calculator provides basic arithmetic operations.
package calculator

import "errors"

// Calculator performs arithmetic operations.
type Calculator struct {
        precision int
}

// New creates a new Calculator with specified precision.
func New(precision int) *Calculator {
        if precision < 0 {
                precision = 2
        }
        return &Calculator{precision: precision}
}

// Add returns the sum of two numbers.
func (c *Calculator) Add(a, b float64) float64 {
        return a + b
}

// Subtract returns the difference of two numbers.
func (c *Calculator) Subtract(a, b float64) float64 {
        return a - b
}

// Multiply returns the product of two numbers.
func (c *Calculator) Multiply(a, b float64) float64 {
        return a * b
}

// Divide returns the quotient of two numbers.
func (c *Calculator) Divide(a, b float64) (float64, error) {
        if b == 0 {
                return 0, errors.New("division by zero")
        }
        return a / b, nil
}

// Power returns a raised to the power of n.
func (c *Calculator) Power(a float64, n int) float64 {
        result := 1.0
        for i := 0; i < n; i++ {
                result *= a
        }
        return result
}
