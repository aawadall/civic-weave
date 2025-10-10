-- CivicWeave Database Initialization
-- This file runs when PostgreSQL container starts

-- Create database if it doesn't exist (handled by Docker)
-- CREATE DATABASE civicweave;

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE civicweave TO civicweave;
