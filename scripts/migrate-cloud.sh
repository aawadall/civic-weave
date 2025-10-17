#!/bin/bash
# Cloud Migration Script for CivicWeave
# Usage: ./migrate-cloud.sh [source-cloud] [target-cloud]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 is not installed. Please install it first."
        exit 1
    fi
}

# Function to validate database connection
validate_connection() {
    local host=$1
    local user=$2
    local db=$3
    local password=$4
    
    print_status "Testing connection to $host..."
    
    if PGPASSWORD=$password psql -h $host -U $user -d $db -c "SELECT 1;" &> /dev/null; then
        print_success "Connection to $host successful"
        return 0
    else
        print_error "Failed to connect to $host"
        return 1
    fi
}

# Function to create backup
create_backup() {
    local host=$1
    local user=$2
    local db=$3
    local password=$4
    local backup_file=$5
    
    print_status "Creating backup of $db..."
    
    PGPASSWORD=$password pg_dump -h $host -U $user -d $db \
        --no-owner --no-privileges --clean --if-exists \
        --file="$backup_file"
    
    if [ $? -eq 0 ]; then
        print_success "Backup created: $backup_file"
    else
        print_error "Backup failed"
        exit 1
    fi
}

# Function to export database
export_database() {
    local host=$1
    local user=$2
    local db=$3
    local password=$4
    local export_file=$5
    
    print_status "Exporting database from $host..."
    
    PGPASSWORD=$password pg_dump -h $host -U $user -d $db \
        --no-owner --no-privileges --clean --if-exists \
        --file="$export_file"
    
    if [ $? -eq 0 ]; then
        print_success "Database exported: $export_file"
    else
        print_error "Export failed"
        exit 1
    fi
}

# Function to import database
import_database() {
    local host=$1
    local user=$2
    local db=$3
    local password=$4
    local import_file=$5
    
    print_status "Importing database to $host..."
    
    PGPASSWORD=$password psql -h $host -U $user -d $db \
        -f "$import_file"
    
    if [ $? -eq 0 ]; then
        print_success "Database imported successfully"
    else
        print_error "Import failed"
        exit 1
    fi
}

# Function to validate migration
validate_migration() {
    local host=$1
    local user=$2
    local db=$3
    local password=$4
    
    print_status "Validating migration..."
    
    # Check if key tables exist
    local tables=("users" "projects" "volunteers" "project_messages" "schema_migrations")
    
    for table in "${tables[@]}"; do
        if PGPASSWORD=$password psql -h $host -U $user -d $db -c "SELECT 1 FROM $table LIMIT 1;" &> /dev/null; then
            print_success "Table $table exists and is accessible"
        else
            print_error "Table $table is missing or not accessible"
            exit 1
        fi
    done
    
    # Check migration history
    print_status "Checking migration history..."
    PGPASSWORD=$password psql -h $host -U $user -d $db -c "SELECT * FROM schema_migrations ORDER BY version;"
    
    print_success "Migration validation complete"
}

# Main migration function
migrate_database() {
    local source_host=$1
    local source_user=$2
    local source_db=$3
    local source_password=$4
    local target_host=$5
    local target_user=$6
    local target_db=$7
    local target_password=$8
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="backup_${timestamp}.sql"
    local export_file="civicweave_export_${timestamp}.sql"
    
    print_status "Starting database migration..."
    print_status "Source: $source_host/$source_db"
    print_status "Target: $target_host/$target_db"
    
    # Step 1: Validate connections
    validate_connection $source_host $source_user $source_db $source_password
    validate_connection $target_host $target_user $target_db $target_password
    
    # Step 2: Create backup
    create_backup $source_host $source_user $source_db $source_password $backup_file
    
    # Step 3: Export database
    export_database $source_host $source_user $source_db $source_password $export_file
    
    # Step 4: Import database
    import_database $target_host $target_user $target_db $target_password $export_file
    
    # Step 5: Validate migration
    validate_migration $target_host $target_user $target_db $target_password
    
    print_success "Migration completed successfully!"
    print_status "Backup file: $backup_file"
    print_status "Export file: $export_file"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --source-host HOST        Source database host"
    echo "  --source-user USER        Source database user"
    echo "  --source-db DB            Source database name"
    echo "  --source-password PASS    Source database password"
    echo "  --target-host HOST        Target database host"
    echo "  --target-user USER        Target database user"
    echo "  --target-db DB            Target database name"
    echo "  --target-password PASS    Target database password"
    echo "  --help                    Show this help message"
    echo ""
    echo "Example:"
    echo "  $0 --source-host old-cloud.com --source-user postgres --source-db civicweave --source-password pass1 \\"
    echo "      --target-host new-cloud.com --target-user postgres --target-db civicweave --target-password pass2"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --source-host)
            SOURCE_HOST="$2"
            shift 2
            ;;
        --source-user)
            SOURCE_USER="$2"
            shift 2
            ;;
        --source-db)
            SOURCE_DB="$2"
            shift 2
            ;;
        --source-password)
            SOURCE_PASSWORD="$2"
            shift 2
            ;;
        --target-host)
            TARGET_HOST="$2"
            shift 2
            ;;
        --target-user)
            TARGET_USER="$2"
            shift 2
            ;;
        --target-db)
            TARGET_DB="$2"
            shift 2
            ;;
        --target-password)
            TARGET_PASSWORD="$2"
            shift 2
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Check if all required parameters are provided
if [[ -z "$SOURCE_HOST" || -z "$SOURCE_USER" || -z "$SOURCE_DB" || -z "$SOURCE_PASSWORD" || 
      -z "$TARGET_HOST" || -z "$TARGET_USER" || -z "$TARGET_DB" || -z "$TARGET_PASSWORD" ]]; then
    print_error "Missing required parameters"
    show_usage
    exit 1
fi

# Check if required commands are available
check_command "psql"
check_command "pg_dump"

# Confirm migration
print_warning "This will migrate the database from $SOURCE_HOST to $TARGET_HOST"
print_warning "This operation cannot be undone. Make sure you have backups!"
echo ""
read -p "Are you sure you want to continue? (y/N): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_status "Migration cancelled"
    exit 0
fi

# Run migration
migrate_database "$SOURCE_HOST" "$SOURCE_USER" "$SOURCE_DB" "$SOURCE_PASSWORD" \
                 "$TARGET_HOST" "$TARGET_USER" "$TARGET_DB" "$TARGET_PASSWORD"
