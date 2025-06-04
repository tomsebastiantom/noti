-- Remove trigger and function
DROP TRIGGER IF EXISTS update_user_preferences_updated_at ON user_preferences;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop the user_preferences table
DROP TABLE IF EXISTS user_preferences;
