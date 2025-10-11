-- UP
-- Rename Initiatives to Projects Migration

-- First, add new project-specific columns to initiatives table
ALTER TABLE initiatives ADD COLUMN team_lead_id UUID REFERENCES users(id);
ALTER TABLE initiatives ADD COLUMN project_status VARCHAR(20) DEFAULT 'draft' CHECK (project_status IN ('draft', 'recruiting', 'active', 'completed', 'archived'));

-- Rename the table from initiatives to projects
ALTER TABLE initiatives RENAME TO projects;

-- Update the initiative_skill_requirements table name and references
ALTER TABLE initiative_skill_requirements RENAME TO project_skill_requirements;
ALTER TABLE project_skill_requirements RENAME COLUMN initiative_id TO project_id;
ALTER TABLE project_skill_requirements DROP CONSTRAINT initiative_skill_requirements_initiative_id_fkey;
ALTER TABLE project_skill_requirements ADD CONSTRAINT project_skill_requirements_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

-- Update applications table references
ALTER TABLE applications RENAME COLUMN initiative_id TO project_id;
ALTER TABLE applications DROP CONSTRAINT applications_initiative_id_fkey;
ALTER TABLE applications ADD CONSTRAINT applications_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

-- Update volunteer_ratings table references (from previous migration)
ALTER TABLE volunteer_ratings RENAME COLUMN project_id TO project_id_temp;
ALTER TABLE volunteer_ratings ADD COLUMN project_id UUID REFERENCES projects(id) ON DELETE SET NULL;
UPDATE volunteer_ratings SET project_id = project_id_temp;
ALTER TABLE volunteer_ratings DROP COLUMN project_id_temp;

-- Update project_team_members table references (from previous migration)
ALTER TABLE project_team_members RENAME COLUMN project_id TO project_id_temp;
ALTER TABLE project_team_members ADD COLUMN project_id UUID REFERENCES projects(id) ON DELETE CASCADE;
UPDATE project_team_members SET project_id = project_id_temp;
ALTER TABLE project_team_members DROP COLUMN project_id_temp;
ALTER TABLE project_team_members DROP CONSTRAINT project_team_members_project_id_temp_fkey;
ALTER TABLE project_team_members ADD CONSTRAINT project_team_members_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

-- Update indexes
DROP INDEX IF EXISTS idx_initiatives_location;
DROP INDEX IF EXISTS idx_initiatives_status;
DROP INDEX IF EXISTS idx_initiatives_required_skills;
DROP INDEX IF EXISTS idx_initiative_skill_requirements_initiative_id;
DROP INDEX IF EXISTS idx_initiative_skill_requirements_embedding;
DROP INDEX IF EXISTS idx_applications_initiative_id;

-- Create new indexes with project naming
CREATE INDEX idx_projects_location ON projects USING GIST (ST_Point(location_lng, location_lat));
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_project_status ON projects(project_status);
CREATE INDEX idx_projects_required_skills ON projects USING GIN (required_skills);
CREATE INDEX idx_projects_team_lead_id ON projects(team_lead_id);
CREATE INDEX idx_projects_created_by_admin_id ON projects(created_by_admin_id);

CREATE INDEX idx_project_skill_requirements_project_id ON project_skill_requirements(project_id);
CREATE INDEX idx_project_skill_requirements_embedding ON project_skill_requirements USING ivfflat (required_vector vector_cosine_ops) WITH (lists = 100);

CREATE INDEX idx_applications_project_id ON applications(project_id);

-- Update the updated_at trigger
DROP TRIGGER IF EXISTS update_initiative_skill_requirements_updated_at ON project_skill_requirements;
CREATE TRIGGER update_project_skill_requirements_updated_at BEFORE UPDATE ON project_skill_requirements FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- DOWN
-- Reverse the migration
DROP TRIGGER IF EXISTS update_project_skill_requirements_updated_at ON project_skill_requirements;
CREATE TRIGGER update_initiative_skill_requirements_updated_at BEFORE UPDATE ON initiative_skill_requirements FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP INDEX IF EXISTS idx_applications_project_id;
DROP INDEX IF EXISTS idx_project_skill_requirements_embedding;
DROP INDEX IF EXISTS idx_project_skill_requirements_project_id;
DROP INDEX IF EXISTS idx_projects_created_by_admin_id;
DROP INDEX IF EXISTS idx_projects_team_lead_id;
DROP INDEX IF EXISTS idx_projects_required_skills;
DROP INDEX IF EXISTS idx_projects_project_status;
DROP INDEX IF EXISTS idx_projects_status;
DROP INDEX IF EXISTS idx_projects_location;

-- Restore original indexes
CREATE INDEX idx_initiative_skill_requirements_embedding ON initiative_skill_requirements USING ivfflat (required_vector vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_initiative_skill_requirements_initiative_id ON initiative_skill_requirements(initiative_id);
CREATE INDEX idx_applications_initiative_id ON applications(initiative_id);
CREATE INDEX idx_initiatives_required_skills ON initiatives USING GIN (required_skills);
CREATE INDEX idx_initiatives_status ON initiatives(status);
CREATE INDEX idx_initiatives_location ON initiatives USING GIST (ST_Point(location_lng, location_lat));

-- Restore table names and column references
ALTER TABLE project_team_members DROP CONSTRAINT project_team_members_project_id_fkey;
ALTER TABLE project_team_members ADD CONSTRAINT project_team_members_project_id_temp_fkey FOREIGN KEY (project_id) REFERENCES initiatives(id) ON DELETE CASCADE;
ALTER TABLE project_team_members ADD COLUMN project_id_temp UUID REFERENCES initiatives(id) ON DELETE CASCADE;
UPDATE project_team_members SET project_id_temp = project_id;
ALTER TABLE project_team_members DROP COLUMN project_id;
ALTER TABLE project_team_members RENAME COLUMN project_id_temp TO project_id;

ALTER TABLE volunteer_ratings DROP CONSTRAINT volunteer_ratings_project_id_fkey;
ALTER TABLE volunteer_ratings ADD COLUMN project_id_temp UUID REFERENCES initiatives(id) ON DELETE SET NULL;
UPDATE volunteer_ratings SET project_id_temp = project_id;
ALTER TABLE volunteer_ratings DROP COLUMN project_id;
ALTER TABLE volunteer_ratings RENAME COLUMN project_id_temp TO project_id;

ALTER TABLE applications DROP CONSTRAINT applications_project_id_fkey;
ALTER TABLE applications ADD CONSTRAINT applications_initiative_id_fkey FOREIGN KEY (project_id) REFERENCES initiatives(id) ON DELETE CASCADE;
ALTER TABLE applications RENAME COLUMN project_id TO initiative_id;

ALTER TABLE project_skill_requirements DROP CONSTRAINT project_skill_requirements_project_id_fkey;
ALTER TABLE project_skill_requirements ADD CONSTRAINT initiative_skill_requirements_initiative_id_fkey FOREIGN KEY (project_id) REFERENCES initiatives(id) ON DELETE CASCADE;
ALTER TABLE project_skill_requirements RENAME COLUMN project_id TO initiative_id;

ALTER TABLE project_skill_requirements RENAME TO initiative_skill_requirements;
ALTER TABLE projects RENAME TO initiatives;

-- Remove added columns
ALTER TABLE initiatives DROP COLUMN IF EXISTS project_status;
ALTER TABLE initiatives DROP COLUMN IF EXISTS team_lead_id;
