#!/bin/bash
# Install Git hooks for Glide development

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOOKS_DIR="$(git rev-parse --git-dir)/hooks"

echo "📦 Installing Git hooks..."

# Install pre-commit hook
if [ -f "$HOOKS_DIR/pre-commit" ]; then
    echo "⚠️  Pre-commit hook already exists. Backing up to pre-commit.backup"
    cp "$HOOKS_DIR/pre-commit" "$HOOKS_DIR/pre-commit.backup"
fi

cp "$SCRIPT_DIR/pre-commit" "$HOOKS_DIR/pre-commit"
chmod +x "$HOOKS_DIR/pre-commit"

echo "✅ Pre-commit hook installed"
echo ""
echo "The pre-commit hook will now:"
echo "  • Automatically format Go files with gofmt"
echo "  • Ensure go.mod and go.sum are tidy"
echo ""
echo "To bypass the hook temporarily, use: git commit --no-verify"
