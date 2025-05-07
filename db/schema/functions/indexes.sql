-- Performance Indexes for MamaCare SL
-- Additional indexes beyond the basic ones already defined in table creation scripts

-- Composite indexes for common query patterns
CREATE INDEX idx_visits_upcoming ON visits (mother_id, scheduled_date)
WHERE status = 'SCHEDULED' AND scheduled_date > CURRENT_DATE;

CREATE INDEX idx_visits_recent_completed ON visits (mother_id, completed_date)
WHERE status = 'COMPLETED' AND completed_date IS NOT NULL;

CREATE INDEX idx_child_upcoming_vaccinations ON children (mother_id, date_of_birth);

CREATE INDEX idx_users_by_area ON users (assigned_area)
WHERE role = 'MOTHER';

-- Partial indexes for frequently filtered subsets
CREATE INDEX idx_users_active_mothers ON users (phone_number, full_name)
WHERE role = 'MOTHER' AND account_status = 'active';

CREATE INDEX idx_high_risk_screenings ON screener_results (mother_id, screened_at)
WHERE risk_level = 'RED';

CREATE INDEX idx_visits_needing_reminder ON visits (scheduled_date)
WHERE reminder_sent = FALSE AND status = 'SCHEDULED' AND scheduled_date > CURRENT_DATE;

CREATE INDEX idx_active_ambulances ON ambulances (status, current_location)
WHERE status = 'AVAILABLE' AND is_active = TRUE;

-- Expression indexes for computed values
CREATE INDEX idx_child_age_months ON children (
  EXTRACT(YEAR FROM age(CURRENT_DATE, date_of_birth)) * 12 +
  EXTRACT(MONTH FROM age(CURRENT_DATE, date_of_birth))
);

-- Full text search indexes
CREATE INDEX idx_facilities_name_trgm ON healthcare_facilities USING gin (name gin_trgm_ops);
CREATE INDEX idx_users_fullname_trgm ON users USING gin (full_name gin_trgm_ops);

-- GiST indexes for more location-based queries
CREATE INDEX idx_user_nearest_facility ON healthcare_facilities USING GIST (location);
CREATE INDEX idx_ambulance_tracking ON ambulances USING GIST (current_location);
CREATE INDEX idx_sos_locations ON sos_events USING GIST (location);

-- BRIN indexes for large append-only tables (more efficient than B-tree for time-series data)
CREATE INDEX idx_notification_history_time_brin ON notification_history USING BRIN (sent_at);
CREATE INDEX idx_growth_measurements_time_brin ON growth_measurements USING BRIN (measured_at);

-- Enable index-only scans where possible by including additional columns
CREATE INDEX idx_growth_combined ON growth_measurements (child_id, measured_at) INCLUDE (weight_grams, height_cm, growth_status);
CREATE INDEX idx_immunization_combined ON immunization_records (child_id, administered_date) INCLUDE (vaccine_name, vaccine_dose);

-- Comments for documentation
COMMENT ON INDEX idx_visits_upcoming IS 'Optimizes queries for upcoming scheduled visits';
COMMENT ON INDEX idx_high_risk_screenings IS 'Speeds up queries for high-risk (RED) screening results';
COMMENT ON INDEX idx_child_age_months IS 'Supports efficient querying of children by age in months';
COMMENT ON INDEX idx_user_nearest_facility IS 'Enables spatial queries for finding nearest healthcare facilities';
