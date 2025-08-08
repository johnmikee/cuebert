-- Migration: Convert to timezone-aware timestamps
-- This ensures proper timezone handling across all timestamp fields

-- 1. Convert all timestamp columns to TIMESTAMPTZ (timestamp with time zone)

-- updates_os_users table
ALTER TABLE updates_os_users 
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- updates_os_exclusions table
ALTER TABLE updates_os_exclusions 
    ALTER COLUMN until TYPE TIMESTAMPTZ USING until AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- updates_os_task table
ALTER TABLE updates_os_task 
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- updates_os_tracking table (most critical)
ALTER TABLE updates_os_tracking 
    ALTER COLUMN first_ack_time TYPE TIMESTAMPTZ USING first_ack_time AT TIME ZONE 'UTC',
    ALTER COLUMN first_message_sent_at TYPE TIMESTAMPTZ USING first_message_sent_at AT TIME ZONE 'UTC',
    ALTER COLUMN user_reminder_at TYPE TIMESTAMPTZ USING user_reminder_at AT TIME ZONE 'UTC',
    ALTER COLUMN final_message_sent_at TYPE TIMESTAMPTZ USING final_message_sent_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- updates_os_updates table
ALTER TABLE updates_os_updates 
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- updates_os_devices table
ALTER TABLE updates_os_devices 
    ALTER COLUMN last_check_in TYPE TIMESTAMPTZ USING last_check_in AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- 2. Add created_at columns where missing (with timezone awareness)
ALTER TABLE updates_os_users 
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();

ALTER TABLE updates_os_exclusions 
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();

ALTER TABLE updates_os_task 
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();

ALTER TABLE updates_os_tracking 
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();

ALTER TABLE updates_os_devices 
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();

-- 3. Update the trigger function to use timezone-aware timestamp
CREATE OR REPLACE FUNCTION log_last_updated()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = NOW(); -- NOW() returns TIMESTAMPTZ
        RETURN NEW;
    END;
    $$ language 'plpgsql';

-- 4. Create helper functions for timezone conversions
CREATE OR REPLACE FUNCTION convert_to_user_timezone(
    timestamp_utc TIMESTAMPTZ,
    tz_offset INTEGER -- offset in hours from UTC
) RETURNS TIMESTAMPTZ AS $$
BEGIN
    RETURN timestamp_utc AT TIME ZONE 'UTC' + (tz_offset || ' hours')::INTERVAL;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION convert_from_user_timezone(
    timestamp_local TIMESTAMP,
    tz_offset INTEGER -- offset in hours from UTC
) RETURNS TIMESTAMPTZ AS $$
BEGIN
    RETURN (timestamp_local - (tz_offset || ' hours')::INTERVAL) AT TIME ZONE 'UTC';
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- 5. Create views for timezone-aware queries
CREATE OR REPLACE VIEW tracking_user_local_time AS
SELECT 
    t.*,
    convert_to_user_timezone(t.first_message_sent_at, t.tz_offset) as first_message_sent_at_local,
    convert_to_user_timezone(t.final_message_sent_at, t.tz_offset) as final_message_sent_at_local,
    convert_to_user_timezone(t.user_reminder_at, t.tz_offset) as user_reminder_at_local,
    convert_to_user_timezone(t.updated_at, t.tz_offset) as updated_at_local
FROM updates_os_tracking t;

-- 6. Add indexes on new created_at columns
CREATE INDEX IF NOT EXISTS idx_users_created_at ON updates_os_users(created_at);
CREATE INDEX IF NOT EXISTS idx_tracking_created_at ON updates_os_tracking(created_at);
CREATE INDEX IF NOT EXISTS idx_devices_created_at ON updates_os_devices(created_at);
CREATE INDEX IF NOT EXISTS idx_exclusions_created_at ON updates_os_exclusions(created_at);

-- 7. Add comment documentation
COMMENT ON COLUMN updates_os_tracking.first_message_sent_at IS 'UTC timestamp when first message was sent';
COMMENT ON COLUMN updates_os_tracking.final_message_sent_at IS 'UTC timestamp when final message was sent';
COMMENT ON COLUMN updates_os_tracking.user_reminder_at IS 'UTC timestamp for scheduled reminder';
COMMENT ON COLUMN updates_os_tracking.tz_offset IS 'User timezone offset in hours from UTC';
COMMENT ON FUNCTION convert_to_user_timezone IS 'Convert UTC timestamp to user local time based on tz_offset';
COMMENT ON FUNCTION convert_from_user_timezone IS 'Convert user local time to UTC timestamp based on tz_offset';
COMMENT ON VIEW tracking_user_local_time IS 'View of tracking table with timestamps converted to user local time';