# Installation Guide

Getting Glide up and running takes less than 2 minutes.

## System Requirements

- **Operating System**: macOS, Linux, or Windows
- **Architecture**: AMD64 or ARM64
- **Disk Space**: ~50MB for the binary

## Installation Methods

### Quick Install (Recommended)

For macOS and Linux users, use our installation script:

```bash
curl -sSL https://raw.githubusercontent.com/ivannovak/glide/main/install.sh | bash
```

This script will:
1. Detect your operating system and architecture
2. Download the appropriate binary
3. Install it to `/usr/local/bin`
4. Verify the installation

### Manual Installation

1. Download the appropriate binary from the [releases page](https://github.com/ivannovak/glide/releases)
2. Extract the archive
3. Move the binary to your PATH:

```bash
# Example for macOS/Linux
tar -xzf glide_*.tar.gz
chmod +x glid
sudo mv glide/usr/local/bin/
```

### Homebrew (Coming Soon)

```bash
# This will be available soon
brew install ivannovak/tap/glide
```

## Verify Installation

After installation, verify Glide is working:

```bash
glide version
```

You should see output like:
```
Glide CLI v0.9.0 (2025-09-11)
```

## Shell Completion (Optional)

Enable tab completion for your shell:

```bash
# Bash
glide completion bash > ~/.glide-completion.bash
echo "source ~/.glide-completion.bash" >> ~/.bashrc

# Zsh
glide completion zsh > ~/.glide-completion.zsh
echo "source ~/.glide-completion.zsh" >> ~/.zshrc

# Fish
glide completion fish > ~/.config/fish/completions/glide.fish
```

## Updating Glide

Glide can update itself:

```bash
glide self-update
```

## Troubleshooting

### Permission Denied

If you get a permission error during installation:

```bash
# Make the binary executable
chmod +x glid

# Use sudo for system-wide installation
sudo mv glide/usr/local/bin/
```

### Command Not Found

If `glid` is not found after installation:

1. Check if the binary is in your PATH:
   ```bash
   which glid
   ```

2. Add `/usr/local/bin` to your PATH if needed:
   ```bash
   echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```

## Next Steps

Now that Glide is installed, continue to [First Steps](first-steps.md) to learn the essential commands.
