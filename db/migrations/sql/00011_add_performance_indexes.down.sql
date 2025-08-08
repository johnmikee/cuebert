-- Rollback migration: Remove performance indexes

-- Drop tracking table indexes
DROP INDEX IF EXISTS idx_tracking_user_slack;
DROP INDEX IF EXISTS idx_tracking_complete;
DROP INDEX IF EXISTS idx_tracking_message_status;
DROP INDEX IF EXISTS idx_tracking_need_first_msg;
DROP INDEX IF EXISTS idx_tracking_need_final_msg;
DROP INDEX IF EXISTS idx_tracking_reminders;
DROP INDEX IF EXISTS idx_tracking_updated;
DROP INDEX IF EXISTS idx_tracking_first_sent_at;
DROP INDEX IF EXISTS idx_tracking_covering;

-- Drop device table indexes
DROP INDEX IF EXISTS idx_device_user;
DROP INDEX IF EXISTS idx_device_checkin;
DROP INDEX IF EXISTS idx_device_os_version;
DROP INDEX IF EXISTS idx_device_compliance;

-- Drop exclusions table indexes
DROP INDEX IF EXISTS idx_exclusion_active;
DROP INDEX IF EXISTS idx_exclusion_user;
DROP INDEX IF EXISTS idx_exclusion_validation;

-- Drop task table indexes
DROP INDEX IF EXISTS idx_task_complete;
DROP INDEX IF EXISTS idx_task_name;

-- Drop updates table indexes
DROP INDEX IF EXISTS idx_updates_week;
DROP INDEX IF EXISTS idx_updates_item;

-- Drop overdue users table indexes
DROP INDEX IF EXISTS idx_overdue_task;
DROP INDEX IF EXISTS idx_overdue_email;