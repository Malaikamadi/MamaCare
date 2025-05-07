-- Database Performance Optimizations for MamaCare SL
-- Configuration settings and query optimizations

-- Enable query statistics collection for performance monitoring
ALTER DATABASE current_database() SET track_io_timing = ON;
ALTER DATABASE current_database() SET track_functions = 'all';

-- Increase work_mem for complex query operations
-- This helps with in-memory sorting and hash operations
ALTER DATABASE current_database() SET work_mem = '16MB';

-- Optimize for SSD storage by reducing random page cost
-- This makes the planner more likely to use index scans
ALTER DATABASE current_database() SET random_page_cost = 1.1;

-- Set effective_cache_size based on expected available memory
-- This helps the planner understand how much memory is available for caching
ALTER DATABASE current_database() SET effective_cache_size = '2GB';

-- Use parallel query processing when appropriate
ALTER DATABASE current_database() SET max_parallel_workers_per_gather = 2;

-- Create tablespaces for different types of data (if needed in production)
-- Note: For development, this can remain commented out
-- CREATE TABLESPACE fastdata LOCATION '/path/to/fast/storage';
-- CREATE TABLESPACE archivedata LOCATION '/path/to/archive/storage';

-- Create partitioned tables for large time-series data
-- Example for notification history partitioning (can be implemented when data grows)
/*
CREATE TABLE notification_history_partitioned (
  -- Same schema as notification_history
) PARTITION BY RANGE (sent_at);

CREATE TABLE notification_history_y2023m01 PARTITION OF notification_history_partitioned
  FOR VALUES FROM ('2023-01-01') TO ('2023-02-01');
*/

-- Analyze tables to update statistics for the query planner
ANALYZE users;
ANALYZE children;
ANALYZE visits;
ANALYZE screener_results;
ANALYZE screener_answers;
ANALYZE sos_events;
ANALYZE notification_history;

-- Vacuum tables to reclaim space and update visibility information
VACUUM ANALYZE users;
VACUUM ANALYZE children;
VACUUM ANALYZE visits;
VACUUM ANALYZE screener_results;
VACUUM ANALYZE growth_measurements;
VACUUM ANALYZE immunization_records;

-- Create query optimization rules
-- Add table statistics for better query planning
ALTER TABLE users ALTER COLUMN role SET STATISTICS 1000;
ALTER TABLE visits ALTER COLUMN status SET STATISTICS 1000;
ALTER TABLE screener_results ALTER COLUMN risk_level SET STATISTICS 1000;

-- Set storage parameters for heavily-accessed tables
ALTER TABLE users SET (
  autovacuum_vacuum_scale_factor = 0.05,
  autovacuum_analyze_scale_factor = 0.02
);

ALTER TABLE visits SET (
  autovacuum_vacuum_scale_factor = 0.05,
  autovacuum_analyze_scale_factor = 0.02
);

ALTER TABLE sos_events SET (
  autovacuum_vacuum_scale_factor = 0.05,
  autovacuum_analyze_scale_factor = 0.02
);

-- Create aggregate view for KPI dashboard
CREATE OR REPLACE VIEW kpi_dashboard AS
SELECT
  -- Users and coverage
  (SELECT COUNT(*) FROM users WHERE role = 'MOTHER') AS registered_mothers,
  (SELECT COUNT(*) FROM children) AS registered_children,
  (SELECT COUNT(*) FROM users WHERE role = 'CHW') AS active_chws,
  
  -- Visit statistics
  (SELECT COUNT(*) FROM visits WHERE status = 'COMPLETED' AND completed_date >= CURRENT_DATE - INTERVAL '30 days') AS completed_visits_last_30d,
  (SELECT COUNT(*) FROM visits WHERE status = 'MISSED' AND scheduled_date >= CURRENT_DATE - INTERVAL '30 days') AS missed_visits_last_30d,
  
  -- Vaccination statistics
  (SELECT COUNT(*) FROM immunization_records WHERE administered_date >= CURRENT_DATE - INTERVAL '30 days') AS vaccines_given_last_30d,
  
  -- Health screening statistics
  (SELECT COUNT(*) FROM screener_results WHERE risk_level = 'RED' AND screened_at >= CURRENT_DATE - INTERVAL '30 days') AS red_flag_cases_last_30d,
  (SELECT COUNT(*) FROM screener_results WHERE risk_level = 'YELLOW' AND screened_at >= CURRENT_DATE - INTERVAL '30 days') AS yellow_flag_cases_last_30d,
  
  -- SOS statistics
  (SELECT COUNT(*) FROM sos_events WHERE created_at >= CURRENT_DATE - INTERVAL '30 days') AS sos_events_last_30d,
  (SELECT AVG(EXTRACT(EPOCH FROM (resolved_at - created_at))/60) FROM sos_events 
   WHERE status = 'CLOSED' AND created_at >= CURRENT_DATE - INTERVAL '30 days') AS avg_sos_resolution_minutes
;

-- Add comments for documentation
COMMENT ON VIEW kpi_dashboard IS 'Real-time KPI metrics for the dashboard';
COMMENT ON DATABASE current_database() IS 'MamaCare SL: Maternal and Child Health Platform for Sierra Leone';
