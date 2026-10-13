-- Run while connected AS a runtime role (cc_app, cc_webhook, cc_worker).
-- Fails (ON_ERROR_STOP) unless the role is a plain login: no superuser, no BYPASSRLS,
-- not the database owner, no DDL on schema public.
\set ON_ERROR_STOP on

SELECT r.rolname, r.rolsuper, r.rolbypassrls, r.rolcreaterole, r.rolcreatedb,
       pg_get_userbyid(d.datdba) AS db_owner
FROM pg_roles r
JOIN pg_database d ON d.datname = current_database()
WHERE r.rolname = current_user;

DO $$
DECLARE
  r pg_roles%ROWTYPE;
BEGIN
  SELECT * INTO r FROM pg_roles WHERE rolname = current_user;
  IF r.rolsuper THEN RAISE EXCEPTION '% is SUPERUSER', current_user; END IF;
  IF r.rolbypassrls THEN RAISE EXCEPTION '% has BYPASSRLS', current_user; END IF;
  IF r.rolcreaterole OR r.rolcreatedb THEN RAISE EXCEPTION '% can create roles or databases', current_user; END IF;
  IF (SELECT pg_get_userbyid(datdba) FROM pg_database WHERE datname = current_database()) = current_user THEN
    RAISE EXCEPTION '% owns the database', current_user;
  END IF;
  IF has_schema_privilege(current_user, 'public', 'CREATE') THEN
    RAISE EXCEPTION '% can CREATE in schema public', current_user;
  END IF;
  RAISE NOTICE 'OK: % is NOSUPERUSER, NOBYPASSRLS, not owner, no DDL on public', current_user;
END
$$;
