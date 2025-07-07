-- Master Seed File for MamaCare SL
-- Coordinates loading of all seed data in the correct order

-- Start transaction
BEGIN;

-- Load reference data first (required for the system to function)
\echo 'Loading healthcare facilities...'
\i healthcare_facilities.sql

\echo 'Loading vaccine schedules...'
\i vaccine_schedules.sql

\echo 'Loading prenatal visit schedules...'
\i prenatal_visit_schedules.sql

\echo 'Loading screener questions...'
\i screener_questions.sql

-- Load test data for development environments only
-- Comment these out when seeding production databases
\echo 'Loading test users...'
\i test_users.sql

\echo 'Loading test children and health records...'
\i test_children.sql

-- Initialize materialized views
\echo 'Refreshing materialized views...'
SELECT refresh_all_materialized_views();

-- Commit transaction
COMMIT;

\echo 'Seed data loading complete!';
