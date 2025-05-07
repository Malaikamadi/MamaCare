-- MamaCare SL - Revert Initial Database Setup Migration
-- This script reverses the changes made in the up.sql migration

-- Drop the function for refreshing materialized views
DROP FUNCTION IF EXISTS refresh_all_materialized_views();

-- Remove schema version tracking
DROP TABLE IF EXISTS schema_version;

-- Note: We do not disable extensions as they might be used by other applications
-- If you want to disable extensions, uncomment the following lines:
-- DROP EXTENSION IF EXISTS "uuid-ossp";
-- DROP EXTENSION IF EXISTS "postgis";
-- DROP EXTENSION IF EXISTS "pg_trgm";
-- DROP EXTENSION IF EXISTS "pgcrypto";
