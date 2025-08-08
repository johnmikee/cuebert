-- Migration: Fix data type inconsistencies
-- This migration corrects mismatched data types across related tables

-- 1. Fix user_service_id type inconsistency
-- Currently: INTEGER in updates_os_users, VARCHAR(255) in tracking and task tables

-- First, ensure all user_service_id values can be converted to INTEGER
-- Check for non-numeric values
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM updates_os_tracking 
        WHERE user_service_id !~ '^\d+$'
    ) THEN
        RAISE EXCEPTION 'Non-numeric user_service_id found in updates_os_tracking';
    END IF;
    
    IF EXISTS (
        SELECT 1 FROM updates_os_task 
        WHERE user_service_id !~ '^\d+$'
    ) THEN
        RAISE EXCEPTION 'Non-numeric user_service_id found in updates_os_task';
    END IF;
END $$;

-- Convert VARCHAR to INTEGER for user_service_id
ALTER TABLE updates_os_tracking 
    ALTER COLUMN user_service_id TYPE INTEGER USING user_service_id::INTEGER;

ALTER TABLE updates_os_task 
    ALTER COLUMN user_service_id TYPE INTEGER USING user_service_id::INTEGER;

-- Now we can add the foreign key constraints for users
ALTER TABLE updates_os_tracking 
    ADD CONSTRAINT fk_tracking_user 
    FOREIGN KEY (user_service_id) 
    REFERENCES updates_os_users(user_service_id) 
    ON DELETE CASCADE;

ALTER TABLE updates_os_task 
    ADD CONSTRAINT fk_task_user 
    FOREIGN KEY (user_service_id) 
    REFERENCES updates_os_users(user_service_id) 
    ON DELETE CASCADE;

-- 2. Optimize VARCHAR sizes based on actual usage

-- Slack IDs are typically 11 characters (e.g., U0123456789)
ALTER TABLE updates_os_users 
    ALTER COLUMN user_slack_id TYPE VARCHAR(15);

ALTER TABLE updates_os_tracking 
    ALTER COLUMN user_slack_id TYPE VARCHAR(15);

-- Serial numbers are typically 12-20 characters
ALTER TABLE updates_os_devices 
    ALTER COLUMN serial_number TYPE VARCHAR(30);

ALTER TABLE updates_os_tracking 
    ALTER COLUMN serial_number TYPE VARCHAR(30);

ALTER TABLE updates_os_exclusions 
    ALTER COLUMN serial_number TYPE VARCHAR(30);

-- Platform names are short (macOS, Windows, Linux, etc.)
ALTER TABLE updates_os_devices 
    ALTER COLUMN platform TYPE VARCHAR(20);

-- OS versions are typically short (e.g., "14.1.1", "Windows 11")
ALTER TABLE updates_os_devices 
    ALTER COLUMN os_version TYPE VARCHAR(20);

-- Device names can be reasonably sized
ALTER TABLE updates_os_devices 
    ALTER COLUMN device_name TYPE VARCHAR(100);

-- Model names are typically under 50 chars
ALTER TABLE updates_os_devices 
    ALTER COLUMN model TYPE VARCHAR(50);

-- Task names and descriptions can be shorter
ALTER TABLE updates_os_task 
    ALTER COLUMN item_name TYPE VARCHAR(100),
    ALTER COLUMN task_description TYPE VARCHAR(200);

ALTER TABLE updates_os_tracking 
    ALTER COLUMN task_name TYPE VARCHAR(100),
    ALTER COLUMN user_full_name TYPE VARCHAR(100);

-- Email addresses have a reasonable max length
ALTER TABLE updates_os_users 
    ALTER COLUMN user_email TYPE VARCHAR(100);

ALTER TABLE updates_os_tracking 
    ALTER COLUMN user_email TYPE VARCHAR(100);

ALTER TABLE updates_os_exclusions 
    ALTER COLUMN user_email TYPE VARCHAR(100);

-- 3. Add NOT NULL constraints where appropriate
ALTER TABLE updates_os_tracking 
    ALTER COLUMN task_complete SET NOT NULL,
    ALTER COLUMN first_ack SET NOT NULL,
    ALTER COLUMN first_message_sent SET NOT NULL,
    ALTER COLUMN first_message_waiting SET NOT NULL,
    ALTER COLUMN reminder_sent SET NOT NULL,
    ALTER COLUMN reminder_waiting SET NOT NULL;

-- 4. Set better defaults
ALTER TABLE updates_os_tracking 
    ALTER COLUMN first_ack SET DEFAULT FALSE,
    ALTER COLUMN first_message_sent SET DEFAULT FALSE,
    ALTER COLUMN first_message_waiting SET DEFAULT FALSE,
    ALTER COLUMN reminder_sent SET DEFAULT FALSE,
    ALTER COLUMN reminder_waiting SET DEFAULT FALSE,
    ALTER COLUMN final_message_sent SET DEFAULT FALSE;

-- 5. Create indexes for foreign key columns
CREATE INDEX IF NOT EXISTS idx_tracking_user 
    ON updates_os_tracking(user_service_id);

CREATE INDEX IF NOT EXISTS idx_task_user 
    ON updates_os_task(user_service_id);