-- Annual depreciation tax input (ANN-03): the catalog forward migration of
-- docs/project-specs/MODULE-annual.md §2.6 ("New key annual_depreciation (money,
-- optional) for 612 PF, 601 PM, 626 PM: the year's depreciation from the CPA's
-- fixed-asset schedule, typed only, no asset register").
--
--   * appended LAST to regimes.required_inputs of 612 fisica, 601 moral and 626 moral, so
--     the existing form order (TaxInputs.values) does not move;
--   * kind money, required false: the monthly estimate never asks for it, the annual
--     estimate (§2.7) subtracts it when present;
--   * guarded by NOT @> on the single key, so a re-run adds it once.
-- The applied catalog migrations are never edited. taxpayer_tax_inputs derives kind from
-- this definition on every write (taxpayer_tax_inputs_derive_kind), so no stored kind
-- needs re-deriving here.
--
-- RLS: regimes is global config (ENABLE, SELECT-only policy for cc_app, no write grant);
-- this migration runs as the owner and changes no policy or grant. Down removes the key
-- and the values stored under it: taxpayer_tax_inputs has FORCE ROW LEVEL SECURITY and no
-- owner policy, so the owner lifts FORCE for that one DELETE and restores it in the same
-- transaction (otherwise the DELETE silently matches no row).

-- +goose Up

UPDATE regimes r SET required_inputs = r.required_inputs || jsonb_build_array(jsonb_build_object(
  'key', 'annual_depreciation', 'label', 'Depreciación fiscal del ejercicio', 'kind', 'money', 'required', false))
WHERE (r.code, r.person_type) IN (('612', 'fisica'), ('601', 'moral'), ('626', 'moral'))
  AND NOT r.required_inputs @> '[{"key": "annual_depreciation"}]';

-- +goose Down

ALTER TABLE taxpayer_tax_inputs NO FORCE ROW LEVEL SECURITY;
DELETE FROM taxpayer_tax_inputs WHERE input_key = 'annual_depreciation';
ALTER TABLE taxpayer_tax_inputs FORCE ROW LEVEL SECURITY;

UPDATE regimes r SET required_inputs = (
  SELECT coalesce(jsonb_agg(e.def ORDER BY e.n), '[]'::jsonb)
  FROM jsonb_array_elements(r.required_inputs) WITH ORDINALITY AS e(def, n)
  WHERE e.def ->> 'key' <> 'annual_depreciation')
WHERE r.required_inputs @> '[{"key": "annual_depreciation"}]';
