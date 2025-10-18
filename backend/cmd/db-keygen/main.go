package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

// APIKey represents a generated API key pair
type APIKey struct {
	ID          string    `json:"id"`
	PublicKey   string    `json:"public_key"`
	PrivateKey  string    `json:"private_key"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Permissions []string  `json:"permissions"`
}

// ServerKeyConfig represents server-side key configuration
type ServerKeyConfig struct {
	Keys []ServerKey `json:"keys"`
}

// ServerKey represents a server-side key entry
type ServerKey struct {
	ID          string    `json:"id"`
	PublicKey   string    `json:"public_key"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Permissions []string  `json:"permissions"`
	IsActive    bool      `json:"is_active"`
}

// ClientKeyConfig represents client-side key configuration
type ClientKeyConfig struct {
	AgentURL   string    `json:"agent_url"`
	ClientID   string    `json:"client_id"`
	PrivateKey string    `json:"private_key"`
	PublicKey  string    `json:"public_key"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func main() {
	var (
		description = flag.String("description", "Generated API Key", "Description for the API key")
		duration    = flag.String("duration", "1y", "Key validity duration (e.g., 1y, 6m, 30d)")
		permissions = flag.String("permissions", "read,write,deploy,bootstrap", "Comma-separated permissions")
		output      = flag.String("output", "", "Output directory for key files")
		serverMode  = flag.Bool("server", false, "Generate server-side configuration")
		clientMode  = flag.Bool("client", false, "Generate client-side configuration")
		agentURL    = flag.String("agent-url", "", "Agent URL for client configuration")
		list        = flag.Bool("list", false, "List existing keys")
		revoke      = flag.String("revoke", "", "Revoke key by ID")
		help        = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *list {
		listKeys(*output)
		return
	}

	if *revoke != "" {
		revokeKey(*revoke, *output)
		return
	}

	// Parse duration
	durationParsed, err := parseDuration(*duration)
	if err != nil {
		log.Fatalf("Invalid duration: %v", err)
	}

	// Parse permissions
	permissionsList := parsePermissions(*permissions)

	// Generate key pair
	apiKey, err := generateAPIKey(*description, durationParsed, permissionsList)
	if err != nil {
		log.Fatalf("Failed to generate API key: %v", err)
	}

	// Determine output directory
	outputDir := *output
	if outputDir == "" {
		if *serverMode {
			outputDir = "./keys/server"
		} else if *clientMode {
			outputDir = "./keys/client"
		} else {
			outputDir = "./keys"
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	if *serverMode {
		err = saveServerConfig(apiKey, outputDir)
	} else if *clientMode {
		if *agentURL == "" {
			log.Fatal("Agent URL is required for client mode")
		}
		err = saveClientConfig(apiKey, *agentURL, outputDir)
	} else {
		err = saveKeyPair(apiKey, outputDir)
	}

	if err != nil {
		log.Fatalf("Failed to save key configuration: %v", err)
	}

	fmt.Printf("✅ API key generated successfully!\n")
	fmt.Printf("📝 ID: %s\n", apiKey.ID)
	fmt.Printf("📁 Output directory: %s\n", outputDir)
	fmt.Printf("⏰ Expires: %s\n", apiKey.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("🔑 Permissions: %v\n", apiKey.Permissions)
}

func showHelp() {
	fmt.Println("Database Agent API Key Generator")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/db-keygen/main.go [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -description string")
	fmt.Println("        Description for the API key (default \"Generated API Key\")")
	fmt.Println("  -duration string")
	fmt.Println("        Key validity duration (e.g., 1y, 6m, 30d) (default \"1y\")")
	fmt.Println("  -permissions string")
	fmt.Println("        Comma-separated permissions (default \"read,write,deploy,bootstrap\")")
	fmt.Println("  -output string")
	fmt.Println("        Output directory for key files")
	fmt.Println("  -server")
	fmt.Println("        Generate server-side configuration")
	fmt.Println("  -client")
	fmt.Println("        Generate client-side configuration")
	fmt.Println("  -agent-url string")
	fmt.Println("        Agent URL for client configuration")
	fmt.Println("  -list")
	fmt.Println("        List existing keys")
	fmt.Println("  -revoke string")
	fmt.Println("        Revoke key by ID")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Generate a server key")
	fmt.Println("  go run cmd/db-keygen/main.go -server -description \"Production Server\"")
	fmt.Println("")
	fmt.Println("  # Generate a client key")
	fmt.Println("  go run cmd/db-keygen/main.go -client -agent-url \"localhost:50051\" -description \"CI/CD Client\"")
	fmt.Println("")
	fmt.Println("  # List existing keys")
	fmt.Println("  go run cmd/db-keygen/main.go -list")
	fmt.Println("")
	fmt.Println("  # Revoke a key")
	fmt.Println("  go run cmd/db-keygen/main.go -revoke \"key-id\"")
}

func generateAPIKey(description string, duration time.Duration, permissions []string) (*APIKey, error) {
	// Generate random key material
	keyMaterial := make([]byte, 32)
	if _, err := rand.Read(keyMaterial); err != nil {
		return nil, fmt.Errorf("failed to generate random key material: %w", err)
	}

	// Generate public key (hash of key material)
	publicKeyHash := sha256.Sum256(keyMaterial)
	publicKey := base64.StdEncoding.EncodeToString(publicKeyHash[:])

	// Generate private key (argon2 hash with salt)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	privateKeyHash := argon2.IDKey(keyMaterial, salt, 1, 64*1024, 4, 32)
	privateKey := base64.StdEncoding.EncodeToString(privateKeyHash)

	now := time.Now()
	return &APIKey{
		ID:          uuid.New().String(),
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
		Description: description,
		CreatedAt:   now,
		ExpiresAt:   now.Add(duration),
		Permissions: permissions,
	}, nil
}

func parseDuration(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return 365 * 24 * time.Hour, nil // Default to 1 year
	}

	// Parse duration string (e.g., "1y", "6m", "30d", "24h")
	var duration time.Duration
	var err error

	switch {
	case len(durationStr) > 0 && durationStr[len(durationStr)-1] == 'y':
		var years int
		if _, err := fmt.Sscanf(durationStr, "%dy", &years); err != nil {
			return 0, fmt.Errorf("invalid year format: %s", durationStr)
		}
		duration = time.Duration(years) * 365 * 24 * time.Hour
	case len(durationStr) > 0 && durationStr[len(durationStr)-1] == 'm':
		var months int
		if _, err := fmt.Sscanf(durationStr, "%dm", &months); err != nil {
			return 0, fmt.Errorf("invalid month format: %s", durationStr)
		}
		duration = time.Duration(months) * 30 * 24 * time.Hour
	case len(durationStr) > 0 && durationStr[len(durationStr)-1] == 'd':
		var days int
		if _, err := fmt.Sscanf(durationStr, "%dd", &days); err != nil {
			return 0, fmt.Errorf("invalid day format: %s", durationStr)
		}
		duration = time.Duration(days) * 24 * time.Hour
	case len(durationStr) > 0 && durationStr[len(durationStr)-1] == 'h':
		duration, err = time.ParseDuration(durationStr)
		if err != nil {
			return 0, fmt.Errorf("invalid hour format: %s", durationStr)
		}
	default:
		duration, err = time.ParseDuration(durationStr)
		if err != nil {
			return 0, fmt.Errorf("invalid duration format: %s", durationStr)
		}
	}

	return duration, nil
}

func parsePermissions(permissionsStr string) []string {
	if permissionsStr == "" {
		return []string{"read", "write", "deploy", "bootstrap"}
	}

	var permissions []string
	for _, perm := range splitAndTrim(permissionsStr, ",") {
		if perm != "" {
			permissions = append(permissions, perm)
		}
	}
	return permissions
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, item := range split(s, sep) {
		if trimmed := trim(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func split(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)

	// Trim leading whitespace
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	// Trim trailing whitespace
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

func saveKeyPair(apiKey *APIKey, outputDir string) error {
	// Save full key pair
	fullKeyFile := filepath.Join(outputDir, fmt.Sprintf("key-%s.json", apiKey.ID))
	fullKeyData, err := json.MarshalIndent(apiKey, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal key data: %w", err)
	}

	if err := os.WriteFile(fullKeyFile, fullKeyData, 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	// Save public key only
	publicKeyFile := filepath.Join(outputDir, fmt.Sprintf("public-%s.key", apiKey.ID))
	if err := os.WriteFile(publicKeyFile, []byte(apiKey.PublicKey), 0644); err != nil {
		return fmt.Errorf("failed to write public key file: %w", err)
	}

	// Save private key only
	privateKeyFile := filepath.Join(outputDir, fmt.Sprintf("private-%s.key", apiKey.ID))
	if err := os.WriteFile(privateKeyFile, []byte(apiKey.PrivateKey), 0600); err != nil {
		return fmt.Errorf("failed to write private key file: %w", err)
	}

	return nil
}

func saveServerConfig(apiKey *APIKey, outputDir string) error {
	// Load existing server config or create new
	configFile := filepath.Join(outputDir, "server-keys.json")
	config := &ServerKeyConfig{Keys: []ServerKey{}}

	if data, err := os.ReadFile(configFile); err == nil {
		if err := json.Unmarshal(data, config); err != nil {
			log.Printf("Warning: Failed to parse existing config, creating new one: %v", err)
		}
	}

	// Add new key to config
	serverKey := ServerKey{
		ID:          apiKey.ID,
		PublicKey:   apiKey.PublicKey,
		Description: apiKey.Description,
		CreatedAt:   apiKey.CreatedAt,
		ExpiresAt:   apiKey.ExpiresAt,
		Permissions: apiKey.Permissions,
		IsActive:    true,
	}
	config.Keys = append(config.Keys, serverKey)

	// Save updated config
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal server config: %w", err)
	}

	if err := os.WriteFile(configFile, configData, 0644); err != nil {
		return fmt.Errorf("failed to write server config: %w", err)
	}

	// Also save individual key file
	keyFile := filepath.Join(outputDir, fmt.Sprintf("server-key-%s.json", apiKey.ID))
	keyData, err := json.MarshalIndent(serverKey, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal server key: %w", err)
	}

	if err := os.WriteFile(keyFile, keyData, 0644); err != nil {
		return fmt.Errorf("failed to write server key file: %w", err)
	}

	return nil
}

func saveClientConfig(apiKey *APIKey, agentURL, outputDir string) error {
	config := &ClientKeyConfig{
		AgentURL:   agentURL,
		ClientID:   apiKey.ID,
		PrivateKey: apiKey.PrivateKey,
		PublicKey:  apiKey.PublicKey,
		CreatedAt:  apiKey.CreatedAt,
		ExpiresAt:  apiKey.ExpiresAt,
	}

	// Save client config
	configFile := filepath.Join(outputDir, "client-config.json")
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client config: %w", err)
	}

	if err := os.WriteFile(configFile, configData, 0600); err != nil {
		return fmt.Errorf("failed to write client config: %w", err)
	}

	return nil
}

func listKeys(outputDir string) {
	if outputDir == "" {
		outputDir = "./keys"
	}

	// List server keys
	serverConfigFile := filepath.Join(outputDir, "server", "server-keys.json")
	if data, err := os.ReadFile(serverConfigFile); err == nil {
		var config ServerKeyConfig
		if err := json.Unmarshal(data, &config); err == nil {
			fmt.Println("🔑 Server Keys:")
			for _, key := range config.Keys {
				status := "✅ Active"
				if !key.IsActive || time.Now().After(key.ExpiresAt) {
					status = "❌ Inactive/Expired"
				}
				fmt.Printf("  %s - %s (%s) - %s\n", key.ID[:8], key.Description, status, key.CreatedAt.Format("2006-01-02"))
			}
		}
	}

	// List client configs
	clientConfigFile := filepath.Join(outputDir, "client", "client-config.json")
	if data, err := os.ReadFile(clientConfigFile); err == nil {
		var config ClientKeyConfig
		if err := json.Unmarshal(data, &config); err == nil {
			fmt.Println("💻 Client Keys:")
			status := "✅ Active"
			if time.Now().After(config.ExpiresAt) {
				status = "❌ Expired"
			}
			fmt.Printf("  %s - %s (%s) - %s\n", config.ClientID[:8], config.AgentURL, status, config.CreatedAt.Format("2006-01-02"))
		}
	}
}

func revokeKey(keyID, outputDir string) {
	if outputDir == "" {
		outputDir = "./keys"
	}

	// Revoke server key
	serverConfigFile := filepath.Join(outputDir, "server", "server-keys.json")
	if data, err := os.ReadFile(serverConfigFile); err == nil {
		var config ServerKeyConfig
		if err := json.Unmarshal(data, &config); err == nil {
			for i, key := range config.Keys {
				if key.ID == keyID {
					config.Keys[i].IsActive = false

					// Save updated config
					configData, err := json.MarshalIndent(config, "", "  ")
					if err != nil {
						log.Printf("Failed to update server config: %v", err)
						return
					}

					if err := os.WriteFile(serverConfigFile, configData, 0644); err != nil {
						log.Printf("Failed to save server config: %v", err)
						return
					}

					fmt.Printf("✅ Key %s revoked successfully\n", keyID)
					return
				}
			}
		}
	}

	fmt.Printf("❌ Key %s not found\n", keyID)
}
