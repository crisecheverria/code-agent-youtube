# Painika

An AI-powered coding assistant that automatically manages its client-server architecture.

## ✨ Features

- 🤖 **AI-powered coding assistance** using Groq API
- 🔧 **Built-in tools** for file operations and shell commands  
- 💬 **Interactive TUI** with conversation history
- 📊 **Token usage tracking** and cost estimation
- 🚀 **Cross-platform support** (Linux, macOS, Windows)
- ⚡ **Auto server management** - no manual setup needed
- 🔄 **Smart port detection** - works even if port 3000 is busy
- 🧹 **Automatic cleanup** - server stops when you quit
- 📂 **Shell config integration** - reads API keys from .zshrc/.bashrc

## 🚀 Getting Started

### Clone and Build
```bash
git clone https://github.com/crisecheverria/painika.git
cd painika
bun install
bun run build
```

### Setup API Key
Add your Groq API key to your shell config:

```bash
# Add to ~/.zshrc (or ~/.bashrc)
echo 'export GROQ_API_KEY="your_groq_api_key_here"' >> ~/.zshrc
source ~/.zshrc
```

Get your free API key at: [console.groq.com/keys](https://console.groq.com/keys)

## ⚙️ Configuration

Painika automatically detects configuration from multiple sources:

### API Key Sources (in priority order):
1. **Environment variable**: `export GROQ_API_KEY="..."`  
2. **Shell config files**: `~/.zshrc`, `~/.bashrc`, `~/.bash_profile`, `~/.profile`
3. **Local .env file**: `.env` in current directory

### Optional Settings
```bash
# AI model to use (default: llama-3.3-70b-versatile)
export MODEL="llama-3.1-8b-instant"

# Custom server URL (auto-detected by default)
export SERVER_URL="http://localhost:3000"  
```

### Available Groq Models
- `llama-3.3-70b-versatile` (default - smartest)
- `llama-3.1-8b-instant` (fastest) 
- `llama-3.1-70b-versatile`
- `mixtral-8x7b-32768`
- `gemma2-9b-it`

💡 **Pro tip**: Add your API key to `~/.zshrc` once and forget about it!

## 💬 Commands

Once inside Painika, you can use these commands:

| Command | Description |
|---------|-------------|
| `help`, `h` | Show help message |
| `tokens`, `t` | Show token usage & cost estimate |
| `history`, `hist` | Show conversation history |
| `clear`, `c` | Clear the screen |
| `reset`, `r` | Reset conversation history |
| `quit`, `q` | Exit (automatically stops server) |

### Example Session
```bash
💬 > help me optimize this Python function
🤖 I'd be happy to help optimize your Python function! Could you share the code?

💬 > tokens
📊 Token Usage Statistics:
   Input tokens:  150
   Output tokens: 45
   Total tokens:  195
   Estimated cost: $0.0001

💬 > quit
👋 Goodbye!
🧹 Stopping server...
✅ Server stopped
```

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

## 🏗️ Architecture

Painika uses a smart client-server architecture:

- **Single Binary**: Contains both client and embedded server
- **Go Client**: Terminal interface with automatic server management  
- **Embedded TypeScript Server**: Bun-powered API server bundled inside the binary
- **Auto-Discovery**: Client detects server port and manages lifecycle
- **Zero Config**: Works out of the box with sensible defaults

### How It Works
1. Run `painika` → Client checks if server is running
2. If not → Client extracts and starts embedded server 
3. Server finds available port (3000, 3001, 3002...)
4. Client connects to server's actual port
5. You chat with AI → Server handles Groq API calls
6. Type `quit` → Client stops server and cleans up


## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 🔧 Troubleshooting

### "GROQ_API_KEY environment variable is required"
```bash
# Add API key to your shell config
echo 'export GROQ_API_KEY="your_key_here"' >> ~/.zshrc
source ~/.zshrc
```

### "Server failed to start"
- **Port conflict**: Painika automatically finds available ports (3000-3100)
- **Missing dependencies**: Make sure `bun` is installed for server functionality
- **Permissions**: Run `chmod +x ~/.painika/bin/painika` if needed

### Manual Server Management
```bash
# Start server only (if needed)
painika server

# Check server health
curl http://localhost:3000/health  # or whatever port is shown
```

### Reset Everything
```bash
# Kill any stuck processes
pkill -f painika
pkill -f "bun run"

# Restart fresh
painika
```

