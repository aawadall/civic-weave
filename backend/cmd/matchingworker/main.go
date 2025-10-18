package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"civicweave/backend/config"
	"civicweave/backend/database"
	"civicweave/backend/services"
)

func main() {
	log.Println("🚀 Starting CivicWeave Matching Worker...")

	// Load configuration
	cfg := config.Load()
	log.Printf("📋 Configuration loaded: DB=%s:%s", cfg.Database.Host, cfg.Database.Port)

	// Connect to database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Database connected successfully")

	// Create skill matching service
	matchingService := services.NewSkillMatchingService(db)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run initial calculation
	log.Println("🔄 Running initial match calculation...")
	if err := matchingService.BatchCalculateProjectMatches(); err != nil {
		log.Printf("❌ Initial calculation failed: %v", err)
	} else {
		log.Println("✅ Initial calculation completed successfully")
	}

	// Set up ticker for periodic calculations
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	log.Println("⏰ Starting periodic calculations (every 15 minutes)")

	// Main loop
	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Shutdown signal received, stopping worker...")
			return
		case <-ticker.C:
			log.Println("🔄 Running periodic match calculation...")
			if err := matchingService.BatchCalculateProjectMatches(); err != nil {
				log.Printf("❌ Periodic calculation failed: %v", err)
			} else {
				log.Println("✅ Periodic calculation completed successfully")
			}
		case sig := <-sigChan:
			log.Printf("🛑 Received signal %v, initiating graceful shutdown...", sig)
			cancel()
		}
	}
}
