-- Vaccine schedules table for MamaCare SL
-- Defines standard vaccination schedules based on national guidelines

CREATE TABLE IF NOT EXISTS vaccine_schedules (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Vaccine information
  vaccine_name non_empty_text NOT NULL,
  vaccine_dose TEXT NOT NULL, -- e.g., "1", "2", "booster"
  
  -- When to administer
  age_months INTEGER NOT NULL, -- Can be 0 for birth vaccines
  window_start_days INTEGER NOT NULL, -- Days before recommended age when vaccine can be given
  window_end_days INTEGER NOT NULL, -- Days after recommended age when vaccine should be given
  
  -- Vaccine details
  disease_target TEXT NOT NULL, -- What disease this prevents
  is_required BOOLEAN NOT NULL DEFAULT TRUE, -- Whether this is a required vaccine
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  
  -- Make sure window_end_days ≥ window_start_days
  CONSTRAINT valid_window CHECK (window_end_days >= window_start_days),
  
  -- Unique constraint to prevent duplicate entries
  UNIQUE (vaccine_name, vaccine_dose)
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_vaccine_schedules_updated_at
BEFORE UPDATE ON vaccine_schedules
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Since this is reference data, no row-level security is needed
-- Only admins should be able to modify this via Hasura permissions

-- Create indexes for common queries
CREATE INDEX idx_vaccine_schedules_age ON vaccine_schedules (age_months);
CREATE INDEX idx_vaccine_schedules_vaccine ON vaccine_schedules (vaccine_name, vaccine_dose);

-- Add comments for documentation
COMMENT ON TABLE vaccine_schedules IS 'Standard vaccination schedules based on Sierra Leone national guidelines';
COMMENT ON COLUMN vaccine_schedules.age_months IS 'Recommended age in months for this vaccine';
COMMENT ON COLUMN vaccine_schedules.window_start_days IS 'Days before recommended age when vaccine can be given';
COMMENT ON COLUMN vaccine_schedules.window_end_days IS 'Days after recommended age when vaccine should be given';
