-- Device Tokens table for MamaCare SL
-- Stores push notification tokens for mobile devices

CREATE TABLE IF NOT EXISTS device_tokens (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- User relationship
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  -- Device information
  device_id TEXT NOT NULL, -- Unique identifier for the device
  device_name TEXT, -- Human-readable device name
  device_platform device_platform NOT NULL,
  
  -- Push token details
  push_token TEXT NOT NULL,
  push_token_valid BOOLEAN NOT NULL DEFAULT TRUE,
  
  -- App information
  app_version TEXT,
  os_version TEXT,
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_used_at TIMESTAMP WITH TIME ZONE,
  
  -- Each user can have multiple devices, but each device has one unique token
  UNIQUE (user_id, device_id)
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_device_tokens_updated_at
BEFORE UPDATE ON device_tokens
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Row-level security policies for Hasura
ALTER TABLE device_tokens ENABLE ROW LEVEL SECURITY;

-- Users can manage their own device tokens
CREATE POLICY user_manage_own_tokens ON device_tokens
  USING (user_id::text = current_setting('hasura.user.id', true))
  WITH CHECK (user_id::text = current_setting('hasura.user.id', true));

-- Admins can view all device tokens (for troubleshooting)
CREATE POLICY admin_view_all_tokens ON device_tokens
  USING (current_setting('hasura.user.role', true) = 'ADMIN');

-- Create indexes for common queries
CREATE INDEX idx_device_tokens_user_id ON device_tokens (user_id);
CREATE INDEX idx_device_tokens_platform ON device_tokens (device_platform);
CREATE INDEX idx_device_tokens_valid ON device_tokens (push_token_valid);

-- Add comments for documentation
COMMENT ON TABLE device_tokens IS 'Stores push notification tokens for mobile devices';
COMMENT ON COLUMN device_tokens.push_token IS 'Expo Push Token or Firebase Cloud Messaging token';
COMMENT ON COLUMN device_tokens.device_id IS 'Unique identifier for the physical device';
COMMENT ON COLUMN device_tokens.push_token_valid IS 'Whether the token is still valid for sending notifications';
