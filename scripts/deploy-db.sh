#!/bin/bash

# Database Deployment Script
# This script helps deploy database migrations to production

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENV_FILE=".env.production"
DRY_RUN=false
TARGET_VERSION=""
ROLLBACK_VERSION=""
SHOW_STATUS=false

# Function to show usage
show_usage() {
    echo "Database Deployment Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -e, --env FILE        Environment file (default: .env.production)"
    echo "  -d, --dry-run         Show what would be executed without running"
    echo "  -v, --version VER     Run migration up to specific version (e.g., 011)"
    echo "  -r, --rollback VER    Rollback to specific version (e.g., 010)"
    echo "  -s, --status          Show migration status"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  # Show migration status"
    echo "  $0 --status"
    echo ""
    echo "  # Run all pending migrations"
    echo "  $0"
    echo ""
    echo "  # Run migrations up to version 011"
    echo "  $0 --version 011"
    echo ""
    echo "  # Dry run to see what would be executed"
    echo "  $0 --dry-run"
    echo ""
    echo "  # Rollback to version 010"
    echo "  $0 --rollback 010"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--env)
            ENV_FILE="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -v|--version)
            TARGET_VERSION="$2"
            shift 2
            ;;
        -r|--rollback)
            ROLLBACK_VERSION="$2"
            shift 2
            ;;
        -s|--status)
            SHOW_STATUS=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_usage
            exit 1
            ;;
    esac
done

# Check if environment file exists
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}Error: Environment file '$ENV_FILE' not found${NC}"
    echo ""
    echo "Create a production environment file with your database credentials:"
    echo "  cp .env.example $ENV_FILE"
    echo "  # Edit $ENV_FILE with your production database settings"
    exit 1
fi

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "backend/cmd/db-deploy/main.go" ]; then
    echo -e "${RED}Error: Please run this script from the project root directory${NC}"
    exit 1
fi

echo -e "${BLUE}🚀 Database Deployment Utility${NC}"
echo "=================================="
echo "Environment file: $ENV_FILE"
echo "Working directory: $(pwd)"
echo ""

# Build the deployment tool
echo -e "${YELLOW}🔨 Building database deployment tool...${NC}"
cd backend
if ! go build -o ../db-deploy cmd/db-deploy/main.go; then
    echo -e "${RED}Error: Failed to build deployment tool${NC}"
    exit 1
fi
cd ..

# Prepare command arguments
ARGS=("-env" "$ENV_FILE")

if [ "$DRY_RUN" = true ]; then
    ARGS+=("-dry-run")
fi

if [ -n "$TARGET_VERSION" ]; then
    ARGS+=("-version" "$TARGET_VERSION")
fi

if [ -n "$ROLLBACK_VERSION" ]; then
    ARGS+=("-rollback" "$ROLLBACK_VERSION")
fi

if [ "$SHOW_STATUS" = true ]; then
    ARGS+=("-status")
fi

# Run the deployment tool
echo -e "${YELLOW}🔧 Running database deployment...${NC}"
echo "Command: ./db-deploy ${ARGS[*]}"
echo ""

if ./db-deploy "${ARGS[@]}"; then
    echo ""
    echo -e "${GREEN}✅ Database deployment completed successfully!${NC}"
else
    echo ""
    echo -e "${RED}❌ Database deployment failed!${NC}"
    exit 1
fi

# Cleanup
rm -f db-deploy

echo ""
echo -e "${BLUE}🎉 Deployment complete!${NC}"
