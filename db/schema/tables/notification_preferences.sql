-- Notification Preferences table for MamaCare SL
-- Manages user communication preferences

CREATE TABLE IF NOT EXISTS notification_preferences (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- User relationship
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  -- Channel preferences
  enable_push_notifications BOOLEAN NOT NULL DEFAULT TRUE,
  enable_sms BOOLEAN NOT NULL DEFAULT TRUE,
  
  -- Type preferences
  enable_visit_reminders BOOLEAN NOT NULL DEFAULT TRUE,
  enable_vaccine_reminders BOOLEAN NOT NULL DEFAULT TRUE,
  enable_educational_content BOOLEAN NOT NULL DEFAULT TRUE,
  enable_emergency_alerts BOOLEAN NOT NULL DEFAULT TRUE,
  
  -- Timing preferences
  preferred_reminder_days_before INTEGER NOT NULL DEFAULT 1 CHECK (preferred_reminder_days_before BETWEEN 0 AND 7),
  quiet_hours_start TIME,
  quiet_hours_end TIME,
  
  -- Language preference
  preferred_language language_preference NOT NULL DEFAULT 'ENGLISH',
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  -- Each user has one preference record
  UNIQUE (user_id)
);

-- Automatically update the updated_at timestamp
CREATE TRIGGER update_notification_preferences_updated_at
BEFORE UPDATE ON notification_preferences
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Row-level security policies for Hasura
ALTER TABLE notification_preferences ENABLE ROW LEVEL SECURITY;

-- Users can manage their own notification preferences
CREATE POLICY user_manage_own_preferences ON notification_preferences
  USING (user_id::text = current_setting('hasura.user.id', true))
  WITH CHECK (user_id::text = current_setting('hasura.user.id', true));

-- Healthcare providers can view preferences (for patient communication)
CREATE POLICY healthcare_view_preferences ON notification_preferences
  USING (
    current_setting('hasura.user.role', true) IN ('CHW', 'CLINICIAN', 'ADMIN')
  );

-- Create indexes for common queries
CREATE INDEX idx_notification_preferences_user_id ON notification_preferences (user_id);
CREATE INDEX idx_notification_preferences_language ON notification_preferences (preferred_language);

-- Add comments for documentation
COMMENT ON TABLE notification_preferences IS 'Manages user communication preferences';
COMMENT ON COLUMN notification_preferences.enable_push_notifications IS 'Whether to send push notifications to mobile devices';
COMMENT ON COLUMN notification_preferences.enable_sms IS 'Whether to send SMS to phone number';
COMMENT ON COLUMN notification_preferences.preferred_reminder_days_before IS 'How many days before an event to send reminders';
