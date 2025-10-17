# Cloud Migration Examples

## Example 1: Google Cloud to AWS RDS

```bash
# Set environment variables
export SOURCE_DB_HOST="your-gcp-instance-ip"
export SOURCE_DB_USER="postgres"
export SOURCE_DB_NAME="civicweave"
export SOURCE_DB_PASSWORD="your-gcp-password"

export NEW_DB_HOST="your-rds-endpoint.amazonaws.com"
export NEW_DB_USER="postgres"
export NEW_DB_NAME="civicweave"
export NEW_DB_PASSWORD="your-aws-password"

# Run migration
./scripts/migrate-with-env.sh
```

## Example 2: AWS RDS to Azure Database

```bash
# Set environment variables
export SOURCE_DB_HOST="your-rds-endpoint.amazonaws.com"
export SOURCE_DB_USER="postgres"
export SOURCE_DB_NAME="civicweave"
export SOURCE_DB_PASSWORD="your-aws-password"

export NEW_DB_HOST="your-server.postgres.database.azure.com"
export NEW_DB_USER="postgres@your-server"
export NEW_DB_NAME="civicweave"
export NEW_DB_PASSWORD="your-azure-password"

# Run migration
./scripts/migrate-with-env.sh
```

## Example 3: Using the Advanced Script

```bash
# Run with command line arguments
./scripts/migrate-cloud.sh \
  --source-host "old-cloud.com" \
  --source-user "postgres" \
  --source-db "civicweave" \
  --source-password "old-password" \
  --target-host "new-cloud.com" \
  --target-user "postgres" \
  --target-db "civicweave" \
  --target-password "new-password"
```

## Example 4: Using Your Idempotent Migrations

If you prefer to rebuild from scratch using your existing migrations:

```bash
# 1. Create new database on target cloud
# 2. Set environment variables for new cloud
export DB_HOST="new-cloud-host"
export DB_USER="postgres"
export DB_NAME="civicweave"
export DB_PASSWORD="new-password"

# 3. Run all migrations (they're idempotent!)
make db-migrate

# 4. Export only data from old database
pg_dump -h old-cloud-host -U postgres -d civicweave \
  --data-only --no-owner --no-privileges \
  --file=data_only.sql

# 5. Import data to new database
psql -h new-cloud-host -U postgres -d civicweave -f data_only.sql
```

## Example 5: Zero-Downtime Migration

For zero-downtime migration:

```bash
# 1. Set up read replica on new cloud
# 2. Sync data continuously
# 3. When ready, switch application to new cloud
# 4. Stop old cloud

# Step 1: Create read replica
# (This depends on your cloud provider's specific tools)

# Step 2: Sync data
while true; do
  pg_dump -h old-cloud-host -U postgres -d civicweave \
    --data-only --no-owner --no-privileges \
    --file=incremental_sync.sql
  
  psql -h new-cloud-host -U postgres -d civicweave \
    -f incremental_sync.sql
  
  sleep 300  # Sync every 5 minutes
done
```

## Example 6: Migration with Validation

```bash
#!/bin/bash
# migration-with-validation.sh

set -e

echo "🚀 Starting validated migration..."

# Step 1: Pre-migration validation
echo "🔍 Pre-migration validation..."
psql -h $SOURCE_DB_HOST -U $SOURCE_DB_USER -d $SOURCE_DB_NAME \
  -c "SELECT COUNT(*) as user_count FROM users;"
psql -h $SOURCE_DB_HOST -U $SOURCE_DB_USER -d $SOURCE_DB_NAME \
  -c "SELECT COUNT(*) as project_count FROM projects;"

# Step 2: Run migration
./scripts/migrate-with-env.sh

# Step 3: Post-migration validation
echo "🔍 Post-migration validation..."
psql -h $NEW_DB_HOST -U $NEW_DB_USER -d $NEW_DB_NAME \
  -c "SELECT COUNT(*) as user_count FROM users;"
psql -h $NEW_DB_HOST -U $NEW_DB_USER -d $NEW_DB_NAME \
  -c "SELECT COUNT(*) as project_count FROM projects;"

# Step 4: Test application connectivity
echo "🔍 Testing application connectivity..."
cd backend
export DB_HOST=$NEW_DB_HOST
export DB_USER=$NEW_DB_USER
export DB_NAME=$NEW_DB_NAME
export DB_PASSWORD=$NEW_DB_PASSWORD

# Test database connection
go run cmd/server/main.go --dry-run

echo "✅ Migration with validation complete!"
```

## Example 7: Rollback Strategy

```bash
#!/bin/bash
# rollback-migration.sh

echo "🔄 Rolling back migration..."

# Restore from backup
psql -h $SOURCE_DB_HOST -U $SOURCE_DB_USER -d $SOURCE_DB_NAME \
  -f "backup_$(date +%Y%m%d_%H%M%S).sql"

echo "✅ Rollback complete!"
```

## Example 8: Cloud-Specific Examples

### Google Cloud SQL

```bash
# Using gcloud CLI
gcloud sql connect civicweave-postgres --user=postgres --database=civicweave

# Export
pg_dump -h your-instance-ip -U postgres -d civicweave \
  --no-owner --no-privileges --clean --if-exists \
  --file=civicweave_export.sql
```

### AWS RDS

```bash
# Using AWS CLI
aws rds describe-db-instances --db-instance-identifier your-instance

# Export
pg_dump -h your-rds-endpoint.amazonaws.com -U postgres -d civicweave \
  --no-owner --no-privileges --clean --if-exists \
  --file=civicweave_export.sql
```

### Azure Database

```bash
# Using Azure CLI
az postgres server show --name your-server --resource-group your-rg

# Export
pg_dump -h your-server.postgres.database.azure.com -U postgres@your-server -d civicweave \
  --no-owner --no-privileges --clean --if-exists \
  --file=civicweave_export.sql
```

## Best Practices

1. **Always test on staging first**
2. **Create backups before migration**
3. **Validate data integrity after migration**
4. **Have rollback plans ready**
5. **Monitor during migration**
6. **Use your idempotent migrations when possible**

## Troubleshooting

### Common Issues

1. **Connection timeout**: Increase connection timeout
2. **Permission denied**: Check user permissions
3. **Database doesn't exist**: Create database first
4. **Schema conflicts**: Use `--clean --if-exists` flags

### Debug Commands

```bash
# Test connection
psql -h $HOST -U $USER -d $DB -c "SELECT 1;"

# Check database size
psql -h $HOST -U $USER -d $DB -c "SELECT pg_size_pretty(pg_database_size('$DB'));"

# Check table counts
psql -h $HOST -U $USER -d $DB -c "SELECT schemaname,tablename,n_tup_ins,n_tup_upd,n_tup_del FROM pg_stat_user_tables;"
```
