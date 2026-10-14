-- 606 PF quarterly option (M5-12): LISR art. 116 ¶3 lets a lessor whose only income is
-- rent of up to 10 monthly minimum wages (read as UMA, research §A.2.4 step 5 ⚠️) pay
-- quarterly; LIVA 5-F makes the IVA follow. The quarterly table is not transcribed, so
-- the 606 strategy answers a quarterly payer with the informational estimate
-- (ADR-0002 §2). The threshold reading is already in docs/reviews/CPA-REVIEW.md.
--
-- The option is data (contract RegimeOptionDef, shape unchanged): an optional boolean
-- defaulting to false, appended to 606's option_defs. Stored 606 regimes stay valid:
-- an omitted optional option is simply absent and the engine reads it as monthly.
-- The applied catalog migration (20261013180000_regimes_catalog.sql) is never edited.
--
-- RLS: regimes is global config (ENABLE, SELECT-only policy for cc_app, no write grant);
-- this migration runs as the owner and changes no policy or grant.

-- +goose Up
UPDATE regimes
SET option_defs = option_defs || '[{"key": "pays_quarterly",
  "label": "Hago pagos provisionales trimestrales (art. 116 LISR)",
  "type": "boolean", "default": false, "required": false}]'::jsonb
WHERE code = '606' AND person_type = 'fisica'
  AND NOT option_defs @> '[{"key": "pays_quarterly"}]'::jsonb;

-- +goose Down
UPDATE regimes
SET option_defs = (
  SELECT coalesce(jsonb_agg(d ORDER BY n), '[]'::jsonb)
  FROM jsonb_array_elements(option_defs) WITH ORDINALITY AS e(d, n)
  WHERE d ->> 'key' <> 'pays_quarterly')
WHERE code = '606' AND person_type = 'fisica';
