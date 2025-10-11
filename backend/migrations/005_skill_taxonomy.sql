-- UP
-- Skill Taxonomy Migration: Sparse Vector System
-- Replaces dense embedding approach with variable-length sparse vectors

-- Global skill taxonomy (ordered, growing list of all skills)
CREATE TABLE skill_taxonomy (
    id SERIAL PRIMARY KEY,
    skill_name VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Junction table: volunteer → skills WITH WEIGHTS
-- Weight = 0.5 initially, adjusted by TLs based on performance
CREATE TABLE volunteer_skills (
    volunteer_id UUID REFERENCES volunteers(id) ON DELETE CASCADE,
    skill_id INTEGER REFERENCES skill_taxonomy(id) ON DELETE CASCADE,
    skill_weight DECIMAL(3,2) NOT NULL DEFAULT 0.5 CHECK (skill_weight >= 0.1 AND skill_weight <= 1.0),
    proficiency_level VARCHAR(20) CHECK (proficiency_level IN ('beginner', 'intermediate', 'advanced', 'expert')),
    years_experience INTEGER,
    last_used_year INTEGER,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (volunteer_id, skill_id)
);

-- Junction table: initiative → required skills  
CREATE TABLE initiative_required_skills (
    initiative_id UUID REFERENCES initiatives(id) ON DELETE CASCADE,
    skill_id INTEGER REFERENCES skill_taxonomy(id) ON DELETE CASCADE,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (initiative_id, skill_id)
);

-- Pre-calculated match scores (populated by hourly Python batch job)
CREATE TABLE volunteer_initiative_matches (
    volunteer_id UUID REFERENCES volunteers(id) ON DELETE CASCADE,
    initiative_id UUID REFERENCES initiatives(id) ON DELETE CASCADE,
    match_score DECIMAL(5,4) NOT NULL CHECK (match_score >= 0.0 AND match_score <= 1.0),
    jaccard_index DECIMAL(5,4) NOT NULL CHECK (jaccard_index >= 0.0 AND jaccard_index <= 1.0),
    matched_skill_ids INTEGER[] NOT NULL,
    matched_skill_count INTEGER NOT NULL,
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (volunteer_id, initiative_id)
);

-- Track weight adjustments by TLs (audit trail)
CREATE TABLE volunteer_skill_weight_overrides (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    volunteer_id UUID REFERENCES volunteers(id) ON DELETE CASCADE,
    skill_id INTEGER REFERENCES skill_taxonomy(id) ON DELETE CASCADE,
    original_weight DECIMAL(3,2) NOT NULL,
    override_weight DECIMAL(3,2) NOT NULL CHECK (override_weight >= 0.1 AND override_weight <= 1.0),
    adjusted_by_admin_id UUID REFERENCES admins(id),
    initiative_id UUID REFERENCES initiatives(id),  -- context: which project prompted adjustment
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for fast lookups and matching queries
CREATE INDEX idx_volunteer_skills_volunteer ON volunteer_skills(volunteer_id);
CREATE INDEX idx_volunteer_skills_skill ON volunteer_skills(skill_id);
CREATE INDEX idx_initiative_skills_initiative ON initiative_required_skills(initiative_id);
CREATE INDEX idx_initiative_skills_skill ON initiative_required_skills(skill_id);
CREATE INDEX idx_matches_volunteer ON volunteer_initiative_matches(volunteer_id, match_score DESC);
CREATE INDEX idx_matches_initiative ON volunteer_initiative_matches(initiative_id, match_score DESC);
CREATE INDEX idx_matches_score ON volunteer_initiative_matches(match_score DESC);
CREATE INDEX idx_weight_overrides_volunteer ON volunteer_skill_weight_overrides(volunteer_id);
CREATE INDEX idx_weight_overrides_admin ON volunteer_skill_weight_overrides(adjusted_by_admin_id);

-- Seed with predefined SKILL_OPTIONS (from CreateInitiativePage.jsx)
INSERT INTO skill_taxonomy (skill_name) VALUES 
  ('Event Planning'),
  ('Marketing'),
  ('Social Media'),
  ('Content Creation'),
  ('Graphic Design'),
  ('Photography'),
  ('Videography'),
  ('Writing'),
  ('Public Speaking'),
  ('Community Outreach'),
  ('Fundraising'),
  ('Project Management'),
  ('Data Analysis'),
  ('Research'),
  ('Teaching/Training'),
  ('Translation'),
  ('Legal'),
  ('Accounting'),
  ('Technology'),
  ('Web Development'),
  ('Mobile App Development'),
  ('Cybersecurity'),
  ('Healthcare'),
  ('Mental Health'),
  ('Elderly Care'),
  ('Childcare'),
  ('Environmental'),
  ('Sustainability'),
  ('Gardening'),
  ('Construction'),
  ('Handyman'),
  ('Transportation'),
  ('Cooking'),
  ('Catering'),
  ('Music'),
  ('Art'),
  ('Theater'),
  ('Sports'),
  ('Fitness'),
  ('Other');

-- DOWN
DROP TABLE IF EXISTS volunteer_skill_weight_overrides;
DROP TABLE IF EXISTS volunteer_initiative_matches;
DROP TABLE IF EXISTS initiative_required_skills;
DROP TABLE IF EXISTS volunteer_skills;
DROP TABLE IF EXISTS skill_taxonomy;

