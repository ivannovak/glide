// Package performance provides performance budgets and benchmarking utilities.
//
// This package defines performance targets for critical operations and
// provides utilities for measuring and validating performance against
// those targets. It's used for CI/CD performance regression detection.
//
// # Performance Budgets
//
// Define maximum acceptable durations for operations:
//
//	budgets := performance.ListBudgets()
//	for _, budget := range budgets {
//	    fmt.Printf("%s: max %v (%s priority)\n",
//	        budget.Name,
//	        budget.MaxDuration,
//	        budget.Priority)
//	}
//
// # Budget Checking
//
// Validate operations against budgets:
//
//	budget, ok := performance.GetBudget("context_detection")
//	if ok {
//	    duration := measureContextDetection()
//	    if duration > budget.MaxDuration {
//	        log.Warn("Performance budget exceeded",
//	            "operation", budget.Name,
//	            "actual", duration,
//	            "budget", budget.MaxDuration)
//	    }
//	}
//
// # Standard Budgets
//
// Pre-defined budgets for Glide operations:
//
//	context_detection   < 100ms   (Critical)
//	config_load         < 50ms    (Critical)
//	plugin_discovery    < 500ms   (Critical)
//	startup_total       < 200ms   (Critical)
//
// # CI Integration
//
// Use in benchmarks for regression detection:
//
//	func BenchmarkContextDetection(b *testing.B) {
//	    budget := performance.MustGetBudget("context_detection")
//	    for i := 0; i < b.N; i++ {
//	        detector.Detect()
//	    }
//	    // CI compares against budget
//	}
//
// # Custom Budgets
//
// Register application-specific budgets:
//
//	performance.RegisterBudget(performance.Budget{
//	    Name:        "my_operation",
//	    MaxDuration: 50 * time.Millisecond,
//	    Priority:    "critical",
//	    Description: "Maximum time for my operation",
//	})
//
// See docs/PERFORMANCE.md for complete performance documentation.
package performance
