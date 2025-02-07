package optimizer

import (
	"github.com/llir/llvm/ir"
)

// OptimizationLevel represents different levels of optimization
type OptimizationLevel int

const (
	O0 OptimizationLevel = iota // No optimizations
	O1                          // Basic optimizations
	O2                          // Moderate optimizations
	O3                          // Aggressive optimizations
)

// Convenience exported constants
const (
	NoOptimization     = O0
	BasicOptimization  = O1
	MediumOptimization = O2
	HighOptimization   = O3
)

// Optimizer interface defines the optimization process
type Optimizer interface {
	Optimize(*ir.Module) (*ir.Module, error)
	SetLevel(OptimizationLevel)
}

// DefaultOptimizer provides a basic implementation with no optimizations
type DefaultOptimizer struct {
	level OptimizationLevel
}

// NewOptimizer creates a new default optimizer
func NewOptimizer() Optimizer {
	return &DefaultOptimizer{level: O0}
}

// Optimize implements the Optimizer interface but currently performs no optimizations
func (o *DefaultOptimizer) Optimize(module *ir.Module) (*ir.Module, error) {
	// No optimizations performed, just return the module as-is
	return module, nil
}

// SetLevel sets the optimization level
func (o *DefaultOptimizer) SetLevel(level OptimizationLevel) {
	o.level = level
}
