-- Visits table for MamaCare SL
-- Tracks all healthcare appointments, both scheduled and completed

CREATE TABLE IF NOT EXISTS visits (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Visit subject (either mother or child)
  mother_id UUID REFERENCES users(id) ON DELETE CASCADE,
  child_id UUID REFERENCES children(id) ON DELETE CASCADE,
  
  -- Visit details
  visit_type visit_type NOT NULL,
  status visit_status NOT NULL DEFAULT 'SCHEDULED',
  
  -- Scheduling information
  scheduled_date DATE NOT NULL,
  scheduled_time TIME,
  duration_minutes INTEGER,
  
  -- Completion details (null if not completed)
  completed_date DATE,
  completed_time TIME,
  completed_by_user_id UUID REFERENCES users(id),
  
  -- Location details
  facility_id UUID REFERENCES healthcare_facilities(id),
  facility_name TEXT, -- For locations not in our database
  
  -- Visit notes and outcomes
  chief_complaint TEXT,
  diagnosis TEXT,
  treatment TEXT,
  notes TEXT,
  followup_needed BOOLEAN DEFAULT FALSE,
  followup_visit_id UUID REFERENCES visits(id),
  
  -- Reminder status
  reminder_sent BOOLEAN NOT NULL DEFAULT FALSE,
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  -- Either mother_id OR child_id must be provided
  CONSTRAINT visit_has_subject CHECK (
    (mother_id IS NOT NULL AND child_id IS NULL) OR
    (mother_id IS NULL AND child_id IS NOT NULL)
  ),
  
  -- Completed visits must have completed_date
  CONSTRAINT valid_completed_visit CHECK (
    status != 'COMPLETED' OR completed_date IS NOT NULL
  ),
  
  -- Cannot have followup_visit_id unless followup_needed is true
  CONSTRAINT valid_followup CHECK (
    NOT followup_needed OR followup_visit_id IS NOT NULL
  )
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_visits_updated_at
BEFORE UPDATE ON visits
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Row-level security policies for Hasura
ALTER TABLE visits ENABLE ROW LEVEL SECURITY;

-- Mothers can view and modify their own visits and their children's visits
CREATE POLICY mother_manage_own_visits ON visits
  USING (
    mother_id::text = current_setting('hasura.user.id', true) OR
    EXISTS (
      SELECT 1 FROM children c
      WHERE c.id = visits.child_id AND c.mother_id::text = current_setting('hasura.user.id', true)
    )
  )
  WITH CHECK (
    mother_id::text = current_setting('hasura.user.id', true) OR
    EXISTS (
      SELECT 1 FROM children c
      WHERE c.id = visits.child_id AND c.mother_id::text = current_setting('hasura.user.id', true)
    )
  );

-- Fathers can view their children's visits
CREATE POLICY father_view_children_visits ON visits
  USING (
    EXISTS (
      SELECT 1 FROM children c
      WHERE c.id = visits.child_id AND c.father_id::text = current_setting('hasura.user.id', true)
    )
  );

-- Healthcare providers can view and modify all visits
CREATE POLICY healthcare_manage_visits ON visits
  USING (
    current_setting('hasura.user.role', true) IN ('CHW', 'CLINICIAN', 'ADMIN')
  )
  WITH CHECK (
    current_setting('hasura.user.role', true) IN ('CHW', 'CLINICIAN', 'ADMIN')
  );

-- Create indexes for common queries
CREATE INDEX idx_visits_mother_id ON visits (mother_id);
CREATE INDEX idx_visits_child_id ON visits (child_id);
CREATE INDEX idx_visits_scheduled_date ON visits (scheduled_date);
CREATE INDEX idx_visits_status ON visits (status);
CREATE INDEX idx_visits_type ON visits (visit_type);
CREATE INDEX idx_visits_reminder_sent ON visits (reminder_sent) WHERE reminder_sent = FALSE;

-- Add comments for documentation
COMMENT ON TABLE visits IS 'Tracks all healthcare appointments, both scheduled and completed';
COMMENT ON COLUMN visits.visit_type IS 'Type of visit (antenatal, vaccination, etc.)';
COMMENT ON COLUMN visits.status IS 'Current status of the visit (scheduled, completed, etc.)';
COMMENT ON COLUMN visits.reminder_sent IS 'Whether a reminder has been sent for this visit';
