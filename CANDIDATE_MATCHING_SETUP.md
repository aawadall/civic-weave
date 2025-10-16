# Candidate Matching & Notification System

This document describes the automated candidate matching and notification system for CivicWeave projects.

## Overview

The system automatically:
1. Identifies top matching volunteers for recruiting projects
2. Notifies volunteers about projects they're a good fit for
3. Notifies project team leads about available candidates
4. Allows team leads to edit their projects
5. Enables volunteers to apply to projects

## Components

### 1. Database Migration

**File**: `backend/migrations/009_candidate_notifications.sql`

Creates the `candidate_notifications` table to track which volunteers have been notified about which projects. This prevents duplicate notifications.

**Schema**:
```sql
candidate_notifications (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects,
    volunteer_id UUID REFERENCES volunteers,
    match_score DECIMAL(5,4),
    notified_at TIMESTAMP,
    notification_batch_id UUID,
    UNIQUE(project_id, volunteer_id, notification_batch_id)
)
```

### 2. Python Batch Job

**File**: `backend/jobs/notify_project_matches.py`

Daily job that:
- Queries all projects with status 'recruiting' or 'active'
- For each project, finds top K candidates (default: 10) with match score ≥ 0.6
- Sends notification messages to candidates via `project_messages` table
- Sends summary to team lead listing all top candidates
- Records notifications in `candidate_notifications` table

**Configuration via Environment Variables**:
```bash
TOP_K_CANDIDATES=10          # Number of top candidates to notify per project
MIN_MATCH_SCORE=0.6          # Minimum match score (0.0-1.0)
SYSTEM_USER_ID=00000000-0000-0000-0000-000000000000
```

### 3. Team Lead Project Editing

**File**: `backend/handlers/project.go` (UpdateProject handler)

Modified to allow:
- Admins can edit any project
- Team leads can only edit their own projects
- Uses `IsTeamLead()` service method to verify ownership

**Authorization Flow**:
1. Check if user is authenticated
2. Verify user is admin OR team lead of the specific project
3. If authorized, allow all field updates
4. Return 403 Forbidden if not authorized

### 4. Volunteer Applications

**File**: `backend/handlers/application.go` (CreateApplication handler)

Existing functionality verified:
- Volunteers can apply via `POST /api/applications`
- Applications default to "pending" status
- Duplicate applications are prevented
- Database trigger auto-enrolls volunteers when application is accepted

## Setup Instructions

### Step 1: Run Database Migration

```bash
cd backend
make db-migrate
```

Or manually:
```bash
psql -h localhost -U postgres -d civicweave -f backend/migrations/009_candidate_notifications.sql
```

### Step 2: Configure Environment Variables

Add to `backend/.env`:
```bash
# Matching configuration
TOP_K_CANDIDATES=10
MIN_MATCH_SCORE=0.6
SYSTEM_USER_ID=00000000-0000-0000-0000-000000000000

# Database (if not already set)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=civicweave
DB_USER=postgres
DB_PASSWORD=your_password
```

### Step 3: Set Up Python Environment

```bash
cd backend
python3 -m venv venv
source venv/bin/activate
pip install psycopg2-binary python-dotenv
```

### Step 4: Test the Notification Job

Run manually to verify:
```bash
cd backend
./jobs/run_daily_matching.sh
```

Expected output:
```
=== Project Candidate Notification System ===
Started at 2025-10-13 09:00:00
Batch ID: abc123...
Configuration: TOP_K=10, MIN_SCORE=0.6
Found X recruiting/active projects

📋 Processing project: Community Garden (...)
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

### Step 5: Schedule Daily Job

#### Option A: Cron (Linux/Mac)

Add to crontab (`crontab -e`):
```cron
# Run daily at 9 AM
0 9 * * * cd /path/to/CivicWeave/backend && ./jobs/run_daily_matching.sh >> /var/log/civicweave/notify_matches.log 2>&1
```

#### Option B: Docker

Add to `docker-compose.yml`:
```yaml
cron-jobs:
  build:
    context: ./backend
    dockerfile: Dockerfile
  command: sh -c "while true; do ./jobs/run_daily_matching.sh && sleep 86400; done"
  environment:
    - DB_HOST=postgres
    - DB_PORT=5432
    - TOP_K_CANDIDATES=10
    - MIN_MATCH_SCORE=0.6
  depends_on:
    - postgres
```

#### Option C: GitHub Actions (Cloud Run)

Create `.github/workflows/daily-matching.yml`:
```yaml
name: Daily Candidate Matching
on:
  schedule:
    - cron: '0 9 * * *'  # 9 AM UTC daily
  workflow_dispatch:  # Allow manual trigger

jobs:
  notify-matches:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
        with:
          python-version: '3.10'
      - name: Install dependencies
        run: |
          pip install psycopg2-binary python-dotenv
      - name: Run notification job
        env:
          DB_HOST: ${{ secrets.DB_HOST }}
          DB_PORT: ${{ secrets.DB_PORT }}
          DB_NAME: ${{ secrets.DB_NAME }}
          DB_USER: ${{ secrets.DB_USER }}
          DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
          TOP_K_CANDIDATES: 10
          MIN_MATCH_SCORE: 0.6
        run: |
          cd backend
          python3 jobs/notify_project_matches.py
```

## Testing

### 1. Test Team Lead Project Editing

```bash
# As team lead (should succeed)
curl -X PUT http://localhost:8080/api/projects/{project_id} \
  -H "Authorization: Bearer {team_lead_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Title",
    "description": "Updated description"
  }'

# As non-team-lead volunteer (should fail with 403)
curl -X PUT http://localhost:8080/api/projects/{project_id} \
  -H "Authorization: Bearer {other_user_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Title"
  }'
```

Expected responses:
- Team lead: `200 OK` with updated project
- Non-team-lead: `403 Forbidden` with error message

### 2. Test Volunteer Application

```bash
# Apply to a project
curl -X POST http://localhost:8080/api/applications \
  -H "Authorization: Bearer {volunteer_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "initiative_id": "{project_id}",
    "message": "I would love to contribute to this project!"
  }'
```

Expected: `201 Created` with application object

### 3. Test Notification Job

Prerequisites:
- Run `calculate_matches.py` first to populate match scores
- Ensure at least one project has status 'recruiting' or 'active'
- Ensure volunteers have `skills_visible = true`

```bash
cd backend
python3 jobs/notify_project_matches.py
```

Verify:
1. Check console output for candidates notified
2. Query `project_messages` table for new notifications
3. Query `candidate_notifications` table for tracking records

```sql
-- Check recent notifications
SELECT * FROM project_messages 
WHERE created_at > NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC;

-- Check tracking records
SELECT * FROM candidate_notifications
WHERE notified_at > NOW() - INTERVAL '1 hour';
```

### 4. Verify No Duplicate Notifications

Run the job twice in succession:
```bash
python3 jobs/notify_project_matches.py
python3 jobs/notify_project_matches.py
```

Second run should show: "No new candidates found" for all projects.

## Monitoring

### Check Job Execution Logs

```bash
# If running via cron
tail -f /var/log/civicweave/notify_matches.log

# If running in Docker
docker logs civicweave-cron-jobs -f
```

### Database Queries for Monitoring

```sql
-- Total notifications sent
SELECT COUNT(*) FROM candidate_notifications;

-- Notifications by project
SELECT 
    p.title,
    COUNT(cn.id) as notification_count,
    MAX(cn.notified_at) as last_notified
FROM candidate_notifications cn
JOIN projects p ON cn.project_id = p.id
GROUP BY p.id, p.title
ORDER BY notification_count DESC;

-- Top matched volunteers
SELECT 
    v.name,
    COUNT(cn.id) as projects_matched,
    AVG(cn.match_score) as avg_match_score
FROM candidate_notifications cn
JOIN volunteers v ON cn.volunteer_id = v.id
GROUP BY v.id, v.name
ORDER BY projects_matched DESC
LIMIT 10;

-- Recent notification activity
SELECT 
    DATE(notified_at) as notification_date,
    COUNT(*) as notifications_sent
FROM candidate_notifications
WHERE notified_at > NOW() - INTERVAL '30 days'
GROUP BY DATE(notified_at)
ORDER BY notification_date DESC;
```

## Troubleshooting

### Issue: No candidates being notified

**Possible causes**:
1. `volunteer_initiative_matches` table is empty
   - **Solution**: Run `calculate_matches.py` first
2. No projects have status 'recruiting' or 'active'
   - **Solution**: Update project status in database
3. Match scores are below threshold
   - **Solution**: Lower `MIN_MATCH_SCORE` environment variable
4. Volunteers have `skills_visible = false`
   - **Solution**: Update volunteer records

### Issue: Job fails with database connection error

**Possible causes**:
1. Incorrect database credentials
   - **Solution**: Verify environment variables
2. Database not accessible
   - **Solution**: Check network/firewall settings
3. PostgreSQL not running
   - **Solution**: Start PostgreSQL service

### Issue: Duplicate notifications

**Check**:
- Verify `UNIQUE` constraint on `candidate_notifications` table
- Check if multiple job instances are running simultaneously
- Review batch ID generation logic

## Future Enhancements

- [ ] Add email notifications (in addition to in-app messages)
- [ ] Implement user preferences for notification frequency
- [ ] Add unsubscribe option for match notifications
- [ ] Create admin dashboard for monitoring
- [ ] Implement retry logic for failed notifications
- [ ] Add notification templates with customization options
- [ ] Track notification open/click rates
- [ ] A/B test different notification message formats

## API Endpoints

### Project Management
- `GET /api/projects` - List all projects
- `GET /api/projects/:id` - Get project details
- `PUT /api/projects/:id` - Update project (team lead or admin only)
- `GET /api/projects/:id/details` - Get project with team members and applications

### Applications
- `POST /api/applications` - Create application (volunteer only)
- `GET /api/applications` - List applications (with filters)
- `PUT /api/applications/:id` - Update application status (admin/team lead)

### Messages
- `GET /api/projects/:id/messages` - Get project messages
- `POST /api/projects/:id/messages` - Send message
- `GET /api/messages/unread` - Get unread message count

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Review logs for error messages
3. Verify database schema is up to date
4. Ensure all environment variables are set correctly



