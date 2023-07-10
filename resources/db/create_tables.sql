--
-- Name: bot_results; Type: TABLE; Schema: public; Owner: cue
--

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
    reminder_interval int,
    reminder_waiting boolean,
    serial_number character varying(255),
    tz_offset int,
    created_at timestamp,
    updated_at timestamp,
    PRIMARY KEY (serial_number)
);

ALTER TABLE bot_results OWNER TO cue;

--
-- TOC entry 214 (class 1259 OID 16400)
-- Name: devices; Type: TABLE; Schema: public; Owner: cue
--

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

--
-- TOC entry 214 (class 1259 OID 16400)
-- Name: exclusions; Type: TABLE; Schema: public; Owner: cue
--

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


ALTER TABLE devices OWNER TO cue;

--
-- TOC entry 215 (class 1259 OID 16405)
-- Name: users; Type: TABLE; Schema: public; Owner: cue
--

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


ALTER TABLE users OWNER TO cue;
