-- Rate-limit buckets (M1-10, ADR-0005): the auth limits of MODULE-auth §5 shared by every
-- API instance. GCRA (generic cell rate algorithm): one row per (scope, key_hash) holds
-- the bucket's theoretical arrival time (tat) instead of a counter and a refill stamp.
-- rate_limit_take decides with one UPDATE ... WHERE, which re-checks the condition on
-- the latest row version when it waits on a concurrent take (READ COMMITTED), so N
-- instances racing on one key admit exactly the burst. A denied take writes nothing.
--
-- Keys are 32-byte HMACs computed app-side with a server key (never an email or an IP).
-- The caller passes its clock (injected, so tests are deterministic); instance clocks
-- are NTP-synced and the windows are minutes, so skew is immaterial.
--
-- Not user-scoped (keys are client addresses and email HMACs, many with no account), so
-- there is no user policy: ENABLE + FORCE with permissive cc_app policies, and grants
-- narrowed to what take and reset need. No DELETE for any runtime role: buckets whose
-- tat has passed are full again (the same state as a missing row) and the worker purges
-- them in M9 (there is deliberately no index on tat: it would make every take a non-HOT
-- update; the purge scans).

-- +goose Up
CREATE TABLE rate_limit_buckets (
  scope    text        NOT NULL,
  key_hash bytea       NOT NULL,
  tat      timestamptz NOT NULL,
  PRIMARY KEY (scope, key_hash),
  CONSTRAINT rate_limit_buckets_scope_format CHECK (scope ~ '^[a-z][a-z0-9_]{0,62}$'),
  -- 32 = crypto.HashSize (HMAC-SHA256).
  CONSTRAINT rate_limit_buckets_key_hash_size CHECK (octet_length(key_hash) = 32)
) WITH (fillfactor = 90); -- room on each page for HOT updates of tat
COMMENT ON TABLE rate_limit_buckets IS
  'GCRA rate-limit buckets shared by API instances (M1-10, ADR-0005). Keys are HMACs. Purged by the worker (M9).';

ALTER TABLE rate_limit_buckets ENABLE ROW LEVEL SECURITY;
ALTER TABLE rate_limit_buckets FORCE ROW LEVEL SECURITY;
CREATE POLICY rate_limit_buckets_app_select ON rate_limit_buckets FOR SELECT TO cc_app USING (true);
CREATE POLICY rate_limit_buckets_app_insert ON rate_limit_buckets FOR INSERT TO cc_app WITH CHECK (true);
CREATE POLICY rate_limit_buckets_app_update ON rate_limit_buckets FOR UPDATE TO cc_app USING (true) WITH CHECK (true);
GRANT SELECT, INSERT ON rate_limit_buckets TO cc_app;
GRANT UPDATE (tat) ON rate_limit_buckets TO cc_app;

-- rate_limit_take spends one event of (p_scope, p_key)'s bucket at p_now: a burst of
-- p_burst, refilled one event per p_interval. It returns 0 when the event is admitted,
-- otherwise how long until it would be (nothing is spent).
-- +goose StatementBegin
CREATE FUNCTION rate_limit_take(p_scope text, p_key bytea, p_now timestamptz, p_interval interval, p_burst integer)
  RETURNS interval
  LANGUAGE plpgsql
  SECURITY INVOKER
  SET search_path = pg_catalog, pg_temp
AS $$
DECLARE
  v_tolerance interval;
  v_tat       timestamptz;
BEGIN
  IF p_scope IS NULL OR p_scope !~ '^[a-z][a-z0-9_]{0,62}$' THEN
    RAISE EXCEPTION 'scope must be snake_case' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF p_key IS NULL OR octet_length(p_key) <> 32 THEN
    RAISE EXCEPTION 'key must be a 32-byte digest' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF p_now IS NULL OR p_interval IS NULL OR p_interval <= interval '0' OR p_burst IS NULL OR p_burst < 1 THEN
    RAISE EXCEPTION 'interval and burst must be positive' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  v_tolerance := p_interval * p_burst;

  -- A new key starts with a full bucket (tat = now).
  INSERT INTO public.rate_limit_buckets (scope, key_hash, tat) VALUES (p_scope, p_key, p_now)
  ON CONFLICT (scope, key_hash) DO NOTHING;

  UPDATE public.rate_limit_buckets
     SET tat = greatest(tat, p_now) + p_interval
   WHERE scope = p_scope AND key_hash = p_key
     AND greatest(tat, p_now) + p_interval <= p_now + v_tolerance;
  IF FOUND THEN
    RETURN interval '0';
  END IF;

  SELECT tat INTO v_tat FROM public.rate_limit_buckets WHERE scope = p_scope AND key_hash = p_key;
  RETURN greatest(v_tat, p_now) + p_interval - (p_now + v_tolerance);
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION rate_limit_take(text, bytea, timestamptz, interval, integer) IS
  'GCRA: spends one event of a bucket; 0 when admitted, else the wait. One atomic UPDATE decides.';

-- rate_limit_reset fills a bucket again (a successful login clears its failures).
-- +goose StatementBegin
CREATE FUNCTION rate_limit_reset(p_scope text, p_key bytea) RETURNS void
  LANGUAGE sql
  SECURITY INVOKER
  SET search_path = pg_catalog, pg_temp
AS $$
  UPDATE public.rate_limit_buckets SET tat = '-infinity' WHERE scope = p_scope AND key_hash = p_key;
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION rate_limit_reset(text, bytea) IS 'Fills a rate-limit bucket again (e.g. after a successful login).';

REVOKE ALL ON FUNCTION rate_limit_take(text, bytea, timestamptz, interval, integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION rate_limit_reset(text, bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION rate_limit_take(text, bytea, timestamptz, interval, integer) TO cc_app;
GRANT EXECUTE ON FUNCTION rate_limit_reset(text, bytea) TO cc_app;

-- +goose Down
DROP FUNCTION rate_limit_reset(text, bytea);
DROP FUNCTION rate_limit_take(text, bytea, timestamptz, interval, integer);
DROP TABLE rate_limit_buckets;
