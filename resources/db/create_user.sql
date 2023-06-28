-- Role: cue
-- DROP ROLE IF EXISTS cue;

CREATE ROLE cue WITH
  LOGIN
  SUPERUSER
  INHERIT
  CREATEDB
  CREATEROLE
  NOREPLICATION
;

-- PLEASE CHANGE THIS IN PROD
ALTER USER cue WITH PASSWORD 'cue';
