package config

// Load loads the configuration from the config file
// This is a convenience function for backward compatibility
func Load() (*Config, error) {
	loader := NewLoader()
	return loader.Load()
}
