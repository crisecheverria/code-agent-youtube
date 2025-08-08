package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

//go:embed server.js
var serverBundle string

// Global server process for cleanup
var globalServerCmd *exec.Cmd

// Configuration structure
type Config struct {
	ServerURL string
	Token     string
	Model     string
}

// HTTP client wrapper
type Client struct {
	config Config
	client *http.Client
}

// Message structure (matching TypeScript)
type Message struct {
	ID        string `json:"id"`
	Role      string `json:"role"` // "user" or "assistant"
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"` // ISO 8601 format
}

// Converation structure
type Conversation struct {
	ID          string    `json:"id"`
	Messages    []Message `json:"messages"`
	TotalTokens struct {
		Input  int `json:"input"`
		Output int `json:"output"`
	} `json:"totalTokens"`
	CreatedAt string `json:"createdAt"` // ISO 8601 format
	UpdatedAt string `json:"updatedAt"` // ISO 8601 format
}

// Token usage structure
type TokenUsage struct {
	Input  int `json:"input"`
	Output int `json:"output"`
	Total  int `json:"total"`
}

// Session response structure
type SessionResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"sessionId"`
	Error     string `json:"error,omitempty"`
}

// Chat response structure
type ChatResponse struct {
	Success  bool      `json:"success"`
	Messages []Message `json:"messages"`
	Error    string    `json:"error,omitempty"`
}

// Create a new client
func NewClient(config Config) *Client {
	return &Client{
		config: config,
		client: &http.Client{},
	}
}

func (c *Client) InitSession() error {
	payload := map[string]interface{}{
		"groq": map[string]string{
			"token":   c.config.Token,
			"model":   c.config.Model,
			"baseURL": "https://api.groq.com/openai",
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(c.config.ServerURL+"/session", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("failed to initialize session: %s", result.Error)
	}

	return nil
}

func (c *Client) SendMessage(content string) (*ChatResponse, error) {
	payload := map[string]string{
		"content": content,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.config.ServerURL+"/message", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to send message: %s", result.Error)
	}

	return &result, nil
}

func (c *Client) GetConversation() (*Conversation, error) {
	resp, err := c.client.Get(c.config.ServerURL + "/conversation")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Success      bool          `json:"success"`
		Conversation *Conversation `json:"conversation"`
		Error        string        `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to get conversation: %s", result.Error)
	}

	return result.Conversation, nil
}

func (c *Client) GetTokenUsage() (*TokenUsage, error) {
	resp, err := c.client.Get(c.config.ServerURL + "/tokens")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool        `json:"success"`
		Usage   *TokenUsage `json:"usage"`
		Error   string      `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to get token usage: %s", result.Error)
	}

	return result.Usage, nil
}

func (c *Client) ClearConversation() error {
	resp, err := c.client.Post(c.config.ServerURL+"/session", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("failed to clear conversation: %s", result.Error)
	}

	return nil
}

// Get environment variables for configuration, checking shell config files
func getEnv(key, defaultValue string) string {
	// First check system environment
	value := os.Getenv(key)
	if value != "" {
		return value
	}

	// If not found, try to read from shell config files
	value = getEnvFromShellConfig(key)
	if value != "" {
		return value
	}

	return defaultValue
}

// Read environment variable from shell configuration files
func getEnvFromShellConfig(key string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Common shell config files to check
	configFiles := []string{
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".bash_profile"),
		filepath.Join(homeDir, ".profile"),
	}

	for _, configFile := range configFiles {
		if value := extractEnvFromFile(configFile, key); value != "" {
			return value
		}
	}

	return ""
}

// Extract environment variable from a config file
func extractEnvFromFile(filename, key string) string {
	file, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Look for export statements
		if strings.HasPrefix(line, "export "+key+"=") {
			value := strings.TrimPrefix(line, "export "+key+"=")
			return cleanEnvValue(value)
		}

		// Look for direct assignments
		if strings.HasPrefix(line, key+"=") {
			value := strings.TrimPrefix(line, key+"=")
			return cleanEnvValue(value)
		}
	}

	return ""
}

// Clean environment variable value (remove quotes, etc.)
func cleanEnvValue(value string) string {
	// Remove surrounding quotes
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return value[1 : len(value)-1]
	}
	return value
}

func main() {
	// Load .env file if it exists
	// Try loading from current directory first, then from packages/tui/
	if err := godotenv.Load(); err != nil {
		// Try loading from packages/tui directory
		if err2 := godotenv.Load("packages/tui/.env"); err2 != nil {
			// .env file not found in either location - continue with environment variables
			log.Printf("No .env file found or error loading it: %v", err)
		}
	}

	// Check if running as server
	if len(os.Args) > 1 && os.Args[1] == "server" {
		startServer()
		return
	}

	// Check for help flag
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printUsage()
		return
	}

	// Default: run as TUI client
	runTUI()
}

func printUsage() {
	fmt.Println("Painika - AI-powered coding assistant")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  painika          Start the TUI client (default)")
	fmt.Println("  painika server   Start the backend server")
	fmt.Println("  painika --help   Show this help message")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  GROQ_API_KEY        Your Groq API key (required)")
	fmt.Println("  MODEL               AI model to use (default: llama-3.3-70b-versatile)")
	fmt.Println("  SERVER_URL          Server URL (default: http://localhost:3000)")
	fmt.Println()
}

func startServer() {
	fmt.Println("🚀 Starting Code Agent server...")

	// Create a temporary file for the server bundle
	tempFile, err := ioutil.TempFile("", "server-*.js")
	if err != nil {
		log.Fatalf("❌ Failed to create temporary file: %v", err)
	}
	tempFileName := tempFile.Name()

	// Write the embedded server bundle to the temporary file
	if _, err := tempFile.WriteString(serverBundle); err != nil {
		tempFile.Close()
		os.Remove(tempFileName)
		log.Fatalf("❌ Failed to write server bundle: %v", err)
	}
	tempFile.Close()

	// Clean up temp file when server exits
	defer os.Remove(tempFileName)

	fmt.Printf("📦 Server bundle extracted to: %s\n", tempFileName)

	// Start the Bun server
	cmd := exec.Command("bun", "run", tempFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func runTUI() {
	// Load configuration from environment variables
	config := Config{
		ServerURL: getEnv("SERVER_URL", "http://localhost:3000"),
		Token:     getEnv("GROQ_API_KEY", ""),
		Model:     getEnv("MODEL", "llama-3.3-70b-versatile"),
	}

	// Validate configuration
	if config.Token == "" {
		fmt.Println("❌ GROQ_API_KEY environment variable is required")
		fmt.Println("Please set it before running the application.")
		fmt.Println()
		fmt.Println("Example:")
		if runtime.GOOS == "windows" {
			fmt.Println("  set GROQ_API_KEY=your_api_key_here")
		} else {
			fmt.Println("  export GROQ_API_KEY=your_api_key_here")
		}
		fmt.Println()
		fmt.Println("Get your API key from: https://console.groq.com/keys")
		os.Exit(1)
	}

	// Set up signal handling for cleanup
	setupCleanupHandlers()

	// Create client
	client := NewClient(config)

	// Check if server is running, if not start it automatically
	if !isServerRunning(config.ServerURL) {
		fmt.Println("🔄 Server not running, starting automatically...")
		
		// Start server in background and get the actual port
		actualPort, serverCmd, err := startServerInBackgroundWithPort()
		if err != nil {
			fmt.Printf("❌ Failed to start server: %v\n", err)
			fmt.Println("💡 Try starting the server manually with: painika server")
			os.Exit(1)
		}

		// Store server process globally for cleanup
		globalServerCmd = serverCmd

		// Update config to use actual server port
		config.ServerURL = fmt.Sprintf("http://localhost:%d", actualPort)

		// Wait for server to be ready
		fmt.Print("⏳ Waiting for server to start")
		for i := 0; i < 30; i++ { // Wait up to 15 seconds
			if isServerRunning(config.ServerURL) {
				fmt.Println(" ✅")
				break
			}
			fmt.Print(".")
			time.Sleep(500 * time.Millisecond)
			
			if i == 29 {
				fmt.Println(" ❌")
				fmt.Println("❌ Server failed to start within 15 seconds")
				if serverCmd != nil && serverCmd.Process != nil {
					serverCmd.Process.Kill()
				}
				os.Exit(1)
			}
		}
	}

	// Initialize session
	fmt.Println("🚀 Initializing AI session...")
	if err := client.InitSession(); err != nil {
		log.Fatalf("❌ Failed to initialize session: %v", err)
	}

	// Welcome message
	fmt.Println("🤖 Code Agent initialized successfully!")
	fmt.Printf("   Model: %s\n", config.Model)
	fmt.Printf("   Server: %s\n", config.ServerURL)
	fmt.Println()
	fmt.Println("💡 Type 'help' for commands, 'quit' to exit")
	fmt.Println("📝 Start chatting with the AI...")
	fmt.Println()

	// Interactive loop
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("💬 > ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		// Handle special commands
		switch strings.ToLower(input) {
		case "quit", "exit", "q":
			fmt.Println("👋 Goodbye!")
			cleanupAndExit()
			return
		case "help", "h":
			printHelp()
		case "tokens", "t":
			showTokenUsage(client)
		case "history", "hist":
			showConversationHistory(client)
		case "clear", "c":
			clearScreen()
		case "reset", "r":
			resetConversation(client)
		default:
			// Send message to AI
			handleMessage(client, input)
		}
	}
}

func startServerInBackground() (*exec.Cmd, error) {
	// Create a temporary file for the server bundle
	tempFile, err := ioutil.TempFile("", "server-*.js")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	tempFileName := tempFile.Name()

	// Write the embedded server bundle to the temporary file
	if _, err := tempFile.WriteString(serverBundle); err != nil {
		tempFile.Close()
		os.Remove(tempFileName)
		return nil, fmt.Errorf("failed to write server bundle: %v", err)
	}
	tempFile.Close()

	// Start the Bun server in background
	cmd := exec.Command("bun", "run", tempFileName)
	cmd.Env = os.Environ()

	// Start the process without waiting
	if err := cmd.Start(); err != nil {
		os.Remove(tempFileName)
		return nil, fmt.Errorf("failed to start server: %v", err)
	}

	// Clean up temp file when process exits (in a goroutine)
	go func() {
		cmd.Wait() // Wait for process to finish
		os.Remove(tempFileName)
	}()

	return cmd, nil
}

func startServerInBackgroundWithPort() (int, *exec.Cmd, error) {
	// Create a temporary file for the server bundle
	tempFile, err := ioutil.TempFile("", "server-*.js")
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	tempFileName := tempFile.Name()

	// Write the embedded server bundle to the temporary file
	if _, err := tempFile.WriteString(serverBundle); err != nil {
		tempFile.Close()
		os.Remove(tempFileName)
		return 0, nil, fmt.Errorf("failed to write server bundle: %v", err)
	}
	tempFile.Close()

	// Start the Bun server in background and capture output
	cmd := exec.Command("bun", "run", tempFileName)
	cmd.Env = os.Environ()
	
	// Capture stdout to parse the port
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		os.Remove(tempFileName)
		return 0, nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		os.Remove(tempFileName)
		return 0, nil, fmt.Errorf("failed to start server: %v", err)
	}

	// Read server output to get the actual port
	portChan := make(chan int, 1)
	errorChan := make(chan error, 1)
	
	go func() {
		defer stdout.Close()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			// Look for line like "🚀 Code Agent server starting on port 3001"
			if strings.Contains(line, "server starting on port") {
				parts := strings.Split(line, "port ")
				if len(parts) >= 2 {
					portStr := strings.TrimSpace(parts[1])
					if port, err := fmt.Sscanf(portStr, "%d", new(int)); err == nil && port == 1 {
						var actualPort int
						fmt.Sscanf(portStr, "%d", &actualPort)
						portChan <- actualPort
						return
					}
				}
			}
		}
		errorChan <- fmt.Errorf("could not parse server port from output")
	}()

	// Wait for port or timeout
	select {
	case port := <-portChan:
		// Clean up temp file when process exits (in a goroutine)
		go func() {
			cmd.Wait() // Wait for process to finish
			os.Remove(tempFileName)
		}()
		return port, cmd, nil
	case err := <-errorChan:
		cmd.Process.Kill()
		os.Remove(tempFileName)
		return 0, nil, err
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		os.Remove(tempFileName)
		return 0, nil, fmt.Errorf("timeout waiting for server to start")
	}
}

func isServerRunning(serverURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(serverURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// Setup signal handlers for graceful cleanup
func setupCleanupHandlers() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		fmt.Println("\n🛑 Received interrupt signal, cleaning up...")
		cleanupAndExit()
	}()
}

// Cleanup server and exit
func cleanupAndExit() {
	if globalServerCmd != nil && globalServerCmd.Process != nil {
		fmt.Println("🧹 Stopping server...")
		globalServerCmd.Process.Kill()
		globalServerCmd.Wait() // Wait for process to finish
		fmt.Println("✅ Server stopped")
	}
	os.Exit(0)
}

// Handle regular chat message
func handleMessage(client *Client, input string) {
	fmt.Print("🤖 ")

	// Show thinking indicator
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Print(".")
			}
		}
	}()

	// Send message
	response, err := client.SendMessage(input)
	done <- true

	if err != nil {
		fmt.Printf("\n❌ Error: %v\n\n", err)
		return
	}

	// Clear thinking dots and show response
	if len(response.Messages) > 0 {
		fmt.Printf("\r🤖 %s\n", response.Messages[len(response.Messages)-1].Content)
	} else {
		fmt.Printf("\r🤖 No response received\n")
	}
	fmt.Println()
}

// Show help information
func printHelp() {
	fmt.Println("📖 Available Commands:")
	fmt.Println("  help, h      - Show this help message")
	fmt.Println("  tokens, t    - Show token usage statistics")
	fmt.Println("  history, hist - Show conversation history")
	fmt.Println("  clear, c     - Clear the screen")
	fmt.Println("  reset, r     - Reset conversation history")
	fmt.Println("  quit, q      - Exit the application")
	fmt.Println()
	fmt.Println("🔧 Available AI Tools:")
	fmt.Println("  • bash         - Execute shell commands")
	fmt.Println("  • read_file    - Read file contents")
	fmt.Println("  • write_file   - Create/modify files")
	fmt.Println("  • list_files   - List directory contents")
	fmt.Println()
	fmt.Println("💡 The AI will automatically use tools when needed!")
	fmt.Println()
}

// Show token usage statistics
func showTokenUsage(client *Client) {
	usage, err := client.GetTokenUsage()
	if err != nil {
		fmt.Printf("❌ Error getting token usage: %v\n", err)
		return
	}

	fmt.Printf("📊 Token Usage Statistics:\n")
	fmt.Printf("   Input tokens:  %d\n", usage.Input)
	fmt.Printf("   Output tokens: %d\n", usage.Output)
	fmt.Printf("   Total tokens:  %d\n", usage.Total)

	// Rough cost estimation (approximate)
	estimatedCost := float64(usage.Total) * 0.00027 / 1000 // Rough estimate for Groq
	fmt.Printf("   Estimated cost: $%.4f\n", estimatedCost)
	fmt.Println()
}

// Show conversation history
func showConversationHistory(client *Client) {
	conversation, err := client.GetConversation()
	if err != nil {
		fmt.Printf("❌ Error getting conversation: %v\n", err)
		return
	}

	fmt.Printf("📚 Conversation History (%d messages):\n", len(conversation.Messages))

	if len(conversation.Messages) == 0 {
		fmt.Println("   No messages yet. Start chatting!")
		fmt.Println()
		return
	}

	for i, msg := range conversation.Messages {
		icon := "💬"
		if msg.Role == "assistant" {
			icon = "🤖"
		} else if msg.Role == "tool" {
			icon = "🔧"
		} else if msg.Role == "system" {
			continue // Skip system messages in history
		}

		// Parse timestamp from ISO 8601 format
		parsedTime, err := time.Parse(time.RFC3339, msg.Timestamp)
		var timestamp string
		if err != nil {
			timestamp = "unknown"
		} else {
			timestamp = parsedTime.Format("15:04:05")
		}

		// Truncate long messages
		content := msg.Content
		if len(content) > 100 {
			content = content[:97] + "..."
		}

		fmt.Printf("   %d. %s [%s] %s\n", i+1, icon, timestamp, content)
	}
	fmt.Println()
}

// Clear the screen
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// Reset conversation
func resetConversation(client *Client) {
	err := client.ClearConversation()
	if err != nil {
		fmt.Printf("❌ Error clearing conversation: %v\n", err)
		return
	}

	fmt.Println("🧹 Conversation history cleared!")
	fmt.Println()
}
