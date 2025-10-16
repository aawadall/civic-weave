#!/bin/bash
# Daily batch job to notify candidates about project matches
# This script should be run via cron or scheduled task

set -e

# Change to backend directory
cd "$(dirname "$0")/.."

echo "Starting daily candidate notification job at $(date)"

# Activate virtual environment or create one if it doesn't exist
if [ ! -d "venv" ]; then
    echo "Creating Python virtual environment..."
    python3 -m venv venv
fi

source venv/bin/activate

# Install dependencies if needed
pip install -q psycopg2-binary python-dotenv 2>/dev/null || true

# Load environment variables if .env file exists
if [ -f ".env" ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Run the notification job
python3 jobs/notify_project_matches.py

echo "Daily candidate notification job completed at $(date)"



