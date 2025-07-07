-- Screener results table for MamaCare SL
-- Stores health assessment responses and risk evaluations

CREATE TABLE IF NOT EXISTS screener_results (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Who was screened
  mother_id UUID REFERENCES users(id) ON DELETE SET NULL,
  child_id UUID REFERENCES children(id) ON DELETE SET NULL,
  
  -- Screening session details
  screened_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  screened_by_user_id UUID REFERENCES users(id),
  screened_by_name TEXT, -- If screener is not a system user
  completed BOOLEAN NOT NULL DEFAULT TRUE,
  
  -- Location information
  location_latitude latitude,
  location_longitude longitude,
  facility_id UUID REFERENCES healthcare_facilities(id),
  
  -- Risk assessment results
  risk_level risk_level NOT NULL,
  primary_concern TEXT, -- Main symptom or issue
  danger_signs_detected BOOLEAN NOT NULL DEFAULT FALSE,
  followup_recommended BOOLEAN NOT NULL DEFAULT FALSE,
  
  -- Response actions
  action_taken TEXT,
  referral_facility_id UUID REFERENCES healthcare_facilities(id),
  followup_visit_id UUID REFERENCES visits(id),
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  -- Ensure we have screened either a mother or a child but not both
  CONSTRAINT valid_screening_subject CHECK (
    (mother_id IS NOT NULL AND child_id IS NULL) OR
    (mother_id IS NULL AND child_id IS NOT NULL)
  ),
  
  -- Ensure we know who performed the screening
  CONSTRAINT valid_screener_info CHECK (
    screened_by_user_id IS NOT NULL OR screened_by_name IS NOT NULL
  )
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_screener_results_updated_at
BEFORE UPDATE ON screener_results
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Row-level security policies for Hasura
ALTER TABLE screener_results ENABLE ROW LEVEL SECURITY;

-- Mothers can view their own screenings and their children's screenings
CREATE POLICY mother_view_own_screenings ON screener_results
  USING (
    mother_id::text = current_setting('hasura.user.id', true) OR
    EXISTS (
      SELECT 1 FROM children c
      WHERE c.id = screener_results.child_id AND c.mother_id::text = current_setting('hasura.user.id', true)
    )
  );

-- Fathers can view their children's screenings
CREATE POLICY father_view_children_screenings ON screener_results
  USING (
    EXISTS (
      SELECT 1 FROM children c
      WHERE c.id = screener_results.child_id AND c.father_id::text = current_setting('hasura.user.id', true)
    )
  );

-- Healthcare providers can view and create screenings
CREATE POLICY healthcare_manage_screenings ON screener_results
  USING (
    current_setting('hasura.user.role', true) IN ('CHW', 'CLINICIAN', 'ADMIN')
  )
  WITH CHECK (
    current_setting('hasura.user.role', true) IN ('CHW', 'CLINICIAN', 'ADMIN')
  );

-- Create indexes for common queries
CREATE INDEX idx_screener_results_mother_id ON screener_results (mother_id);
CREATE INDEX idx_screener_results_child_id ON screener_results (child_id);
CREATE INDEX idx_screener_results_risk_level ON screener_results (risk_level);
CREATE INDEX idx_screener_results_danger_signs ON screener_results (danger_signs_detected);
CREATE INDEX idx_screener_results_date ON screener_results (screened_at);

-- Add comments for documentation
COMMENT ON TABLE screener_results IS 'Stores health assessment responses and risk evaluations';
COMMENT ON COLUMN screener_results.risk_level IS 'Overall risk assessment result (RED, YELLOW, GREEN)';
COMMENT ON COLUMN screener_results.danger_signs_detected IS 'Whether any danger signs were detected during screening';
COMMENT ON COLUMN screener_results.followup_recommended IS 'Whether followup care is recommended based on results';
