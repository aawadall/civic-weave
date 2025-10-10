package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"civicweave/backend/config"
	"civicweave/backend/models"
	"civicweave/backend/services"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to database successfully")

	// Initialize services
	embeddingService := services.NewEmbeddingService(cfg.OpenAI.APIKey, cfg.OpenAI.EmbeddingModel)
	skillClaimService := models.NewSkillClaimService(db)
	vectorAggregationService := services.NewVectorAggregationService(db, skillClaimService)

	log.Println("✅ Services initialized")

	// Check if we have OpenAI API key
	if cfg.OpenAI.APIKey == "" {
		log.Fatalf("❌ OpenAI API key is required for backfill script")
	}

	// Run the backfill process
	if err := backfillSkillVectors(db, embeddingService, skillClaimService, vectorAggregationService); err != nil {
		log.Fatalf("❌ Backfill failed: %v", err)
	}

	log.Println("🎉 Backfill completed successfully!")
}

func backfillSkillVectors(
	db *sql.DB,
	embeddingService *services.EmbeddingService,
	skillClaimService *models.SkillClaimService,
	vectorAggregationService *services.VectorAggregationService,
) error {
	// First, check if there are any volunteers with JSONB skills to migrate
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM volunteers WHERE skills IS NOT NULL AND jsonb_array_length(skills) > 0`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count volunteers with skills: %w", err)
	}

	if count == 0 {
		log.Println("ℹ️  No volunteers with JSONB skills found to migrate")
		return nil
	}

	log.Printf("📊 Found %d volunteers with JSONB skills to migrate", count)

	// Get all volunteers with JSONB skills
	query := `
		SELECT id, user_id, name, skills 
		FROM volunteers 
		WHERE skills IS NOT NULL AND jsonb_array_length(skills) > 0
		ORDER BY created_at ASC`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query volunteers: %w", err)
	}
	defer rows.Close()

	var processedCount int
	var errorCount int

	for rows.Next() {
		var volunteerID uuid.UUID
		var userID uuid.UUID
		var name string
		var skillsJSON []byte

		if err := rows.Scan(&volunteerID, &userID, &name, &skillsJSON); err != nil {
			log.Printf("❌ Failed to scan volunteer row: %v", err)
			errorCount++
			continue
		}

		// Parse JSONB skills array
		var skills []string
		if err := json.Unmarshal(skillsJSON, &skills); err != nil {
			log.Printf("❌ Failed to unmarshal skills for volunteer %s: %v", name, err)
			errorCount++
			continue
		}

		if len(skills) == 0 {
			continue
		}

		log.Printf("🔄 Processing volunteer: %s (%d skills)", name, len(skills))

		// Check if volunteer already has skill claims
		existingClaims, err := skillClaimService.GetSkillClaimsByVolunteerID(volunteerID)
		if err != nil {
			log.Printf("❌ Failed to check existing claims for volunteer %s: %v", name, err)
			errorCount++
			continue
		}

		if len(existingClaims) > 0 {
			log.Printf("⚠️  Volunteer %s already has %d skill claims, skipping", name, len(existingClaims))
			continue
		}

		// Process each skill
		var successCount int
		for _, skill := range skills {
			if strings.TrimSpace(skill) == "" {
				continue
			}

			// Generate embedding for the skill
			embedding, err := embeddingService.GenerateEmbedding(strings.TrimSpace(skill))
			if err != nil {
				log.Printf("❌ Failed to generate embedding for skill '%s' (volunteer %s): %v", skill, name, err)
				errorCount++
				continue
			}

			// Create skill claim
			claim := &models.SkillClaim{
				VolunteerID:      volunteerID,
				SkillName:        strings.ToLower(strings.TrimSpace(skill)),
				Embedding:        embedding,
				ProficiencyLevel: 3, // Default proficiency level for backfilled skills
				ClaimWeight:      0.5, // Default initial weight
				IsActive:         true,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}

			if err := skillClaimService.CreateSkillClaim(claim); err != nil {
				log.Printf("❌ Failed to create skill claim for '%s' (volunteer %s): %v", skill, name, err)
				errorCount++
				continue
			}

			successCount++
			log.Printf("✅ Created skill claim: '%s' for volunteer %s", skill, name)
		}

		if successCount > 0 {
			// Aggregate volunteer's skill vector
			if err := vectorAggregationService.AggregateVolunteerSkills(volunteerID); err != nil {
				log.Printf("⚠️  Failed to aggregate skill vector for volunteer %s: %v", name, err)
			} else {
				log.Printf("✅ Aggregated skill vector for volunteer %s", name)
			}
		}

		processedCount++
		log.Printf("✅ Completed volunteer %s (%d/%d skills processed)", name, successCount, len(skills))
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating through volunteers: %w", err)
	}

	log.Printf("📊 Backfill Summary:")
	log.Printf("   - Volunteers processed: %d", processedCount)
	log.Printf("   - Errors encountered: %d", errorCount)

	if errorCount > 0 {
		log.Printf("⚠️  %d errors occurred during backfill. Check logs above for details.", errorCount)
	}

	// Verify the migration
	log.Println("🔍 Verifying migration...")
	
	var totalClaims int
	err = db.QueryRow(`SELECT COUNT(*) FROM skill_claims WHERE created_at >= NOW() - INTERVAL '1 hour'`).Scan(&totalClaims)
	if err != nil {
		log.Printf("⚠️  Failed to verify migration: %v", err)
	} else {
		log.Printf("✅ Created %d new skill claims in the last hour", totalClaims)
	}

	var totalVectors int
	err = db.QueryRow(`SELECT COUNT(*) FROM volunteer_skill_vectors WHERE last_updated >= NOW() - INTERVAL '1 hour'`).Scan(&totalVectors)
	if err != nil {
		log.Printf("⚠️  Failed to verify vectors: %v", err)
	} else {
		log.Printf("✅ Created/updated %d volunteer skill vectors in the last hour", totalVectors)
	}

	return nil
}

// Helper function to check if a volunteer has existing skill claims
func volunteerHasSkillClaims(db *sql.DB, volunteerID uuid.UUID) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM skill_claims WHERE volunteer_id = $1`, volunteerID).Scan(&count)
	return count > 0, err
}

// Helper function to clean up old JSONB skills after successful migration
func cleanupOldSkills(db *sql.DB, volunteerID uuid.UUID) error {
	query := `UPDATE volunteers SET skills = NULL WHERE id = $1`
	_, err := db.Exec(query, volunteerID)
	return err
}
