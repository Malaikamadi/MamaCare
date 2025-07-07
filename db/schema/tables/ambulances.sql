-- Ambulances table for MamaCare SL
-- Tracks ambulance vehicles and their current status

CREATE TABLE IF NOT EXISTS ambulances (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Basic information
  vehicle_id TEXT NOT NULL UNIQUE, -- Registration or internal ID
  name TEXT NOT NULL, -- Human-readable name or call sign
  
  -- Operational details
  status ambulance_status NOT NULL DEFAULT 'AVAILABLE',
  facility_id UUID NOT NULL REFERENCES healthcare_facilities(id),
  
  -- Vehicle specifications
  vehicle_type TEXT NOT NULL, -- e.g., "AMBULANCE", "MOTORBIKE", "4X4"
  capacity INTEGER NOT NULL,
  has_oxygen BOOLEAN NOT NULL DEFAULT FALSE,
  has_stretcher BOOLEAN NOT NULL DEFAULT FALSE,
  
  -- Current tracking
  current_location GEOGRAPHY(POINT),
  last_location_update TIMESTAMP WITH TIME ZONE,
  current_driver_name TEXT,
  current_driver_phone phone_number,
  
  -- Current mission (if any)
  current_sos_id UUID REFERENCES sos_events(id),
  dispatched_at TIMESTAMP WITH TIME ZONE,
  estimated_arrival_time TIMESTAMP WITH TIME ZONE,
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  
  -- Constraints
  CONSTRAINT valid_status_sos CHECK (
    (status != 'DISPATCHED' AND status != 'TRANSPORTING' AND current_sos_id IS NULL) OR
    (status IN ('DISPATCHED', 'TRANSPORTING') AND current_sos_id IS NOT NULL)
  )
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_ambulances_updated_at
BEFORE UPDATE ON ambulances
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Row-level security policies for Hasura
ALTER TABLE ambulances ENABLE ROW LEVEL SECURITY;

-- Public view policy (anyone can see available ambulances)
CREATE POLICY public_view_ambulances ON ambulances
  USING (TRUE)
  WITH CHECK (FALSE); -- View only, no modifications

-- Healthcare providers can manage ambulances
CREATE POLICY healthcare_manage_ambulances ON ambulances
  USING (
    current_setting('hasura.user.role', true) IN ('CLINICIAN', 'ADMIN')
  )
  WITH CHECK (
    current_setting('hasura.user.role', true) IN ('CLINICIAN', 'ADMIN')
  );

-- Create indexes for common queries
CREATE INDEX idx_ambulances_status ON ambulances (status);
CREATE INDEX idx_ambulances_facility ON ambulances (facility_id);
CREATE INDEX idx_ambulances_location ON ambulances USING GIST (current_location);
CREATE INDEX idx_ambulances_sos ON ambulances (current_sos_id);

-- Add comments for documentation
COMMENT ON TABLE ambulances IS 'Tracks ambulance vehicles and their current status';
COMMENT ON COLUMN ambulances.status IS 'Current operational status of the ambulance';
COMMENT ON COLUMN ambulances.current_location IS 'Real-time geographic location of the ambulance';
COMMENT ON COLUMN ambulances.estimated_arrival_time IS 'Projected arrival time for current mission';
