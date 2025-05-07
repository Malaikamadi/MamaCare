-- MamaCare SL Database Schema
-- Main schema file that includes all table definitions

-- Enable necessary PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";      -- For UUID generation
CREATE EXTENSION IF NOT EXISTS "postgis";        -- For geospatial features
CREATE EXTENSION IF NOT EXISTS "pg_trgm";        -- For text search

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_version (
    version VARCHAR(50) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);

-- Insert initial schema version
INSERT INTO schema_version (version, description)
VALUES ('0.1.0', 'Initial schema setup');

-- Table definitions will be included below
-- Each section will define a core entity in the MamaCare system
