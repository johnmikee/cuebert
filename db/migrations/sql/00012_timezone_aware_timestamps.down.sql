-- Rollback migration: Revert to non-timezone-aware timestamps

-- Drop views and functions
DROP VIEW IF EXISTS tracking_user_local_time;
DROP FUNCTION IF EXISTS convert_to_user_timezone(TIMESTAMPTZ, INTEGER);
DROP FUNCTION IF EXISTS convert_from_user_timezone(TIMESTAMP, INTEGER);

-- Drop indexes on created_at columns
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_tracking_created_at;
DROP INDEX IF EXISTS idx_devices_created_at;
DROP INDEX IF EXISTS idx_exclusions_created_at;

-- Remove created_at columns
ALTER TABLE updates_os_users DROP COLUMN IF EXISTS created_at;
ALTER TABLE updates_os_exclusions DROP COLUMN IF EXISTS created_at;
ALTER TABLE updates_os_task DROP COLUMN IF EXISTS created_at;
ALTER TABLE updates_os_tracking DROP COLUMN IF EXISTS created_at;
ALTER TABLE updates_os_devices DROP COLUMN IF EXISTS created_at;

-- Revert all timestamp columns back to TIMESTAMP WITHOUT TIME ZONE

-- updates_os_users table
ALTER TABLE updates_os_users 
    ALTER COLUMN updated_at TYPE TIMESTAMP WITHOUT TIME ZONE;

-- updates_os_exclusions table
ALTER TABLE updates_os_exclusions 
    ALTER COLUMN until TYPE TIMESTAMP WITHOUT TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITHOUT TIME ZONE;

-- updates_os_task table
ALTER TABLE updates_os_task 
    ALTER COLUMN updated_at TYPE TIMESTAMP WITHOUT TIME ZONE;

-- updates_os_tracking table
ALTER TABLE updates_os_tracking 
    ALTER COLUMN first_ack_time TYPE TIMESTAMP WITHOUT TIME ZONE,
    ALTER COLUMN first_message_sent_at TYPE TIMESTAMP WITHOUT TIME ZONE,
    ALTER COLUMN user_reminder_at TYPE TIMESTAMP WITHOUT TIME ZONE,
    ALTER COLUMN final_message_sent_at TYPE TIMESTAMP WITHOUT TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITHOUT TIME ZONE;

-- updates_os_updates table
ALTER TABLE updates_os_updates 
    ALTER COLUMN updated_at TYPE TIMESTAMP WITHOUT TIME ZONE;

-- updates_os_devices table
ALTER TABLE updates_os_devices 
    ALTER COLUMN last_check_in TYPE TIMESTAMP WITHOUT TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITHOUT TIME ZONE;

-- Restore original trigger function
CREATE OR REPLACE FUNCTION log_last_updated()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = now();
        RETURN NEW;
    END;
    $$ language 'plpgsql';

-- Remove comments
COMMENT ON COLUMN updates_os_tracking.first_message_sent_at IS NULL;
COMMENT ON COLUMN updates_os_tracking.final_message_sent_at IS NULL;
COMMENT ON COLUMN updates_os_tracking.user_reminder_at IS NULL;
COMMENT ON COLUMN updates_os_tracking.tz_offset IS NULL;