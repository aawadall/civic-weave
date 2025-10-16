# Table Name Fix - volunteer_project_matches

## Issue
The batch jobs were referencing incorrect table names from before the migration from "initiatives" to "projects".

## What Was Fixed

### ❌ Old (Incorrect) Table Names
- `volunteer_initiative_matches` 
- `initiatives`
- `initiative_required_skills`

### ✅ New (Correct) Table Names
- `volunteer_project_matches` ✓
- `projects` ✓
- `project_required_skills` ✓

## Files Updated

### 1. `backend/jobs/calculate_matches.py`
**Fixed**:
- Changed query from `initiatives` table to `projects` table
- Changed `initiative_required_skills` to `project_required_skills`
- Changed `volunteer_initiative_matches` to `volunteer_project_matches`
- Updated status check to use `project_status IN ('recruiting', 'active')`
- Updated all variable names for clarity

### 2. `backend/jobs/notify_project_matches.py`
**Fixed**:
- Changed `volunteer_initiative_matches` to `volunteer_project_matches`
- Changed `initiative_id` column to `project_id`

### 3. `setup_candidate_matching_clean.sql`
**Fixed**:
- Updated all queries to use `volunteer_project_matches`
- Changed `initiative_id` to `project_id` in WHERE clauses

## Database Table Structure

The correct table (from migration 005_skill_taxonomy.sql):

```sql
CREATE TABLE volunteer_project_matches (
    volunteer_id UUID REFERENCES volunteers(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    match_score DECIMAL(5,4) NOT NULL CHECK (match_score >= 0.0 AND match_score <= 1.0),
    jaccard_index DECIMAL(5,4) NOT NULL CHECK (jaccard_index >= 0.0 AND jaccard_index <= 1.0),
    matched_skill_ids INTEGER[] NOT NULL,
    matched_skill_count INTEGER NOT NULL,
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (volunteer_id, project_id)
);
```

## Verification

After running the fixes, verify with:

```sql
-- Check table exists
SELECT COUNT(*) FROM volunteer_project_matches;

-- Check table structure
\d volunteer_project_matches

-- Should return 0 rows initially (populated by calculate_matches.py)
SELECT * FROM volunteer_project_matches LIMIT 5;
```

## Next Steps

1. ✅ All table names corrected
2. ✅ SQL script updated
3. ✅ Python scripts updated
4. 🔄 Ready to run: `make job-calculate-matches`
5. 🔄 Then run: `make job-notify-matches`

## Status

✅ **FIXED** - All files now use correct table names matching the database schema.



