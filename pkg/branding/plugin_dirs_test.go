package branding

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPluginDirName(t *testing.T) {
	tests := []struct {
		name           string
		configFileName string
		want           string
	}{
		{
			name:           "standard glide config",
			configFileName: ".glide.yml",
			want:           ".glide",
		},
		{
			name:           "custom cli config",
			configFileName: ".mycli.yml",
			want:           ".mycli",
		},
		{
			name:           "no extension",
			configFileName: ".myconfig",
			want:           ".myconfig",
		},
		{
			name:           "multiple dots",
			configFileName: ".my.cli.yml",
			want:           ".my.cli",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := ConfigFileName
			defer func() { ConfigFileName = original }()

			// Set test value
			ConfigFileName = tt.configFileName

			if got := GetPluginDirName(); got != tt.want {
				t.Errorf("GetPluginDirName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetGlobalPluginDir(t *testing.T) {
	// Save original value
	original := ConfigFileName
	defer func() { ConfigFileName = original }()

	home, _ := os.UserHomeDir()

	tests := []struct {
		name           string
		configFileName string
		want           string
	}{
		{
			name:           "standard glide",
			configFileName: ".glide.yml",
			want:           filepath.Join(home, ".glide", "plugins"),
		},
		{
			name:           "custom cli",
			configFileName: ".mycli.yml",
			want:           filepath.Join(home, ".mycli", "plugins"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ConfigFileName = tt.configFileName

			if got := GetGlobalPluginDir(); got != tt.want {
				t.Errorf("GetGlobalPluginDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLocalPluginDir(t *testing.T) {
	// Save original value
	original := ConfigFileName
	defer func() { ConfigFileName = original }()

	tests := []struct {
		name           string
		configFileName string
		baseDir        string
		want           string
	}{
		{
			name:           "current directory glide",
			configFileName: ".glide.yml",
			baseDir:        ".",
			want:           filepath.Join(".", ".glide", "plugins"),
		},
		{
			name:           "absolute path",
			configFileName: ".glide.yml",
			baseDir:        "/usr/local",
			want:           filepath.Join("/usr/local", ".glide", "plugins"),
		},
		{
			name:           "custom cli",
			configFileName: ".mycli.yml",
			baseDir:        "/project",
			want:           filepath.Join("/project", ".mycli", "plugins"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ConfigFileName = tt.configFileName

			if got := GetLocalPluginDir(tt.baseDir); got != tt.want {
				t.Errorf("GetLocalPluginDir(%v) = %v, want %v", tt.baseDir, got, tt.want)
			}
		})
	}
}
