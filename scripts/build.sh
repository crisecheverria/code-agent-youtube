#!/bin/bash
set -euo pipefail

# Build script for Code Agent
# Builds both the core server and TUI client for all platforms

echo "🔨 Building Painika..."

# Build the core server
echo "📦 Building core server..."
cd packages/core
bun run build
cd ../..

# Create bin directory if it doesn't exist
mkdir -p bin

# Build for all platforms
echo "🏗️  Building TUI client for all platforms..."

# Linux builds
echo "🐧 Building for Linux..."
cd packages/tui
GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o ../../bin/painika-linux-amd64 ./main.go
GOOS=linux GOARCH=arm64 go build -ldflags='-s -w' -o ../../bin/painika-linux-arm64 ./main.go

# macOS builds
echo "🍎 Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -ldflags='-s -w' -o ../../bin/painika-darwin-amd64 ./main.go
GOOS=darwin GOARCH=arm64 go build -ldflags='-s -w' -o ../../bin/painika-darwin-arm64 ./main.go

# Windows builds
echo "🪟 Building for Windows..."
GOOS=windows GOARCH=amd64 go build -ldflags='-s -w' -o ../../bin/painika-windows-amd64.exe ./main.go

cd ../..

echo "✅ Build complete! Binaries are in the bin/ directory:"
ls -la bin/painika-*

echo ""
echo "📋 Next steps:"
echo "1. Test a binary: ./bin/painika-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/') --help"
echo "2. Create a GitHub release and upload these binaries"
echo "3. Update install.sh REPO variable with your GitHub repo"
echo "4. Test installation: curl -fsSL https://your-domain.com/install.sh | bash"