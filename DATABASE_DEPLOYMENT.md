# 🗄️ Database Deployment Utility

This utility provides a robust way to deploy database migrations to production environments safely and reliably.

## 🚀 Quick Start

### 1. Setup Production Environment

Create a production environment file:

```bash
# Copy the template
cp .env.example .env.production

# Edit with your production database credentials
nano .env.production
```

### 2. Check Migration Status

```bash
# Show current migration status
make db-deploy-status
```

### 3. Deploy Migrations

```bash
# Deploy all pending migrations
make db-deploy

# Or use the script directly
./scripts/deploy-db.sh
```

## 📋 Available Commands

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make db-deploy-status` | Show migration status |
| `make db-deploy-dry` | Dry run (show what would be executed) |
| `make db-deploy` | Deploy all pending migrations |
| `make db-deploy-version` | Deploy up to specific version |
| `make db-rollback` | Rollback to specific version |

### Direct Script Usage

```bash
# Show help
./scripts/deploy-db.sh --help

# Show migration status
./scripts/deploy-db.sh --status

# Dry run (see what would be executed)
./scripts/deploy-db.sh --dry-run

# Deploy all pending migrations
./scripts/deploy-db.sh

# Deploy up to specific version
./scripts/deploy-db.sh --version 011

# Rollback to specific version
./scripts/deploy-db.sh --rollback 010

# Use custom environment file
./scripts/deploy-db.sh --env .env.staging
```

## 🔧 Features

### ✅ Safety Features

- **Dry Run Mode**: See what would be executed without making changes
- **Transaction Safety**: Each migration runs in a transaction
- **Rollback Support**: Rollback migrations if needed
- **Status Tracking**: Track which migrations have been applied
- **Checksum Validation**: Verify migration integrity

### 📊 Migration Tracking

The utility creates a `schema_migrations` table to track:
- Migration version
- Applied timestamp
- Checksum for integrity

### 🔄 Rollback Support

Each migration file must have both UP and DOWN sections:

```sql
-- UP
-- Migration content here
CREATE TABLE new_table (...);

-- DOWN
-- Rollback content here
DROP TABLE new_table;
```

## 🛠️ Environment Configuration

### Required Environment Variables

```bash
# Database Configuration
DB_HOST=your-production-db-host
DB_PORT=5432
DB_NAME=civicweave
DB_USER=civicweave
DB_PASSWORD=your-secure-password
DB_SSLMODE=require

# JWT Configuration
JWT_SECRET=your-jwt-secret
```

### Optional Environment Variables

```bash
# Email Configuration
ENABLE_EMAIL=true
MAILGUN_API_KEY=your-mailgun-key
MAILGUN_DOMAIN=your-domain

# OAuth Configuration
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# OpenAI Configuration
OPENAI_API_KEY=your-openai-key
OPENAI_EMBEDDING_MODEL=text-embedding-3-small
```

## 📁 Migration File Structure

Migration files should be named with the pattern: `{version}_{description}.sql`

Example: `011_task_enhancements.sql`

Each file must contain:

```sql
-- UP
-- Migration content here
CREATE TABLE task_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES project_tasks(id),
    -- ... other columns
);

-- DOWN
-- Rollback content here
DROP TABLE task_comments;
```

## 🚨 Best Practices

### Before Deployment

1. **Test Locally**: Always test migrations in development first
2. **Backup Database**: Create a backup before deploying to production
3. **Dry Run**: Use `--dry-run` to verify what will be executed
4. **Check Status**: Use `--status` to see current migration state

### During Deployment

1. **Monitor Output**: Watch for any errors during execution
2. **Verify Results**: Check that migrations applied correctly
3. **Test Application**: Verify the application works with new schema

### After Deployment

1. **Update Application**: Deploy new application code
2. **Monitor Performance**: Watch for any performance issues
3. **Document Changes**: Update documentation if needed

## 🔍 Troubleshooting

### Common Issues

**Connection Failed**
```bash
# Check environment variables
cat .env.production

# Test database connection
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1;"
```

**Migration Already Applied**
```bash
# Check migration status
./scripts/deploy-db.sh --status

# If needed, manually remove from schema_migrations
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "DELETE FROM schema_migrations WHERE version = '011';"
```

**Rollback Failed**
```bash
# Check if DOWN section exists in migration file
cat backend/migrations/011_task_enhancements.sql

# Manual rollback if needed
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f rollback_script.sql
```

## 📚 Examples

### Deploy Task Management Enhancements

```bash
# 1. Check current status
make db-deploy-status

# 2. Dry run to see what will be executed
make db-deploy-dry

# 3. Deploy the migration
make db-deploy

# 4. Verify deployment
make db-deploy-status
```

### Rollback if Needed

```bash
# Rollback to previous version
make db-rollback
# Enter: 010

# Verify rollback
make db-deploy-status
```

## 🔐 Security Notes

- Never commit production credentials to version control
- Use strong passwords and secrets
- Enable SSL for database connections in production
- Regularly rotate secrets and passwords
- Use environment-specific configuration files

## 📞 Support

If you encounter issues:

1. Check the troubleshooting section above
2. Review the migration file syntax
3. Verify environment configuration
4. Check database connection and permissions
5. Contact the development team for assistance
