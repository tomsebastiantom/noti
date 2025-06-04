-- Create user_preferences table in tenant databases
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    channel_preferences JSONB NOT NULL,
    category_preferences JSONB NOT NULL,
    digest_settings JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Add indexes to improve query performance
CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);

-- Add a uniqueness constraint to ensure one preference set per user
CREATE UNIQUE INDEX idx_user_preferences_user_unique ON user_preferences(user_id);

-- Ensure all existing users have default preferences
INSERT INTO user_preferences (id, user_id, enabled, channel_preferences, category_preferences, digest_settings)
SELECT 
    gen_random_uuid(), 
    id, 
    TRUE, 
    '{"email": true, "sms": true, "push": true, "web-push": true, "in-app": true}'::JSONB, 
    '{}'::JSONB,
    '{"enabled": false, "type": "none", "intervalMinutes": 60, "deliveryHour": 9, "preferredDayOfWeek": 1, "preferredChannel": "email"}'::JSONB
FROM users
WHERE id NOT IN (SELECT user_id FROM user_preferences);

-- Add a trigger to automatically update the updated_at field
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_user_preferences_updated_at
BEFORE UPDATE ON user_preferences
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
