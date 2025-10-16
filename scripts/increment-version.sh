#!/bin/bash

# Script to automatically increment version numbers in VERSION files
# Usage: ./scripts/increment-version.sh [component]
# component: "backend", "frontend", or "both" (default)

COMPONENT=${1:-"both"}

# Function to increment version in a file
increment_version_file() {
    local file=$1
    local current_version
    
    if [ -f "$file" ]; then
        current_version=$(cat "$file" | tr -d '\n\r')
    else
        current_version="1.0.0"
    fi
    
    # Parse version (assume semantic versioning: major.minor.patch)
    IFS='.' read -ra VERSION_PARTS <<< "$current_version"
    major=${VERSION_PARTS[0]:-1}
    minor=${VERSION_PARTS[1]:-0}
    patch=${VERSION_PARTS[2]:-0}
    
    # Increment patch version
    patch=$((patch + 1))
    new_version="$major.$minor.$patch"
    
    # Write new version to file
    echo "$new_version" > "$file"
    echo "Incremented $file: $current_version -> $new_version"
}

# Get the script directory to find the project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Increment versions based on component
case $COMPONENT in
    "backend")
        increment_version_file "$PROJECT_ROOT/backend/VERSION"
        ;;
    "frontend")
        increment_version_file "$PROJECT_ROOT/frontend/VERSION"
        ;;
    "both"|*)
        increment_version_file "$PROJECT_ROOT/backend/VERSION"
        increment_version_file "$PROJECT_ROOT/frontend/VERSION"
        ;;
esac

echo "Version increment complete!"
