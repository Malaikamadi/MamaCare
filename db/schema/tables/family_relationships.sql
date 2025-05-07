-- Family relationships table for MamaCare SL
-- Tracks relationships between users (mothers, fathers, etc.)

CREATE TABLE IF NOT EXISTS family_relationships (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Relationship information
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  related_person_id UUID REFERENCES users(id) ON DELETE SET NULL,
  
  -- For external family members not in the system
  related_person_name TEXT,
  related_person_phone phone_number,
  
  -- Relationship type
  relationship_type TEXT NOT NULL CHECK (
    relationship_type IN (
      'SPOUSE', 'PARTNER', 'FATHER_OF_CHILD', 'MOTHER_OF_CHILD',
      'SIBLING', 'GRANDPARENT', 'OTHER_RELATIVE', 'GUARDIAN'
    )
  ),
  
  -- Additional details
  notes TEXT,
  is_emergency_contact BOOLEAN NOT NULL DEFAULT false,
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  -- Either related_person_id OR (related_person_name AND related_person_phone) must be provided
  CONSTRAINT valid_relation_info CHECK (
    (related_person_id IS NOT NULL) OR 
    (related_person_name IS NOT NULL AND related_person_phone IS NOT NULL)
  )
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_family_relationships_updated_at
BEFORE UPDATE ON family_relationships
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Row-level security policies for Hasura
ALTER TABLE family_relationships ENABLE ROW LEVEL SECURITY;

-- Users can view and edit their own family relationships
CREATE POLICY user_manage_own_relationships ON family_relationships
  USING (user_id::text = current_setting('hasura.user.id', true))
  WITH CHECK (user_id::text = current_setting('hasura.user.id', true));

-- CHWs can view family relationships for mothers in their assigned area
CREATE POLICY chw_view_mother_relationships ON family_relationships
  USING (
    current_setting('hasura.user.role', true) = 'CHW' AND
    EXISTS (
      SELECT 1 FROM users u 
      WHERE u.id = family_relationships.user_id 
      AND u.role = 'MOTHER'
      AND u.assigned_area = (
        SELECT assigned_area FROM users 
        WHERE id::text = current_setting('hasura.user.id', true)
      )
    )
  );

-- Clinicians and admins can view all family relationships
CREATE POLICY admin_clinician_view_all_relationships ON family_relationships
  USING (
    current_setting('hasura.user.role', true) IN ('ADMIN', 'CLINICIAN')
  );

-- Create indexes for common queries
CREATE INDEX idx_family_relationships_user_id ON family_relationships (user_id);
CREATE INDEX idx_family_relationships_related_person_id ON family_relationships (related_person_id);
CREATE INDEX idx_family_relationships_type ON family_relationships (relationship_type);

-- Add comments for documentation
COMMENT ON TABLE family_relationships IS 'Tracks relationships between users, including parents, spouses, and other family members';
COMMENT ON COLUMN family_relationships.related_person_id IS 'If the related person is a system user, link to their ID';
COMMENT ON COLUMN family_relationships.related_person_name IS 'For external family members not in the system (e.g., fathers who don't use the app)';
COMMENT ON COLUMN family_relationships.relationship_type IS 'Type of family relationship, particularly important for identifying fathers of children';
