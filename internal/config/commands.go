package config

import (
	"fmt"
	"strings"
)

// ParseCommands converts CommandMap to normalized Command structures
func ParseCommands(rawCommands CommandMap) (map[string]*Command, error) {
	commands := make(map[string]*Command)

	for name, value := range rawCommands {
		cmd, err := parseCommand(name, value)
		if err != nil {
			return nil, fmt.Errorf("error parsing command %s: %w", name, err)
		}
		commands[name] = cmd
	}

	return commands, nil
}

// parseCommand handles both string and map formats
func parseCommand(name string, value interface{}) (*Command, error) {

	switch v := value.(type) {
	case string:
		// Simple format: commands: build: "docker build"
		return &Command{Cmd: v}, nil

	case CommandMap:
		// Handle CommandMap type (which is map[string]interface{})
		return parseCommand(name, map[string]interface{}(v))

	case map[string]interface{}:
		// Structured format with additional properties
		cmd := &Command{}

		// Parse required cmd field
		if cmdStr, ok := v["cmd"].(string); ok {
			cmd.Cmd = cmdStr
		} else {
			return nil, fmt.Errorf("command must have 'cmd' field")
		}

		// Parse optional fields
		if alias, ok := v["alias"].(string); ok {
			cmd.Alias = alias
		}
		if desc, ok := v["description"].(string); ok {
			cmd.Description = desc
		}
		if help, ok := v["help"].(string); ok {
			cmd.Help = help
		}
		if cat, ok := v["category"].(string); ok {
			cmd.Category = cat
		}

		return cmd, nil

	case map[interface{}]interface{}:
		// YAML unmarshals to map[interface{}]interface{} sometimes
		// Convert to map[string]interface{} and retry
		strMap := make(map[string]interface{})
		for k, v := range v {
			if keyStr, ok := k.(string); ok {
				strMap[keyStr] = v
			}
		}
		return parseCommand(name, strMap)

	default:
		return nil, fmt.Errorf("invalid command format for %s", name)
	}
}

// ExpandCommand prepares a command for execution with parameter substitution
func ExpandCommand(cmd string, args []string) string {
	// Replace positional parameters
	expanded := cmd
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)
		expanded = strings.ReplaceAll(expanded, placeholder, arg)
	}

	// Handle $@ for all arguments
	if strings.Contains(expanded, "$@") {
		expanded = strings.ReplaceAll(expanded, "$@", strings.Join(args, " "))
	}

	// Handle $* as an alias for $@
	if strings.Contains(expanded, "$*") {
		expanded = strings.ReplaceAll(expanded, "$*", strings.Join(args, " "))
	}

	return expanded
}

// ValidateCommand checks if a command is valid
func ValidateCommand(cmd *Command) error {
	if cmd.Cmd == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Check for circular references (basic check)
	if strings.Contains(cmd.Cmd, "glide"+cmd.Alias) || strings.Contains(cmd.Cmd, "glide "+cmd.Alias) {
		return fmt.Errorf("command may contain circular reference")
	}

	return nil
}