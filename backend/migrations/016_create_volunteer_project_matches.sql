-- Migration: Create volunteer_project_matches table
-- Date: 2025-01-XX
-- Description: Store pre-calculated skill match scores for projects

-- Up Migration
CREATE TABLE IF NOT EXISTS volunteer_project_matches (
  volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  match_score DECIMAL(5,4) NOT NULL CHECK (match_score >= 0 AND match_score <= 1),
  jaccard_index DECIMAL(5,4) CHECK (jaccard_index >= 0 AND jaccard_index <= 1),
  matched_skill_ids INTEGER[] NOT NULL DEFAULT '{}',
  matched_skill_count INTEGER NOT NULL DEFAULT 0,
  calculated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (volunteer_id, project_id)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_vpm_project_score 
  ON volunteer_project_matches(project_id, match_score DESC);

CREATE INDEX IF NOT EXISTS idx_vpm_volunteer_score 
  ON volunteer_project_matches(volunteer_id, match_score DESC);

CREATE INDEX IF NOT EXISTS idx_vpm_calculated_at 
  ON volunteer_project_matches(calculated_at DESC);

-- Comments for documentation
COMMENT ON TABLE volunteer_project_matches IS 'Pre-calculated skill match scores between volunteers and projects';
COMMENT ON COLUMN volunteer_project_matches.match_score IS 'Cosine similarity score (0-1)';
COMMENT ON COLUMN volunteer_project_matches.jaccard_index IS 'Jaccard index of skill overlap (0-1)';
