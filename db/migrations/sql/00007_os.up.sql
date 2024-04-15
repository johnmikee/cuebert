CREATE TABLE IF NOT EXISTS updates_os_users (
    user_service_id INTEGER NOT NULL,
    user_long_name VARCHAR(255) NULL,
    user_email VARCHAR(255) NULL,
    user_slack_id VARCHAR(255) NULL,
    tz_offset INTEGER,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    PRIMARY KEY (user_service_id)
);

CREATE TABLE IF NOT EXISTS updates_os_exclusions (
    approved BOOLEAN,
    serial_number VARCHAR(255) NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    reason VARCHAR(255) NOT NULL,
    until TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE TABLE IF NOT EXISTS updates_os_task (
    user_service_id VARCHAR(255) NOT NULL,
    task_id integer NOT NULL,
    item_name VARCHAR(255) NULL,
    task_description VARCHAR(255) NOT NULL,
    task_complete BOOLEAN,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    PRIMARY KEY (user_service_id)
);

CREATE TABLE IF NOT EXISTS updates_os_tracking (
    serial_number VARCHAR(255) NOT NULL,
    user_service_id VARCHAR(255) NOT NULL,
    task_complete BOOLEAN DEFAULT FALSE,
    task_name VARCHAR(255) NOT NULL,
    user_slack_id VARCHAR(255) NULL,
    user_email VARCHAR(255) NULL,
    user_full_name VARCHAR(255) NULL,
    first_ack BOOLEAN DEFAULT FALSE,
    first_ack_time TIMESTAMP WITHOUT TIME ZONE,
    first_message_sent BOOLEAN DEFAULT FALSE,
    first_message_sent_at TIMESTAMP WITHOUT TIME ZONE,
    first_message_waiting BOOLEAN DEFAULT FALSE,
    user_reminder_at TIMESTAMP WITHOUT TIME ZONE,
    user_reminder_date character varying(255) NULL,
    user_reminder_time character varying(255) NULL,
    reminder_sent BOOLEAN DEFAULT FALSE,
    reminder_interval INTEGER,
    reminder_waiting BOOLEAN DEFAULT FALSE,
    tz_offset INTEGER,
    final_message_sent BOOLEAN,
    final_message_sent_at TIMESTAMP WITHOUT TIME ZONE,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    PRIMARY KEY (serial_number, user_service_id)
);

CREATE TABLE IF NOT EXISTS updates_os_updates (
    week integer NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    monday INTEGER,
    tuesday INTEGER,
    wednesday INTEGER,
    thursday INTEGER,
    friday INTEGER,
    updated_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE TABLE IF NOT EXISTS updates_os_devices (
    device_id VARCHAR(255) NOT NULL,
    device_name VARCHAR(255) NULL,
    model VARCHAR(255) NULL,
    serial_number VARCHAR(255) NULL,
    platform VARCHAR(255) NULL,
    os_version VARCHAR(255) NULL,
    device_user VARCHAR(255) NULL,
    user_mdm_id VARCHAR(255) NULL,
    last_check_in TIMESTAMP WITHOUT TIME ZONE,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    PRIMARY KEY (device_id)
);

CREATE INDEX idx_os_user_email ON updates_os_users (user_email);

CREATE INDEX idx_os_device_id ON updates_os_devices (device_id);

CREATE INDEX idx_os_serial_number ON updates_os_exclusions (serial_number);

CREATE INDEX idx_os_user_service_id ON updates_os_task (user_service_id);

CREATE INDEX idx_os_serial_number_user_service_id ON updates_os_tracking (serial_number, user_service_id);

CREATE TRIGGER update_updates_exclusion_time BEFORE INSERT OR UPDATE ON updates_os_exclusions FOR EACH ROW EXECUTE FUNCTION log_last_updated();

CREATE TRIGGER update_updates_task_time BEFORE INSERT OR UPDATE ON updates_os_task FOR EACH ROW EXECUTE FUNCTION log_last_updated();

CREATE TRIGGER update_updates_tracking_time BEFORE INSERT OR UPDATE ON updates_os_tracking FOR EACH ROW EXECUTE FUNCTION log_last_updated();

CREATE TRIGGER update_updates_updates_time BEFORE INSERT OR UPDATE ON updates_os_updates FOR EACH ROW EXECUTE FUNCTION log_last_updated();

CREATE TRIGGER update_updates_user_time BEFORE INSERT OR UPDATE ON updates_os_users FOR EACH ROW EXECUTE FUNCTION log_last_updated();
