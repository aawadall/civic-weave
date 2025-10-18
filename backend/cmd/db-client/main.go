package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"civicweave/backend/pkg/manifest"
	pb "civicweave/backend/proto/dbagent"
	"civicweave/backend/services/dbagent"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ClientConfig represents the client configuration
type ClientConfig struct {
	AgentURL   string `json:"agent_url"`
	ClientID   string `json:"client_id"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	CreatedAt  string `json:"created_at"`
	ExpiresAt  string `json:"expires_at"`
}

// Exit codes for headless mode
const (
	ExitSuccess       = 0
	ExitError         = 1
	ExitDriftDetected = 2
	ExitAuthFailure   = 3
)

func main() {
	var (
		command       = flag.String("command", "", "Command to execute: ping, compare, download, deploy, history, bootstrap")
		agentURL      = flag.String("agent", "", "Agent URL (host:port)")
		manifestPath  = flag.String("manifest", "", "Manifest directory path")
		database      = flag.String("database", "", "Database name")
		output        = flag.String("output", "", "Output path for download command")
		dryRun        = flag.Bool("dry-run", false, "Dry run mode (for deploy command)")
		targetVersion = flag.String("target-version", "", "Target version for deploy command")
		force         = flag.Bool("force", false, "Force deployment (skip safety checks)")
		limit         = flag.Int("limit", 10, "Limit for history command")
		offset        = flag.Int("offset", 0, "Offset for history command")
		includeData   = flag.Bool("include-data", false, "Include data in download/compare commands")
		environment   = flag.String("environment", "production", "Environment for download command")
		configFile    = flag.String("config", "", "Client configuration file path")
		headless      = flag.Bool("headless", false, "Headless mode (machine-readable output)")
		quiet         = flag.Bool("quiet", false, "Quiet mode (suppress output)")
		help          = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Load client configuration
	config, err := loadClientConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load client config: %v", err)
	}

	// Override agent URL if provided
	if *agentURL != "" {
		config.AgentURL = *agentURL
	}

	if config.AgentURL == "" {
		log.Fatal("Agent URL is required (use --agent or set in config file)")
	}

	// Create gRPC connection
	conn, err := createGRPCConnection(config.AgentURL)
	if err != nil {
		log.Fatalf("Failed to connect to agent: %v", err)
	}
	defer conn.Close()

	// Create authenticated client
	client := pb.NewDatabaseAgentClient(conn)
	clientAuth := dbagent.NewClientAuth(config.ClientID, config.PrivateKey)

	// Execute command
	switch *command {
	case "ping":
		err = executePing(client, clientAuth, *headless, *quiet)
	case "compare":
		err = executeCompare(client, clientAuth, *manifestPath, *database, *includeData, *headless, *quiet)
	case "download":
		err = executeDownload(client, clientAuth, *database, *output, *includeData, *environment, *headless, *quiet)
	case "deploy":
		err = executeDeploy(client, clientAuth, *manifestPath, *database, *dryRun, *targetVersion, *force, *headless, *quiet)
	case "history":
		err = executeHistory(client, clientAuth, *database, *limit, *offset, *headless, *quiet)
	case "bootstrap":
		err = executeBootstrap(client, clientAuth, *manifestPath, *database, *headless, *quiet)
	default:
		log.Fatalf("Unknown command: %s", *command)
	}

	if err != nil {
		if *headless {
			os.Exit(ExitError)
		}
		log.Fatalf("Command failed: %v", err)
	}

	if *headless {
		os.Exit(ExitSuccess)
	}
}

func showHelp() {
	fmt.Println("Database Agent Client")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/db-client/main.go -command=<command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  ping      - Health check and connectivity test")
	fmt.Println("  compare   - Compare local manifest to live database")
	fmt.Println("  download  - Extract current schema as manifest")
	fmt.Println("  deploy    - Deploy manifest to database")
	fmt.Println("  history   - Get deployment history")
	fmt.Println("  bootstrap - Initialize new database from scratch")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -command string")
	fmt.Println("        Command to execute (required)")
	fmt.Println("  -agent string")
	fmt.Println("        Agent URL (host:port)")
	fmt.Println("  -manifest string")
	fmt.Println("        Manifest directory path")
	fmt.Println("  -database string")
	fmt.Println("        Database name")
	fmt.Println("  -output string")
	fmt.Println("        Output path for download command")
	fmt.Println("  -dry-run")
	fmt.Println("        Dry run mode (for deploy command)")
	fmt.Println("  -target-version string")
	fmt.Println("        Target version for deploy command")
	fmt.Println("  -force")
	fmt.Println("        Force deployment (skip safety checks)")
	fmt.Println("  -limit int")
	fmt.Println("        Limit for history command (default 10)")
	fmt.Println("  -offset int")
	fmt.Println("        Offset for history command (default 0)")
	fmt.Println("  -include-data")
	fmt.Println("        Include data in download/compare commands")
	fmt.Println("  -environment string")
	fmt.Println("        Environment for download command (default \"production\")")
	fmt.Println("  -config string")
	fmt.Println("        Client configuration file path")
	fmt.Println("  -headless")
	fmt.Println("        Headless mode (machine-readable output)")
	fmt.Println("  -quiet")
	fmt.Println("        Quiet mode (suppress output)")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Test connectivity")
	fmt.Println("  go run cmd/db-client/main.go -command=ping -agent=localhost:50051")
	fmt.Println("")
	fmt.Println("  # Compare manifest to live database")
	fmt.Println("  go run cmd/db-client/main.go -command=compare -manifest=./manifest -database=prod")
	fmt.Println("")
	fmt.Println("  # Deploy manifest (dry run)")
	fmt.Println("  go run cmd/db-client/main.go -command=deploy -manifest=./manifest -database=prod -dry-run")
	fmt.Println("")
	fmt.Println("  # Deploy manifest (actual)")
	fmt.Println("  go run cmd/db-client/main.go -command=deploy -manifest=./manifest -database=prod")
	fmt.Println("")
	fmt.Println("  # Get deployment history")
	fmt.Println("  go run cmd/db-client/main.go -command=history -database=prod -limit=20")
	fmt.Println("")
	fmt.Println("  # Download current schema")
	fmt.Println("  go run cmd/db-client/main.go -command=download -database=prod -output=./manifest")
}

func loadClientConfig(configFile string) (*ClientConfig, error) {
	if configFile == "" {
		configFile = "./keys/client/client-config.json"
	}

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("client config file not found: %s", configFile)
	}

	// Read config file
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config
	var config ClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func createGRPCConnection(agentURL string) (*grpc.ClientConn, error) {
	// Determine if we should use TLS
	var opts []grpc.DialOption
	if isSecureConnection(agentURL) {
		creds := credentials.NewTLS(&tls.Config{})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add keepalive options
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                60 * time.Second,
		Timeout:             5 * time.Second,
		PermitWithoutStream: true,
	}))

	// Connect to agent
	conn, err := grpc.Dial(agentURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent: %w", err)
	}

	return conn, nil
}

func isSecureConnection(agentURL string) bool {
	// Simple check for secure connection
	// In a real implementation, you might want to check the port or use a more sophisticated method
	return false // For now, assume insecure connections
}

func executePing(client pb.DatabaseAgentClient, auth *dbagent.ClientAuth, headless, quiet bool) error {
	ctx := auth.WithAuth(context.Background())

	req := &pb.PingRequest{
		ClientVersion: "1.0.0",
	}

	resp, err := client.Ping(ctx, req)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	if headless {
		// JSON output for headless mode
		output := map[string]interface{}{
			"status":        resp.Status,
			"agent_version": resp.AgentVersion,
			"timestamp":     resp.Timestamp,
		}
		jsonOutput, _ := json.Marshal(output)
		fmt.Println(string(jsonOutput))
	} else if !quiet {
		fmt.Printf("✅ Ping successful\n")
		fmt.Printf("📊 Agent Version: %s\n", resp.AgentVersion)
		fmt.Printf("📊 Status: %s\n", resp.Status)
		fmt.Printf("📊 Timestamp: %d\n", resp.Timestamp)
	}

	return nil
}

func executeCompare(client pb.DatabaseAgentClient, auth *dbagent.ClientAuth, manifestPath, database string, includeData, headless, quiet bool) error {
	if manifestPath == "" {
		return fmt.Errorf("manifest path is required")
	}
	if database == "" {
		return fmt.Errorf("database name is required")
	}

	// Parse manifest
	parser := manifest.NewParser(manifestPath)
	manifestData, err := parser.ParseManifest()
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	ctx := auth.WithAuth(context.Background())

	req := &pb.CompareManifestRequest{
		DatabaseName:    database,
		Manifest:        manifestData,
		IncludeDataDiff: includeData,
	}

	resp, err := client.CompareManifest(ctx, req)
	if err != nil {
		return fmt.Errorf("compare failed: %w", err)
	}

	if headless {
		// JSON output for headless mode
		output := map[string]interface{}{
			"is_identical":    resp.IsIdentical,
			"differences":     resp.Differences,
			"missing_objects": resp.MissingObjects,
			"extra_objects":   resp.ExtraObjects,
			"local_checksum":  resp.LocalChecksum,
			"remote_checksum": resp.RemoteChecksum,
		}
		jsonOutput, _ := json.Marshal(output)
		fmt.Println(string(jsonOutput))
	} else if !quiet {
		if resp.IsIdentical {
			fmt.Printf("✅ Database schema is identical to manifest\n")
		} else {
			fmt.Printf("⚠️  Database schema differs from manifest\n")
			if len(resp.Differences) > 0 {
				fmt.Printf("📋 Differences:\n")
				for _, diff := range resp.Differences {
					fmt.Printf("  • %s\n", diff)
				}
			}
			if len(resp.MissingObjects) > 0 {
				fmt.Printf("📋 Missing Objects:\n")
				for _, obj := range resp.MissingObjects {
					fmt.Printf("  • %s\n", obj)
				}
			}
			if len(resp.ExtraObjects) > 0 {
				fmt.Printf("📋 Extra Objects:\n")
				for _, obj := range resp.ExtraObjects {
					fmt.Printf("  • %s\n", obj)
				}
			}
		}
	}

	return nil
}

func executeDownload(client pb.DatabaseAgentClient, auth *dbagent.ClientAuth, database, output string, includeData bool, environment string, headless, quiet bool) error {
	if database == "" {
		return fmt.Errorf("database name is required")
	}

	ctx := auth.WithAuth(context.Background())

	req := &pb.DownloadManifestRequest{
		DatabaseName: database,
		IncludeData:  includeData,
		Environment:  environment,
	}

	resp, err := client.DownloadManifest(ctx, req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	if headless {
		// JSON output for headless mode
		output := map[string]interface{}{
			"checksum":      resp.Checksum,
			"objects_count": resp.ObjectsCount,
		}
		jsonOutput, _ := json.Marshal(output)
		fmt.Println(string(jsonOutput))
	} else if !quiet {
		fmt.Printf("✅ Schema downloaded successfully\n")
		fmt.Printf("📊 Checksum: %s\n", resp.Checksum)
		fmt.Printf("📊 Objects Count: %d\n", resp.ObjectsCount)
	}

	// Save manifest if output path is provided
	if output != "" {
		parser := manifest.NewParser(output)
		if err := parser.WriteManifest(resp.Manifest); err != nil {
			return fmt.Errorf("failed to save manifest: %w", err)
		}
		if !quiet {
			fmt.Printf("📁 Manifest saved to: %s\n", output)
		}
	}

	return nil
}

func executeDeploy(client pb.DatabaseAgentClient, auth *dbagent.ClientAuth, manifestPath, database string, dryRun bool, targetVersion string, force bool, headless, quiet bool) error {
	if manifestPath == "" {
		return fmt.Errorf("manifest path is required")
	}
	if database == "" {
		return fmt.Errorf("database name is required")
	}

	// Parse manifest
	parser := manifest.NewParser(manifestPath)
	manifestData, err := parser.ParseManifest()
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	ctx := auth.WithAuth(context.Background())

	req := &pb.DeployManifestRequest{
		DatabaseName:  database,
		Manifest:      manifestData,
		DryRun:        dryRun,
		TargetVersion: targetVersion,
		Force:         force,
	}

	resp, err := client.DeployManifest(ctx, req)
	if err != nil {
		return fmt.Errorf("deploy failed: %w", err)
	}

	if headless {
		// JSON output for headless mode
		output := map[string]interface{}{
			"success":           resp.Success,
			"status":            resp.Status,
			"execution_time_ms": resp.ExecutionTimeMs,
			"migrations":        len(resp.Migrations),
			"warnings":          resp.Warnings,
			"errors":            resp.Errors,
		}
		jsonOutput, _ := json.Marshal(output)
		fmt.Println(string(jsonOutput))
	} else if !quiet {
		if resp.Success {
			if dryRun {
				fmt.Printf("✅ Dry run completed successfully\n")
			} else {
				fmt.Printf("✅ Deployment completed successfully\n")
			}
			fmt.Printf("📊 Status: %s\n", resp.Status)
			fmt.Printf("📊 Execution Time: %d ms\n", resp.ExecutionTimeMs)
			fmt.Printf("📊 Migrations: %d\n", len(resp.Migrations))

			if len(resp.Warnings) > 0 {
				fmt.Printf("⚠️  Warnings:\n")
				for _, warning := range resp.Warnings {
					fmt.Printf("  • %s\n", warning)
				}
			}
		} else {
			fmt.Printf("❌ Deployment failed\n")
			if len(resp.Errors) > 0 {
				fmt.Printf("📋 Errors:\n")
				for _, err := range resp.Errors {
					fmt.Printf("  • %s\n", err)
				}
			}
		}
	}

	return nil
}

func executeHistory(client pb.DatabaseAgentClient, auth *dbagent.ClientAuth, database string, limit, offset int, headless, quiet bool) error {
	if database == "" {
		return fmt.Errorf("database name is required")
	}

	ctx := auth.WithAuth(context.Background())

	req := &pb.DeploymentHistoryRequest{
		DatabaseName: database,
		Limit:        int32(limit),
		Offset:       int32(offset),
	}

	resp, err := client.GetDeploymentHistory(ctx, req)
	if err != nil {
		return fmt.Errorf("history failed: %w", err)
	}

	if headless {
		// JSON output for headless mode
		output := map[string]interface{}{
			"deployments": resp.Deployments,
			"total_count": resp.TotalCount,
			"has_more":    resp.HasMore,
		}
		jsonOutput, _ := json.Marshal(output)
		fmt.Println(string(jsonOutput))
	} else if !quiet {
		fmt.Printf("📊 Deployment History for %s\n", database)
		fmt.Printf("📊 Total Count: %d\n", resp.TotalCount)
		fmt.Printf("📊 Has More: %t\n", resp.HasMore)

		if len(resp.Deployments) > 0 {
			fmt.Printf("\n📋 Deployments:\n")
			for _, dep := range resp.Deployments {
				fmt.Printf("  • %s - %s (%s) - %d ms\n",
					dep.Version, dep.Status, dep.AppliedBy, dep.ExecutionTimeMs)
			}
		}
	}

	return nil
}

func executeBootstrap(client pb.DatabaseAgentClient, auth *dbagent.ClientAuth, manifestPath, database string, headless, quiet bool) error {
	if manifestPath == "" {
		return fmt.Errorf("manifest path is required")
	}
	if database == "" {
		return fmt.Errorf("database name is required")
	}

	// Parse manifest
	parser := manifest.NewParser(manifestPath)
	manifestData, err := parser.ParseManifest()
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	ctx := auth.WithAuth(context.Background())

	req := &pb.BootstrapRequest{
		DatabaseName:     database,
		ConnectionString: "placeholder_connection_string", // This should be provided by the client
		Manifest:         manifestData,
		CreateDatabase:   true,
	}

	resp, err := client.Bootstrap(ctx, req)
	if err != nil {
		return fmt.Errorf("bootstrap failed: %w", err)
	}

	if headless {
		// JSON output for headless mode
		output := map[string]interface{}{
			"success":           resp.Success,
			"database_name":     resp.DatabaseName,
			"execution_time_ms": resp.ExecutionTimeMs,
			"migrations":        len(resp.Migrations),
			"warnings":          resp.Warnings,
			"errors":            resp.Errors,
		}
		jsonOutput, _ := json.Marshal(output)
		fmt.Println(string(jsonOutput))
	} else if !quiet {
		if resp.Success {
			fmt.Printf("✅ Bootstrap completed successfully\n")
			fmt.Printf("📊 Database: %s\n", resp.DatabaseName)
			fmt.Printf("📊 Execution Time: %d ms\n", resp.ExecutionTimeMs)
			fmt.Printf("📊 Migrations: %d\n", len(resp.Migrations))

			if len(resp.Warnings) > 0 {
				fmt.Printf("⚠️  Warnings:\n")
				for _, warning := range resp.Warnings {
					fmt.Printf("  • %s\n", warning)
				}
			}
		} else {
			fmt.Printf("❌ Bootstrap failed\n")
			if len(resp.Errors) > 0 {
				fmt.Printf("📋 Errors:\n")
				for _, err := range resp.Errors {
					fmt.Printf("  • %s\n", err)
				}
			}
		}
	}

	return nil
}
