package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"civicweave/backend/pkg/metadb"
	pb "civicweave/backend/proto/dbagent"
	"civicweave/backend/services/dbagent"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// Configuration represents the agent configuration
type Configuration struct {
	// Server configuration
	Port        int    `json:"port"`
	Host        string `json:"host"`
	TLSCertFile string `json:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file"`
	EnableTLS   bool   `json:"enable_tls"`

	// Database configuration
	MetaDBHost     string `json:"meta_db_host"`
	MetaDBPort     int    `json:"meta_db_port"`
	MetaDBUser     string `json:"meta_db_user"`
	MetaDBPassword string `json:"meta_db_password"`
	MetaDBName     string `json:"meta_db_name"`
	MetaDBSSLMode  string `json:"meta_db_ssl_mode"`

	// Security configuration
	EnableAuth      bool `json:"enable_auth"`
	EnableRateLimit bool `json:"enable_rate_limit"`
	RateLimitRPS    int  `json:"rate_limit_rps"`

	// Logging configuration
	LogLevel string `json:"log_level"`
}

func main() {
	var (
		configFile = flag.String("config", "", "Configuration file path")
		port       = flag.Int("port", 50051, "Server port")
		host       = flag.String("host", "0.0.0.0", "Server host")
		enableTLS  = flag.Bool("tls", false, "Enable TLS")
		certFile   = flag.String("cert", "", "TLS certificate file")
		keyFile    = flag.String("key", "", "TLS key file")
		envFile    = flag.String("env", ".env", "Environment file path")
		help       = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Load environment variables
	if err := godotenv.Load(*envFile); err != nil {
		log.Printf("No %s file found, using system environment variables", *envFile)
	}

	// Load configuration
	config := loadConfiguration(*configFile, *port, *host, *enableTLS, *certFile, *keyFile)

	// Initialize metadata database
	metaDB, err := initializeMetadataDatabase(config)
	if err != nil {
		log.Fatalf("Failed to initialize metadata database: %v", err)
	}
	defer metaDB.Close()

	// Initialize metadata repository
	metaRepo := metadb.NewRepository(metaDB)

	// Initialize schema if needed
	if err := metaRepo.InitializeSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Initialize audit logger
	auditLogger := dbagent.NewAuditLogger(metaRepo)

	// Initialize agent service
	agentService := dbagent.NewAgentService(metaRepo, auditLogger)

	// Initialize gRPC server
	server := initializeGRPCServer(config, agentService, auditLogger)

	// Start server
	if err := startServer(config, server); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func showHelp() {
	fmt.Println("Database Agent Server")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/db-agent/main.go [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -config string")
	fmt.Println("        Configuration file path")
	fmt.Println("  -port int")
	fmt.Println("        Server port (default 50051)")
	fmt.Println("  -host string")
	fmt.Println("        Server host (default \"0.0.0.0\")")
	fmt.Println("  -tls")
	fmt.Println("        Enable TLS encryption")
	fmt.Println("  -cert string")
	fmt.Println("        TLS certificate file")
	fmt.Println("  -key string")
	fmt.Println("        TLS key file")
	fmt.Println("  -env string")
	fmt.Println("        Environment file path (default \".env\")")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  METADB_HOST     - Metadata database host")
	fmt.Println("  METADB_PORT     - Metadata database port")
	fmt.Println("  METADB_USER     - Metadata database user")
	fmt.Println("  METADB_PASSWORD - Metadata database password")
	fmt.Println("  METADB_NAME     - Metadata database name")
	fmt.Println("  METADB_SSL_MODE - Metadata database SSL mode")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Start server with default settings")
	fmt.Println("  go run cmd/db-agent/main.go")
	fmt.Println("")
	fmt.Println("  # Start server with TLS")
	fmt.Println("  go run cmd/db-agent/main.go -tls -cert server.crt -key server.key")
	fmt.Println("")
	fmt.Println("  # Start server on custom port")
	fmt.Println("  go run cmd/db-agent/main.go -port 8080")
}

func loadConfiguration(configFile string, port int, host string, enableTLS bool, certFile, keyFile string) *Configuration {
	config := &Configuration{
		Port:            port,
		Host:            host,
		EnableTLS:       enableTLS,
		TLSCertFile:     certFile,
		TLSKeyFile:      keyFile,
		EnableAuth:      true,
		EnableRateLimit: true,
		RateLimitRPS:    100,
		LogLevel:        "info",
	}

	// Load from environment variables
	config.MetaDBHost = getEnvOrDefault("METADB_HOST", "localhost")
	config.MetaDBPort = getEnvIntOrDefault("METADB_PORT", 5432)
	config.MetaDBUser = getEnvOrDefault("METADB_USER", "postgres")
	config.MetaDBPassword = getEnvOrDefault("METADB_PASSWORD", "password")
	config.MetaDBName = getEnvOrDefault("METADB_NAME", "db_agent_metadata")
	config.MetaDBSSLMode = getEnvOrDefault("METADB_SSL_MODE", "disable")

	return config
}

func initializeMetadataDatabase(config *Configuration) (*sql.DB, error) {
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.MetaDBHost, config.MetaDBPort, config.MetaDBUser,
		config.MetaDBPassword, config.MetaDBName, config.MetaDBSSLMode)

	// Connect to metadata database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to metadata database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping metadata database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Printf("✅ Connected to metadata database: %s:%d/%s", config.MetaDBHost, config.MetaDBPort, config.MetaDBName)
	return db, nil
}

func initializeGRPCServer(config *Configuration, agentService *dbagent.AgentService, auditLogger *dbagent.AuditLogger) *grpc.Server {
	// Create server options
	var opts []grpc.ServerOption

	// Add TLS credentials if enabled
	if config.EnableTLS {
		if config.TLSCertFile == "" || config.TLSKeyFile == "" {
			log.Fatal("TLS enabled but certificate or key file not provided")
		}

		creds, err := credentials.NewServerTLSFromFile(config.TLSCertFile, config.TLSKeyFile)
		if err != nil {
			log.Fatalf("Failed to load TLS credentials: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Add keepalive options
	opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
		Time:    60 * time.Second,
		Timeout: 5 * time.Second,
	}))

	// Add keepalive enforcement
	opts = append(opts, grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		MinTime:             30 * time.Second,
		PermitWithoutStream: true,
	}))

	// Create gRPC server
	server := grpc.NewServer(opts...)

	// Register services
	pb.RegisterDatabaseAgentServer(server, agentService)

	// Enable reflection for debugging
	reflection.Register(server)

	log.Printf("✅ gRPC server initialized with TLS: %t", config.EnableTLS)
	return server
}

func startServer(config *Configuration, server *grpc.Server) error {
	// Create listener
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Create context for graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Starting Database Agent server on %s", addr)
		if err := server.Serve(listener); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("🛑 Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✅ Server shutdown complete")
	case <-shutdownCtx.Done():
		log.Println("⚠️  Forced shutdown due to timeout")
		server.Stop()
	}

	return nil
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
