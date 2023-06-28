CREATE TABLE bot_results (
    slack_id character varying(255) NOT NULL,
    manager_slack_id character varying(255),
    first_ack boolean,
    first_ack_time timestamp NOT NULL,
    first_message_sent boolean,
    delay_at timestamp NOT NULL,
    delay_date character varying(255),
    delay_time character varying(255),
    serial_number character varying(255),
    created_at timestamp,
    updated_at timestamp,
    PRIMARY KEY (serial_number)
);

ALTER TABLE bot_results OWNER TO cue;

CREATE TABLE devices (
    device_id character varying(255) NOT NULL,
    device_name character varying(255) NOT NULL,
    model character varying(255) NOT NULL,
    serial_number character varying(255) NOT NULL,
    platform character varying(255) NOT NULL,
    os_version character varying(255) NOT NULL,
    user_name character varying(255) NOT NULL,
    user_mdm_id character varying(255) NOT NULL,
    last_check_in timestamp,
    created_at timestamp,
    updated_at timestamp,
    PRIMARY KEY (device_id)
);

ALTER TABLE devices OWNER TO cue;

CREATE TABLE exclusions (
    serial_number character varying(255) NOT NULL,
    user_email character varying(255) NOT NULL,
    reason character varying(255) NOT NULL,
    until timestamp NOT NULL,
    created_at timestamp,
    updated_at timestamp,
    PRIMARY KEY (serial_number)
);

ALTER TABLE devices OWNER TO cue;

CREATE TABLE users (
    user_mdm_id character varying(255) NOT NULL,
    user_long_name character varying(255),
    user_email character varying(255),
    user_slack_id character varying(255),
    created_at timestamp,
    updated_at timestamp,
    PRIMARY KEY (user_slack_id)
);

ALTER TABLE users OWNER TO cue;

CREATE OR REPLACE FUNCTION notify_devices_event() RETURNS TRIGGER AS $$

    DECLARE
        data json;
        notification json;

    BEGIN
        -- Convert the old or new row to JSON, based on the kind of action.
        -- Action = DELETE?             -> OLD row
        -- Action = INSERT or UPDATE?   -> NEW row
        IF (TG_OP = 'DELETE') THEN
            data = row_to_json(OLD);
        ELSE
            data = row_to_json(NEW);
        END IF;

        -- Contruct the notification as a JSON string.
        notification = json_build_object(
                          'table',TG_TABLE_NAME,
                          'action', TG_OP,
                          'data', data);

        -- Execute pg_notify(channel, notification)
        PERFORM pg_notify('devices',notification::text);

        -- Result is ignored since this is an AFTER trigger
        RETURN NULL;
    END;

$$ LANGUAGE plpgsql;

CREATE TRIGGER devices_notify_event
AFTER INSERT OR UPDATE OR DELETE ON devices
    FOR EACH ROW EXECUTE PROCEDURE notify_devices_event();

CREATE OR REPLACE FUNCTION notify_bot_event() RETURNS TRIGGER AS $$

    DECLARE
        data json;
        notification json;

    BEGIN
        -- Convert the old or new row to JSON, based on the kind of action.
        -- Action = DELETE?             -> OLD row
        -- Action = INSERT or UPDATE?   -> NEW row
        IF (TG_OP = 'DELETE') THEN
            data = row_to_json(OLD);
        ELSE
            data = row_to_json(NEW);
        END IF;

        -- Contruct the notification as a JSON string.
        notification = json_build_object(
                          'table',TG_TABLE_NAME,
                          'action', TG_OP,
                          'data', data);

        -- Execute pg_notify(channel, notification)
        PERFORM pg_notify('bot_results',notification::text);

        -- Result is ignored since this is an AFTER trigger
        RETURN NULL;
    END;

$$ LANGUAGE plpgsql;

CREATE TRIGGER bot_notify_event
AFTER INSERT OR UPDATE OR DELETE ON bot_results
    FOR EACH ROW EXECUTE PROCEDURE notify_bot_event();

CREATE OR REPLACE FUNCTION log_last_updated()
  RETURNS TRIGGER AS $$
  BEGIN
      NEW.updated_at = now();
      RETURN NEW;
  END;
  $$ language 'plpgsql';

CREATE TRIGGER update_device_time
BEFORE INSERT or UPDATE ON devices
FOR EACH ROW
EXECUTE PROCEDURE log_last_updated();

CREATE TRIGGER update_user_time
BEFORE INSERT or UPDATE ON users
FOR EACH ROW
EXECUTE PROCEDURE log_last_updated();

CREATE TRIGGER update_bot_time
BEFORE INSERT or UPDATE ON bot_results
FOR EACH ROW
EXECUTE PROCEDURE log_last_updated();

CREATE TRIGGER update_bot_time
BEFORE INSERT or UPDATE ON exclusions
FOR EACH ROW
EXECUTE PROCEDURE log_last_updated();
