package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Connect to database
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Starting migration to skill taxonomy system...")

	// 1. Migrate volunteer skills from JSON array to taxonomy
	if err := migrateVolunteerSkills(db); err != nil {
		log.Fatal("Failed to migrate volunteer skills:", err)
	}

	// 2. Migrate initiative required skills from JSON array to taxonomy
	if err := migrateInitiativeSkills(db); err != nil {
		log.Fatal("Failed to migrate initiative skills:", err)
	}

	log.Println("Migration completed successfully!")
}

func migrateVolunteerSkills(db *sql.DB) error {
	log.Println("Migrating volunteer skills...")

	// Get all volunteers with their skills JSON
	rows, err := db.Query(`
		SELECT id, skills 
		FROM volunteers 
		WHERE skills IS NOT NULL AND skills != '[]'
	`)
	if err != nil {
		return fmt.Errorf("failed to query volunteers: %w", err)
	}
	defer rows.Close()

	var migratedCount int
	for rows.Next() {
		var volunteerID uuid.UUID
		var skillsJSON []byte

		err := rows.Scan(&volunteerID, &skillsJSON)
		if err != nil {
			return fmt.Errorf("failed to scan volunteer: %w", err)
		}

		// Parse skills JSON array
		var skills []string
		if err := json.Unmarshal(skillsJSON, &skills); err != nil {
			log.Printf("Warning: Failed to parse skills for volunteer %s: %v", volunteerID, err)
			continue
		}

		// Add each skill to taxonomy and volunteer_skills
		for _, skillName := range skills {
			if strings.TrimSpace(skillName) == "" {
				continue
			}

			// Add skill to taxonomy (if not exists)
			var skillID int
			err = db.QueryRow(`
				INSERT INTO skill_taxonomy (skill_name)
				VALUES ($1)
				ON CONFLICT (skill_name) DO UPDATE SET skill_name = EXCLUDED.skill_name
				RETURNING id
			`, skillName).Scan(&skillID)
			if err != nil {
				log.Printf("Warning: Failed to add skill '%s' to taxonomy: %v", skillName, err)
				continue
			}

			// Add volunteer skill with default weight 0.5
			_, err = db.Exec(`
				INSERT INTO volunteer_skills (volunteer_id, skill_id, skill_weight)
				VALUES ($1, $2, $3)
				ON CONFLICT (volunteer_id, skill_id) DO NOTHING
			`, volunteerID, skillID, 0.5)
			if err != nil {
				log.Printf("Warning: Failed to add volunteer skill: %v", err)
				continue
			}
		}

		migratedCount++
		if migratedCount%100 == 0 {
			log.Printf("Migrated %d volunteers...", migratedCount)
		}
	}

	log.Printf("Successfully migrated %d volunteers with skills", migratedCount)
	return nil
}

func migrateInitiativeSkills(db *sql.DB) error {
	log.Println("Migrating initiative required skills...")

	// Get all initiatives with their required_skills JSON
	rows, err := db.Query(`
		SELECT id, required_skills 
		FROM initiatives 
		WHERE required_skills IS NOT NULL AND required_skills != '[]'
	`)
	if err != nil {
		return fmt.Errorf("failed to query initiatives: %w", err)
	}
	defer rows.Close()

	var migratedCount int
	for rows.Next() {
		var initiativeID uuid.UUID
		var skillsJSON []byte

		err := rows.Scan(&initiativeID, &skillsJSON)
		if err != nil {
			return fmt.Errorf("failed to scan initiative: %w", err)
		}

		// Parse skills JSON array
		var skills []string
		if err := json.Unmarshal(skillsJSON, &skills); err != nil {
			log.Printf("Warning: Failed to parse required skills for initiative %s: %v", initiativeID, err)
			continue
		}

		// Add each skill to taxonomy and initiative_required_skills
		for _, skillName := range skills {
			if strings.TrimSpace(skillName) == "" {
				continue
			}

			// Add skill to taxonomy (if not exists)
			var skillID int
			err = db.QueryRow(`
				INSERT INTO skill_taxonomy (skill_name)
				VALUES ($1)
				ON CONFLICT (skill_name) DO UPDATE SET skill_name = EXCLUDED.skill_name
				RETURNING id
			`, skillName).Scan(&skillID)
			if err != nil {
				log.Printf("Warning: Failed to add skill '%s' to taxonomy: %v", skillName, err)
				continue
			}

			// Add initiative required skill
			_, err = db.Exec(`
				INSERT INTO initiative_required_skills (initiative_id, skill_id)
				VALUES ($1, $2)
				ON CONFLICT (initiative_id, skill_id) DO NOTHING
			`, initiativeID, skillID)
			if err != nil {
				log.Printf("Warning: Failed to add initiative required skill: %v", err)
				continue
			}
		}

		migratedCount++
		if migratedCount%50 == 0 {
			log.Printf("Migrated %d initiatives...", migratedCount)
		}
	}

	log.Printf("Successfully migrated %d initiatives with required skills", migratedCount)
	return nil
}
