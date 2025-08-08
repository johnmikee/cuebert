-- Rollback migration: Restore original data types

-- Drop foreign key constraints that depend on INTEGER type
ALTER TABLE updates_os_tracking 
    DROP CONSTRAINT IF EXISTS fk_tracking_user;

ALTER TABLE updates_os_task 
    DROP CONSTRAINT IF EXISTS fk_task_user;

-- Revert user_service_id back to VARCHAR(255)
ALTER TABLE updates_os_tracking 
    ALTER COLUMN user_service_id TYPE VARCHAR(255) USING user_service_id::TEXT;

ALTER TABLE updates_os_task 
    ALTER COLUMN user_service_id TYPE VARCHAR(255) USING user_service_id::TEXT;

-- Revert all VARCHAR optimizations back to VARCHAR(255)
ALTER TABLE updates_os_users 
    ALTER COLUMN user_slack_id TYPE VARCHAR(255);

ALTER TABLE updates_os_tracking 
    ALTER COLUMN user_slack_id TYPE VARCHAR(255),
    ALTER COLUMN serial_number TYPE VARCHAR(255),
    ALTER COLUMN task_name TYPE VARCHAR(255),
    ALTER COLUMN user_email TYPE VARCHAR(255),
    ALTER COLUMN user_full_name TYPE VARCHAR(255);

ALTER TABLE updates_os_devices 
    ALTER COLUMN serial_number TYPE VARCHAR(255),
    ALTER COLUMN device_name TYPE VARCHAR(255),
    ALTER COLUMN model TYPE VARCHAR(255),
    ALTER COLUMN platform TYPE VARCHAR(255),
    ALTER COLUMN os_version TYPE VARCHAR(255);

ALTER TABLE updates_os_exclusions 
    ALTER COLUMN serial_number TYPE VARCHAR(255),
    ALTER COLUMN user_email TYPE VARCHAR(255);

ALTER TABLE updates_os_task 
    ALTER COLUMN item_name TYPE VARCHAR(255),
    ALTER COLUMN task_description TYPE VARCHAR(255);

ALTER TABLE updates_os_users 
    ALTER COLUMN user_email TYPE VARCHAR(255);

-- Remove NOT NULL constraints
ALTER TABLE updates_os_tracking 
    ALTER COLUMN task_complete DROP NOT NULL,
    ALTER COLUMN first_ack DROP NOT NULL,
    ALTER COLUMN first_message_sent DROP NOT NULL,
    ALTER COLUMN first_message_waiting DROP NOT NULL,
    ALTER COLUMN reminder_sent DROP NOT NULL,
    ALTER COLUMN reminder_waiting DROP NOT NULL;

-- Remove defaults
ALTER TABLE updates_os_tracking 
    ALTER COLUMN first_ack DROP DEFAULT,
    ALTER COLUMN first_message_sent DROP DEFAULT,
    ALTER COLUMN first_message_waiting DROP DEFAULT,
    ALTER COLUMN reminder_sent DROP DEFAULT,
    ALTER COLUMN reminder_waiting DROP DEFAULT,
    ALTER COLUMN final_message_sent DROP DEFAULT;

-- Drop indexes
DROP INDEX IF EXISTS idx_tracking_user;
DROP INDEX IF EXISTS idx_task_user;