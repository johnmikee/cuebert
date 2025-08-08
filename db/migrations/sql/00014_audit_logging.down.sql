-- Rollback migration: Remove audit logging system

-- Drop triggers
DROP TRIGGER IF EXISTS audit_tracking ON updates_os_tracking;
DROP TRIGGER IF EXISTS audit_users ON updates_os_users;
DROP TRIGGER IF EXISTS audit_devices ON updates_os_devices;
DROP TRIGGER IF EXISTS audit_exclusions ON updates_os_exclusions;
DROP TRIGGER IF EXISTS audit_task ON updates_os_task;
DROP TRIGGER IF EXISTS audit_unified_tracking ON unified_tracking;

-- Drop utility functions
DROP FUNCTION IF EXISTS get_audit_history(VARCHAR, VARCHAR, INTEGER);
DROP FUNCTION IF EXISTS get_user_activity(VARCHAR, DATE);
DROP FUNCTION IF EXISTS cleanup_old_audit_logs(INTEGER);

-- Drop trigger functions
DROP FUNCTION IF EXISTS audit_trigger_function();
DROP FUNCTION IF EXISTS audit_trigger_function_no_id();

-- Drop all audit log partitions
DO $$
DECLARE
    partition_record RECORD;
BEGIN
    FOR partition_record IN 
        SELECT tablename 
        FROM pg_tables 
        WHERE tablename LIKE 'audit_log_%'
    LOOP
        EXECUTE format('DROP TABLE IF EXISTS %I CASCADE', partition_record.tablename);
    END LOOP;
END $$;

-- Drop partition template table
DROP TABLE IF EXISTS audit_log_template CASCADE;

-- Drop the main audit log table
DROP TABLE IF EXISTS audit_log CASCADE;