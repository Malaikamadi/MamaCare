-- Growth measurements table for MamaCare SL
-- Tracks child growth metrics over time

CREATE TABLE IF NOT EXISTS growth_measurements (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Child reference
  child_id UUID NOT NULL REFERENCES children(id) ON DELETE CASCADE,
  
  -- Measurement information
  measured_at DATE NOT NULL DEFAULT CURRENT_DATE,
  
  -- Anthropometric measurements
  weight_grams INTEGER CHECK (weight_grams BETWEEN 500 AND 30000),
  height_cm DECIMAL(5,2) CHECK (height_cm BETWEEN 20 AND 150),
  head_circumference_cm DECIMAL(4,1) CHECK (head_circumference_cm BETWEEN 10 AND 60),
  mid_upper_arm_circumference_cm DECIMAL(4,1),
  
  -- WHO z-scores (standard deviations from median)
  weight_for_age_z DECIMAL(3,2),
  height_for_age_z DECIMAL(3,2),
  weight_for_height_z DECIMAL(3,2),
  
  -- Growth status assessment
  growth_status growth_status NOT NULL,
  
  -- Who took the measurement
  measured_by_user_id UUID REFERENCES users(id),
  measured_by_name TEXT,
  measured_at_facility_id UUID REFERENCES healthcare_facilities(id),
  
  -- Any observations or notes
  notes TEXT,
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  -- Constraints
  CONSTRAINT measurement_has_weight_or_height CHECK (
    weight_grams IS NOT NULL OR 
    height_cm IS NOT NULL OR 
    head_circumference_cm IS NOT NULL
  ),
  
  CONSTRAINT valid_measurer_info CHECK (
    measured_by_user_id IS NOT NULL OR measured_by_name IS NOT NULL
  )
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_growth_measurements_updated_at
BEFORE UPDATE ON growth_measurements
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Row-level security policies for Hasura
ALTER TABLE growth_measurements ENABLE ROW LEVEL SECURITY;

-- Parents can view their children's measurements
CREATE POLICY parents_view_measurements ON growth_measurements
  USING (
    EXISTS (
      SELECT 1 FROM children c
      WHERE c.id = growth_measurements.child_id AND (
        c.mother_id::text = current_setting('hasura.user.id', true) OR
        c.father_id::text = current_setting('hasura.user.id', true)
      )
    )
  );

-- Healthcare providers can add and view measurements
CREATE POLICY healthcare_manage_measurements ON growth_measurements
  USING (
    current_setting('hasura.user.role', true) IN ('CHW', 'CLINICIAN', 'ADMIN')
  )
  WITH CHECK (
    current_setting('hasura.user.role', true) IN ('CHW', 'CLINICIAN', 'ADMIN')
  );

-- Create indexes for common queries
CREATE INDEX idx_growth_measurements_child_id ON growth_measurements (child_id);
CREATE INDEX idx_growth_measurements_date ON growth_measurements (measured_at);
CREATE INDEX idx_growth_measurements_status ON growth_measurements (growth_status);

-- Add comments for documentation
COMMENT ON TABLE growth_measurements IS 'Tracks child growth measurements over time';
COMMENT ON COLUMN growth_measurements.weight_for_age_z IS 'Z-score compared to WHO growth standards';
COMMENT ON COLUMN growth_measurements.growth_status IS 'Assessment based on WHO classification';
COMMENT ON COLUMN growth_measurements.mid_upper_arm_circumference_cm IS 'Important indicator for malnutrition screening';
