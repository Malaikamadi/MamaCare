-- Healthcare facilities table for MamaCare SL
-- Tracks hospitals, clinics, and health centers

CREATE TABLE IF NOT EXISTS healthcare_facilities (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Basic information
  name non_empty_text NOT NULL,
  facility_type TEXT NOT NULL CHECK (facility_type IN ('HOSPITAL', 'CLINIC', 'HEALTH_CENTER', 'HEALTH_POST')),
  
  -- Contact details
  phone_number phone_number,
  email email,
  website TEXT,
  
  -- Location information
  address TEXT NOT NULL,
  district TEXT NOT NULL,
  chiefdom TEXT,
  
  -- Geographic coordinates (for mapping)
  location GEOGRAPHY(POINT) NOT NULL,
  
  -- Operational details
  operating_hours TEXT,
  has_ambulance BOOLEAN NOT NULL DEFAULT false,
  has_maternity_ward BOOLEAN NOT NULL DEFAULT false,
  bed_capacity INTEGER,
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_active BOOLEAN NOT NULL DEFAULT true
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_healthcare_facilities_updated_at
BEFORE UPDATE ON healthcare_facilities
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Create spatial index for location queries
CREATE INDEX idx_healthcare_facilities_location ON healthcare_facilities USING GIST (location);

-- Create index for facility type searches
CREATE INDEX idx_healthcare_facilities_type ON healthcare_facilities (facility_type);

-- Add comments for documentation
COMMENT ON TABLE healthcare_facilities IS 'Healthcare facilities including hospitals, clinics, and health centers in Sierra Leone';
COMMENT ON COLUMN healthcare_facilities.facility_type IS 'Type of healthcare facility - affects services available';
COMMENT ON COLUMN healthcare_facilities.location IS 'Geographic point location for mapping and distance calculations';
COMMENT ON COLUMN healthcare_facilities.has_ambulance IS 'Whether the facility has ambulance services available';
