-- UP
-- Initial database schema for CivicWeave v2 migration system
-- This migration sets up the enhanced migration tracking system

-- Create the enhanced migrations table
CREATE TABLE IF NOT EXISTS schema_migrations_v2 (
    version VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    runtime_version VARCHAR(20),
    execution_time_ms INTEGER,
    status VARCHAR(20) DEFAULT 'applied'
);

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_schema_migrations_v2_applied_at ON schema_migrations_v2(applied_at);
CREATE INDEX IF NOT EXISTS idx_schema_migrations_v2_status ON schema_migrations_v2(status);

-- Insert a record for this initial migration
INSERT INTO schema_migrations_v2 (version, name, checksum, status)
VALUES ('1.0.0', 'initial_schema', 'initial_migration_checksum', 'applied')
ON CONFLICT (version) DO NOTHING;
