-- UP
-- Skill Vector Data Model Migration

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS vector;

-- Add skills visibility toggle to existing volunteers table
ALTER TABLE volunteers ADD COLUMN skills_visible BOOLEAN DEFAULT true;

-- Skill claims table - volunteer free-text skill descriptions with embeddings
CREATE TABLE skill_claims (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    claim_text TEXT NOT NULL,
    embedding vector(384) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Skill weights table - system-managed weights per skill claim
CREATE TABLE skill_weights (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    skill_claim_id UUID NOT NULL REFERENCES skill_claims(id) ON DELETE CASCADE,
    weight DECIMAL(3,2) NOT NULL DEFAULT 0.5 CHECK (weight >= 0.1 AND weight <= 1.0),
    updated_by_admin_id UUID REFERENCES admins(id),
    last_task_id UUID, -- Reference to task/initiative that triggered weight update
    update_reason VARCHAR(100), -- 'initial', 'task_completion', 'admin_review', 'manual'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(skill_claim_id)
);

-- Volunteer skill vectors table - aggregated flat vector per volunteer
CREATE TABLE volunteer_skill_vectors (
    volunteer_id UUID PRIMARY KEY REFERENCES volunteers(id) ON DELETE CASCADE,
    aggregated_vector vector(384) NOT NULL,
    location_point GEOMETRY(POINT, 4326),
    last_aggregated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initiative skill requirements table - required skill vectors for initiatives
CREATE TABLE initiative_skill_requirements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    initiative_id UUID NOT NULL REFERENCES initiatives(id) ON DELETE CASCADE,
    required_vector vector(384) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(initiative_id)
);

-- Indexes for vector similarity search
CREATE INDEX idx_skill_claims_volunteer_id ON skill_claims(volunteer_id);
CREATE INDEX idx_skill_claims_embedding ON skill_claims USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_skill_claims_active ON skill_claims(is_active) WHERE is_active = true;

-- Indexes for skill weights
CREATE INDEX idx_skill_weights_claim_id ON skill_weights(skill_claim_id);
CREATE INDEX idx_skill_weights_weight ON skill_weights(weight);

-- Indexes for volunteer skill vectors
CREATE INDEX idx_volunteer_skill_vectors_embedding ON volunteer_skill_vectors USING ivfflat (aggregated_vector vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_volunteer_skill_vectors_location ON volunteer_skill_vectors USING GIST (location_point);

-- Indexes for initiative skill requirements
CREATE INDEX idx_initiative_skill_requirements_initiative_id ON initiative_skill_requirements(initiative_id);
CREATE INDEX idx_initiative_skill_requirements_embedding ON initiative_skill_requirements USING ivfflat (required_vector vector_cosine_ops) WITH (lists = 100);

-- Triggers for updating timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_skill_claims_updated_at BEFORE UPDATE ON skill_claims FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_skill_weights_updated_at BEFORE UPDATE ON skill_weights FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_volunteer_skill_vectors_updated_at BEFORE UPDATE ON volunteer_skill_vectors FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_initiative_skill_requirements_updated_at BEFORE UPDATE ON initiative_skill_requirements FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- DOWN
-- Drop all tables and extensions in reverse order
DROP TRIGGER IF EXISTS update_initiative_skill_requirements_updated_at ON initiative_skill_requirements;
DROP TRIGGER IF EXISTS update_volunteer_skill_vectors_updated_at ON volunteer_skill_vectors;
DROP TRIGGER IF EXISTS update_skill_weights_updated_at ON skill_weights;
DROP TRIGGER IF EXISTS update_skill_claims_updated_at ON skill_claims;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_initiative_skill_requirements_embedding;
DROP INDEX IF EXISTS idx_initiative_skill_requirements_initiative_id;
DROP INDEX IF EXISTS idx_volunteer_skill_vectors_location;
DROP INDEX IF EXISTS idx_volunteer_skill_vectors_embedding;
DROP INDEX IF EXISTS idx_skill_weights_weight;
DROP INDEX IF EXISTS idx_skill_weights_claim_id;
DROP INDEX IF EXISTS idx_skill_claims_active;
DROP INDEX IF EXISTS idx_skill_claims_embedding;
DROP INDEX IF EXISTS idx_skill_claims_volunteer_id;

DROP TABLE IF EXISTS initiative_skill_requirements;
DROP TABLE IF EXISTS volunteer_skill_vectors;
DROP TABLE IF EXISTS skill_weights;
DROP TABLE IF EXISTS skill_claims;

ALTER TABLE volunteers DROP COLUMN IF EXISTS skills_visible;
