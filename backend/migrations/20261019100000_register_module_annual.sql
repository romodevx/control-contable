-- Register module_annual (ANN-22): the annual estimate's flag, OFF, in demo_full only.
-- Sources: docs/project-specs/MODULE-annual.md §6 and OQ-A7 (DECIDED: demo_full only),
-- docs/project-specs/MODULES.md "Presentation presets".
--
-- Prerequisites (module_tax_engine, module_taxpayers) are declared in code, in
-- flags.Prerequisites (internal/platform/flags/prerequisites.go), the single declaration the
-- route gate, GET /v1/flags and cmd/flags read: with either OFF the flag is implied OFF.
-- The baseline seeded demo_full as "every flag at the time"; this keeps that meaning.

-- +goose Up
INSERT INTO feature_flags (key, description) VALUES
  ('module_annual', 'Annual estimate (declaración anual)');

INSERT INTO flag_preset_flags (preset_name, flag_key) VALUES
  ('demo_full', 'module_annual');

UPDATE flag_presets SET description = 'All modules M1-M9 + annual' WHERE name = 'demo_full';

-- +goose Down
UPDATE flag_presets SET description = 'All modules M1-M9' WHERE name = 'demo_full';

-- The preset membership goes with the flag (ON DELETE CASCADE).
DELETE FROM feature_flags WHERE key = 'module_annual';
