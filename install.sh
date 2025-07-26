#!/bin/bash
set -euo pipefail

# Code Agent Installation Script
# Based on OpenCode's approach

# Configuration
REPO="crisecheverria/code-agent-youtube" # Update with your GitHub repo
INSTALL_DIR="$HOME/.code-agent/bin"
BINARY_NAME="code-agent"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
  echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
  local os
  local arch

  case "$(uname -s)" in
  Darwin*)
    os="darwin"
    ;;
  Linux*)
    os="linux"
    ;;
  CYGWIN* | MINGW* | MSYS*)
    os="windows"
    ;;
  *)
    log_error "Unsupported operating system: $(uname -s)"
    exit 1
    ;;
  esac

  case "$(uname -m)" in
  x86_64 | amd64)
    arch="amd64"
    ;;
  arm64 | aarch64)
    arch="arm64"
    ;;
  armv7l)
    arch="arm"
    ;;
  *)
    log_error "Unsupported architecture: $(uname -m)"
    exit 1
    ;;
  esac

  echo "${os}-${arch}"
}

# Get latest release version from GitHub
get_latest_version() {
  if command -v curl >/dev/null 2>&1; then
    curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
  else
    log_error "Neither curl nor wget is available. Please install one of them."
    exit 1
  fi
}

# Download and install binary
install_binary() {
  local platform=$1
  local version=${2:-$(get_latest_version)}

  if [ -z "$version" ]; then
    log_error "Could not determine latest version"
    exit 1
  fi

  log_info "Installing Code Agent ${version} for ${platform}"

  # Create install directory
  mkdir -p "$INSTALL_DIR"

  # Download URL
  local download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-${platform}"
  if [[ "$platform" == *"windows"* ]]; then
    download_url="${download_url}.exe"
  fi

  local binary_path="$INSTALL_DIR/$BINARY_NAME"
  if [[ "$platform" == *"windows"* ]]; then
    binary_path="${binary_path}.exe"
  fi

  log_info "Downloading from: $download_url"

  # Download binary
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$download_url" -o "$binary_path"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "$download_url" -O "$binary_path"
  else
    log_error "Neither curl nor wget is available"
    exit 1
  fi

  # Make executable
  chmod +x "$binary_path"

  log_success "Binary installed to: $binary_path"
}

# Add to PATH
add_to_path() {
  local shell_config=""
  local shell_name=$(basename "$SHELL")

  # Detect shell and config file
  case "$shell_name" in
  bash)
    if [[ -f "$HOME/.bashrc" ]]; then
      shell_config="$HOME/.bashrc"
    elif [[ -f "$HOME/.bash_profile" ]]; then
      shell_config="$HOME/.bash_profile"
    fi
    ;;
  zsh)
    shell_config="$HOME/.zshrc"
    ;;
  fish)
    shell_config="$HOME/.config/fish/config.fish"
    ;;
  *)
    log_warn "Unknown shell: $shell_name. Please manually add $INSTALL_DIR to your PATH"
    return
    ;;
  esac

  if [[ -z "$shell_config" ]]; then
    log_warn "Could not find shell config file. Please manually add $INSTALL_DIR to your PATH"
    return
  fi

  # Check if already in PATH
  if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
    log_info "Directory already in PATH"
    return
  fi

  # Add to shell config
  if [[ "$shell_name" == "fish" ]]; then
    echo "fish_add_path $INSTALL_DIR" >>"$shell_config"
  else
    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >>"$shell_config"
  fi

  log_success "Added $INSTALL_DIR to PATH in $shell_config"
  log_info "Please restart your shell or run: source $shell_config"
}

# Main installation function
main() {
  log_info "Starting Code Agent installation..."

  local platform
  platform=$(detect_platform)

  log_info "Detected platform: $platform"

  # Install binary
  install_binary "$platform" "$1"

  # Add to PATH
  add_to_path

  log_success "Installation complete!"
  log_info "Run 'code-agent --help' to get started"
  log_info "You may need to restart your shell for PATH changes to take effect"
}

# Run main function with optional version argument
main "${1:-}"

