-- Core users table for MamaCare SL
-- This is the primary authentication and user profile table

CREATE TABLE IF NOT EXISTS users (
  -- Primary identifiers
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  firebase_uid TEXT UNIQUE,  -- Links to Firebase Authentication
  
  -- Basic information
  phone_number phone_number UNIQUE NOT NULL,
  email email UNIQUE,
  full_name non_empty_text NOT NULL,
  
  -- Account details
  role user_role NOT NULL DEFAULT 'MOTHER',
  language language_preference NOT NULL DEFAULT 'ENGLISH',
  profile_picture_url TEXT,
  
  -- Mother-specific fields (null for other roles)
  expected_delivery_date DATE,
  last_menstrual_period DATE,
  blood_type TEXT,
  
  -- CHW/Clinician-specific fields (null for other roles)
  facility_id UUID,  -- References healthcare_facilities table
  assigned_area TEXT, -- Village or geographic area assigned to CHW
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_login TIMESTAMP WITH TIME ZONE,
  account_status TEXT NOT NULL DEFAULT 'active' CHECK (account_status IN ('active', 'inactive', 'suspended')),
  
  -- Emergency contact (for all users)
  emergency_contact_name TEXT,
  emergency_contact_phone phone_number,
  emergency_contact_relationship TEXT
);

-- Automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = CURRENT_TIMESTAMP;
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Row-level security policies for Hasura
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Users can view and edit their own records
CREATE POLICY user_modify_own ON users
  USING (id::text = current_setting('hasura.user.id', true))
  WITH CHECK (id::text = current_setting('hasura.user.id', true));

-- CHWs can view mother records in their assigned area
CREATE POLICY chw_view_mothers ON users
  USING (
    (current_setting('hasura.user.role', true) = 'CHW' AND 
     role = 'MOTHER' AND 
     assigned_area = (SELECT assigned_area FROM users WHERE id::text = current_setting('hasura.user.id', true)))
  );

-- Clinicians can view all mother and CHW records
CREATE POLICY clinician_view_users ON users
  USING (
    (current_setting('hasura.user.role', true) = 'CLINICIAN' AND 
     role IN ('MOTHER', 'CHW'))
  );

-- Admins can view and modify all records
CREATE POLICY admin_manage_all ON users
  USING (current_setting('hasura.user.role', true) = 'ADMIN')
  WITH CHECK (current_setting('hasura.user.role', true) = 'ADMIN');

-- Create index for phone number searches (common login method)
CREATE INDEX idx_users_phone_number ON users (phone_number);

-- Create index for role-based queries
CREATE INDEX idx_users_role ON users (role);

-- Add comments for documentation
COMMENT ON TABLE users IS 'Core user profiles and authentication information for MamaCare SL';
COMMENT ON COLUMN users.firebase_uid IS 'Firebase Authentication UID for phone auth';
COMMENT ON COLUMN users.role IS 'User role determines permissions and access level';
COMMENT ON COLUMN users.expected_delivery_date IS 'Only applicable for pregnant mothers';
COMMENT ON COLUMN users.facility_id IS 'Healthcare facility where CHW or clinician is based';
