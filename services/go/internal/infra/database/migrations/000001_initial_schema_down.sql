-- Rollback Migration for Initial Schema
-- Drops all tables and types created in the initial schema

-- Drop tables in reverse order (to handle foreign key constraints)
DROP TABLE IF EXISTS education_content;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS sos_events;
DROP TABLE IF EXISTS health_metrics;
DROP TABLE IF EXISTS visits;
DROP TABLE IF EXISTS mothers;
DROP TABLE IF EXISTS facilities;
DROP TABLE IF EXISTS users;

-- Drop ENUM types
DROP TYPE IF EXISTS content_type;
DROP TYPE IF EXISTS content_category;
DROP TYPE IF EXISTS notification_priority;
DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS sos_event_nature;
DROP TYPE IF EXISTS sos_event_status;
DROP TYPE IF EXISTS blood_type;
DROP TYPE IF EXISTS visit_type;
DROP TYPE IF EXISTS visit_status;
DROP TYPE IF EXISTS facility_type;
DROP TYPE IF EXISTS risk_level;
DROP TYPE IF EXISTS user_role;
