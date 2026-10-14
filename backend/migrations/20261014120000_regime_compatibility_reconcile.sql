-- Regime compatibility reconcile (M2-13 FU1): MODULE-taxpayers §2.5 refused pairs that the
-- M2-02 seed (20261013180000_regimes_catalog.sql) allowed with a CPA-review warning.
-- Authority order: the law > the spec > the seed. Each pair was decided against the LISR
-- text (docs/research/tax-regimes-mx-2026.md Part E):
--   * the law clearly forbids holding both regimes at once -> refused here (allowed = false,
--     needs_cpa_review = false, review_note NULL, reason and legal basis cited);
--   * the law does not settle it -> the seed row stays as it is (allowed with a warning) and
--     the spec was amended; those pairs are listed in docs/reviews/CPA-REVIEW.md.
-- The applied catalog migration is never edited; this forward migration updates its rows.
--
-- Existing taxpayer_regimes rows are not re-checked: the verdict applies to new writes
-- (create, add regime), as every catalog verdict does. module_taxpayers is OFF in every
-- shared environment until FU1 is settled, so no shared row holds a newly refused pair.
--
-- RLS: regime_compatibility is global config (ENABLE, SELECT-only policy for cc_app, no
-- write grant); this migration runs as the owner and changes no policy or grant.

-- +goose Up

-- Personas físicas.
UPDATE regime_compatibility SET
  allowed = false, needs_cpa_review = false, review_note = NULL,
  reason = 'El RIF era sólo para actividades empresariales sin título profesional; al dejarlo se tributa en 612. No se tienen ambos a la vez.',
  legal_basis = 'LISR art. 111, párrafos primero y décimo quinto (texto vigente hasta el 31-12-2021); Disposiciones Transitorias LISR 2022, art. Segundo fr. IX',
  source_url = 'https://www.diputados.gob.mx/LeyesBiblio/ref/lisr/LISR_orig_11dic13.pdf'
WHERE person_type = 'fisica' AND regime_a = '612' AND regime_b = '621';

UPDATE regime_compatibility SET
  allowed = false, needs_cpa_review = false, review_note = NULL,
  reason = 'Un contribuyente del RIF podía seguir en el RIF o pasar al RESICO, no ambos: el RESICO sólo admite además sueldos e intereses.',
  legal_basis = 'LISR art. 113-E, párrafos primero y sexto; Disposiciones Transitorias LISR 2022 (DOF 12-11-2021), art. Segundo fr. IX',
  source_url = 'https://www.diputados.gob.mx/LeyesBiblio/ref/lisr/LISR_ref07_12nov21.pdf'
WHERE person_type = 'fisica' AND regime_a = '621' AND regime_b = '626';

-- Personas morales: a cooperative that defers (620, Título VII Cap. VII) applies it
-- instead of Título II (601, 622, 623, 624).
UPDATE regime_compatibility SET
  allowed = false, needs_cpa_review = false, review_note = NULL,
  reason = 'Las cooperativas de producción que difieren su impuesto aplican su capítulo en lugar del Título II; no tributan en ambos.',
  legal_basis = 'LISR art. 194, primer párrafo',
  source_url = 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'
WHERE person_type = 'moral' AND (regime_a, regime_b) IN (('601', '620'), ('620', '622'), ('620', '623'), ('620', '624'));

UPDATE regime_compatibility SET
  allowed = false, needs_cpa_review = false, review_note = NULL,
  reason = 'Las cooperativas de producción que difieren su impuesto no pueden tributar en el RESICO.',
  legal_basis = 'LISR art. 206 fr. IV',
  source_url = 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'
WHERE person_type = 'moral' AND regime_a = '620' AND regime_b = '626';

UPDATE regime_compatibility SET
  allowed = false, needs_cpa_review = false, review_note = NULL,
  reason = 'Una persona moral dedicada exclusivamente a actividades agrícolas, ganaderas, silvícolas o pesqueras debe tributar en ese capítulo, no en el régimen general.',
  legal_basis = 'LISR art. 74, primer párrafo',
  source_url = 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'
WHERE person_type = 'moral' AND regime_a = '601' AND regime_b = '622';

UPDATE regime_compatibility SET
  allowed = false, needs_cpa_review = false, review_note = NULL,
  reason = 'El régimen de actividades agrícolas, ganaderas, silvícolas y pesqueras no aplica a las sociedades del régimen opcional para grupos.',
  legal_basis = 'LISR art. 74, segundo párrafo',
  source_url = 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'
WHERE person_type = 'moral' AND regime_a = '622' AND regime_b = '623';

UPDATE regime_compatibility SET
  allowed = false, needs_cpa_review = false, review_note = NULL,
  reason = 'Cada régimen exige que al menos el 90 % de los ingresos venga de su actividad (agropecuaria o autotransporte); no se cumplen ambos a la vez.',
  legal_basis = 'LISR arts. 72 y 74',
  source_url = 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'
WHERE person_type = 'moral' AND regime_a = '622' AND regime_b = '624';

-- +goose Down

-- Back to the M2-02 seed rows: allowed with a CPA-review warning.
UPDATE regime_compatibility SET
  allowed = true, needs_cpa_review = true,
  reason = 'Combinación permitida, pendiente de confirmar con un contador.',
  legal_basis = 'LISR art. 111 (texto vigente hasta el 31-12-2021)',
  source_url = 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf',
  review_note = 'La investigación no cubre qué otros ingresos podía tener un contribuyente del RIF (art. 111 derogado).'
WHERE person_type = 'fisica' AND (regime_a, regime_b) IN (('612', '621'), ('621', '626'));

UPDATE regime_compatibility SET
  allowed = true, needs_cpa_review = true,
  reason = 'Combinación permitida, pendiente de confirmar con un contador.',
  legal_basis = 'LISR Títulos II, III y VII',
  source_url = 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf',
  review_note = 'La investigación no establece si estos regímenes de personas morales son excluyentes entre sí.'
WHERE person_type = 'moral' AND (regime_a, regime_b) IN (
  ('601', '620'), ('620', '622'), ('620', '623'), ('620', '624'), ('620', '626'),
  ('601', '622'), ('622', '623'), ('622', '624'));
