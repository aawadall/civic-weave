# Enhanced Migration System (v2)

This directory contains the enhanced migration system for CivicWeave with semantic versioning, compatibility tracking, and idempotent operations.

## Migration Structure

Each migration is organized in its own directory with the following structure:

```
v1.0.0_migration_name/
├── metadata.json    # Migration metadata and version info
├── up.sql          # Forward migration SQL
└── down.sql        # Rollback migration SQL (optional)
```

## Migration Metadata

The `metadata.json` file contains:

```json
{
  "version": "1.0.0",                    # Semantic version
  "name": "migration_name",               # Human-readable name
  "description": "What this migration does",
  "min_runtime_version": "0.1.0",        # Minimum runtime version required
  "max_runtime_version": "",             # Maximum runtime version (optional)
  "dependencies": ["1.0.0"],             # Required migration versions
  "author": "Author Name",                # Migration author
  "created_at": "2025-01-18T00:00:00Z"   # Creation timestamp
}
```

## Version Compatibility

- **min_runtime_version**: The minimum application version required to run this migration
- **max_runtime_version**: The maximum application version allowed (optional)
- **dependencies**: List of migration versions that must be applied first

## Idempotency

All migrations are designed to be idempotent:

- Use `CREATE TABLE IF NOT EXISTS` for table creation
- Use `CREATE INDEX IF NOT EXISTS` for index creation
- Use `INSERT ... ON CONFLICT DO NOTHING` for seed data
- Use `ALTER TABLE ... IF EXISTS/IF NOT EXISTS` for schema changes

## Usage

### CLI Commands

```bash
# Apply pending migrations
make db-migrate-v2

# Check migration status
make db-migrate-status

# View compatibility matrix
make db-migrate-compat

# Rollback to specific version
cd backend && go run cmd/migrate/main.go down 1.0.0
```

### Programmatic API

```go
// Check compatibility
compat, err := database.CheckCompatibility(db, "1.2.3")

// Auto-migrate with options
options := &database.MigrationHookOptions{
    DryRun: false,
    FailOnIncompatible: true,
    RuntimeVersion: "1.2.3",
}
err := database.AutoMigrate(db, "1.2.3", options)

// Get migration status
status, err := database.GetMigrationStatus(db)
```

## Migration Guidelines

1. **Versioning**: Use semantic versioning (MAJOR.MINOR.PATCH)
2. **Naming**: Use descriptive names that explain what the migration does
3. **Dependencies**: Always specify dependencies for complex migrations
4. **Rollback**: Provide rollback SQL when possible
5. **Testing**: Test both up and down migrations thoroughly

## Best Practices

- Keep migrations small and focused
- Always provide rollback SQL
- Test migrations on a copy of production data
- Use transactions for complex migrations
- Document breaking changes in migration descriptions
- Consider runtime version requirements carefully
