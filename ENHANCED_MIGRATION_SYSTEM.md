# Enhanced Migration System with Semantic Versioning

## Overview

This PR introduces a comprehensive enhanced migration system for CivicWeave that provides semantic versioning, compatibility tracking, idempotent operations, and programmatic APIs for headless operations.

## Key Features

### 🎯 Semantic Versioning
- **Version Format**: Uses semantic versioning (MAJOR.MINOR.PATCH) for migrations
- **Compatibility Matrix**: Tracks compatibility between database and runtime versions
- **Dependency Management**: Supports migration dependencies and validation

### 🔄 Idempotent Operations
- **Safe Re-runs**: Migrations can be run multiple times safely
- **Checksum Validation**: Prevents execution of modified migrations
- **Transaction Support**: All migrations run in transactions with rollback on failure

### 🛠️ Enhanced CLI
- **Multiple Commands**: `up`, `down`, `status`, `compatibility`, `validate`, `check`
- **CI/CD Friendly**: Exit codes and quiet mode for automated environments
- **Rich Output**: Human-readable and JSON output formats

### 🔌 Programmatic API
- **Headless Operations**: Full API for Docker, Kubernetes, and CI/CD integration
- **Health Checks**: Built-in health check functionality
- **Compatibility Checking**: Runtime version validation

## Architecture

### Migration Structure
```
migrations_v2/
├── v1.0.0_initial_schema/
│   ├── metadata.json    # Version, dependencies, compatibility
│   ├── up.sql          # Forward migration
│   └── down.sql        # Rollback migration
└── README.md           # Documentation
```

### Database Schema
```sql
CREATE TABLE schema_migrations_v2 (
    version VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    runtime_version VARCHAR(20),
    execution_time_ms INTEGER,
    status VARCHAR(20) DEFAULT 'applied'
);
```

## Usage Examples

### CLI Commands
```bash
# Apply pending migrations
make db-migrate-v2

# Check migration status
make db-migrate-status

# View compatibility matrix
make db-migrate-compat

# Validate migration files
make db-migrate-validate

# Health check for CI/CD
make db-migrate-check

# Rollback to specific version
make db-migrate-rollback
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

// Health check for CI/CD
exitCode, message, err := database.CheckMigrationHealth(db, "1.2.3")
```

## Migration Guide

### For Existing Users
- **Backward Compatible**: Existing migration system continues to work
- **Gradual Migration**: New system runs alongside legacy system
- **No Breaking Changes**: All existing functionality preserved

### Creating New Migrations
1. Create directory: `migrations_v2/v1.1.0_feature_name/`
2. Add `metadata.json` with version and compatibility info
3. Add `up.sql` with forward migration
4. Add `down.sql` with rollback migration (optional)

### Example Migration
```json
{
  "version": "1.1.0",
  "name": "add_user_preferences",
  "description": "Add user preferences table",
  "min_runtime_version": "1.0.0",
  "dependencies": ["1.0.0"],
  "author": "Developer Name"
}
```

## Benefits

### 🚀 Developer Experience
- **Clear Versioning**: Semantic versions make compatibility obvious
- **Rich CLI**: Comprehensive commands for all migration needs
- **Documentation**: Built-in help and examples

### 🔒 Production Safety
- **Idempotent**: Safe to run multiple times
- **Checksum Validation**: Prevents accidental re-runs of modified migrations
- **Compatibility Checking**: Prevents runtime/database mismatches

### 🤖 Automation Ready
- **CI/CD Integration**: Exit codes and quiet mode
- **Health Checks**: Built-in monitoring capabilities
- **Programmatic API**: Full control for automation

## Testing

### Manual Testing
```bash
# Test migration system
make db-migrate-v2
make db-migrate-status
make db-migrate-compat

# Test rollback
make db-migrate-rollback  # Enter version when prompted
```

### Automated Testing
```bash
# Health check in CI/CD
make db-migrate-check
echo $?  # Should be 0 for compatible, 1 for needs migration, 2 for incompatible
```

## Files Added/Modified

### New Files
- `backend/database/migration_metadata.go` - Metadata structures and registry
- `backend/database/migrate_v2.go` - Enhanced migration engine
- `backend/database/compatibility.go` - Version compatibility logic
- `backend/database/hooks.go` - Programmatic API for headless operations
- `backend/migrations_v2/` - New migration directory structure
- `backend/migrations_v2/README.md` - Documentation

### Modified Files
- `backend/cmd/migrate/main.go` - Enhanced CLI with new commands
- `backend/cmd/server/main.go` - Added compatibility checking on startup
- `Makefile` - Added new migration targets
- `backend/go.mod` - Added semver dependency

## Backward Compatibility

- ✅ Existing migration system continues to work unchanged
- ✅ All existing Makefile targets preserved
- ✅ No breaking changes to existing APIs
- ✅ Gradual migration path available

## Future Enhancements

- **Migration Templates**: Scaffold new migrations with templates
- **Migration Testing**: Built-in testing framework for migrations
- **Migration Analytics**: Track migration performance and usage
- **Migration Rollback Plans**: Automated rollback strategies

## Conclusion

This enhanced migration system provides a robust, production-ready solution for database schema management with semantic versioning, compatibility tracking, and comprehensive automation support. It maintains full backward compatibility while offering powerful new capabilities for modern development workflows.
