package create

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

/*
this package is used to build the database tables. it is called once when the program starts up.
*/

var db *pgxpool.Conn

func Build(d *pgxpool.Pool, l *logger.Logger) error {
	conn, err := d.Acquire(context.Background())
	if err != nil {
		l.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return err
	}

	db = conn
	defer db.Release()

	if err := users(); err != nil {
		l.Info().AnErr("creating users table", err).Msg("failed to create users table")
		return err
	}

	if err := devices(); err != nil {
		l.Info().AnErr("creating devices table", err).Msg("failed to create devices table")
		return err
	}

	if err := exclusions(); err != nil {
		l.Info().AnErr("creating exclusions table", err).Msg("failed to create exclusions table")
		return err
	}

	if err := bots(); err != nil {
		l.Info().AnErr("creating bots table", err).Msg("failed to create bots table")
		return err
	}

	if err := triggers(); err != nil {
		l.Info().AnErr("creating triggers table", err).Msg("failed to create triggers table")
	}

	return nil

}

func exec(statement string) error {
	_, err := db.Exec(context.Background(), statement)
	return err
}

func bots() error {
	statement := `
DROP TABLE IF EXISTS bot_results;	
CREATE TABLE bot_results (
	slack_id character varying(255) NOT NULL,
	user_email character varying(255),
	manager_slack_id character varying(255),
	first_ack boolean,
	first_ack_time timestamp NOT NULL,
	first_message_sent boolean,
	first_message_sent_at timestamp NOT NULL,
	first_message_waiting boolean,
	manager_message_sent boolean,
	manager_message_sent_at timestamp NOT NULL,
	full_name character varying(255),
	delay_at timestamp NOT NULL,
	delay_date character varying(255),
	delay_time character varying(255),
	delay_sent boolean,
	serial_number character varying(255),
	tz_offset int,
	created_at timestamp,
	updated_at timestamp,
	PRIMARY KEY (serial_number)
);
	`
	return exec(statement)
}

func users() error {
	statement := `
DROP TABLE IF EXISTS users;
CREATE TABLE users (
	user_mdm_id character varying(255) NOT NULL,
	user_long_name character varying(255),
	user_email character varying(255),
	user_slack_id character varying(255),
	tz_offset int,
	created_at timestamp,
	updated_at timestamp,
	PRIMARY KEY (user_slack_id)
);


ALTER TABLE users OWNER TO postgres;
	`
	return exec(statement)
}

func devices() error {
	statement := `
DROP TABLE IF EXISTS devices;
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
	`
	return exec(statement)
}

func exclusions() error {
	statement := `
DROP TABLE IF EXISTS exclusions;
CREATE TABLE exclusions (
	approved boolean,
	serial_number character varying(255) NOT NULL,
	user_email character varying(255) NOT NULL,
	reason character varying(255) NOT NULL,
	until timestamp NOT NULL,
	created_at timestamp,
	updated_at timestamp,
	PRIMARY KEY (serial_number)
);
	`
	return exec(statement)
}

func triggers() error {
	statement := `
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
`
	return exec(statement)
}
