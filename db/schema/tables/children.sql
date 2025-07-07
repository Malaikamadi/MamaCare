-- Children table for MamaCare SL
-- Tracks child information, growth metrics, and relationships to parents

CREATE TABLE IF NOT EXISTS children (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Basic information
  first_name non_empty_text NOT NULL,
  middle_name TEXT,
  last_name non_empty_text NOT NULL,
  
  -- Birth information
  date_of_birth DATE NOT NULL,
  birth_weight_grams INTEGER,
  birth_length_cm INTEGER,
  place_of_birth TEXT,
  delivery_type TEXT CHECK (delivery_type IN ('VAGINAL', 'C_SECTION', 'ASSISTED')),
  gestational_age_weeks INTEGER CHECK (gestational_age_weeks BETWEEN 20 AND 45),
  
  -- Parent relationships
  mother_id UUID REFERENCES users(id) ON DELETE SET NULL,
  -- Father can either be a user in the system or just recorded by name
  father_id UUID REFERENCES users(id) ON DELETE SET NULL,
  father_name TEXT, -- Used when father is not a system user
  
  -- Health information
  blood_type TEXT,
  allergies TEXT,
  disabilities TEXT,
  chronic_conditions TEXT,
  
  -- Current status
  current_growth_status growth_status,
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_active BOOLEAN NOT NULL DEFAULT true,
  
  -- At least mother_id OR (father_id OR father_name) must be provided
  CONSTRAINT valid_parent_info CHECK (
    mother_id IS NOT NULL OR father_id IS NOT NULL OR father_name IS NOT NULL
  )
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_children_updated_at
BEFORE UPDATE ON children
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Row-level security policies for Hasura
ALTER TABLE children ENABLE ROW LEVEL SECURITY;

-- Parents can view and edit their children's records
CREATE POLICY parents_manage_children ON children
  USING (
    mother_id::text = current_setting('hasura.user.id', true) OR
    father_id::text = current_setting('hasura.user.id', true)
  )
  WITH CHECK (
    mother_id::text = current_setting('hasura.user.id', true) OR
    father_id::text = current_setting('hasura.user.id', true)
  );

-- CHWs can view children in their assigned area
CREATE POLICY chw_view_area_children ON children
  USING (
    current_setting('hasura.user.role', true) = 'CHW' AND
    EXISTS (
      SELECT 1 FROM users u 
      WHERE u.id = children.mother_id 
      AND u.assigned_area = (
        SELECT assigned_area FROM users 
        WHERE id::text = current_setting('hasura.user.id', true)
      )
    )
  );

-- Clinicians and admins can view all children
CREATE POLICY admin_clinician_view_all_children ON children
  USING (
    current_setting('hasura.user.role', true) IN ('ADMIN', 'CLINICIAN')
  );

-- Create indexes for common queries
CREATE INDEX idx_children_mother_id ON children (mother_id);
CREATE INDEX idx_children_father_id ON children (father_id);
CREATE INDEX idx_children_dob ON children (date_of_birth);
CREATE INDEX idx_children_growth_status ON children (current_growth_status);

-- Add comments for documentation
COMMENT ON TABLE children IS 'Records for children tracked in the MamaCare SL system';
COMMENT ON COLUMN children.mother_id IS 'Reference to mother in users table';
COMMENT ON COLUMN children.father_id IS 'Reference to father in users table, if registered';
COMMENT ON COLUMN children.father_name IS 'Name of father when not a registered user';
COMMENT ON COLUMN children.current_growth_status IS 'Current growth status based on WHO standards';
