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

---

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
