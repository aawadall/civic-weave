-- Metadata Database Schema for Database Agent System
-- This database tracks deployment history, migration versions, and audit logs

-- Create the databases table to track managed databases
CREATE TABLE IF NOT EXISTS databases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    connection_string_encrypted TEXT NOT NULL,
    description TEXT,
    environment VARCHAR(50) DEFAULT 'production',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    tags TEXT[] DEFAULT '{}'
);

-- Create the deployments table to track deployment versions
CREATE TABLE IF NOT EXISTS deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    manifest_version VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, applied, failed, rolled_back
    applied_at TIMESTAMP,
    applied_by VARCHAR(255),
    checksum VARCHAR(64),
    execution_time_ms INTEGER,
    error_message TEXT,
    dry_run BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(database_id, version)
);

-- Create the migrations table to track individual migration execution
CREATE TABLE IF NOT EXISTS migrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, applied, failed, skipped
    execution_time_ms INTEGER,
    error_message TEXT,
    applied_at TIMESTAMP,
    rollback_at TIMESTAMP,
    UNIQUE(deployment_id, version)
);

-- Create the audit_log table for comprehensive audit trail
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID REFERENCES databases(id) ON DELETE SET NULL,
    deployment_id UUID REFERENCES deployments(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL, -- ping, compare, download, deploy, bootstrap, rollback
    user_agent VARCHAR(255),
    client_ip INET,
    request_id UUID,
    status_code INTEGER,
    error_message TEXT,
    execution_time_ms INTEGER,
    request_size_bytes INTEGER,
    response_size_bytes INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}'
);

-- Create the api_keys table for authentication tracking
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id VARCHAR(255) NOT NULL UNIQUE,
    public_key_hash VARCHAR(64) NOT NULL,
    description TEXT,
    permissions TEXT[] DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255)
);

-- Create the schema_snapshots table for schema drift detection
CREATE TABLE IF NOT EXISTS schema_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    deployment_id UUID REFERENCES deployments(id) ON DELETE SET NULL,
    schema_hash VARCHAR(64) NOT NULL,
    table_count INTEGER,
    index_count INTEGER,
    function_count INTEGER,
    trigger_count INTEGER,
    schema_dump TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_databases_name ON databases(name);
CREATE INDEX IF NOT EXISTS idx_databases_environment ON databases(environment);
CREATE INDEX IF NOT EXISTS idx_databases_active ON databases(is_active);

CREATE INDEX IF NOT EXISTS idx_deployments_database_id ON deployments(database_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
CREATE INDEX IF NOT EXISTS idx_deployments_applied_at ON deployments(applied_at);

CREATE INDEX IF NOT EXISTS idx_migrations_deployment_id ON migrations(deployment_id);
CREATE INDEX IF NOT EXISTS idx_migrations_version ON migrations(version);
CREATE INDEX IF NOT EXISTS idx_migrations_status ON migrations(status);

CREATE INDEX IF NOT EXISTS idx_audit_log_database_id ON audit_log(database_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_log_request_id ON audit_log(request_id);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_id ON api_keys(key_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active);
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys(expires_at);

CREATE INDEX IF NOT EXISTS idx_schema_snapshots_database_id ON schema_snapshots(database_id);
CREATE INDEX IF NOT EXISTS idx_schema_snapshots_created_at ON schema_snapshots(created_at);

-- Create functions for common operations

-- Function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_databases_updated_at BEFORE UPDATE ON databases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to log API key usage
CREATE OR REPLACE FUNCTION log_api_key_usage(key_id_param VARCHAR(255))
RETURNS BOOLEAN AS $$
BEGIN
    UPDATE api_keys 
    SET last_used_at = CURRENT_TIMESTAMP, usage_count = usage_count + 1
    WHERE api_keys.key_id = key_id_param AND is_active = true;
    
    RETURN FOUND;
END;
$$ language 'plpgsql';

-- Function to get deployment history for a database
CREATE OR REPLACE FUNCTION get_deployment_history(db_name VARCHAR(255), limit_count INTEGER DEFAULT 10)
RETURNS TABLE (
    deployment_id UUID,
    version VARCHAR(50),
    status VARCHAR(20),
    applied_at TIMESTAMP,
    applied_by VARCHAR(255),
    execution_time_ms INTEGER,
    migration_count INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        d.id,
        d.version,
        d.status,
        d.applied_at,
        d.applied_by,
        d.execution_time_ms,
        COUNT(m.id)::INTEGER as migration_count
    FROM deployments d
    JOIN databases db ON d.database_id = db.id
    LEFT JOIN migrations m ON d.id = m.deployment_id
    WHERE db.name = db_name
    GROUP BY d.id, d.version, d.status, d.applied_at, d.applied_by, d.execution_time_ms
    ORDER BY d.applied_at DESC NULLS LAST, d.created_at DESC
    LIMIT limit_count;
END;
$$ language 'plpgsql';

-- Function to get current database version
CREATE OR REPLACE FUNCTION get_current_database_version(db_name VARCHAR(255))
RETURNS VARCHAR(50) AS $$
DECLARE
    current_version VARCHAR(50);
BEGIN
    SELECT d.version INTO current_version
    FROM deployments d
    JOIN databases db ON d.database_id = db.id
    WHERE db.name = db_name 
    AND d.status = 'applied'
    ORDER BY d.applied_at DESC
    LIMIT 1;
    
    RETURN COALESCE(current_version, '0.0.0');
END;
$$ language 'plpgsql';

-- Function to check if migration version exists
CREATE OR REPLACE FUNCTION migration_version_exists(db_name VARCHAR(255), migration_version VARCHAR(50))
RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS(
        SELECT 1 
        FROM migrations m
        JOIN deployments d ON m.deployment_id = d.id
        JOIN databases db ON d.database_id = db.id
        WHERE db.name = db_name 
        AND m.version = migration_version
        AND m.status IN ('applied', 'skipped')
    );
END;
$$ language 'plpgsql';

-- Insert default system database entry
INSERT INTO databases (name, connection_string_encrypted, description, environment, created_by)
VALUES (
    'system', 
    'encrypted_placeholder_for_metadata_db', 
    'Metadata database for tracking deployments', 
    'system', 
    'system'
) ON CONFLICT (name) DO NOTHING;

-- Create a view for deployment status summary
CREATE OR REPLACE VIEW deployment_status_summary AS
SELECT 
    db.name as database_name,
    db.environment,
    COUNT(d.id) as total_deployments,
    COUNT(CASE WHEN d.status = 'applied' THEN 1 END) as successful_deployments,
    COUNT(CASE WHEN d.status = 'failed' THEN 1 END) as failed_deployments,
    get_current_database_version(db.name) as current_version,
    MAX(d.applied_at) as last_deployment_at,
    db.is_active
FROM databases db
LEFT JOIN deployments d ON db.id = d.database_id
GROUP BY db.id, db.name, db.environment, db.is_active
ORDER BY db.name;

-- Grant permissions (adjust as needed for your security model)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO db_agent_user;
-- GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO db_agent_user;
-- GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO db_agent_user;
