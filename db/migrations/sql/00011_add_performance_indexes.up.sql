-- Migration: Add missing indexes for query performance optimization
-- These indexes target the most frequently queried columns based on application patterns

-- 1. Tracking table indexes (most critical for performance)

-- Index for user lookups (very frequent)
CREATE INDEX IF NOT EXISTS idx_tracking_user_slack 
    ON updates_os_tracking(user_slack_id);

-- Index for task completion status (filtered in most queries)
CREATE INDEX IF NOT EXISTS idx_tracking_complete 
    ON updates_os_tracking(task_complete) 
    WHERE task_complete = FALSE; -- Partial index for incomplete tasks

-- Composite index for status queries
CREATE INDEX IF NOT EXISTS idx_tracking_message_status 
    ON updates_os_tracking(task_complete, first_message_sent, final_message_sent) 
    WHERE task_complete = FALSE;

-- Index for finding users needing first message
CREATE INDEX IF NOT EXISTS idx_tracking_need_first_msg 
    ON updates_os_tracking(first_message_sent, task_complete) 
    WHERE first_message_sent = FALSE AND task_complete = FALSE;

-- Index for finding users needing final message
CREATE INDEX IF NOT EXISTS idx_tracking_need_final_msg 
    ON updates_os_tracking(final_message_sent, first_message_sent, task_complete) 
    WHERE final_message_sent = FALSE AND first_message_sent = TRUE AND task_complete = FALSE;

-- Index for reminder queries
CREATE INDEX IF NOT EXISTS idx_tracking_reminders 
    ON updates_os_tracking(reminder_waiting, user_reminder_at) 
    WHERE reminder_waiting = TRUE;

-- Index for timestamp-based queries (for reporting and cleanup)
CREATE INDEX IF NOT EXISTS idx_tracking_updated 
    ON updates_os_tracking(updated_at);

CREATE INDEX IF NOT EXISTS idx_tracking_first_sent_at 
    ON updates_os_tracking(first_message_sent_at) 
    WHERE first_message_sent_at IS NOT NULL;

-- 2. Device table indexes

-- Index for user device lookups
CREATE INDEX IF NOT EXISTS idx_device_user 
    ON updates_os_devices(device_user);

-- Index for last check-in queries (for stale device detection)
CREATE INDEX IF NOT EXISTS idx_device_checkin 
    ON updates_os_devices(last_check_in);

-- Index for OS version queries
CREATE INDEX IF NOT EXISTS idx_device_os_version 
    ON updates_os_devices(os_version);

-- Composite index for device compliance queries
CREATE INDEX IF NOT EXISTS idx_device_compliance 
    ON updates_os_devices(platform, os_version);

-- 3. Exclusions table indexes

-- Index for active exclusions
CREATE INDEX IF NOT EXISTS idx_exclusion_active 
    ON updates_os_exclusions(until) 
    WHERE until > NOW();

-- Index for user exclusion lookups
CREATE INDEX IF NOT EXISTS idx_exclusion_user 
    ON updates_os_exclusions(user_email);

-- Composite index for exclusion validation
CREATE INDEX IF NOT EXISTS idx_exclusion_validation 
    ON updates_os_exclusions(serial_number, until) 
    WHERE approved = TRUE;

-- 4. Task table indexes

-- Index for task completion status
CREATE INDEX IF NOT EXISTS idx_task_complete 
    ON updates_os_task(task_complete) 
    WHERE task_complete = FALSE;

-- Index for task name lookups
CREATE INDEX IF NOT EXISTS idx_task_name 
    ON updates_os_task(item_name);

-- 5. Updates table indexes (for weekly scheduling)

-- Index for week lookups
CREATE INDEX IF NOT EXISTS idx_updates_week 
    ON updates_os_updates(week);

-- Index for item name lookups
CREATE INDEX IF NOT EXISTS idx_updates_item 
    ON updates_os_updates(item_name);

-- 6. Overdue users table indexes

-- Index for task lookups
CREATE INDEX IF NOT EXISTS idx_overdue_task 
    ON overdue_users(task);

-- Index for user email lookups
CREATE INDEX IF NOT EXISTS idx_overdue_email 
    ON overdue_users(user_email);

-- 7. Create statistics for better query planning
ANALYZE updates_os_tracking;
ANALYZE updates_os_devices;
ANALYZE updates_os_users;
ANALYZE updates_os_exclusions;
ANALYZE updates_os_task;

-- 8. Create a covering index for the most common query pattern
-- This assumes queries often need user info, device info, and status together
CREATE INDEX IF NOT EXISTS idx_tracking_covering 
    ON updates_os_tracking(
        user_slack_id,
        task_complete,
        first_message_sent,
        final_message_sent,
        serial_number
    ) 
    INCLUDE (
        user_email,
        user_full_name,
        first_message_sent_at,
        final_message_sent_at
    ) 
    WHERE task_complete = FALSE;