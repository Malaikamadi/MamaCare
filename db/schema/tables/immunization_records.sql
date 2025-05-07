-- Immunization records table for MamaCare SL
-- Tracks child vaccinations and immunization history

CREATE TABLE IF NOT EXISTS immunization_records (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Child reference
  child_id UUID NOT NULL REFERENCES children(id) ON DELETE CASCADE,
  
  -- Vaccine information
  vaccine_name non_empty_text NOT NULL,
  vaccine_dose TEXT NOT NULL, -- e.g., "1", "2", "booster"
  
  -- Administration details
  administered_date DATE NOT NULL,
  administered_by_user_id UUID REFERENCES users(id),
  administered_by_name TEXT,
  administered_at_facility_id UUID REFERENCES healthcare_facilities(id),
  
  -- Vaccine details
  batch_number TEXT,
  manufacturer TEXT,
  
  -- Any reactions or notes
  adverse_reactions TEXT,
  notes TEXT,
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_verified BOOLEAN NOT NULL DEFAULT true,
  
  -- Constraints
  CONSTRAINT valid_administrator_info CHECK (
    administered_by_user_id IS NOT NULL OR administered_by_name IS NOT NULL
  ),
  
  -- Unique constraint to prevent duplicate vaccinations
  UNIQUE (child_id, vaccine_name, vaccine_dose)
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_immunization_records_updated_at
BEFORE UPDATE ON immunization_records
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Row-level security policies for Hasura
ALTER TABLE immunization_records ENABLE ROW LEVEL SECURITY;

-- Parents can view their children's immunization records
CREATE POLICY parents_view_immunizations ON immunization_records
  USING (
    EXISTS (
      SELECT 1 FROM children c
      WHERE c.id = immunization_records.child_id AND (
        c.mother_id::text = current_setting('hasura.user.id', true) OR
        c.father_id::text = current_setting('hasura.user.id', true)
      )
    )
  );

-- Healthcare providers can add and view immunization records
CREATE POLICY healthcare_manage_immunizations ON immunization_records
  USING (
    current_setting('hasura.user.role', true) IN ('CHW', 'CLINICIAN', 'ADMIN')
  )
  WITH CHECK (
    current_setting('hasura.user.role', true) IN ('CHW', 'CLINICIAN', 'ADMIN')
  );

-- Create indexes for common queries
CREATE INDEX idx_immunization_records_child_id ON immunization_records (child_id);
CREATE INDEX idx_immunization_records_date ON immunization_records (administered_date);
CREATE INDEX idx_immunization_records_vaccine ON immunization_records (vaccine_name, vaccine_dose);

-- Add comments for documentation
COMMENT ON TABLE immunization_records IS 'Tracks child vaccinations and immunization history';
COMMENT ON COLUMN immunization_records.vaccine_name IS 'Name of vaccine administered';
COMMENT ON COLUMN immunization_records.vaccine_dose IS 'Dose number or type (e.g., "1", "2", "booster")';
COMMENT ON COLUMN immunization_records.is_verified IS 'Whether this record has been verified by a healthcare provider';
