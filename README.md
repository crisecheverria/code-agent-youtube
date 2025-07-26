# Painika

An AI-powered coding assistant with a client-server architecture built using TypeScript (Bun) for the server and Go for the TUI client.

## Features

- 🤖 AI-powered coding assistance using Groq API
- 🔧 Built-in tools for file operations and shell commands
- 💬 Interactive TUI with conversation history
- 📊 Token usage tracking
- 🚀 Cross-platform support (Linux, macOS, Windows)
- 📦 Multiple installation methods (shell script, npm, direct download)

## Quick Start

### Option 1: Install Script (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/crisecheverria/painika/main/install.sh | bash
```

### Option 2: Build from Source

```bash
git clone https://github.com/crisecheverria/painika.git
cd painika
bun install
./scripts/build.sh
```

## Usage

1. **Set your API key:**

   ```bash
   export GROQ_API_KEY="your_groq_api_key_here"
   ```

2. **Start the server:**

   ```bash
   painika server
   ```

3. **In another terminal, start the TUI client:**

   ```bash
   painika
   ```

## Configuration

Environment variables:

- `GROQ_API_KEY` - Your Groq API key (required)
- `MODEL` - AI model to use (default: `llama-3.3-70b-versatile`)
- `SERVER_URL` - Server URL (default: `http://localhost:3000`)

## Available Commands

- `help, h` - Show help message
- `tokens, t` - Show token usage statistics
- `history, hist` - Show conversation history
- `clear, c` - Clear the screen
- `reset, r` - Reset conversation history
- `quit, q` - Exit the application

## Development

### Prerequisites

- [Bun](https://bun.sh/) - JavaScript runtime and package manager
- [Go](https://golang.org/) 1.21 or later
- Groq API key

### Setup

```bash
git clone https://github.com/crisecheverria/painika.git
cd painika
bun install
```

### Development Mode

Start the server:

```bash
bun run dev
```

In another terminal, run the TUI:

```bash
cd packages/tui
go run main.go
```

### Building

Build for all platforms:

```bash
./scripts/build.sh
```

## Architecture

- **Server** (`packages/core`): TypeScript/Bun HTTP server with Hono framework
- **Client** (`packages/tui`): Go-based terminal user interface
- **Communication**: REST API with JSON payloads

## Installation Methods

### 1. Shell Script Installation

- Cross-platform installation script
- Automatic platform detection
- PATH configuration
- Inspired by OpenCode's approach

### 2. npm Global Package

- Standard Node.js package manager
- Universal launcher script
- Works with existing Node.js workflows

### 3. GitHub Releases

- Pre-built binaries for all platforms
- Automated releases via GitHub Actions
- Direct download and execution

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Troubleshooting

### Server not running

```bash
# Check if server is running
curl http://localhost:3000/health

# Start server if needed
painika server
```

### Permission denied

```bash
# Make sure scripts are executable
chmod +x install.sh scripts/build.sh bin/painika
```

### Binary not found

```bash
# Build the project first
./scripts/build.sh
```

