// Package shell provides shell command execution for Glide.
//
// This package handles executing shell commands with proper handling of
// arguments, environment variables, timeouts, and output streaming.
//
// # Basic Execution
//
// Execute a simple command:
//
//	executor := shell.NewExecutor()
//	result, err := executor.Execute("ls", []string{"-la"})
//	if err != nil {
//	    return err
//	}
//	fmt.Println(result.Stdout)
//
// # Context-Aware Execution
//
// Execute with context for cancellation and timeout:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	result, err := executor.ExecuteContext(ctx, "long-running-command", nil)
//
// # Environment Variables
//
// Pass custom environment variables:
//
//	result, err := executor.Execute("script.sh", nil,
//	    shell.WithEnv("MY_VAR", "value"),
//	    shell.WithEnv("ANOTHER", "value"),
//	)
//
// # Working Directory
//
// Execute in a specific directory:
//
//	result, err := executor.Execute("npm", []string{"install"},
//	    shell.WithDir("/path/to/project"),
//	)
//
// # Output Handling
//
// The Result struct contains execution details:
//
//	type Result struct {
//	    ExitCode int
//	    Stdout   string
//	    Stderr   string
//	    Duration time.Duration
//	}
//
// # Streaming Output
//
// Stream output in real-time:
//
//	executor := shell.NewExecutor(shell.WithStreaming(os.Stdout, os.Stderr))
//	result, err := executor.Execute("make", []string{"build"})
//
// # Error Handling
//
// Non-zero exit codes are returned as errors:
//
//	result, err := executor.Execute("failing-command", nil)
//	if err != nil {
//	    if result != nil {
//	        fmt.Printf("Exit code: %d\n", result.ExitCode)
//	        fmt.Printf("Stderr: %s\n", result.Stderr)
//	    }
//	}
package shell
