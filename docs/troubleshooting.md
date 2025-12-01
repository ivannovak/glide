# Glide Troubleshooting Guide

## Quick Diagnostics

Check your Glide installation and context:

```bash
glide version          # Verify Glide is installed
glide context          # Show detected project context
glide help             # See available commands
```

## Common Issues and Solutions

### Installation Issues

#### "Command not found: glid"

**Problem:** Glide is not in your PATH or not installed.

**Solutions:**
1. Check if Glide is installed:
   ```bash
   ls -la /usr/local/bin/glid
   which glid
   ```

2. Add to PATH if installed elsewhere:
   ```bash
   export PATH="$PATH:/path/to/glide"
   echo 'export PATH="$PATH:/path/to/glide"' >> ~/.bashrc
   ```

3. Reinstall Glide:
   ```bash
   # Download latest release
   curl -L https://github.com/ivannovak/glide/releases/latest/download/glid-$(uname -s)-$(uname -m) -o glid
   chmod +x glid
   sudo mv glide/usr/local/bin/
   ```

#### "Permission denied" when running glid

**Problem:** Glide binary lacks execute permission.

**Solution:**
```bash
chmod +x /usr/local/bin/glid
```

### Context Detection Issues

#### "No project detected"

**Problem:** Glide cannot find a project in the current directory.

**Solutions:**

1. **For Git repositories:**
   ```bash
   # Ensure you're in a Git repository
   git status
   # If not, initialize Git
   git init
   ```

2. **For standalone projects (no Git):**
   ```bash
   # Create a .glide.yml file
   cat > .glide.yml << 'EOF'
   commands:
     hello: echo "Hello from Glide!"
   EOF
   ```

3. **Navigate to project root:**
   ```bash
   # Find your project root
   find . -name ".git" -o -name ".glide.yml"
   cd /path/to/project
   ```

#### "Command only available in multi-worktree mode"

**Problem:** Trying to use `glide project` commands in single-repo mode.

**Solutions:**
1. Check current mode:
   ```bash
   glide context
   ```

2. Convert to multi-worktree mode:
   ```bash
   glide setup --mode multi
   ```

3. Or stay in single-repo mode and avoid `project` commands.

### YAML Command Issues

#### "Command not found" for YAML-defined command

**Problem:** Your YAML command isn't being recognized.

**Solutions:**

1. **Check YAML syntax:**
   ```yaml
   # .glide.yml - Correct format
   commands:
     build: docker build .
     test: npm test
   ```

2. **Verify file location:**
   ```bash
   # Must be in current or parent directory
   ls -la .glide.yml
   cat .glide.yml
   ```

3. **Check for typos:**
   ```bash
   # List available commands
   glide help
   # Your YAML commands should appear in the list
   ```

#### YAML command fails to execute

**Problem:** Command is recognized but execution fails.

**Solutions:**

1. **Test command directly:**
   ```bash
   # If this works:
   docker build .

   # But this doesn't:
   glidebuild  # where build: docker build .

   # Check for shell issues
   ```

2. **Use shell explicitly for complex commands:**
   ```yaml
   commands:
     # Instead of:
     complex: cd dir && npm install

     # Use:
     complex: |
       cd dir
       npm install
   ```

3. **Debug parameter substitution:**
   ```yaml
   commands:
     # Test with echo first
     deploy: echo "Would deploy to: $1"
   ```

### Plugin Issues

#### "Plugin not found" after installation

**Problem:** Installed plugin isn't recognized.

**Solutions:**

1. **Verify installation:**
   ```bash
   glideplugins list
   ls -la ~/.glide/plugins/
   ```

2. **Check plugin binary name:**
   ```bash
   # Plugin binaries should be executable
   chmod +x ~/.glide/plugins/plugin-name
   ```

3. **Ensure plugin is valid:**
   ```bash
   # Get plugin info
   glideplugins info plugin-name
   ```

#### Cannot install plugin

**Problem:** `glideplugins install` fails.

**Solutions:**

1. **Verify plugin binary exists:**
   ```bash
   ls -la /path/to/plugin
   file /path/to/plugin  # Should show executable
   ```

2. **Check plugin directory permissions:**
   ```bash
   mkdir -p ~/.glide/plugins
   chmod 755 ~/.glide/plugins
   ```

3. **Install with full path:**
   ```bash
   glideplugins install /absolute/path/to/plugin
   ```

### Multi-Worktree Issues

#### "Cannot create worktree"

**Problem:** `glide project worktree` fails.

**Solutions:**

1. **Ensure you're in multi-worktree mode:**
   ```bash
   glide context  # Should show "multi-worktree"
   # If not:
   glide setup --mode multi
   ```

2. **Check if worktree already exists:**
   ```bash
   git worktree list
   ls worktrees/
   ```

3. **Remove stale worktree:**
   ```bash
   git worktree remove worktrees/branch-name
   rm -rf worktrees/branch-name
   ```

4. **Ensure branch name is valid:**
   ```bash
   # Use valid Git branch names
   glide project worktree feature/my-feature  # Good
   glide project worktree "my feature"        # Bad (spaces)
   ```

#### Wrong directory structure after setup

**Problem:** Multi-worktree setup created unexpected structure.

**Expected structure:**
```
project/
├── vcs/          # Main repository
│   └── .git/
└── worktrees/    # Feature branches
    └── feature-a/
```

**Solutions:**
1. **Verify structure:**
   ```bash
   ls -la
   # Should show vcs/ and worktrees/ directories
   ```

2. **Manual fix if needed:**
   ```bash
   # If setup failed midway
   mv .git vcs/.git
   mv * vcs/ 2>/dev/null || true
   mkdir -p worktrees
   ```

### Shell Completion Issues

#### Completion not working

**Problem:** Tab completion doesn't work after installation.

**Solutions:**

1. **Generate completion for your shell:**
   ```bash
   # Bash
   glide completion bash > /tmp/glide.bash
   source /tmp/glide.bash

   # Zsh
   glide completion zsh > ~/.zsh/completions/_glide
   source ~/.zshrc

   # Fish
   glide completion fish > ~/.config/fish/completions/glide.fish
   ```

2. **Verify completion is loaded:**
   ```bash
   # Bash
   complete -p | grep glid

   # Zsh
   print -l $_comps | grep glid
   ```

### Configuration Issues

#### Cannot find configuration

**Problem:** Glide can't find `.glide.yml`.

**Solutions:**

1. **Check current directory:**
   ```bash
   pwd
   ls -la .glide.yml
   ```

2. **Check parent directories:**
   ```bash
   # Glide searches up to 5 levels
   find . -maxdepth 5 -name ".glide.yml"
   ```

3. **Create minimal config:**
   ```bash
   cat > .glide.yml << 'EOF'
   # Minimal Glide configuration
   commands:
     init: echo "Initialized"
   EOF
   ```

### Update Issues

#### Self-update fails

**Problem:** `glide self-update` doesn't work.

**Solutions:**

1. **Check GitHub connectivity:**
   ```bash
   curl -I https://api.github.com
   ```

2. **Manual update:**
   ```bash
   # Download specific version
   VERSION=v1.0.0  # Replace with desired version
   curl -L "https://github.com/ivannovak/glide/releases/download/$VERSION/glid-$(uname -s)-$(uname -m)" -o glid
   chmod +x glid
   sudo mv glide/usr/local/bin/
   ```

3. **Check current version:**
   ```bash
   glide version
   ```

## Debug Mode

Get more information about issues:

```bash
# Show detailed context information
glide context --json

# Check what commands are available
glide help

# See if YAML commands are loaded
grep -A 10 "^commands:" .glide.yml
```

## Getting Help

If issues persist:

1. **Gather debug information:**
   ```bash
   glide version > debug.txt
   glide context >> debug.txt
   cat .glide.yml >> debug.txt
   ```

2. **Check for known issues:**
   - Visit: https://github.com/ivannovak/glide/issues

3. **Report new issue with:**
   - OS and version (`uname -a`)
   - Glide version (`glide version`)
   - Project structure (`ls -la`)
   - Debug output from above

## Emergency Recovery

### Reset Glide Configuration

```bash
# Backup existing config
mv ~/.glide ~/.glide.backup

# Remove project config
mv .glide.yml .glide.yml.backup

# Start fresh
glide setup
```

### Manual Fallback

If Glide won't run, execute your commands directly:

```bash
# Instead of YAML command
# glidebuild
# Run the actual command:
docker build .

# Instead of plugin command
# Just run the underlying tool directly
```

## Platform-Specific Issues

### macOS
- **Binary not trusted:** System Preferences → Security & Privacy → Allow glid
- **PATH issues:** Add to `.zshrc` instead of `.bashrc`

### Linux
- **Permission issues:** Ensure user owns `~/.glide/` directory
- **Binary architecture:** Download correct version (amd64 vs arm64)

### Windows (WSL2)
- **Line endings:** Use Unix line endings in `.glide.yml`
- **Path format:** Use Unix paths (`/mnt/c/...` not `C:\...`)

## Maintenance

Keep Glide running smoothly:

```bash
# Check for updates
glide self-update --check

# Clean old plugins
ls -la ~/.glide/plugins/
# Remove unused plugins manually

# Verify setup
glide context
glide help
```
