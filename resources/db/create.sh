#!/bin/sh

REQUIRED=( psql )
for r in "${REQUIRED[@]}"; do
    command -v "$r" &> /dev/null || { echo >&2 "This requires $r but it's not installed.  Exiting."; exit 1; }
done

base_dir=$(git rev-parse --show-toplevel)

ALL=false
DB=false
DBNAME=postgres
HOST=localhost
PORT=5432
TABLES=false
USER=false
USERNAME=postgres

while getopts a:d:h:n:p:t:u:un flag
do
    case "${flag}" in
        a) ALL=${OPTARG};;
        d) DB=${OPTARG};;
        h) HOST=${OPTARG};;
        n) DBNAME=${OPTARG};;
        p) PORT=${OPTARG};;
        t) TABLES=${OPTARG};;
        tg) TRIGGERS=${OPTARG};;
        u) USER=${OPTARG};;
        un) USERNAME=${OPTARG};;
    esac
done

cd "$base_dir"/db/

db() {
    psql --username=$USERNAME --host=$HOST --port=$PORT < "$base_dir"/resources/db/create_db.sql
}

tables() {
    psql --user=cue --host=$HOST --dbname=$DBNAME --port=$PORT < "$base_dir"/resources/db/create_tables.sql
}

triggers() {
    psql --user=cue --host=$HOST --dbname=$DBNAME --port=$PORT < "$base_dir"/resources/db/create_triggers.sql
}

user() {
    psql --username=$USERNAME --host=$HOST --port=$PORT < "$base_dir"/resources/db/create_user.sql
}

all() {
    user
    db
    tables
    triggers
}

if $ALL; then
    all
    exit 0
elif $USER; then
    user
elif $DB; then
    db
elif $TABLES; then
    tables
elif $TRIGGERS; then
    triggers
else
    echo "please check the arguements passed and try again"
fi
