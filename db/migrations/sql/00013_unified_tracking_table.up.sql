-- Migration: Create unified tracking table to consolidate OS updates and compliance tracking
-- This eliminates duplication and provides a single source of truth

-- 1. Create the unified tracking table
CREATE TABLE IF NOT EXISTS unified_tracking (
    -- Primary key
    id BIGSERIAL PRIMARY KEY,
    
    -- Tracking type to distinguish between different tracking purposes
    tracking_type VARCHAR(20) NOT NULL CHECK (tracking_type IN ('os_update', 'compliance', 'security', 'custom')),
    
    -- User information
    user_id INTEGER NOT NULL,
    user_slack_id VARCHAR(15),
    user_email VARCHAR(100),
    user_full_name VARCHAR(100),
    
    -- Device information (optional for non-device tasks)
    device_serial VARCHAR(30),
    device_os VARCHAR(20),
    device_platform VARCHAR(20),
    
    -- Task information
    task_id VARCHAR(50) NOT NULL,
    task_name VARCHAR(100) NOT NULL,
    task_description TEXT,
    task_complete BOOLEAN NOT NULL DEFAULT FALSE,
    task_deadline TIMESTAMPTZ,
    
    -- Message tracking
    first_message_sent BOOLEAN NOT NULL DEFAULT FALSE,
    first_message_sent_at TIMESTAMPTZ,
    first_ack BOOLEAN NOT NULL DEFAULT FALSE,
    first_ack_at TIMESTAMPTZ,
    
    final_message_sent BOOLEAN NOT NULL DEFAULT FALSE,
    final_message_sent_at TIMESTAMPTZ,
    final_ack BOOLEAN NOT NULL DEFAULT FALSE,
    final_ack_at TIMESTAMPTZ,
    
    -- Reminder tracking
    reminder_scheduled_at TIMESTAMPTZ,
    reminder_sent_at TIMESTAMPTZ,
    reminder_count INTEGER DEFAULT 0,
    reminder_interval INTEGER, -- in minutes
    reminder_waiting BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Manager escalation
    manager_notified BOOLEAN NOT NULL DEFAULT FALSE,
    manager_notified_at TIMESTAMPTZ,
    manager_slack_id VARCHAR(15),
    
    -- Timezone and scheduling
    tz_offset INTEGER DEFAULT 0, -- hours from UTC
    
    -- Flexible metadata storage for task-specific data
    metadata JSONB DEFAULT '{}',
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(100),
    updated_by VARCHAR(100),
    
    -- Constraints
    CONSTRAINT uk_user_task UNIQUE(user_id, task_id, tracking_type),
    CONSTRAINT chk_ack_time CHECK (first_ack = FALSE OR first_ack_at IS NOT NULL),
    CONSTRAINT chk_sent_time CHECK (first_message_sent = FALSE OR first_message_sent_at IS NOT NULL),
    CONSTRAINT chk_final_sent_time CHECK (final_message_sent = FALSE OR final_message_sent_at IS NOT NULL),
    CONSTRAINT chk_manager_notified CHECK (manager_notified = FALSE OR manager_notified_at IS NOT NULL)
);

-- 2. Create indexes for optimal query performance
CREATE INDEX idx_unified_tracking_type ON unified_tracking(tracking_type);
CREATE INDEX idx_unified_user_id ON unified_tracking(user_id);
CREATE INDEX idx_unified_user_slack ON unified_tracking(user_slack_id);
CREATE INDEX idx_unified_device ON unified_tracking(device_serial) WHERE device_serial IS NOT NULL;
CREATE INDEX idx_unified_task_id ON unified_tracking(task_id);
CREATE INDEX idx_unified_task_complete ON unified_tracking(task_complete) WHERE task_complete = FALSE;
CREATE INDEX idx_unified_deadline ON unified_tracking(task_deadline) WHERE task_deadline IS NOT NULL;
CREATE INDEX idx_unified_created_at ON unified_tracking(created_at);
CREATE INDEX idx_unified_reminder ON unified_tracking(reminder_scheduled_at) WHERE reminder_waiting = TRUE;
CREATE INDEX idx_unified_needs_first ON unified_tracking(tracking_type, task_complete, first_message_sent) 
    WHERE task_complete = FALSE AND first_message_sent = FALSE;
CREATE INDEX idx_unified_needs_final ON unified_tracking(tracking_type, task_complete, first_message_sent, final_message_sent) 
    WHERE task_complete = FALSE AND first_message_sent = TRUE AND final_message_sent = FALSE;

-- 3. Add foreign key constraints
ALTER TABLE unified_tracking 
    ADD CONSTRAINT fk_unified_user 
    FOREIGN KEY (user_id) 
    REFERENCES updates_os_users(user_service_id) 
    ON DELETE CASCADE;

-- 4. Create trigger for updated_at
CREATE TRIGGER update_unified_tracking_time 
    BEFORE UPDATE ON unified_tracking 
    FOR EACH ROW 
    EXECUTE FUNCTION log_last_updated();

-- 5. Create partitioning for scalability (by month)
-- This allows for efficient archiving of old data
CREATE TABLE unified_tracking_template (LIKE unified_tracking INCLUDING ALL) PARTITION BY RANGE (created_at);

-- Create initial partitions for the next 6 months
DO $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
BEGIN
    FOR i IN 0..5 LOOP
        start_date := DATE_TRUNC('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        partition_name := 'unified_tracking_' || TO_CHAR(start_date, 'YYYY_MM');
        
        EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF unified_tracking_template FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date);
    END LOOP;
END $$;

-- 6. Create function to automatically create new partitions
CREATE OR REPLACE FUNCTION create_monthly_tracking_partition()
RETURNS void AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    -- Create partition for next month if it doesn't exist
    start_date := DATE_TRUNC('month', CURRENT_DATE + INTERVAL '1 month');
    end_date := start_date + INTERVAL '1 month';
    partition_name := 'unified_tracking_' || TO_CHAR(start_date, 'YYYY_MM');
    
    -- Check if partition exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_tables 
        WHERE tablename = partition_name
    ) THEN
        EXECUTE format('CREATE TABLE %I PARTITION OF unified_tracking_template FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date);
        RAISE NOTICE 'Created partition: %', partition_name;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- 7. Schedule monthly partition creation (requires pg_cron extension or external scheduler)
-- This is a comment for documentation - actual scheduling depends on your setup
-- SELECT cron.schedule('create-tracking-partition', '0 0 1 * *', 'SELECT create_monthly_tracking_partition()');

-- 8. Create views for backward compatibility during migration
CREATE OR REPLACE VIEW v_os_update_tracking AS
SELECT 
    id,
    user_id as user_service_id,
    user_slack_id,
    user_email,
    user_full_name,
    device_serial as serial_number,
    task_name,
    task_complete,
    first_message_sent,
    first_message_sent_at,
    first_ack,
    first_ack_at as first_ack_time,
    final_message_sent,
    final_message_sent_at,
    reminder_scheduled_at as user_reminder_at,
    reminder_waiting,
    reminder_interval,
    tz_offset,
    updated_at
FROM unified_tracking
WHERE tracking_type = 'os_update';

CREATE OR REPLACE VIEW v_compliance_tracking AS
SELECT 
    id,
    user_id as user_service_id,
    user_slack_id,
    user_email,
    user_full_name,
    task_id,
    task_name,
    task_description,
    task_complete,
    task_deadline,
    first_message_sent,
    first_message_sent_at,
    first_ack,
    first_ack_at,
    final_message_sent,
    final_message_sent_at,
    manager_notified,
    manager_notified_at,
    metadata,
    updated_at
FROM unified_tracking
WHERE tracking_type = 'compliance';

-- 9. Add helpful comments
COMMENT ON TABLE unified_tracking IS 'Unified tracking table for all task types (OS updates, compliance, etc.)';
COMMENT ON COLUMN unified_tracking.tracking_type IS 'Type of tracking: os_update, compliance, security, custom';
COMMENT ON COLUMN unified_tracking.metadata IS 'JSON storage for task-specific data that doesn''t fit the standard schema';
COMMENT ON COLUMN unified_tracking.tz_offset IS 'User timezone offset in hours from UTC';