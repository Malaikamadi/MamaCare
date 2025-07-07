-- SOS Escalations table for MamaCare SL
-- Tracks escalation path for emergency situations

CREATE TABLE IF NOT EXISTS sos_escalations (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Relationship to SOS event
  sos_id UUID NOT NULL REFERENCES sos_events(id) ON DELETE CASCADE,
  
  -- Escalation details
  escalation_level INTEGER NOT NULL, -- 1, 2, 3, etc.
  escalation_reason TEXT NOT NULL,
  escalated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  escalated_by_user_id UUID REFERENCES users(id),
  
  -- Who was contacted in escalation
  escalated_to_facility_id UUID REFERENCES healthcare_facilities(id),
  escalated_to_user_id UUID REFERENCES users(id),
  escalated_to_role user_role,
  escalated_to_name TEXT,
  escalated_to_phone phone_number,
  
  -- Response tracking
  response_received BOOLEAN NOT NULL DEFAULT FALSE,
  response_time TIMESTAMP WITH TIME ZONE,
  response_notes TEXT,
  
  -- Outcome
  resolved_via_escalation BOOLEAN,
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  -- Constraints
  CONSTRAINT valid_escalation_target CHECK (
    escalated_to_facility_id IS NOT NULL OR 
    escalated_to_user_id IS NOT NULL OR
    (escalated_to_name IS NOT NULL AND escalated_to_phone IS NOT NULL)
  )
);

-- No updated_at trigger as escalations should not be modified once created
-- (audit trail integrity)

-- Row-level security policies for Hasura
ALTER TABLE sos_escalations ENABLE ROW LEVEL SECURITY;

-- Escalations inherit access permissions from parent SOS events
CREATE POLICY inherit_sos_permissions ON sos_escalations
  USING (
    EXISTS (
      SELECT 1 FROM sos_events se
      WHERE se.id = sos_escalations.sos_id
    )
  );

-- Create indexes for common queries
CREATE INDEX idx_sos_escalations_sos_id ON sos_escalations (sos_id);
CREATE INDEX idx_sos_escalations_level ON sos_escalations (escalation_level);
CREATE INDEX idx_sos_escalations_date ON sos_escalations (escalated_at);

-- Add comments for documentation
COMMENT ON TABLE sos_escalations IS 'Tracks escalation path for emergency situations';
COMMENT ON COLUMN sos_escalations.escalation_level IS 'Level of escalation (higher number = more urgent)';
COMMENT ON COLUMN sos_escalations.response_received IS 'Whether a response was received to this escalation';
COMMENT ON COLUMN sos_escalations.resolved_via_escalation IS 'Whether this escalation led to resolution';
