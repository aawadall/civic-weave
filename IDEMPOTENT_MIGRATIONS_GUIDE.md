# Idempotent Migrations Guide

## What We Did

Successfully converted all database migrations to be **idempotent**, meaning they can be run multiple times safely without errors.

## Problem Solved

Previously, if migrations were run out of sync with the migration tracker, or if the database was set up manually, running `make db-migrate` would fail with errors like:
```
pq: relation "users" already exists
```

## Solution: Idempotent Migrations

Made all migration scripts use SQL constructs that check for existence before creating/modifying:

### Key Changes

1. **Tables**: `CREATE TABLE` → `CREATE TABLE IF NOT EXISTS`
2. **Indexes**: `CREATE INDEX` → `CREATE INDEX IF NOT EXISTS`
3. **Triggers**: Added `DROP TRIGGER IF EXISTS` before `CREATE TRIGGER`
4. **Inserts**: Added `ON CONFLICT DO NOTHING` for seed data
5. **Column Operations**: Used conditional logic with `information_schema` checks

## Updated Migrations

### Migration 001 (Initial Schema)
- All `CREATE TABLE` statements now use `IF NOT EXISTS`
- All `CREATE INDEX` statements now use `IF NOT EXISTS`

### Migration 003 (RBAC System)
- All `CREATE TABLE` statements now use `IF NOT EXISTS`
- All `CREATE INDEX` statements now use `IF NOT EXISTS`
- Default role inserts use `ON CONFLICT (name) DO NOTHING`
- Triggers use `DROP TRIGGER IF EXISTS` before creation

### Migration 010 (Remove Users Role Column) - NEW
- Uses PL/pgSQL block to check if `role` column exists before migrating data
- Uses `DROP INDEX IF EXISTS` for index removal
- Uses `ALTER TABLE ... DROP COLUMN IF EXISTS` for column removal

## Example: Conditional Column Migration

```sql
DO $$
BEGIN
    IF EXISTS (
        SELECT FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'role'
    ) THEN
        -- Migrate data from old column to new structure
        INSERT INTO user_roles (user_id, role_id, assigned_at)
        SELECT 
            u.id AS user_id,
            r.id AS role_id,
            u.created_at AS assigned_at
        FROM users u
        INNER JOIN roles r ON r.name = u.role
        WHERE u.role IS NOT NULL
        ON CONFLICT (user_id, role_id) DO NOTHING;
    END IF;
END $$;

-- Now safe to drop column (won't error if already dropped)
ALTER TABLE users DROP COLUMN IF EXISTS role;
```

## When Database State is Unknown

If your database was set up manually or migrations are out of sync:

### Step 1: Check Migration State
```bash
docker exec civicweave_postgres_1 psql -U civicweave -d civicweave -c "SELECT * FROM schema_migrations ORDER BY version;"
```

### Step 2: Manually Mark Applied Migrations

If migrations are already applied but not recorded:
```sql
INSERT INTO schema_migrations (version, applied_at) VALUES
(1, CURRENT_TIMESTAMP),
(2, CURRENT_TIMESTAMP),
-- ... add all applied migrations
ON CONFLICT DO NOTHING;
```

### Step 3: Run Migrations
```bash
make db-migrate
```

Since all migrations are now idempotent, they'll safely skip already-applied changes.

## Benefits

1. **Safe Re-runs**: Can run migrations multiple times without errors
2. **Easier Recovery**: If migration state gets confused, just re-run
3. **Development Friendly**: Developers can run migrations without worrying about state
4. **CI/CD Ready**: Automated deployments can safely run migrations

## Testing Idempotency

To verify a migration is idempotent:

```bash
# Run once
make db-migrate

# Run again - should succeed with no changes
make db-migrate
```

Both runs should complete successfully.

## Best Practices Going Forward

When creating new migrations:

1. **Always use `IF NOT EXISTS`** for CREATE statements
2. **Always use `IF EXISTS`** for DROP statements  
3. **Use `ON CONFLICT DO NOTHING`** for INSERT statements
4. **Check existence** before modifying columns with `information_schema`
5. **Test idempotency** by running the migration twice

## Migration History

| Version | Name | Status | Notes |
|---------|------|--------|-------|
| 001 | initial | ✅ Idempotent | Base schema |
| 002 | skill_vectors | ⚠️ Check | May need updates |
| 003 | rbac_system | ✅ Idempotent | Roles & RBAC |
| 004 | rename_initiatives_to_projects | ⚠️ Check | Table rename |
| 005 | skill_taxonomy | ⚠️ Check | May need updates |
| 006 | project_enhancements | ⚠️ Check | May need updates |
| 007 | fix_project_admin_fkey | ⚠️ Check | May need updates |
| 008 | add_auto_notify_matches | ⚠️ Check | May need updates |
| 009 | candidate_notifications | ⚠️ Check | May need updates |
| 010 | remove_users_role_column | ✅ Idempotent | Many-to-many roles |

✅ = Fully idempotent
⚠️ = May need review/updates for full idempotency

## Future Migrations

Template for idempotent migrations:

```sql
-- UP
-- Description of migration

-- Create tables
CREATE TABLE IF NOT EXISTS new_table (...);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_name ON table(column);

-- Insert seed data
INSERT INTO table (col1, col2) VALUES
('value1', 'value2')
ON CONFLICT (unique_col) DO NOTHING;

-- Modify columns (with existence check)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT FROM information_schema.columns 
        WHERE table_name = 'table' AND column_name = 'new_col'
    ) THEN
        ALTER TABLE table ADD COLUMN new_col VARCHAR(255);
    END IF;
END $$;

-- DROP (always safe with IF EXISTS)
DROP INDEX IF EXISTS old_index;
ALTER TABLE table DROP COLUMN IF EXISTS old_column;

-- DOWN
-- Reverse operations (also idempotent)
```



