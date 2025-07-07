-- MamaCare SL - Initial Database Setup Migration
-- This migration initializes the database with required extensions and configurations

-- Enable necessary PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";      -- For UUID generation
CREATE EXTENSION IF NOT EXISTS "postgis";        -- For geospatial features
CREATE EXTENSION IF NOT EXISTS "pg_trgm";        -- For text search
CREATE EXTENSION IF NOT EXISTS "pgcrypto";       -- For cryptographic functions

-- Set timezone to UTC for consistent timestamp handling
SET timezone = 'UTC';

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_version (
    version VARCHAR(50) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);

-- Insert initial schema version
INSERT INTO schema_version (version, description)
VALUES ('1.0.0', 'Initial schema migration')
ON CONFLICT (version) DO NOTHING;

-- Performance and security settings
ALTER DATABASE CURRENT SET search_path = "$user", public;

-- Set up connection pooling optimization
ALTER DATABASE CURRENT SET idle_in_transaction_session_timeout = '5min';

-- Configure geospatial settings for Sierra Leone region
ALTER DATABASE CURRENT SET postgis.gdal_enabled_drivers = 'ENABLE_ALL';

-- Create a function to refresh all materialized views
CREATE OR REPLACE FUNCTION refresh_all_materialized_views()
RETURNS INTEGER AS $$
DECLARE
    r RECORD;
    counter INTEGER := 0;
BEGIN
    FOR r IN SELECT schemaname, matviewname 
             FROM pg_matviews 
             WHERE schemaname = 'public'
    LOOP
        EXECUTE 'REFRESH MATERIALIZED VIEW ' || quote_ident(r.schemaname) || '.' || quote_ident(r.matviewname);
        counter := counter + 1;
    END LOOP;
    RETURN counter;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
