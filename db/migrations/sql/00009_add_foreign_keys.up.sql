-- Migration: Add foreign key constraints for data integrity
-- This migration adds critical foreign key relationships that were missing

-- First, we need to ensure data consistency before adding constraints
-- Clean up any orphaned records

-- Remove tracking records with non-existent users
DELETE FROM updates_os_tracking 
WHERE user_service_id::TEXT NOT IN (
    SELECT user_service_id::TEXT FROM updates_os_users
);

-- Remove tracking records with non-existent devices
DELETE FROM updates_os_tracking 
WHERE serial_number NOT IN (
    SELECT serial_number FROM updates_os_devices WHERE serial_number IS NOT NULL
);

-- Remove tasks with non-existent users
DELETE FROM updates_os_task 
WHERE user_service_id NOT IN (
    SELECT user_service_id::TEXT FROM updates_os_users
);

-- Remove exclusions with non-existent devices
DELETE FROM updates_os_exclusions 
WHERE serial_number NOT IN (
    SELECT serial_number FROM updates_os_devices WHERE serial_number IS NOT NULL
);

-- Now add the foreign key constraints

-- 1. Tracking table foreign keys
ALTER TABLE updates_os_tracking 
    ADD CONSTRAINT fk_tracking_device 
    FOREIGN KEY (serial_number) 
    REFERENCES updates_os_devices(serial_number) 
    ON DELETE CASCADE;

-- Note: Cannot add user foreign key yet due to type mismatch (INTEGER vs VARCHAR)
-- This will be fixed in the next migration

-- 2. Task table foreign keys
-- Note: Cannot add user foreign key yet due to type mismatch

-- 3. Exclusions table foreign keys
ALTER TABLE updates_os_exclusions 
    ADD CONSTRAINT fk_exclusion_device 
    FOREIGN KEY (serial_number) 
    REFERENCES updates_os_devices(serial_number) 
    ON DELETE CASCADE;

-- 4. Add unique constraints where needed
ALTER TABLE updates_os_devices 
    ADD CONSTRAINT uk_device_serial UNIQUE (serial_number);

ALTER TABLE updates_os_users 
    ADD CONSTRAINT uk_user_email UNIQUE (user_email);

-- 5. Add check constraints for data validation
ALTER TABLE updates_os_exclusions 
    ADD CONSTRAINT chk_exclusion_until CHECK (until > NOW());

ALTER TABLE updates_os_tracking 
    ADD CONSTRAINT chk_tracking_ack_time CHECK (
        first_ack = FALSE OR first_ack_time IS NOT NULL
    );

ALTER TABLE updates_os_tracking 
    ADD CONSTRAINT chk_tracking_sent_time CHECK (
        first_message_sent = FALSE OR first_message_sent_at IS NOT NULL
    );

-- Create indexes for foreign key columns
CREATE INDEX IF NOT EXISTS idx_tracking_device 
    ON updates_os_tracking(serial_number);

CREATE INDEX IF NOT EXISTS idx_exclusion_device 
    ON updates_os_exclusions(serial_number);