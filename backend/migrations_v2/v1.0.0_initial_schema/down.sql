-- DOWN
-- Rollback initial schema migration
-- This removes the enhanced migration tracking system

-- Drop indexes first
DROP INDEX IF EXISTS idx_schema_migrations_v2_status;
DROP INDEX IF EXISTS idx_schema_migrations_v2_applied_at;

-- Drop the enhanced migrations table
DROP TABLE IF EXISTS schema_migrations_v2;
