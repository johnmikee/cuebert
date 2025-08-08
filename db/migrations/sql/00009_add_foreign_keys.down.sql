-- Rollback migration: Remove foreign key constraints

-- Drop foreign key constraints
ALTER TABLE updates_os_tracking 
    DROP CONSTRAINT IF EXISTS fk_tracking_device;

ALTER TABLE updates_os_exclusions 
    DROP CONSTRAINT IF EXISTS fk_exclusion_device;

-- Drop unique constraints
ALTER TABLE updates_os_devices 
    DROP CONSTRAINT IF EXISTS uk_device_serial;

ALTER TABLE updates_os_users 
    DROP CONSTRAINT IF EXISTS uk_user_email;

-- Drop check constraints
ALTER TABLE updates_os_exclusions 
    DROP CONSTRAINT IF EXISTS chk_exclusion_until;

ALTER TABLE updates_os_tracking 
    DROP CONSTRAINT IF EXISTS chk_tracking_ack_time;

ALTER TABLE updates_os_tracking 
    DROP CONSTRAINT IF EXISTS chk_tracking_sent_time;

-- Drop indexes created for foreign keys
DROP INDEX IF EXISTS idx_tracking_device;
DROP INDEX IF EXISTS idx_exclusion_device;