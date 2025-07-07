-- Notification History table for MamaCare SL
-- Tracks all sent notifications for audit and troubleshooting

CREATE TABLE IF NOT EXISTS notification_history (
  -- Primary identifier
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  
  -- Recipient information
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  -- Notification details
  notification_type notification_type NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  data JSONB, -- Additional payload data
  
  -- Delivery channel
  channel TEXT NOT NULL CHECK (channel IN ('PUSH', 'SMS', 'BOTH')),
  device_token_id UUID REFERENCES device_tokens(id) ON DELETE SET NULL,
  phone_number phone_number,
  
  -- Related records
  related_visit_id UUID REFERENCES visits(id) ON DELETE SET NULL,
  related_child_id UUID REFERENCES children(id) ON DELETE SET NULL,
  related_sos_id UUID REFERENCES sos_events(id) ON DELETE SET NULL,
  
  -- Delivery status
  sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  delivered BOOLEAN,
  delivered_at TIMESTAMP WITH TIME ZONE,
  opened BOOLEAN DEFAULT FALSE,
  opened_at TIMESTAMP WITH TIME ZONE,
  error_message TEXT,
  
  -- System fields
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  -- Ensure we have appropriate delivery channel information
  CONSTRAINT valid_channel_info CHECK (
    (channel = 'PUSH' AND device_token_id IS NOT NULL) OR
    (channel = 'SMS' AND phone_number IS NOT NULL) OR
    (channel = 'BOTH' AND device_token_id IS NOT NULL AND phone_number IS NOT NULL)
  )
);

-- No updated_at trigger as notifications are immutable once sent
-- (audit trail integrity)

-- Row-level security policies for Hasura
ALTER TABLE notification_history ENABLE ROW LEVEL SECURITY;

-- Users can view their own notifications
CREATE POLICY user_view_own_notifications ON notification_history
  USING (user_id::text = current_setting('hasura.user.id', true));

-- Healthcare providers and admins can view all notifications
CREATE POLICY admin_view_all_notifications ON notification_history
  USING (current_setting('hasura.user.role', true) IN ('CHW', 'CLINICIAN', 'ADMIN'));

-- Create indexes for common queries
CREATE INDEX idx_notification_history_user_id ON notification_history (user_id);
CREATE INDEX idx_notification_history_type ON notification_history (notification_type);
CREATE INDEX idx_notification_history_sent_at ON notification_history (sent_at);
CREATE INDEX idx_notification_history_delivered ON notification_history (delivered);
CREATE INDEX idx_notification_history_opened ON notification_history (opened);

-- Add comments for documentation
COMMENT ON TABLE notification_history IS 'Tracks all sent notifications for audit and troubleshooting';
COMMENT ON COLUMN notification_history.data IS 'JSON payload with additional notification data';
COMMENT ON COLUMN notification_history.channel IS 'Communication channel used (PUSH, SMS, or BOTH)';
COMMENT ON COLUMN notification_history.delivered IS 'Whether the notification was successfully delivered';
