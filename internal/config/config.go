package config

// Load loads the configuration from the config file
// This is a convenience function for backward compatibility
func Load() (*Config, error) {
	loader := NewLoader()
	return loader.Load()
}

// Save saves the configuration to the config file
// This is a convenience function for backward compatibility
func Save(cfg *Config) error {
	loader := NewLoader()
	return loader.Save(cfg)
}