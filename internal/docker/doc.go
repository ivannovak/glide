// Package docker provides Docker integration for Glide.
//
// This package handles Docker daemon communication, container management,
// and Docker Compose operations. It provides a high-level API for Docker
// operations used by Glide commands and plugins.
//
// # Docker Availability
//
// Check if Docker is available:
//
//	if docker.IsAvailable() {
//	    // Docker daemon is running
//	}
//
// # Container Operations
//
// Execute commands in containers:
//
//	client := docker.NewClient()
//
//	result, err := client.Exec("mycontainer", []string{"npm", "test"})
//	if err != nil {
//	    return err
//	}
//	fmt.Println(result.Output)
//
// # Docker Compose
//
// Work with Docker Compose files:
//
//	compose := docker.NewCompose("docker-compose.yml")
//
//	err := compose.Up(docker.UpOptions{
//	    Detach: true,
//	    Build:  true,
//	})
//
//	err = compose.Down(docker.DownOptions{
//	    Volumes: true,
//	})
//
// # Container Information
//
// Get information about containers:
//
//	containers, err := client.ListContainers()
//	for _, c := range containers {
//	    fmt.Printf("%s: %s\n", c.Name, c.Status)
//	}
//
// # Lazy Initialization
//
// For performance, Docker checks can be deferred:
//
//	// Fast startup - Docker check deferred
//	detector := context.NewDetectorFast()
//
//	// Check Docker when actually needed
//	if docker.IsAvailable() {
//	    // Perform Docker operations
//	}
//
// # Error Handling
//
// Docker-specific errors:
//
//	err := client.Exec(...)
//	if docker.IsNotRunning(err) {
//	    fmt.Println("Please start Docker Desktop")
//	}
package docker
