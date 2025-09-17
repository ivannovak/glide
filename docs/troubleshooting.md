# Glide Troubleshooting Guide

## Quick Diagnostics

Before diving into specific issues, run Glide's built-in diagnostics:

```bash
glide doctor          # Run diagnostic checks
glide doctor --fix    # Attempt automatic fixes
glide context         # Show current context
glide version --full  # Display detailed version info
```

## Common Issues and Solutions

### Installation Issues

#### "Command not found: glide"

**Problem:** Glide is not in your PATH or not installed.

**Solutions:**
1. Check if Glide is installed:
   ```bash
   ls -la /usr/local/bin/glide
   ```

2. Add to PATH if installed elsewhere:
   ```bash
   export PATH="$PATH:/path/to/glide"
   echo 'export PATH="$PATH:/path/to/glide"' >> ~/.bashrc
   ```

3. Reinstall Glide:
   ```bash
   curl -sSL https://glide.dev/install | bash
   ```

#### "Permission denied" when running glide

**Problem:** Glide binary lacks execute permission.

**Solution:**
```bash
chmod +x /usr/local/bin/glide
```

### Docker Issues

#### "Docker daemon is not running"

**Problem:** Docker Desktop is not started or Docker service is stopped.

**Solutions:**

**macOS:**
```bash
# Start Docker Desktop
open -a Docker
# Wait for Docker to start, then verify
docker info
```

**Linux:**
```bash
sudo systemctl start docker
sudo systemctl enable docker
```

#### "Cannot connect to Docker daemon"

**Problem:** User lacks permission to access Docker socket.

**Solution:**
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
# Log out and back in, then verify
docker run hello-world
```

#### "Port already in use"

**Problem:** Required ports are occupied by other services.

**Solutions:**
1. Find what's using the port:
   ```bash
   lsof -i :80    # Check port 80
   lsof -i :3306  # Check port 3306
   ```

2. Stop conflicting service or change Glide's ports:
   ```yaml
   # docker-compose.override.yml
   services:
     nginx:
       ports:
         - "8080:80"  # Use 8080 instead
   ```

#### "No such service" error

**Problem:** Trying to execute commands in non-existent container.

**Solutions:**
1. Check available services:
   ```bash
   glide ps
   docker-compose ps
   ```

2. Ensure services are running:
   ```bash
   glide up
   ```

### Context Detection Issues

#### "Not in a Glide project directory"

**Problem:** Glide cannot detect project structure.

**Solutions:**
1. Navigate to project directory:
   ```bash
   cd ~/Code/myproject
   ```

2. Initialize project if new:
   ```bash
   glide setup
   ```

3. Check for `.git` directory:
   ```bash
   ls -la .git
   # If missing, initialize Git
   git init
   ```

#### "Command only available in multi-worktree mode"

**Problem:** Trying to use multi-worktree commands in single-repo mode.

**Solutions:**
1. Check current mode:
   ```bash
   glide context
   ```

2. Convert to multi-worktree (if desired):
   ```bash
   glide setup --mode multi
   ```

3. Use appropriate commands for your mode:
   ```bash
   # Single-repo mode
   glide test
   
   # Multi-worktree mode
   glide project test
   ```

### Configuration Issues

#### "Configuration file not found"

**Problem:** Glide cannot find or read configuration.

**Solutions:**
1. Create default configuration:
   ```bash
   glide config init
   ```

2. Check configuration location:
   ```bash
   echo $GLIDE_CONFIG
   ls -la ~/.glide.yml
   ```

3. Validate existing configuration:
   ```bash
   glide config validate
   ```

#### "Invalid configuration"

**Problem:** Configuration file has syntax errors or invalid values.

**Solutions:**
1. Validate configuration:
   ```bash
   glide config validate
   ```

2. Reset to defaults:
   ```bash
   mv ~/.glide.yml ~/.glide.yml.backup
   glide config init
   ```

3. Fix YAML syntax:
   ```yaml
   # Correct indentation and structure
   projects:
     myproject:
       path: /path/to/project
       mode: multi
   ```

### Test Execution Issues

#### "No test framework detected"

**Problem:** Glide cannot identify test runner.

**Solutions:**
1. Specify test framework explicitly:
   ```bash
   glide test --framework phpunit
   glide test --framework jest
   ```

2. Ensure test framework is installed:
   ```bash
   # PHP
   composer require --dev phpunit/phpunit
   
   # JavaScript
   npm install --save-dev jest
   ```

#### "Tests failing in Docker"

**Problem:** Tests pass locally but fail in containers.

**Solutions:**
1. Ensure database is migrated:
   ```bash
   glide exec php artisan migrate --env=testing
   ```

2. Check environment variables:
   ```bash
   glide exec php printenv | grep DB_
   ```

3. Clear caches:
   ```bash
   glide exec php artisan cache:clear
   glide exec php composer dump-autoload
   ```

### Database Issues

#### "Connection refused" to database

**Problem:** Cannot connect to MySQL/PostgreSQL.

**Solutions:**
1. Ensure database container is running:
   ```bash
   glide ps
   glide up mysql
   ```

2. Check connection parameters:
   ```bash
   glide exec php printenv | grep DB_
   ```

3. Wait for database to be ready:
   ```bash
   glide exec mysql mysqladmin ping -h mysql --wait
   ```

#### "Access denied for user"

**Problem:** Database credentials are incorrect.

**Solutions:**
1. Verify credentials in `.env`:
   ```bash
   grep DB_ .env
   ```

2. Grant permissions:
   ```bash
   glide exec mysql mysql -u root -p
   # In MySQL:
   GRANT ALL ON *.* TO 'user'@'%';
   FLUSH PRIVILEGES;
   ```

### Performance Issues

#### "Glide commands are slow"

**Problem:** Commands take too long to execute.

**Solutions:**
1. Check Docker resource allocation:
   - Docker Desktop → Preferences → Resources
   - Increase CPU and Memory limits

2. Clean Docker system:
   ```bash
   docker system prune -a
   docker volume prune
   ```

3. Disable unnecessary services:
   ```yaml
   # docker-compose.override.yml
   services:
     redis:
       profiles: ["cache"]  # Only start when needed
   ```

#### "High memory usage"

**Problem:** Docker containers consuming too much memory.

**Solutions:**
1. Limit container memory:
   ```yaml
   services:
     php:
       mem_limit: 512m
   ```

2. Check for memory leaks:
   ```bash
   glide exec php php -i | grep memory_limit
   docker stats
   ```

### Plugin Issues

#### "Plugin not found"

**Problem:** Installed plugin is not recognized.

**Solutions:**
1. Check plugin installation:
   ```bash
   glide plugins list
   ls -la ~/.glide/plugins/
   ```

2. Ensure plugin has correct name:
   ```bash
   # Must start with 'glide-plugin-'
   mv myplugin glide-plugin-myplugin
   ```

3. Verify plugin permissions:
   ```bash
   chmod +x ~/.glide/plugins/glide-plugin-*
   ```

#### "Plugin command failed"

**Problem:** Plugin executes but returns errors.

**Solutions:**
1. Check plugin logs:
   ```bash
   GLIDE_PLUGIN_DEBUG=1 glide myplugin command
   ```

2. Verify plugin dependencies:
   ```bash
   glide plugins info myplugin
   ```

3. Reinstall plugin:
   ```bash
   glide plugins uninstall myplugin
   glide plugins install /path/to/plugin
   ```

### Multi-Worktree Issues

#### "Cannot create worktree"

**Problem:** Git worktree creation fails.

**Solutions:**
1. Check if worktree already exists:
   ```bash
   git worktree list
   ```

2. Remove stale worktree:
   ```bash
   git worktree remove worktrees/feature-name
   rm -rf worktrees/feature-name
   ```

3. Ensure branch doesn't exist:
   ```bash
   git branch -a | grep feature-name
   git branch -D feature-name  # If local
   ```

#### "Worktree Docker conflicts"

**Problem:** Multiple worktrees' Docker containers conflict.

**Solutions:**
1. Stop all containers:
   ```bash
   glide project down
   ```

2. Use unique project names:
   ```bash
   # In each worktree's .env
   COMPOSE_PROJECT_NAME=project_feature_name
   ```

3. Use different ports per worktree:
   ```yaml
   # worktrees/feature1/docker-compose.override.yml
   services:
     nginx:
       ports:
         - "8081:80"
   ```

### File Permission Issues

#### "Permission denied" on files

**Problem:** Cannot read/write project files.

**Solutions:**
1. Fix ownership:
   ```bash
   sudo chown -R $(whoami):$(whoami) .
   ```

2. Fix permissions:
   ```bash
   find . -type f -exec chmod 644 {} \;
   find . -type d -exec chmod 755 {} \;
   chmod +x vendor/bin/*
   ```

3. In Docker, match user IDs:
   ```dockerfile
   ARG UID=1000
   ARG GID=1000
   RUN usermod -u $UID www-data && groupmod -g $GID www-data
   ```

### Network Issues

#### "Could not resolve host"

**Problem:** DNS resolution failing in containers.

**Solutions:**
1. Check Docker DNS settings:
   ```json
   // ~/.docker/daemon.json
   {
     "dns": ["8.8.8.8", "8.8.4.4"]
   }
   ```

2. Restart Docker:
   ```bash
   # macOS
   osascript -e 'quit app "Docker"'
   open -a Docker
   
   # Linux
   sudo systemctl restart docker
   ```

#### "Connection timeout"

**Problem:** Network requests timing out.

**Solutions:**
1. Check proxy settings:
   ```bash
   echo $HTTP_PROXY
   echo $HTTPS_PROXY
   ```

2. Configure Docker proxy:
   ```json
   // ~/.docker/config.json
   {
     "proxies": {
       "default": {
         "httpProxy": "http://proxy:8080",
         "httpsProxy": "http://proxy:8080",
         "noProxy": "localhost,127.0.0.1"
       }
     }
   }
   ```

## Debug Mode

Enable debug output for more information:

```bash
# Verbose output
glide --verbose test

# Debug plugins
GLIDE_PLUGIN_DEBUG=1 glide myplugin command

# Docker compose debug
COMPOSE_DEBUG=1 glide up
```

## Getting Help

If issues persist after trying these solutions:

1. **Check documentation:**
   ```bash
   glide help troubleshooting
   glide help <command>
   ```

2. **Run diagnostics:**
   ```bash
   glide doctor --verbose
   ```

3. **Gather debug information:**
   ```bash
   glide context --json > glide-context.json
   glide doctor --json > glide-doctor.json
   docker info > docker-info.txt
   ```

4. **Report issue:**
   - Include output from debug commands
   - Describe steps to reproduce
   - Mention OS and Glide version
   - File issue at: https://github.com/ivannovak/glide/issues

## Emergency Recovery

If Glide becomes completely non-functional:

### Reset Everything

```bash
# Stop all Docker containers
docker stop $(docker ps -aq)

# Remove Glide configuration
rm -rf ~/.glide.yml
rm -rf ~/.glide/

# Clean Docker
docker system prune -a --volumes

# Reinstall Glide
curl -sSL https://glide.dev/install | bash

# Reconfigure
glide setup
```

### Manual Fallback

If Glide won't run at all, use manual commands:

```bash
# Instead of 'glide up'
docker-compose up -d

# Instead of 'glide test'
docker-compose exec php vendor/bin/phpunit

# Instead of 'glide exec'
docker-compose exec php bash
```

## Platform-Specific Issues

### macOS

- **Slow file system:** Use delegated mounts in docker-compose.yml
- **Port conflicts:** Check for services using `lsof -i :<port>`
- **Docker Desktop issues:** Reset Docker Desktop from preferences

### Linux

- **SELinux conflicts:** Add `:Z` to volume mounts
- **Systemd issues:** Check `journalctl -u docker`
- **Permission issues:** Add user to docker group

### Windows (WSL2)

- **Path issues:** Use Unix-style paths in WSL
- **Line endings:** Configure Git: `git config --global core.autocrlf input`
- **Performance:** Keep projects in WSL filesystem, not Windows

## Maintenance Commands

Regular maintenance can prevent issues:

```bash
# Weekly cleanup
glide clean --all
docker system prune
glide update

# Check for updates
glide upgrade --check

# Validate setup
glide doctor
glide config validate
```