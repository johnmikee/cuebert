-- Rollback migration: Remove unified tracking table and related objects

-- Drop views
DROP VIEW IF EXISTS v_compliance_tracking;
DROP VIEW IF EXISTS v_os_update_tracking;

-- Drop function for partition creation
DROP FUNCTION IF EXISTS create_monthly_tracking_partition();

-- Drop all partitions
DO $$
DECLARE
    partition_record RECORD;
BEGIN
    FOR partition_record IN 
        SELECT tablename 
        FROM pg_tables 
        WHERE tablename LIKE 'unified_tracking_%'
    LOOP
        EXECUTE format('DROP TABLE IF EXISTS %I CASCADE', partition_record.tablename);
    END LOOP;
END $$;

-- Drop partition template table
DROP TABLE IF EXISTS unified_tracking_template CASCADE;

-- Drop the main unified tracking table
DROP TABLE IF EXISTS unified_tracking CASCADE;