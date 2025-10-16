-- UP
-- RBAC Multi-Role System Migration

-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    permissions JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create user_roles table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by_admin_id UUID REFERENCES users(id),
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- Create volunteer_ratings table for scorecard system
CREATE TABLE IF NOT EXISTS volunteer_ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    skill_claim_id UUID REFERENCES skill_claims(id) ON DELETE SET NULL,
    rated_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES initiatives(id) ON DELETE SET NULL, -- Will be updated to projects in next migration
    rating VARCHAR(10) NOT NULL CHECK (rating IN ('up', 'down', 'neutral')),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create campaigns table for outreach
CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    target_roles JSONB DEFAULT '[]',
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'sending', 'sent', 'failed')),
    email_subject VARCHAR(255) NOT NULL,
    email_body TEXT NOT NULL,
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    scheduled_at TIMESTAMP,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create campaign_recipients table
CREATE TABLE IF NOT EXISTS campaign_recipients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sent_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'delivered', 'opened', 'clicked', 'failed')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(campaign_id, user_id)
);

-- Create project_team_members table
CREATE TABLE IF NOT EXISTS project_team_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES initiatives(id) ON DELETE CASCADE, -- Will be updated to projects in next migration
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'invited' CHECK (status IN ('invited', 'active', 'completed', 'removed')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, volunteer_id)
);

-- Insert default roles (idempotent - only insert if not exists)
INSERT INTO roles (name, description, permissions) VALUES
('volunteer', 'Can view and apply to projects, manage own profile', '["view_projects", "apply_to_projects", "view_own_profile", "rate_volunteers"]'),
('team_lead', 'Can create/edit projects, manage teams, rate volunteers', '["view_projects", "create_projects", "edit_projects", "manage_teams", "rate_volunteers", "view_volunteers"]'),
('campaign_manager', 'Can create and send email campaigns', '["view_projects", "view_volunteers", "create_campaigns", "send_campaigns", "view_campaigns"]'),
('admin', 'Full system access including user and role management', '["*"]')
ON CONFLICT (name) DO NOTHING;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_assigned_by ON user_roles(assigned_by_admin_id);

CREATE INDEX IF NOT EXISTS idx_volunteer_ratings_volunteer_id ON volunteer_ratings(volunteer_id);
CREATE INDEX IF NOT EXISTS idx_volunteer_ratings_rated_by ON volunteer_ratings(rated_by_user_id);
CREATE INDEX IF NOT EXISTS idx_volunteer_ratings_project_id ON volunteer_ratings(project_id);
CREATE INDEX IF NOT EXISTS idx_volunteer_ratings_skill_claim_id ON volunteer_ratings(skill_claim_id);

CREATE INDEX IF NOT EXISTS idx_campaigns_created_by ON campaigns(created_by_user_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
CREATE INDEX IF NOT EXISTS idx_campaigns_scheduled_at ON campaigns(scheduled_at);

CREATE INDEX IF NOT EXISTS idx_campaign_recipients_campaign_id ON campaign_recipients(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_recipients_user_id ON campaign_recipients(user_id);
CREATE INDEX IF NOT EXISTS idx_campaign_recipients_status ON campaign_recipients(status);

CREATE INDEX IF NOT EXISTS idx_project_team_members_project_id ON project_team_members(project_id);
CREATE INDEX IF NOT EXISTS idx_project_team_members_volunteer_id ON project_team_members(volunteer_id);
CREATE INDEX IF NOT EXISTS idx_project_team_members_status ON project_team_members(status);

-- Add triggers for updated_at timestamps (idempotent)
DROP TRIGGER IF EXISTS update_campaigns_updated_at ON campaigns;
CREATE TRIGGER update_campaigns_updated_at BEFORE UPDATE ON campaigns FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_project_team_members_updated_at ON project_team_members;
CREATE TRIGGER update_project_team_members_updated_at BEFORE UPDATE ON project_team_members FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- DOWN
-- Drop all tables and indexes in reverse order
DROP TRIGGER IF EXISTS update_project_team_members_updated_at ON project_team_members;
DROP TRIGGER IF EXISTS update_campaigns_updated_at ON campaigns;

DROP INDEX IF EXISTS idx_project_team_members_status;
DROP INDEX IF EXISTS idx_project_team_members_volunteer_id;
DROP INDEX IF EXISTS idx_project_team_members_project_id;

DROP INDEX IF EXISTS idx_campaign_recipients_status;
DROP INDEX IF EXISTS idx_campaign_recipients_user_id;
DROP INDEX IF EXISTS idx_campaign_recipients_campaign_id;

DROP INDEX IF EXISTS idx_campaigns_scheduled_at;
DROP INDEX IF EXISTS idx_campaigns_status;
DROP INDEX IF EXISTS idx_campaigns_created_by;

DROP INDEX IF EXISTS idx_volunteer_ratings_skill_claim_id;
DROP INDEX IF EXISTS idx_volunteer_ratings_project_id;
DROP INDEX IF EXISTS idx_volunteer_ratings_rated_by;
DROP INDEX IF EXISTS idx_volunteer_ratings_volunteer_id;

DROP INDEX IF EXISTS idx_user_roles_assigned_by;
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_user_id;

DROP TABLE IF EXISTS project_team_members;
DROP TABLE IF EXISTS campaign_recipients;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS volunteer_ratings;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
