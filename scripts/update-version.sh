#!/bin/bash

# Script to update version numbers for both frontend and backend
# Usage: ./scripts/update-version.sh [version] [environment]
# If no version provided, auto-generate based on git commit count

VERSION=${1:-""}
ENVIRONMENT=${2:-"development"}
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Auto-generate version if not provided
if [ -z "$VERSION" ]; then
    COMMIT_COUNT=$(git rev-list --count HEAD 2>/dev/null || echo "1")
    VERSION="${COMMIT_COUNT}.0.0-${GIT_COMMIT}"
fi

echo "Updating version to $VERSION ($ENVIRONMENT) - Git commit: $GIT_COMMIT"

# Update backend version
echo "Updating backend version..."
sed -i "s/Version   = getEnvOrDefault(\"VERSION\", \".*\")/Version   = getEnvOrDefault(\"VERSION\", \"$VERSION\")/" backend/cmd/server/version.go

# Update frontend version
echo "Updating frontend version..."
sed -i "s/export const VERSION = import.meta.env.VITE_VERSION || \".*\"/export const VERSION = import.meta.env.VITE_VERSION || \"$VERSION\"/" frontend/src/version.js

echo "Version updated successfully!"
echo "Frontend: $VERSION"
echo "Backend: $VERSION"
echo "Environment: $ENVIRONMENT"
echo "Git Commit: $GIT_COMMIT"
echo ""
echo "Next steps:"
echo "1. Run 'make build-push' to build with new version"
echo "2. Run 'make deploy-app' to deploy to production"
