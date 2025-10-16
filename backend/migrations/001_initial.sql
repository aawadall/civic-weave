-- UP
-- Initial database schema for CivicWeave

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- CREATE EXTENSION IF NOT EXISTS "postgis";  -- Commented out for local dev without PostGIS

-- Users table (unified authentication)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255), -- nullable for OAuth-only users
    email_verified BOOLEAN DEFAULT FALSE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'volunteer')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- OAuth accounts table (Google OAuth linking)
CREATE TABLE IF NOT EXISTS oauth_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_user_id)
);

-- Email verification tokens table
CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Password reset tokens table
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Volunteers table
CREATE TABLE IF NOT EXISTS volunteers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    location_lat DECIMAL(10, 8),
    location_lng DECIMAL(11, 8),
    location_address TEXT,
    skills JSONB DEFAULT '[]',
    availability JSONB DEFAULT '{}',
    consent_given BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Admins table
CREATE TABLE IF NOT EXISTS admins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initiatives table
CREATE TABLE IF NOT EXISTS initiatives (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    required_skills JSONB DEFAULT '[]',
    location_lat DECIMAL(10, 8),
    location_lng DECIMAL(11, 8),
    location_address TEXT,
    start_date DATE,
    end_date DATE,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'closed')),
    created_by_admin_id UUID NOT NULL REFERENCES admins(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Applications table
CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    initiative_id UUID NOT NULL REFERENCES initiatives(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    admin_notes TEXT,
    UNIQUE(volunteer_id, initiative_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_volunteers_skills ON volunteers USING GIN (skills);
-- CREATE INDEX IF NOT EXISTS idx_volunteers_location ON volunteers USING GIST (ST_Point(location_lng, location_lat));  -- Requires PostGIS
CREATE INDEX IF NOT EXISTS idx_initiatives_required_skills ON initiatives USING GIN (required_skills);
CREATE INDEX IF NOT EXISTS idx_initiatives_status ON initiatives(status);
-- CREATE INDEX IF NOT EXISTS idx_initiatives_location ON initiatives USING GIST (ST_Point(location_lng, location_lat));  -- Requires PostGIS
CREATE INDEX IF NOT EXISTS idx_applications_volunteer_id ON applications(volunteer_id);
CREATE INDEX IF NOT EXISTS idx_applications_initiative_id ON applications(initiative_id);
CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);

-- DOWN
-- Drop all tables in reverse order
DROP INDEX IF EXISTS idx_applications_status;
DROP INDEX IF EXISTS idx_applications_initiative_id;
DROP INDEX IF EXISTS idx_applications_volunteer_id;
-- DROP INDEX IF EXISTS idx_initiatives_location;  -- Commented out (PostGIS)
DROP INDEX IF EXISTS idx_initiatives_status;
DROP INDEX IF EXISTS idx_initiatives_required_skills;
-- DROP INDEX IF EXISTS idx_volunteers_location;  -- Commented out (PostGIS)
DROP INDEX IF EXISTS idx_volunteers_skills;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS initiatives;
DROP TABLE IF EXISTS admins;
DROP TABLE IF EXISTS volunteers;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS email_verification_tokens;
DROP TABLE IF EXISTS oauth_accounts;
DROP TABLE IF EXISTS users;
