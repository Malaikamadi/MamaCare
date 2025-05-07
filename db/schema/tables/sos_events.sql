-- SOS Events table for MamaCare SL
-- Tracks emergency assistance requests and response coordination

CREATE TABLE IF NOT EXISTS sos_events (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Who initiated the SOS
  user_id UUID NOT NULL REFERENCES users(id),
  
  -- Emergency details
  emergency_type TEXT NOT NULL, -- e.g., "MEDICAL", "TRANSPORTATION", "SECURITY"
  description TEXT,
  status sos_status NOT NULL DEFAULT 'OPEN',
  
  -- Subject of emergency (who needs help)
  for_self BOOLEAN NOT NULL DEFAULT TRUE,
  for_child_id UUID REFERENCES children(id),
  for_other_person TEXT, -- Description if not for self or registered child
  
  -- Location information (crucial for response)
  location GEOGRAPHY(POINT) NOT NULL,
  location_address TEXT,
  location_description TEXT, -- Landmarks, directions, etc.
  
  -- Timing information
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  accepted_at TIMESTAMP WITH TIME ZONE,
  resolved_at TIMESTAMP WITH TIME ZONE,
  
  -- Response details
  responding_facility_id UUID REFERENCES healthcare_facilities(id),
  responding_user_id UUID REFERENCES users(id),
  ambulance_dispatched BOOLEAN DEFAULT FALSE,
  ambulance_id UUID, -- Reference to ambulance table (created later)
  
  -- Communication logs
  notes TEXT,
  communication_log JSONB, -- Array of timestamped communication events
  
  -- Outcome tracking
  outcome TEXT,
  followup_needed BOOLEAN DEFAULT FALSE,
  followup_visit_id UUID REFERENCES visits(id),
  
  -- Constraints
  CONSTRAINT valid_emergency_subject CHECK (
    for_self = TRUE OR for_child_id IS NOT NULL OR for_other_person IS NOT NULL
  ),
  
  CONSTRAINT valid_status_timing CHECK (
    (status = 'OPEN' AND accepted_at IS NULL AND resolved_at IS NULL) OR
    (status = 'ACCEPTED' AND accepted_at IS NOT NULL AND resolved_at IS NULL) OR
    (status = 'CLOSED' AND accepted_at IS NOT NULL AND resolved_at IS NOT NULL)
  ),
  
  CONSTRAINT valid_response_info CHECK (
    status != 'ACCEPTED' OR (responding_facility_id IS NOT NULL OR responding_user_id IS NOT NULL)
  )
);

-- No trigger for updated_at as we track state changes explicitly with timestamps

-- Row-level security policies for Hasura
ALTER TABLE sos_events ENABLE ROW LEVEL SECURITY;

-- Users can view their own SOS events
CREATE POLICY user_view_own_sos ON sos_events
  USING (user_id::text = current_setting('hasura.user.id', true));

-- Users can view SOS events for their children
CREATE POLICY parent_view_child_sos ON sos_events
  USING (
    EXISTS (
      SELECT 1 FROM children c
      WHERE c.id = sos_events.for_child_id AND 
            (c.mother_id::text = current_setting('hasura.user.id', true) OR
             c.father_id::text = current_setting('hasura.user.id', true))
    )
  );

-- Healthcare providers can view and manage SOS events
CREATE POLICY healthcare_manage_sos ON sos_events
  USING (
    current_setting('hasura.user.role', true) IN ('CHW', 'CLINICIAN', 'ADMIN')
  )
  WITH CHECK (
    current_setting('hasura.user.role', true) IN ('CHW', 'CLINICIAN', 'ADMIN')
  );

-- Create indexes for common queries
CREATE INDEX idx_sos_events_user_id ON sos_events (user_id);
CREATE INDEX idx_sos_events_status ON sos_events (status);
CREATE INDEX idx_sos_events_location ON sos_events USING GIST (location);
CREATE INDEX idx_sos_events_date ON sos_events (created_at);
CREATE INDEX idx_sos_events_facility ON sos_events (responding_facility_id);

-- Add comments for documentation
COMMENT ON TABLE sos_events IS 'Tracks emergency assistance requests and response coordination';
COMMENT ON COLUMN sos_events.location IS 'Geographic point for mapping and distance calculations';
COMMENT ON COLUMN sos_events.status IS 'Current status of the SOS (OPEN, ACCEPTED, CLOSED)';
COMMENT ON COLUMN sos_events.communication_log IS 'JSON array of communication events with timestamps';
