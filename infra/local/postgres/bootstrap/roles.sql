-- Role bootstrap for the LOCAL stack (P0-04). Roles: docs/architecture/data-model.md.
-- Run by initdb/10-roles.sh with psql variables :db_name and :cc_*_password.
--
-- Least privilege:
--   cc_owner    owns the database and schema public; runs goose migrations only.
--   cc_app      API. LOGIN, not owner, NOBYPASSRLS. CONNECT + USAGE only here.
--   cc_webhook  RevenueCat webhook pool. Same posture as cc_app.
--   cc_worker   Worker. Same posture as cc_app.
--   cc_auth_definer
--               NOLOGIN, NOBYPASSRLS. Owns the auth SECURITY DEFINER functions (M1-03)
--               and reaches only the rows its narrow RLS policies allow. cc_owner is a
--               member WITH INHERIT FALSE, SET TRUE: it may SET ROLE to it (needed to hand
--               it the functions and to drop them on Down) but never inherits its
--               privileges or policies. No runtime role is a member.
-- Table, sequence and function grants are NOT given here: they belong to the versioned
-- baseline migration (P0-10), which cc_owner runs, so every environment gets the same grants.
-- None of the roles is SUPERUSER, CREATEDB, CREATEROLE, REPLICATION or BYPASSRLS.

\set ON_ERROR_STOP on

CREATE ROLE cc_owner   LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS INHERIT
  PASSWORD :'cc_owner_password';
CREATE ROLE cc_app     LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS NOINHERIT
  PASSWORD :'cc_app_password';
CREATE ROLE cc_webhook LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS NOINHERIT
  PASSWORD :'cc_webhook_password';
CREATE ROLE cc_worker  LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS NOINHERIT
  PASSWORD :'cc_worker_password';
CREATE ROLE cc_auth_definer NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS NOINHERIT;
GRANT cc_auth_definer TO cc_owner WITH INHERIT FALSE, SET TRUE;

-- Runtime roles fail fast instead of piling up on locks or idle transactions.
ALTER ROLE cc_app     SET statement_timeout = '15s';
ALTER ROLE cc_app     SET idle_in_transaction_session_timeout = '30s';
ALTER ROLE cc_webhook SET statement_timeout = '15s';
ALTER ROLE cc_webhook SET idle_in_transaction_session_timeout = '30s';
ALTER ROLE cc_worker  SET statement_timeout = '60s';
ALTER ROLE cc_worker  SET idle_in_transaction_session_timeout = '60s';

ALTER DATABASE :"db_name" OWNER TO cc_owner;
REVOKE ALL ON DATABASE :"db_name" FROM PUBLIC;
GRANT CONNECT ON DATABASE :"db_name" TO cc_app, cc_webhook, cc_worker;

-- The maintenance databases keep PostgreSQL's default ACL (PUBLIC: CONNECT, TEMP),
-- which would let every cc_* role log in there and create temp objects. Only the
-- bootstrap superuser needs them (P0-30, PR #5 follow-up 2).
REVOKE ALL ON DATABASE postgres FROM PUBLIC;
REVOKE ALL ON DATABASE template1 FROM PUBLIC;

\connect :"db_name"

ALTER SCHEMA public OWNER TO cc_owner;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO cc_app, cc_webhook, cc_worker;

-- Functions are executable by PUBLIC by default; SECURITY DEFINER functions must be
-- granted explicitly by the migration that creates them.
ALTER DEFAULT PRIVILEGES FOR ROLE cc_owner REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
