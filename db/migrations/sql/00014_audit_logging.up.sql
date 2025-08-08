-- Migration: Create comprehensive audit logging system
-- This provides complete tracking of all database changes for compliance and debugging

-- 1. Create the audit log table
CREATE TABLE IF NOT EXISTS audit_log (
    id BIGSERIAL PRIMARY KEY,
    
    -- What was changed
    table_name VARCHAR(50) NOT NULL,
    record_id VARCHAR(100) NOT NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE', 'TRUNCATE')),
    
    -- Change details
    old_values JSONB,
    new_values JSONB,
    changed_fields TEXT[], -- Array of field names that changed
    
    -- Who made the change
    changed_by VARCHAR(100), -- Application user or system process
    application_name VARCHAR(100), -- From application_name connection parameter
    session_user VARCHAR(100) DEFAULT CURRENT_USER,
    client_addr INET DEFAULT inet_client_addr(),
    
    -- When it happened
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Additional context
    transaction_id BIGINT DEFAULT txid_current(),
    statement_timestamp TIMESTAMPTZ DEFAULT statement_timestamp(),
    correlation_id VARCHAR(50), -- For tracking related changes
    metadata JSONB -- Additional context from application
);

-- 2. Create indexes for audit queries
CREATE INDEX idx_audit_table_record ON audit_log(table_name, record_id);
CREATE INDEX idx_audit_changed_at ON audit_log(changed_at DESC);
CREATE INDEX idx_audit_changed_by ON audit_log(changed_by) WHERE changed_by IS NOT NULL;
CREATE INDEX idx_audit_action ON audit_log(action);
CREATE INDEX idx_audit_transaction ON audit_log(transaction_id);
CREATE INDEX idx_audit_correlation ON audit_log(correlation_id) WHERE correlation_id IS NOT NULL;

-- 3. Create partitioning for audit log (by month for efficient archival)
CREATE TABLE audit_log_template (LIKE audit_log INCLUDING ALL) PARTITION BY RANGE (changed_at);

-- Create initial partitions
DO $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
BEGIN
    FOR i IN -1..2 LOOP -- Create partitions from last month to 2 months ahead
        start_date := DATE_TRUNC('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        partition_name := 'audit_log_' || TO_CHAR(start_date, 'YYYY_MM');
        
        EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_log_template FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date);
    END LOOP;
END $$;

-- 4. Create generic audit trigger function
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
DECLARE
    audit_row audit_log;
    changed_fields TEXT[];
    old_value TEXT;
    new_value TEXT;
    col_name TEXT;
BEGIN
    audit_row.table_name = TG_TABLE_NAME;
    audit_row.action = TG_OP;
    audit_row.changed_by = current_setting('app.current_user', true);
    audit_row.application_name = current_setting('application_name', true);
    audit_row.correlation_id = current_setting('app.correlation_id', true);
    
    IF TG_OP = 'DELETE' THEN
        audit_row.record_id = OLD.id::TEXT;
        audit_row.old_values = to_jsonb(OLD);
        INSERT INTO audit_log_template SELECT audit_row.*;
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        audit_row.record_id = NEW.id::TEXT;
        audit_row.old_values = to_jsonb(OLD);
        audit_row.new_values = to_jsonb(NEW);
        
        -- Identify changed fields
        changed_fields := ARRAY[]::TEXT[];
        FOR col_name IN SELECT column_name 
                        FROM information_schema.columns 
                        WHERE table_name = TG_TABLE_NAME 
                        AND table_schema = TG_TABLE_SCHEMA
        LOOP
            EXECUTE format('SELECT ($1).%I::TEXT', col_name) INTO old_value USING OLD;
            EXECUTE format('SELECT ($1).%I::TEXT', col_name) INTO new_value USING NEW;
            
            IF old_value IS DISTINCT FROM new_value THEN
                changed_fields := array_append(changed_fields, col_name);
            END IF;
        END LOOP;
        
        audit_row.changed_fields = changed_fields;
        
        -- Only log if something actually changed
        IF array_length(changed_fields, 1) > 0 THEN
            INSERT INTO audit_log_template SELECT audit_row.*;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'INSERT' THEN
        audit_row.record_id = NEW.id::TEXT;
        audit_row.new_values = to_jsonb(NEW);
        INSERT INTO audit_log_template SELECT audit_row.*;
        RETURN NEW;
    ELSIF TG_OP = 'TRUNCATE' THEN
        audit_row.record_id = 'ALL';
        INSERT INTO audit_log_template SELECT audit_row.*;
        RETURN NULL;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 5. Create specific audit trigger for tables without 'id' column
CREATE OR REPLACE FUNCTION audit_trigger_function_no_id()
RETURNS TRIGGER AS $$
DECLARE
    audit_row audit_log;
    record_key TEXT;
BEGIN
    audit_row.table_name = TG_TABLE_NAME;
    audit_row.action = TG_OP;
    audit_row.changed_by = current_setting('app.current_user', true);
    audit_row.application_name = current_setting('application_name', true);
    
    -- Build a composite key for tables without 'id'
    IF TG_TABLE_NAME = 'updates_os_tracking' THEN
        IF TG_OP = 'DELETE' THEN
            record_key = OLD.serial_number || ':' || OLD.user_service_id;
        ELSE
            record_key = NEW.serial_number || ':' || NEW.user_service_id;
        END IF;
    ELSIF TG_TABLE_NAME = 'updates_os_users' THEN
        IF TG_OP = 'DELETE' THEN
            record_key = OLD.user_service_id::TEXT;
        ELSE
            record_key = NEW.user_service_id::TEXT;
        END IF;
    ELSE
        -- Generic fallback
        record_key = md5(to_jsonb(COALESCE(NEW, OLD))::TEXT);
    END IF;
    
    audit_row.record_id = record_key;
    
    IF TG_OP = 'DELETE' THEN
        audit_row.old_values = to_jsonb(OLD);
    ELSIF TG_OP = 'UPDATE' THEN
        audit_row.old_values = to_jsonb(OLD);
        audit_row.new_values = to_jsonb(NEW);
    ELSIF TG_OP = 'INSERT' THEN
        audit_row.new_values = to_jsonb(NEW);
    END IF;
    
    INSERT INTO audit_log_template SELECT audit_row.*;
    
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- 6. Apply audit triggers to critical tables
CREATE TRIGGER audit_tracking AFTER INSERT OR UPDATE OR DELETE ON updates_os_tracking
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function_no_id();

CREATE TRIGGER audit_users AFTER INSERT OR UPDATE OR DELETE ON updates_os_users
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function_no_id();

CREATE TRIGGER audit_devices AFTER INSERT OR UPDATE OR DELETE ON updates_os_devices
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function_no_id();

CREATE TRIGGER audit_exclusions AFTER INSERT OR UPDATE OR DELETE ON updates_os_exclusions
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function_no_id();

CREATE TRIGGER audit_task AFTER INSERT OR UPDATE OR DELETE ON updates_os_task
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function_no_id();

-- Apply to unified tracking if it exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'unified_tracking') THEN
        EXECUTE 'CREATE TRIGGER audit_unified_tracking AFTER INSERT OR UPDATE OR DELETE ON unified_tracking
            FOR EACH ROW EXECUTE FUNCTION audit_trigger_function()';
    END IF;
END $$;

-- 7. Create utility functions for audit queries
CREATE OR REPLACE FUNCTION get_audit_history(
    p_table_name VARCHAR,
    p_record_id VARCHAR,
    p_limit INTEGER DEFAULT 100
) RETURNS TABLE (
    action VARCHAR,
    changed_at TIMESTAMPTZ,
    changed_by VARCHAR,
    old_values JSONB,
    new_values JSONB,
    changed_fields TEXT[]
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        a.action,
        a.changed_at,
        a.changed_by,
        a.old_values,
        a.new_values,
        a.changed_fields
    FROM audit_log_template a
    WHERE a.table_name = p_table_name
    AND a.record_id = p_record_id
    ORDER BY a.changed_at DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_user_activity(
    p_user VARCHAR,
    p_date DATE DEFAULT CURRENT_DATE
) RETURNS TABLE (
    table_name VARCHAR,
    action VARCHAR,
    record_id VARCHAR,
    changed_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        a.table_name,
        a.action,
        a.record_id,
        a.changed_at
    FROM audit_log_template a
    WHERE a.changed_by = p_user
    AND DATE(a.changed_at) = p_date
    ORDER BY a.changed_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 8. Create maintenance function for audit log cleanup
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs(
    retention_months INTEGER DEFAULT 12
) RETURNS INTEGER AS $$
DECLARE
    cutoff_date DATE;
    deleted_count INTEGER := 0;
    partition_name TEXT;
BEGIN
    cutoff_date := DATE_TRUNC('month', CURRENT_DATE - (retention_months || ' months')::INTERVAL);
    
    -- Drop old partitions
    FOR partition_name IN 
        SELECT tablename 
        FROM pg_tables 
        WHERE tablename LIKE 'audit_log_%'
        AND tablename < 'audit_log_' || TO_CHAR(cutoff_date, 'YYYY_MM')
    LOOP
        EXECUTE format('DROP TABLE IF EXISTS %I', partition_name);
        deleted_count := deleted_count + 1;
        RAISE NOTICE 'Dropped partition: %', partition_name;
    END LOOP;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 9. Add helpful comments
COMMENT ON TABLE audit_log IS 'Comprehensive audit trail of all database changes';
COMMENT ON FUNCTION audit_trigger_function IS 'Generic audit trigger for tables with id column';
COMMENT ON FUNCTION audit_trigger_function_no_id IS 'Audit trigger for tables without id column';
COMMENT ON FUNCTION get_audit_history IS 'Get audit history for a specific record';
COMMENT ON FUNCTION get_user_activity IS 'Get all changes made by a specific user on a given date';
COMMENT ON FUNCTION cleanup_old_audit_logs IS 'Remove audit logs older than specified months';