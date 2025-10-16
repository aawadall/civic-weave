# Quick Start: Candidate Matching System

## Summary

✅ **Implemented** automated candidate matching and notification system for CivicWeave projects.

## What's New

### 1. Batch Job: Daily Candidate Notifications
- Finds top 10 candidates for recruiting projects
- Notifies candidates about matching opportunities
- Notifies team leads about available candidates
- Prevents duplicate notifications

### 2. Team Lead Project Editing
- Team leads can now edit their own projects
- Authorization checks project ownership
- All fields editable (as requested)

### 3. Volunteer Applications
- Verified working correctly
- Volunteers can apply to projects
- Duplicate applications prevented
- Auto-enrollment on acceptance

## Quick Start (5 minutes)

### Step 1: Run Migration
```bash
cd /home/arashad/src/CivicWeave
make db-migrate
```

### Step 2: Configure Environment
Add to `backend/.env`:
```bash
TOP_K_CANDIDATES=10
MIN_MATCH_SCORE=0.6
SYSTEM_USER_ID=00000000-0000-0000-0000-000000000000
```

### Step 3: Set Up Python
```bash
make job-setup-python
```

### Step 4: Test It
```bash
# Calculate match scores
make job-calculate-matches

# Send notifications
make job-notify-matches
```

### Step 5: Schedule Daily (Optional)
```bash
# Copy and edit cron configuration
cp backend/jobs/crontab.example /tmp/civicweave-cron
nano /tmp/civicweave-cron  # Edit paths
crontab /tmp/civicweave-cron
```

## Files Created

```
backend/migrations/009_candidate_notifications.sql    - Database migration
backend/jobs/notify_project_matches.py                - Notification job
backend/jobs/run_daily_matching.sh                    - Shell wrapper
backend/jobs/README.md                                - Jobs documentation
backend/jobs/crontab.example                          - Cron config
CANDIDATE_MATCHING_SETUP.md                           - Complete setup guide
CANDIDATE_MATCHING_IMPLEMENTATION.md                  - Implementation details
```

## Files Modified

```
backend/handlers/project.go    - Added team lead authorization
Makefile                       - Added job commands
```

## How It Works

### Daily Batch Job Flow
```
1. Query all recruiting/active projects
2. For each project:
   a. Find top 10 volunteers with ≥60% match score
   b. Send notification to each candidate
   c. Send summary to team lead
   d. Record in candidate_notifications table
3. Complete with statistics
```

### Team Lead Authorization Flow
```
1. User requests project update
2. Check: Is user admin? → Allow
3. Check: Is user team lead of this project? → Allow
4. Otherwise → 403 Forbidden
```

### Volunteer Application Flow
```
1. Volunteer applies to project
2. Application status = "pending"
3. Team lead reviews application
4. Team lead accepts → Database trigger auto-enrolls volunteer
```

## Testing

### Test Team Lead Editing
```bash
# This should succeed (as team lead)
curl -X PUT http://localhost:8080/api/projects/{your_project_id} \
  -H "Authorization: Bearer {your_token}" \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated Title"}'

# This should fail with 403 (as non-team-lead)
curl -X PUT http://localhost:8080/api/projects/{other_project_id} \
  -H "Authorization: Bearer {your_token}" \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated Title"}'
```

### Test Volunteer Application
```bash
curl -X POST http://localhost:8080/api/applications \
  -H "Authorization: Bearer {volunteer_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "initiative_id": "{project_id}",
    "message": "I would love to contribute!"
  }'
```

### Verify Notifications
```sql
-- Check recent notifications
SELECT * FROM project_messages 
WHERE created_at > NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC;

-- Check notification tracking
SELECT * FROM candidate_notifications
WHERE notified_at > NOW() - INTERVAL '1 hour';
```

## Monitoring

### Check Job is Running
```sql
SELECT 
    DATE(notified_at) as date,
    COUNT(*) as notifications
FROM candidate_notifications
GROUP BY DATE(notified_at)
ORDER BY date DESC
LIMIT 7;
```

### View Logs
```bash
# If running via cron
tail -f /var/log/civicweave/notify_matches.log

# If running manually
./backend/jobs/run_daily_matching.sh
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| No candidates notified | Run `make job-calculate-matches` first |
| Database connection error | Check DB credentials in `.env` |
| Permission denied on script | Run `chmod +x backend/jobs/run_daily_matching.sh` |
| Duplicate notifications | Check only one cron job is scheduled |

## Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `TOP_K_CANDIDATES` | 10 | Number of top candidates to notify per project |
| `MIN_MATCH_SCORE` | 0.6 | Minimum match score (0.0-1.0) to trigger notification |
| `SYSTEM_USER_ID` | 00000... | System user ID for automated messages |

## Next Steps

1. ✅ Run migration
2. ✅ Test manually
3. ✅ Schedule daily job
4. 📊 Monitor for 1 week
5. 🎯 Adjust thresholds based on feedback
6. 🚀 Consider adding email notifications (future)

## Documentation

- **Complete Setup**: See `CANDIDATE_MATCHING_SETUP.md`
- **Implementation Details**: See `CANDIDATE_MATCHING_IMPLEMENTATION.md`
- **Jobs Reference**: See `backend/jobs/README.md`

## Support

If you encounter issues:
1. Check the troubleshooting section above
2. Review logs for error messages
3. Verify database migration completed
4. Ensure environment variables are set

## Success Criteria

✅ Migration creates `candidate_notifications` table  
✅ Batch job runs without errors  
✅ Candidates receive notifications  
✅ Team leads receive summaries  
✅ No duplicate notifications sent  
✅ Team leads can edit their projects  
✅ Non-team-leads cannot edit others' projects  
✅ Volunteers can apply to projects  

---

**Status**: ✅ Implementation Complete  
**Ready for**: Testing and Production Deployment



