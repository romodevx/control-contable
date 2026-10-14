-- Tax input kinds (M5-20): the catalog forward migration of docs/project-specs/
-- MODULE-tax-engine.md §9 item 6 and OQ-T1, matching the frozen contract (0.6.0,
-- docs/api/openapi.yaml TaxInputDef.kind: money | money_monthly | rate | boolean).
--
--   * regime_required_inputs_valid() admits the two new kinds;
--   * prior_provisional_payments becomes money_monthly (one amount per month, OQ-T1) in
--     every regime that has it: 612 PF, 626 PM, 601 PM and 622 PM. 622 is not estimated
--     in v1 but keeps the 601 definitions (§2.8), so one key never has two kinds;
--   * 601 PM gains first_fiscal_year (boolean, optional), second after
--     coeficiente_utilidad as in the contract example regime601. Only 601: §2.8 and §9
--     name no other regime.
-- The applied catalog migration (20261013180000_regimes_catalog.sql) is never edited.
-- taxpayer_tax_inputs (next migration) does not exist yet, so no stored kind needs
-- re-deriving here; a later kind change must re-derive taxpayer_tax_inputs.kind.
--
-- RLS: regimes is global config (ENABLE, SELECT-only policy for cc_app, no write grant);
-- this migration runs as the owner and changes no policy or grant.

-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION regime_required_inputs_valid(inputs jsonb) RETURNS boolean
  LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE
  SET search_path = pg_catalog, pg_temp
AS $$
DECLARE
  def  jsonb;
  seen text[] := '{}';
BEGIN
  IF jsonb_typeof(inputs) IS DISTINCT FROM 'array' THEN
    RETURN false;
  END IF;
  FOR def IN SELECT value FROM jsonb_array_elements(inputs) LOOP
    IF jsonb_typeof(def) <> 'object' THEN
      RETURN false;
    END IF;
    IF (SELECT array_agg(k ORDER BY k) FROM jsonb_object_keys(def) AS k)
       IS DISTINCT FROM ARRAY['key', 'kind', 'label', 'required'] THEN
      RETURN false;
    END IF;
    IF jsonb_typeof(def -> 'key') <> 'string'
       OR jsonb_typeof(def -> 'label') <> 'string'
       OR jsonb_typeof(def -> 'required') <> 'boolean' THEN
      RETURN false;
    END IF;
    IF (def ->> 'key') !~ '^[a-z][a-z0-9]*(_[a-z0-9]+)*$'
       OR length(def ->> 'key') > 64
       OR (def ->> 'key') = ANY (seen)
       OR btrim(def ->> 'label') = ''
       OR (def -> 'kind') NOT IN ('"money"'::jsonb, '"money_monthly"'::jsonb, '"rate"'::jsonb, '"boolean"'::jsonb) THEN
      RETURN false;
    END IF;
    seen := seen || (def ->> 'key');
  END LOOP;
  RETURN true;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION regime_required_inputs_valid(jsonb) IS
  'CHECK of regimes.required_inputs: an array of openapi TaxInputDef objects (kind money | money_monthly | rate | boolean).';

-- Order of the definitions is kept (WITH ORDINALITY): it is the order of the form and of
-- TaxInputs.values.
UPDATE regimes r SET required_inputs = (
  SELECT jsonb_agg(CASE WHEN e.def ->> 'key' = 'prior_provisional_payments'
                        THEN jsonb_set(e.def, '{kind}', '"money_monthly"') ELSE e.def END
                   ORDER BY e.n)
  FROM jsonb_array_elements(r.required_inputs) WITH ORDINALITY AS e(def, n))
WHERE r.required_inputs @> '[{"key": "prior_provisional_payments", "kind": "money"}]';

UPDATE regimes r SET required_inputs = (
  SELECT jsonb_agg(d.def ORDER BY d.n)
  FROM (SELECT e.def, e.n::numeric AS n
        FROM jsonb_array_elements(r.required_inputs) WITH ORDINALITY AS e(def, n)
        UNION ALL
        SELECT '{"key": "first_fiscal_year", "label": "Primer ejercicio fiscal", "kind": "boolean", "required": false}'::jsonb,
               1.5) AS d)
WHERE r.code = '601' AND r.person_type = 'moral'
  AND NOT r.required_inputs @> '[{"key": "first_fiscal_year"}]';

-- +goose Down

UPDATE regimes r SET required_inputs = (
  SELECT coalesce(jsonb_agg(e.def ORDER BY e.n), '[]'::jsonb)
  FROM jsonb_array_elements(r.required_inputs) WITH ORDINALITY AS e(def, n)
  WHERE e.def ->> 'key' <> 'first_fiscal_year')
WHERE r.code = '601' AND r.person_type = 'moral';

UPDATE regimes r SET required_inputs = (
  SELECT jsonb_agg(CASE WHEN e.def ->> 'key' = 'prior_provisional_payments'
                        THEN jsonb_set(e.def, '{kind}', '"money"') ELSE e.def END
                   ORDER BY e.n)
  FROM jsonb_array_elements(r.required_inputs) WITH ORDINALITY AS e(def, n))
WHERE r.required_inputs @> '[{"key": "prior_provisional_payments", "kind": "money_monthly"}]';

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION regime_required_inputs_valid(inputs jsonb) RETURNS boolean
  LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE
  SET search_path = pg_catalog, pg_temp
AS $$
DECLARE
  def  jsonb;
  seen text[] := '{}';
BEGIN
  IF jsonb_typeof(inputs) IS DISTINCT FROM 'array' THEN
    RETURN false;
  END IF;
  FOR def IN SELECT value FROM jsonb_array_elements(inputs) LOOP
    IF jsonb_typeof(def) <> 'object' THEN
      RETURN false;
    END IF;
    IF (SELECT array_agg(k ORDER BY k) FROM jsonb_object_keys(def) AS k)
       IS DISTINCT FROM ARRAY['key', 'kind', 'label', 'required'] THEN
      RETURN false;
    END IF;
    IF jsonb_typeof(def -> 'key') <> 'string'
       OR jsonb_typeof(def -> 'label') <> 'string'
       OR jsonb_typeof(def -> 'required') <> 'boolean' THEN
      RETURN false;
    END IF;
    IF (def ->> 'key') !~ '^[a-z][a-z0-9]*(_[a-z0-9]+)*$'
       OR length(def ->> 'key') > 64
       OR (def ->> 'key') = ANY (seen)
       OR btrim(def ->> 'label') = ''
       OR (def -> 'kind') NOT IN ('"money"'::jsonb, '"rate"'::jsonb) THEN
      RETURN false;
    END IF;
    seen := seen || (def ->> 'key');
  END LOOP;
  RETURN true;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION regime_required_inputs_valid(jsonb) IS
  'CHECK of regimes.required_inputs: an array of openapi TaxInputDef objects (kind money | rate).';
