FROM postgres:latest

ENV POSTGRES_DB=cue \
    POSTGRES_USER=cue \
    POSTGRES_PASSWORD=cue \
    POSTGRES_PORT=5432

COPY resources/db/cue.sql /docker-entrypoint-initdb.d/
