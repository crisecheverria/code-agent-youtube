package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

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

// Get environment variables for configuration
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	// Load configuration from environment variables
	config := Config{
		ServerURL: getEnv("SERVER_URL", "http://localhost:3000"),
		Token:     getEnv("GROQ_API_KEY", ""),
		Model:     getEnv("MODEL", "llama-3.3-70b-versatile"),
	}

	// Validate configuration
	if config.Token == "" {
		fmt.Println("GROQ_API_KEY environment variable is required")
		fmt.Println("Please set it before running the application.")
		os.Exit(1)
	}

	// Create client
	client := NewClient(config)

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
