-- Prenatal visit schedules table for MamaCare SL
-- Defines recommended schedule for antenatal care visits

CREATE TABLE IF NOT EXISTS prenatal_visit_schedules (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Visit information
  visit_name non_empty_text NOT NULL, -- e.g., "First Trimester Booking", "36-Week Checkup"
  visit_number INTEGER NOT NULL, -- Sequence number of the visit
  
  -- When to schedule
  weeks_gestation INTEGER NOT NULL, -- At what gestational age this visit should occur
  window_start_days INTEGER NOT NULL, -- Days before recommended time when visit can occur
  window_end_days INTEGER NOT NULL, -- Days after recommended time when visit should occur
  
  -- Visit details
  description TEXT NOT NULL, -- What happens at this visit
  is_required BOOLEAN NOT NULL DEFAULT TRUE, -- Whether this is a required visit
  danger_signs_to_check TEXT, -- Specific warning signs to look for
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  
  -- Make sure window_end_days ≥ window_start_days
  CONSTRAINT valid_window CHECK (window_end_days >= window_start_days),
  
  -- Unique constraint to prevent duplicate entries
  UNIQUE (visit_number)
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_prenatal_visit_schedules_updated_at
BEFORE UPDATE ON prenatal_visit_schedules
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Since this is reference data, no row-level security is needed
-- Only admins should be able to modify this via Hasura permissions

-- Create indexes for common queries
CREATE INDEX idx_prenatal_visit_schedules_weeks ON prenatal_visit_schedules (weeks_gestation);
CREATE INDEX idx_prenatal_visit_schedules_required ON prenatal_visit_schedules (is_required);

-- Add comments for documentation
COMMENT ON TABLE prenatal_visit_schedules IS 'Standard antenatal care visit schedule based on WHO recommendations';
COMMENT ON COLUMN prenatal_visit_schedules.weeks_gestation IS 'Recommended gestational age in weeks for this visit';
COMMENT ON COLUMN prenatal_visit_schedules.visit_number IS 'Sequence number of the visit in the standard schedule';
