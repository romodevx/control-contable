-- Regime overlap trigger: narrowing a window is always allowed (M2-13 FU1 follow-up).
-- 20261014120000_regime_compatibility_reconcile.sql refuses pairs the seed used to allow
-- and leaves existing taxpayer_regimes rows as they are. On a database that holds such a
-- pair, ending one of the two regimes (archive: UPDATE end_period) was refused with 23P01,
-- so the user could not leave the state the new verdict forbids. An UPDATE that keeps the
-- same row and only shrinks its window ([new start, new end] inside [old start, old end])
-- cannot create an overlap that was not already there, so it skips the clash lookup.
-- Every other write (INSERT, reopen, moving a start earlier) is checked as before.
-- The function keeps SECURITY INVOKER and its pinned search_path.

-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION taxpayer_regimes_check_compatible() RETURNS trigger
  LANGUAGE plpgsql
  SET search_path = pg_catalog, pg_temp
AS $$
DECLARE
  clash_code   text;
  clash_reason text;
BEGIN
  IF current_setting('transaction_isolation') = 'repeatable read' THEN
    RAISE EXCEPTION 'taxpayer_regimes writes cannot run under REPEATABLE READ'
      USING ERRCODE = '25000',
            HINT = 'Its snapshot would miss an incompatible regime committed while waiting for the lock; use READ COMMITTED.';
  END IF;
  IF TG_OP = 'UPDATE'
     AND NEW.taxpayer_id = OLD.taxpayer_id
     AND NEW.regime_code = OLD.regime_code
     AND NEW.start_period >= OLD.start_period
     AND coalesce(NEW.end_period, '9999-12') <= coalesce(OLD.end_period, '9999-12') THEN
    RETURN NEW;
  END IF;
  PERFORM 1 FROM public.taxpayers t
  WHERE t.id = NEW.taxpayer_id AND t.user_id = NEW.user_id
  FOR NO KEY UPDATE;

  SELECT o.regime_code, c.reason INTO clash_code, clash_reason
  FROM public.taxpayer_regimes o
  JOIN public.regime_compatibility c
    ON c.person_type = NEW.person_type
   AND c.regime_a = least(o.regime_code, NEW.regime_code)
   AND c.regime_b = greatest(o.regime_code, NEW.regime_code)
  WHERE o.taxpayer_id = NEW.taxpayer_id
    AND o.user_id = NEW.user_id
    AND o.regime_code <> NEW.regime_code
    AND NOT c.allowed
    AND o.start_period <= coalesce(NEW.end_period, '9999-12')
    AND NEW.start_period <= coalesce(o.end_period, '9999-12')
  ORDER BY o.regime_code
  LIMIT 1;

  IF clash_code IS NOT NULL THEN
    RAISE EXCEPTION 'regimes % and % may not overlap', least(clash_code, NEW.regime_code),
                    greatest(clash_code, NEW.regime_code)
      USING ERRCODE = '23P01', CONSTRAINT = 'taxpayer_regimes_compatible_periods',
            TABLE = 'taxpayer_regimes', DETAIL = clash_reason;
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION taxpayer_regimes_check_compatible() RETURNS trigger
  LANGUAGE plpgsql
  SET search_path = pg_catalog, pg_temp
AS $$
DECLARE
  clash_code   text;
  clash_reason text;
BEGIN
  IF current_setting('transaction_isolation') = 'repeatable read' THEN
    RAISE EXCEPTION 'taxpayer_regimes writes cannot run under REPEATABLE READ'
      USING ERRCODE = '25000',
            HINT = 'Its snapshot would miss an incompatible regime committed while waiting for the lock; use READ COMMITTED.';
  END IF;
  PERFORM 1 FROM public.taxpayers t
  WHERE t.id = NEW.taxpayer_id AND t.user_id = NEW.user_id
  FOR NO KEY UPDATE;

  SELECT o.regime_code, c.reason INTO clash_code, clash_reason
  FROM public.taxpayer_regimes o
  JOIN public.regime_compatibility c
    ON c.person_type = NEW.person_type
   AND c.regime_a = least(o.regime_code, NEW.regime_code)
   AND c.regime_b = greatest(o.regime_code, NEW.regime_code)
  WHERE o.taxpayer_id = NEW.taxpayer_id
    AND o.user_id = NEW.user_id
    AND o.regime_code <> NEW.regime_code
    AND NOT c.allowed
    AND o.start_period <= coalesce(NEW.end_period, '9999-12')
    AND NEW.start_period <= coalesce(o.end_period, '9999-12')
  ORDER BY o.regime_code
  LIMIT 1;

  IF clash_code IS NOT NULL THEN
    RAISE EXCEPTION 'regimes % and % may not overlap', least(clash_code, NEW.regime_code),
                    greatest(clash_code, NEW.regime_code)
      USING ERRCODE = '23P01', CONSTRAINT = 'taxpayer_regimes_compatible_periods',
            TABLE = 'taxpayer_regimes', DETAIL = clash_reason;
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd
