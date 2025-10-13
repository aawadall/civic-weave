# Batch Jobs

This directory contains Python batch jobs for CivicWeave.

## Available Jobs

### 1. Calculate Matches (`calculate_matches.py`)
Pre-calculates volunteer-project match scores and stores them in the `volunteer_initiative_matches` table.

**Schedule**: Hourly (or as needed)
**Purpose**: Keep match scores up-to-date for fast querying

### 2. Notify Project Matches (`notify_project_matches.py`)
Notifies top matching candidates about projects they're a good fit for, and notifies team leads about available candidates.

**Schedule**: Daily
**Purpose**: Keep volunteers informed about relevant opportunities

## Configuration

All jobs use environment variables for configuration:

```bash
# Database connection
DB_HOST=localhost
DB_PORT=5432
DB_NAME=civicweave
DB_USER=postgres
DB_PASSWORD=password

# Job-specific settings
TOP_K_CANDIDATES=10          # Number of top candidates to notify per project
MIN_MATCH_SCORE=0.6          # Minimum match score (0.0-1.0) to trigger notification
SYSTEM_USER_ID=00000000-0000-0000-0000-000000000000  # System user for automated messages
```

## Running Jobs Manually

### Calculate Matches
```bash
cd backend
python3 jobs/calculate_matches.py
```

### Notify Candidates
```bash
cd backend
./jobs/run_daily_matching.sh
```

## Scheduling with Cron

Add these lines to your crontab (`crontab -e`):

```cron
# Calculate matches every hour at minute 0
0 * * * * cd /path/to/CivicWeave/backend && python3 jobs/calculate_matches.py >> /var/log/civicweave/calculate_matches.log 2>&1

# Notify candidates daily at 9 AM
0 9 * * * cd /path/to/CivicWeave/backend && ./jobs/run_daily_matching.sh >> /var/log/civicweave/notify_matches.log 2>&1
```

## Docker Setup

If running in Docker, add to your `docker-compose.yml`:

```yaml
services:
  # ... existing services ...
  
  cron-jobs:
    build:
      context: ./backend
      dockerfile: Dockerfile.jobs
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=civicweave
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD}
      - TOP_K_CANDIDATES=10
      - MIN_MATCH_SCORE=0.6
    depends_on:
      - postgres
    volumes:
      - ./backend/jobs:/app/jobs
```

## Monitoring

Check job logs for:
- Number of projects processed
- Number of candidates notified
- Any errors or warnings
- Execution time

Example output:
```
=== Project Candidate Notification System ===
Started at 2025-10-13 09:00:00
Batch ID: 123e4567-e89b-12d3-a456-426614174000
Configuration: TOP_K=10, MIN_SCORE=0.6
Found 5 recruiting/active projects

📋 Processing project: Community Garden Project (abc123...)
  ✓ Found 8 top candidates
    → Notified John Doe (85% match)
    → Notified Jane Smith (78% match)
    ...
  ✓ Notified team lead about 8 candidates

============================================================
Batch completed at 2025-10-13 09:00:15
Total candidates notified: 32
Total team leads notified: 5
============================================================
```

## Troubleshooting

### Job fails with database connection error
- Verify database credentials in environment variables
- Check if database is accessible from the job execution context
- Ensure PostgreSQL is running

### No candidates being notified
- Check if `volunteer_initiative_matches` table has data (run `calculate_matches.py` first)
- Verify `MIN_MATCH_SCORE` threshold is not too high
- Ensure projects have `project_status` set to 'recruiting' or 'active'
- Check if volunteers have `skills_visible` set to true

### Duplicate notifications
- The system tracks notifications in `candidate_notifications` table
- Each batch run has a unique batch ID to prevent duplicates within the same run
- If running multiple times per day, adjust batch ID logic or notification frequency

## Future Enhancements

- [ ] Add email notifications in addition to in-app messages
- [ ] Implement user preferences for notification frequency
- [ ] Add A/B testing for different notification message formats
- [ ] Create admin dashboard for monitoring job execution
- [ ] Implement retry logic for failed notifications

