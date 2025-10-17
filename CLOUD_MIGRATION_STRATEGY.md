# Cloud-Agnostic Database Migration Strategy

## Overview

This guide provides a stable, idempotent method for migrating your database when changing cloud providers, using your existing migration system as the foundation.

## Migration Strategy Options

### Option 1: Database Dump & Restore (Recommended)

**Best for:** Complete database migration with minimal downtime

```bash
# 1. Export from current cloud
pg_dump -h current-cloud-host -U username -d database_name \
  --no-owner --no-privileges --clean --if-exists \
  --file=civicweave_migration.sql

# 2. Import to new cloud
psql -h new-cloud-host -U username -d database_name \
  -f civicweave_migration.sql
```

### Option 2: Schema + Data Migration (Idempotent)

**Best for:** Zero-downtime migration with validation

```bash
# 1. Export schema only
pg_dump -h current-cloud-host -U username -d database_name \
  --schema-only --no-owner --no-privileges \
  --file=schema_only.sql

# 2. Export data only
pg_dump -h current-cloud-host -U username -d database_name \
  --data-only --no-owner --no-privileges \
  --file=data_only.sql

# 3. Apply to new cloud
psql -h new-cloud-host -U username -d database_name -f schema_only.sql
psql -h new-cloud-host -U username -d database_name -f data_only.sql
```

### Option 3: Migration-Based (Your Current System)

**Best for:** When you want to rebuild from scratch using your idempotent migrations

```bash
# 1. Create fresh database on new cloud
# 2. Run all migrations
make db-migrate
# 3. Import only data (no schema)
```

## Cloud-Specific Migration Scripts

### AWS RDS Migration

```bash
#!/bin/bash
# migrate-to-aws.sh

# Set environment variables
export NEW_DB_HOST="your-rds-endpoint.amazonaws.com"
export NEW_DB_NAME="civicweave"
export NEW_DB_USER="postgres"
export NEW_DB_PASSWORD="your-password"

# Export from current database
pg_dump -h $CURRENT_DB_HOST -U $CURRENT_DB_USER -d $CURRENT_DB_NAME \
  --no-owner --no-privileges --clean --if-exists \
  --file=civicweave_export.sql

# Import to AWS RDS
PGPASSWORD=$NEW_DB_PASSWORD psql -h $NEW_DB_HOST -U $NEW_DB_USER -d $NEW_DB_NAME \
  -f civicweave_export.sql

echo "✅ Migration to AWS RDS complete!"
```

### Google Cloud SQL Migration

```bash
#!/bin/bash
# migrate-to-gcp.sh

# Set environment variables
export NEW_DB_HOST="your-instance-ip"
export NEW_DB_NAME="civicweave"
export NEW_DB_USER="postgres"

# Export from current database
pg_dump -h $CURRENT_DB_HOST -U $CURRENT_DB_USER -d $CURRENT_DB_NAME \
  --no-owner --no-privileges --clean --if-exists \
  --file=civicweave_export.sql

# Import to Google Cloud SQL
psql -h $NEW_DB_HOST -U $NEW_DB_USER -d $NEW_DB_NAME \
  -f civicweave_export.sql

echo "✅ Migration to Google Cloud SQL complete!"
```

### Azure Database Migration

```bash
#!/bin/bash
# migrate-to-azure.sh

# Set environment variables
export NEW_DB_HOST="your-server.postgres.database.azure.com"
export NEW_DB_NAME="civicweave"
export NEW_DB_USER="postgres@your-server"

# Export from current database
pg_dump -h $CURRENT_DB_HOST -U $CURRENT_DB_USER -d $CURRENT_DB_NAME \
  --no-owner --no-privileges --clean --if-exists \
  --file=civicweave_export.sql

# Import to Azure Database
PGPASSWORD=$NEW_DB_PASSWORD psql -h $NEW_DB_HOST -U $NEW_DB_USER -d $NEW_DB_NAME \
  -f civicweave_export.sql

echo "✅ Migration to Azure Database complete!"
```

## Idempotent Migration Validation

### Pre-Migration Checklist

```bash
#!/bin/bash
# validate-migration.sh

echo "🔍 Validating migration readiness..."

# 1. Check current database state
echo "Current database tables:"
psql -h $CURRENT_DB_HOST -U $CURRENT_DB_USER -d $CURRENT_DB_NAME -c "\dt"

# 2. Check migration history
echo "Migration history:"
psql -h $CURRENT_DB_HOST -U $CURRENT_DB_USER -d $CURRENT_DB_NAME -c "SELECT * FROM schema_migrations ORDER BY version;"

# 3. Test idempotency
echo "Testing idempotency..."
make db-migrate
make db-migrate  # Should succeed without errors

echo "✅ Database ready for migration!"
```

### Post-Migration Validation

```bash
#!/bin/bash
# validate-migration-complete.sh

echo "🔍 Validating migration completion..."

# 1. Check all tables exist
echo "Tables in new database:"
psql -h $NEW_DB_HOST -U $NEW_DB_USER -d $NEW_DB_NAME -c "\dt"

# 2. Check migration history
echo "Migration history:"
psql -h $NEW_DB_HOST -U $NEW_DB_USER -d $NEW_DB_NAME -c "SELECT * FROM schema_migrations ORDER BY version;"

# 3. Test application connectivity
echo "Testing application connection..."
cd backend && go run cmd/server/main.go --dry-run

echo "✅ Migration validation complete!"
```

## Environment-Specific Configuration

### Production Migration Script

```bash
#!/bin/bash
# production-migration.sh

set -e  # Exit on any error

echo "🚀 Starting production database migration..."

# Load environment variables
source .env.production

# Backup current database
echo "📦 Creating backup..."
pg_dump -h $CURRENT_DB_HOST -U $CURRENT_DB_USER -d $CURRENT_DB_NAME \
  --no-owner --no-privileges --clean --if-exists \
  --file="backup_$(date +%Y%m%d_%H%M%S).sql"

# Export current state
echo "📤 Exporting current database..."
pg_dump -h $CURRENT_DB_HOST -U $CURRENT_DB_USER -d $CURRENT_DB_NAME \
  --no-owner --no-privileges --clean --if-exists \
  --file=civicweave_migration.sql

# Import to new cloud
echo "📥 Importing to new cloud..."
psql -h $NEW_DB_HOST -U $NEW_DB_USER -d $NEW_DB_NAME \
  -f civicweave_migration.sql

# Validate migration
echo "🔍 Validating migration..."
psql -h $NEW_DB_HOST -U $NEW_DB_USER -d $NEW_DB_NAME -c "SELECT COUNT(*) FROM users;"
psql -h $NEW_DB_HOST -U $NEW_DB_USER -d $NEW_DB_NAME -c "SELECT COUNT(*) FROM projects;"

echo "✅ Production migration complete!"
```

## Rollback Strategy

### Quick Rollback Script

```bash
#!/bin/bash
# rollback-migration.sh

echo "🔄 Rolling back migration..."

# Restore from backup
psql -h $CURRENT_DB_HOST -U $CURRENT_DB_USER -d $CURRENT_DB_NAME \
  -f "backup_$(date +%Y%m%d_%H%M%S).sql"

echo "✅ Rollback complete!"
```

## Best Practices

### 1. Always Use Idempotent Migrations

Your existing system is already idempotent, which is perfect for cloud migrations:

```sql
-- ✅ Good: Idempotent
CREATE TABLE IF NOT EXISTS users (...);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- ❌ Bad: Not idempotent
CREATE TABLE users (...);
CREATE INDEX idx_users_email ON users(email);
```

### 2. Test Migration Process

```bash
# Test on staging first
./migrate-to-staging.sh
# Validate staging works
# Then migrate production
./migrate-to-production.sh
```

### 3. Monitor Migration Progress

```bash
# Monitor during migration
watch -n 5 'psql -h $NEW_DB_HOST -U $NEW_DB_USER -d $NEW_DB_NAME -c "SELECT COUNT(*) FROM users;"'
```

### 4. Use Connection Pooling

```bash
# Use connection pooling for large migrations
pg_dump -h $CURRENT_DB_HOST -U $CURRENT_DB_USER -d $CURRENT_DB_NAME \
  --jobs=4 --no-owner --no-privileges \
  --file=civicweave_export.sql
```

## Migration Timeline

### Phase 1: Preparation (1-2 hours)
- [ ] Create new cloud database
- [ ] Test connectivity
- [ ] Validate current database state
- [ ] Create backup

### Phase 2: Migration (30 minutes - 2 hours)
- [ ] Export current database
- [ ] Import to new cloud
- [ ] Validate data integrity
- [ ] Test application connectivity

### Phase 3: Cutover (5-15 minutes)
- [ ] Update application configuration
- [ ] Deploy to new cloud
- [ ] Monitor for issues
- [ ] Rollback if needed

## Emergency Procedures

### If Migration Fails

1. **Immediate**: Rollback to original database
2. **Investigate**: Check logs for specific errors
3. **Fix**: Address the issue
4. **Retry**: Run migration again

### If Data Corruption Detected

1. **Stop**: Immediately stop application
2. **Restore**: Restore from backup
3. **Investigate**: Find root cause
4. **Fix**: Address the issue
5. **Retry**: Run migration again

## Conclusion

Your existing idempotent migration system is perfect for cloud migrations. The key is to:

1. **Use your existing migrations** as the foundation
2. **Test thoroughly** on staging first
3. **Have rollback plans** ready
4. **Monitor closely** during migration
5. **Validate everything** after migration

This approach gives you a stable, reliable method for migrating between any cloud providers while maintaining data integrity and minimizing downtime.
