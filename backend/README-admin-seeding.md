# Admin User Seeding

This document explains how to create an admin user for the CivicWeave platform.

## Prerequisites

1. Database connection configured
2. Environment variables set up

## Setup

1. **Copy environment file:**
   ```bash
   cp env.example .env
   ```

2. **Edit `.env` file with your admin credentials:**
   ```bash
   ADMIN_EMAIL=admin@yourdomain.com
   ADMIN_PASSWORD=YourSecurePassword123
   ADMIN_NAME=Your Name
   ```

3. **Set database connection:**
   ```bash
   DB_HOST=your_database_host
   DB_PORT=5432
   DB_NAME=civicweave
   DB_USER=civicweave
   DB_PASSWORD=your_database_password
   ```

## Create Admin User

Run the seeding script:

```bash
go run cmd/seed-admin/main.go
```

Or build and run:

```bash
go build -o seed-admin cmd/seed-admin/main.go
./seed-admin
```

## Security Notes

- **Never commit `.env` files** to version control
- Use strong passwords (uppercase, lowercase, numbers, special characters)
- Change default admin credentials after first login
- Consider using environment variables in production

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `ADMIN_EMAIL` | Admin user email | Yes |
| `ADMIN_PASSWORD` | Admin user password | Yes |
| `ADMIN_NAME` | Admin user display name | No (defaults to "System Administrator") |
| `DB_HOST` | Database host | Yes |
| `DB_PORT` | Database port | Yes |
| `DB_NAME` | Database name | Yes |
| `DB_USER` | Database user | Yes |
| `DB_PASSWORD` | Database password | Yes |

## Production Deployment

For production, use Google Cloud Secret Manager:

```bash
# Set admin password in Secret Manager
gcloud secrets create admin-password --data-file=-
# Enter your secure password and press Ctrl+D

# Run seeding with Secret Manager
ADMIN_PASSWORD=$(gcloud secrets versions access latest --secret="admin-password") \
DB_PASSWORD=$(gcloud secrets versions access latest --secret="db-password") \
./seed-admin
```
