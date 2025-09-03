package context

// Detect is a convenience function to detect the current project context
func Detect() *ProjectContext {
	detector, err := NewDetector()
	if err != nil {
		return &ProjectContext{
			WorkingDir: "", // We don't know the working directory
			Error:      err,
		}
	}
	
	ctx, err := detector.Detect()
	if err != nil {
		// Even if detection fails, return the context with basic info
		ctx.Error = err
	}
	return ctx
}