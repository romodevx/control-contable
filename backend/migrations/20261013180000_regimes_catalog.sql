-- Regime catalog (M2-02): the SAT c_RegimenFiscal keys for personas físicas and morales
-- and the matrix of which regimes one taxpayer may hold together. Seeded from
-- docs/research/tax-regimes-mx-2026.md §A.1 (PF) and §B.1 (PM); treatment, option_defs
-- and required_inputs follow ADR-0002 §2.
--
-- Catalog key: (code, person_type), never code alone. 610 and 626 exist for both person
-- types and are different regimes (626 PF is RESICO PF, arts. 113-E to 113-J; 626 PM is
-- RESICO PM, arts. 206-215). taxpayer_regimes (M2-03) references (regime_code, person_type),
-- so a persona moral can never hold a física-only regime.
--
-- treatment is the API's Treatment enum (full | partial | informational); the research
-- calls the last one "explanation only". explanation stays NULL until M5-19 writes the
-- plain-language text in a forward migration.
--
-- Where the research is silent or marked ⚠️, the row says so instead of guessing:
-- needs_cpa_review = true with a review_note (PRD Q11), and a compatibility verdict is
-- left NULL (undetermined). The M2-06 use case decides how to treat an undetermined pair.
--
-- Global configuration, not user data (the P0-10 pattern of feature_flags):
--   * no user_id to key a policy on, and FORCE would lock out cc_owner, the only writer;
--   * cc_app gets SELECT only (GET /v1/regimes, regime assignment checks); cc_worker and
--     cc_webhook get nothing;
--   * RLS is ENABLED (not forced) with one SELECT policy for cc_app and no write policy,
--     so a write granted by mistake later is still refused (INSERT) or matches no row;
--   * writes happen only here and in later migrations, as cc_owner.
-- Both tables hold a few dozen rows: the primary keys cover every lookup, and a sequential
-- scan beats any extra index for GET /v1/regimes?person_type=.

-- +goose Up

-- regime_option_defs_valid: the shape of regimes.option_defs, which must match
-- docs/api/openapi.yaml RegimeOptionDef because GET /v1/regimes returns it as is:
--   [{key, label, type, allowed_values?, default?, required}]
--   * key: snake_case, unique in the array; label: non-blank text; required: boolean;
--   * type enum | multi_enum: allowed_values is a non-empty array of distinct strings;
--     an enum default is one of them, a multi_enum default is null (the API types
--     default as string | boolean | null);
--   * type boolean | rate: no allowed_values; a boolean default is a boolean, a rate
--     default is a DecimalRate string (never a JSON number: no floats);
--   * no other property.
-- Each test is its own IF: SQL does not promise left-to-right evaluation of OR, and a
-- jsonb function applied to the wrong JSON type raises instead of returning false.
-- +goose StatementBegin
CREATE FUNCTION regime_option_defs_valid(defs jsonb) RETURNS boolean
  LANGUAGE plpgsql IMMUTABLE PARALLEL SAFE
  SET search_path = pg_catalog, pg_temp
AS $$
DECLARE
  def      jsonb;
  opt_type text;
  allowed  jsonb;
  dflt     jsonb;
  seen     text[] := '{}';
BEGIN
  IF jsonb_typeof(defs) IS DISTINCT FROM 'array' THEN
    RETURN false;
  END IF;
  FOR def IN SELECT value FROM jsonb_array_elements(defs) LOOP
    IF jsonb_typeof(def) <> 'object' THEN
      RETURN false;
    END IF;
    IF NOT def ?& ARRAY['key', 'label', 'type', 'required'] THEN
      RETURN false;
    END IF;
    IF EXISTS (SELECT 1 FROM jsonb_object_keys(def) AS k
               WHERE k NOT IN ('key', 'label', 'type', 'allowed_values', 'default', 'required')) THEN
      RETURN false;
    END IF;
    IF jsonb_typeof(def -> 'key') <> 'string'
       OR jsonb_typeof(def -> 'label') <> 'string'
       OR jsonb_typeof(def -> 'type') <> 'string'
       OR jsonb_typeof(def -> 'required') <> 'boolean' THEN
      RETURN false;
    END IF;
    IF (def ->> 'key') !~ '^[a-z][a-z0-9]*(_[a-z0-9]+)*$'
       OR (def ->> 'key') = ANY (seen)
       OR btrim(def ->> 'label') = '' THEN
      RETURN false;
    END IF;
    seen := seen || (def ->> 'key');

    opt_type := def ->> 'type';
    allowed := def -> 'allowed_values';               -- SQL NULL when absent
    dflt := nullif(def -> 'default', 'null'::jsonb);  -- absent and JSON null alike
    CASE opt_type
      WHEN 'enum', 'multi_enum' THEN
        IF jsonb_typeof(allowed) IS DISTINCT FROM 'array' THEN
          RETURN false;
        END IF;
        IF jsonb_array_length(allowed) = 0
           OR EXISTS (SELECT 1 FROM jsonb_array_elements(allowed) AS v WHERE jsonb_typeof(v) <> 'string')
           OR (SELECT count(DISTINCT v) <> count(*) FROM jsonb_array_elements(allowed) AS v) THEN
          RETURN false;
        END IF;
        IF dflt IS NOT NULL AND (opt_type = 'multi_enum'
                                 OR jsonb_typeof(dflt) <> 'string'
                                 OR NOT allowed @> jsonb_build_array(dflt)) THEN
          RETURN false;
        END IF;
      WHEN 'boolean' THEN
        IF allowed IS NOT NULL OR (dflt IS NOT NULL AND jsonb_typeof(dflt) <> 'boolean') THEN
          RETURN false;
        END IF;
      WHEN 'rate' THEN
        IF allowed IS NOT NULL THEN
          RETURN false;
        END IF;
        -- docs/api/openapi.yaml DecimalRate.
        IF dflt IS NOT NULL AND (jsonb_typeof(dflt) <> 'string'
                                 OR (dflt #>> '{}') !~ '^[0-9]{1,3}(\.[0-9]{1,6})?$') THEN
          RETURN false;
        END IF;
      ELSE
        RETURN false;
    END CASE;
  END LOOP;
  RETURN true;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION regime_option_defs_valid(jsonb) IS
  'CHECK of regimes.option_defs: an array of openapi RegimeOptionDef objects.';

-- regime_required_inputs_valid: the shape of regimes.required_inputs, the tax-input
-- definitions M5-21 validates PUT .../tax-inputs against (docs/api/openapi.yaml
-- TaxInputDef): [{key, label, kind: money | rate, required}], keys unique, nothing else.
-- +goose StatementBegin
CREATE FUNCTION regime_required_inputs_valid(inputs jsonb) RETURNS boolean
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

CREATE TABLE regimes (
  code             text        NOT NULL,
  person_type      text        NOT NULL,
  name             text        NOT NULL,  -- es-MX, as the SAT catalog spells it
  treatment        text        NOT NULL,
  periodicity      text,                  -- NULL only when the research does not state it
  explanation      text,                  -- plain-language text, written by M5-19
  option_defs      jsonb       NOT NULL DEFAULT '[]',
  required_inputs  jsonb       NOT NULL DEFAULT '[]',
  valid_from       date        NOT NULL,
  valid_to         date,                  -- inclusive; NULL = no end date
  legal_basis      text        NOT NULL,
  source_url       text        NOT NULL,
  needs_cpa_review boolean     NOT NULL DEFAULT false,
  review_note      text,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (code, person_type),
  -- docs/api/openapi.yaml RegimeCode, PersonType and Treatment.
  CONSTRAINT regimes_code_format CHECK (code ~ '^[0-9]{3}$'),
  CONSTRAINT regimes_person_type_valid CHECK (person_type IN ('fisica', 'moral')),
  CONSTRAINT regimes_treatment_valid CHECK (treatment IN ('full', 'partial', 'informational')),
  -- How the taxpayer settles the tax (research §A.1/§B.1 "Periodicity").
  CONSTRAINT regimes_periodicity_valid CHECK (periodicity IN (
    'monthly', 'bimonthly', 'annual', 'per_transaction', 'withheld_at_source', 'none')),
  CONSTRAINT regimes_periodicity_stated CHECK (periodicity IS NOT NULL OR needs_cpa_review),
  CONSTRAINT regimes_name_present CHECK (btrim(name) <> ''),
  CONSTRAINT regimes_explanation_present CHECK (explanation IS NULL OR btrim(explanation) <> ''),
  CONSTRAINT regimes_option_defs_shape CHECK (regime_option_defs_valid(option_defs)),
  CONSTRAINT regimes_required_inputs_shape CHECK (regime_required_inputs_valid(required_inputs)),
  CONSTRAINT regimes_validity_range CHECK (valid_to IS NULL OR valid_to >= valid_from),
  CONSTRAINT regimes_legal_basis_present CHECK (btrim(legal_basis) <> ''),
  CONSTRAINT regimes_source_url_format CHECK (source_url ~ '^https?://[^[:space:]]+$'),
  -- A flagged row says why; an unflagged row carries no stale note.
  CONSTRAINT regimes_review_note_iff_flagged CHECK (
    needs_cpa_review = (review_note IS NOT NULL) AND (review_note IS NULL OR btrim(review_note) <> ''))
);
COMMENT ON TABLE regimes IS
  'SAT c_RegimenFiscal catalog keyed by (code, person_type) (M2-02, ADR-0002). Read by cc_app, written only by migrations.';

CREATE TRIGGER regimes_set_updated_at
  BEFORE UPDATE ON regimes
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE regimes ENABLE ROW LEVEL SECURITY;
-- Intent: cc_app reads the whole catalog (global config); no write policy for any runtime role.
CREATE POLICY regimes_app_read ON regimes FOR SELECT TO cc_app USING (true);
GRANT SELECT ON regimes TO cc_app;

-- regime_compatibility: one row per unordered pair of distinct regimes of the same person
-- type (regime_a < regime_b), so the matrix is complete and a lookup never misses.
-- allowed: true = may coexist, false = may not, NULL = undetermined (needs_cpa_review).
CREATE TABLE regime_compatibility (
  person_type      text        NOT NULL,
  regime_a         text        NOT NULL,
  regime_b         text        NOT NULL,
  allowed          boolean,
  reason           text        NOT NULL,  -- es-MX, user-facing when a combination is refused
  legal_basis      text        NOT NULL,
  source_url       text        NOT NULL,
  needs_cpa_review boolean     NOT NULL DEFAULT false,
  review_note      text,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (person_type, regime_a, regime_b),
  -- Both members belong to the same person type: 626 PF never pairs with 601 PM.
  FOREIGN KEY (regime_a, person_type) REFERENCES regimes (code, person_type),
  FOREIGN KEY (regime_b, person_type) REFERENCES regimes (code, person_type),
  CONSTRAINT regime_compatibility_ordered_pair CHECK (regime_a < regime_b),
  CONSTRAINT regime_compatibility_verdict_or_review CHECK (allowed IS NOT NULL OR needs_cpa_review),
  CONSTRAINT regime_compatibility_reason_present CHECK (btrim(reason) <> ''),
  CONSTRAINT regime_compatibility_legal_basis_present CHECK (btrim(legal_basis) <> ''),
  CONSTRAINT regime_compatibility_source_url_format CHECK (source_url ~ '^https?://[^[:space:]]+$'),
  CONSTRAINT regime_compatibility_review_note_iff_flagged CHECK (
    needs_cpa_review = (review_note IS NOT NULL) AND (review_note IS NULL OR btrim(review_note) <> ''))
);
COMMENT ON TABLE regime_compatibility IS
  'Complete matrix of regime pairs per person type (M2-02). allowed NULL = undetermined, pending CPA review.';
-- The PK (person_type, regime_a, ...) serves the FK on regime_a; this one serves regime_b.
CREATE INDEX regime_compatibility_regime_b_idx ON regime_compatibility (person_type, regime_b);

CREATE TRIGGER regime_compatibility_set_updated_at
  BEFORE UPDATE ON regime_compatibility
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE regime_compatibility ENABLE ROW LEVEL SECURITY;
-- Intent: cc_app reads every pair (global config); no write policy for any runtime role.
CREATE POLICY regime_compatibility_app_read ON regime_compatibility FOR SELECT TO cc_app USING (true);
GRANT SELECT ON regime_compatibility TO cc_app;

-- ---------------------------------------------------------------------------------------
-- Seeds: regimes. Every key is in force since 2022-01-01 with no end date (research §A.1,
-- SAT catCFDI_V_4_20261001.xls). Sources are research Part D.
-- ---------------------------------------------------------------------------------------

-- Personas físicas (research §A.1).
INSERT INTO regimes (code, person_type, name, treatment, periodicity, option_defs, required_inputs,
                     valid_from, legal_basis, source_url, needs_cpa_review, review_note) VALUES
  ('605', 'fisica', 'Sueldos y Salarios e Ingresos Asimilados a Salarios', 'informational',
   'withheld_at_source', '[]', '[]', '2022-01-01', 'LISR arts. 94, 96 y 150',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('606', 'fisica', 'Arrendamiento', 'full', 'monthly',
   '[{"key": "deduction_mode", "label": "Tipo de deducción", "type": "enum",
      "allowed_values": ["blind_35", "actual"], "default": "blind_35", "required": true},
     {"key": "iva_border_region", "label": "Aplico el estímulo de IVA al 8 % en región fronteriza",
      "type": "boolean", "default": false, "required": false}]',
   '[]', '2022-01-01', 'LISR arts. 114-118',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('607', 'fisica', 'Régimen de Enajenación o Adquisición de Bienes', 'informational',
   'per_transaction', '[]', '[]', '2022-01-01', 'LISR arts. 120-132',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('608', 'fisica', 'Demás ingresos', 'informational', 'annual', '[]', '[]', '2022-01-01',
   'LISR art. 141 y siguientes', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('610', 'fisica', 'Residentes en el Extranjero sin Establecimiento Permanente en México', 'informational',
   'withheld_at_source', '[]', '[]', '2022-01-01', 'LISR Título V',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('611', 'fisica', 'Ingresos por Dividendos (socios y accionistas)', 'informational', 'annual',
   '[]', '[]', '2022-01-01', 'LISR art. 140',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('612', 'fisica', 'Personas Físicas con Actividades Empresariales y Profesionales', 'full', 'monthly',
   '[{"key": "iva_fully_taxable", "label": "Todas mis actividades están gravadas con IVA",
      "type": "boolean", "default": true, "required": true},
     {"key": "iva_border_region", "label": "Aplico el estímulo de IVA al 8 % en región fronteriza",
      "type": "boolean", "default": false, "required": false}]',
   '[{"key": "prior_provisional_payments", "label": "Pagos provisionales anteriores del ejercicio",
      "kind": "money", "required": true},
     {"key": "ytd_before_app_income", "label": "Ingresos acumulados antes de usar la app",
      "kind": "money", "required": false},
     {"key": "ytd_before_app_deductions", "label": "Deducciones acumuladas antes de usar la app",
      "kind": "money", "required": false},
     {"key": "ptu_paid", "label": "PTU pagada en el ejercicio", "kind": "money", "required": false},
     {"key": "prior_year_losses", "label": "Pérdidas fiscales de ejercicios anteriores",
      "kind": "money", "required": false}]',
   '2022-01-01', 'LISR arts. 100-110', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('614', 'fisica', 'Ingresos por intereses', 'informational', 'withheld_at_source', '[]', '[]',
   '2022-01-01', 'LISR arts. 54 y 135; LIF 2026 art. 24',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('615', 'fisica', 'Régimen de los ingresos por obtención de premios', 'informational',
   'withheld_at_source', '[]', '[]', '2022-01-01', 'LISR arts. 137-139',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('616', 'fisica', 'Sin obligaciones fiscales', 'informational', 'none', '[]', '[]', '2022-01-01',
   'Catálogo c_RegimenFiscal (SAT, CFDI 4.0)',
   'http://omawww.sat.gob.mx/tramitesyservicios/Paginas/documentos/catCFDI_V_4_20261001.xls', false, NULL),
  ('621', 'fisica', 'Incorporación Fiscal', 'partial', 'bimonthly',
   '[{"key": "rif_year", "label": "Año de tributación en el RIF", "type": "enum",
      "allowed_values": ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10"], "default": null, "required": true},
     {"key": "iva_border_region", "label": "Aplico el estímulo de IVA al 8 % en región fronteriza",
      "type": "boolean", "default": false, "required": false}]',
   '[]', '2022-01-01',
   'LISR art. 111 (texto vigente hasta el 31-12-2021); Transitorio Segundo fr. IX, DOF 12-11-2021',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', true,
   'La tabla de reducción por año en el RIF (100 % a 10 %, años 1 a 10) proviene del art. 111 derogado y no se volvió a verificar (investigación §A.2.7).'),
  ('625', 'fisica', 'Régimen de las Actividades Empresariales con ingresos a través de Plataformas Tecnológicas',
   'full', 'monthly',
   '[{"key": "activities", "label": "Actividades a través de plataformas", "type": "multi_enum",
      "allowed_values": ["transport_delivery", "lodging", "goods_services"], "default": null, "required": true},
     {"key": "definitive_option", "label": "Opto por la retención como pago definitivo",
      "type": "boolean", "default": false, "required": false},
     {"key": "rfc_given_to_platform", "label": "Proporcioné mi RFC a la plataforma",
      "type": "boolean", "default": true, "required": false}]',
   '[{"key": "prior_year_income", "label": "Ingresos del ejercicio anterior", "kind": "money", "required": true}]',
   '2022-01-01', 'LISR arts. 113-A a 113-D; LIF 2026 art. 25 fr. VI; LIVA arts. 18-J, 18-L y 18-M',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', true,
   'IVA definitivo por ingresos cobrados directamente (LIVA 18-M) y reglas RMF 2026 de plataformas sin confirmar contra la fuente oficial (ADR-0002, investigación Parte D).'),
  ('626', 'fisica', 'Régimen Simplificado de Confianza', 'full', 'monthly',
   '[{"key": "iva_border_region", "label": "Aplico el estímulo de IVA al 8 % en región fronteriza",
      "type": "boolean", "default": false, "required": false}]',
   '[]', '2022-01-01', 'LISR arts. 113-E a 113-J',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL);

-- Personas morales (research §B.1).
INSERT INTO regimes (code, person_type, name, treatment, periodicity, option_defs, required_inputs,
                     valid_from, legal_basis, source_url, needs_cpa_review, review_note) VALUES
  ('601', 'moral', 'General de Ley Personas Morales', 'full', 'monthly',
   '[{"key": "iva_fully_taxable", "label": "Todas mis actividades están gravadas con IVA",
      "type": "boolean", "default": true, "required": true},
     {"key": "iva_border_region", "label": "Aplico el estímulo de IVA al 8 % en región fronteriza",
      "type": "boolean", "default": false, "required": false}]',
   '[{"key": "coeficiente_utilidad", "label": "Coeficiente de utilidad", "kind": "rate", "required": true},
     {"key": "prior_provisional_payments", "label": "Pagos provisionales anteriores del ejercicio",
      "kind": "money", "required": true},
     {"key": "ptu_paid", "label": "PTU pagada en el ejercicio", "kind": "money", "required": false},
     {"key": "prior_year_losses", "label": "Pérdidas fiscales de ejercicios anteriores",
      "kind": "money", "required": false},
     {"key": "interest_withholding", "label": "Retención de ISR por intereses", "kind": "money", "required": false}]',
   '2022-01-01', 'LISR arts. 9, 14 y 76 fr. V', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('603', 'moral', 'Personas Morales con Fines no Lucrativos', 'partial', 'annual',
   '[{"key": "iva_fully_taxable", "label": "Todas mis actividades están gravadas con IVA",
      "type": "boolean", "default": true, "required": true},
     {"key": "iva_border_region", "label": "Aplico el estímulo de IVA al 8 % en región fronteriza",
      "type": "boolean", "default": false, "required": false}]',
   '[]', '2022-01-01', 'LISR Título III', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('610', 'moral', 'Residentes en el Extranjero sin Establecimiento Permanente en México', 'informational',
   'withheld_at_source', '[]', '[]', '2022-01-01', 'LISR Título V',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('620', 'moral', 'Sociedades Cooperativas de Producción que optan por diferir sus ingresos', 'informational',
   NULL, '[]', '[]', '2022-01-01', 'LISR arts. 194-199',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', true,
   'La investigación sólo dice "diferimiento especial"; no establece la periodicidad de los pagos.'),
  ('622', 'moral', 'Actividades Agrícolas, Ganaderas, Silvícolas y Pesqueras', 'partial', 'monthly',
   '[{"key": "iva_fully_taxable", "label": "Todas mis actividades están gravadas con IVA",
      "type": "boolean", "default": true, "required": true},
     {"key": "iva_border_region", "label": "Aplico el estímulo de IVA al 8 % en región fronteriza",
      "type": "boolean", "default": false, "required": false}]',
   '[{"key": "coeficiente_utilidad", "label": "Coeficiente de utilidad", "kind": "rate", "required": true},
     {"key": "prior_provisional_payments", "label": "Pagos provisionales anteriores del ejercicio",
      "kind": "money", "required": true},
     {"key": "ptu_paid", "label": "PTU pagada en el ejercicio", "kind": "money", "required": false},
     {"key": "prior_year_losses", "label": "Pérdidas fiscales de ejercicios anteriores",
      "kind": "money", "required": false},
     {"key": "interest_withholding", "label": "Retención de ISR por intereses", "kind": "money", "required": false}]',
   '2022-01-01', 'LISR art. 74', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL),
  ('623', 'moral', 'Opcional para Grupos de Sociedades', 'informational', NULL, '[]', '[]', '2022-01-01',
   'LISR arts. 59-71', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', true,
   'La investigación sólo dice "integración de grupo"; no establece la periodicidad de los pagos.'),
  ('624', 'moral', 'Coordinados', 'informational', NULL, '[]', '[]', '2022-01-01', 'LISR arts. 72-73',
   'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', true,
   'La investigación sólo dice que pagan por cuenta de sus integrantes; no establece la periodicidad.'),
  ('626', 'moral', 'Régimen Simplificado de Confianza', 'full', 'monthly',
   '[{"key": "iva_fully_taxable", "label": "Todas mis actividades están gravadas con IVA",
      "type": "boolean", "default": true, "required": true},
     {"key": "iva_border_region", "label": "Aplico el estímulo de IVA al 8 % en región fronteriza",
      "type": "boolean", "default": false, "required": false}]',
   '[{"key": "prior_provisional_payments", "label": "Pagos provisionales anteriores del ejercicio",
      "kind": "money", "required": true},
     {"key": "ytd_before_app_income", "label": "Ingresos acumulados antes de usar la app",
      "kind": "money", "required": false},
     {"key": "ytd_before_app_deductions", "label": "Deducciones acumuladas antes de usar la app",
      "kind": "money", "required": false},
     {"key": "ptu_paid", "label": "PTU pagada en el ejercicio", "kind": "money", "required": false},
     {"key": "prior_year_losses", "label": "Pérdidas fiscales de ejercicios anteriores",
      "kind": "money", "required": false}]',
   '2022-01-01', 'LISR arts. 206-215', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', false, NULL);

-- ---------------------------------------------------------------------------------------
-- Seeds: regime_compatibility. Explicit rows first (each traced to the research); then
-- set-based rows for the ordinary Título IV pairs, the definitional exclusions (616, 610)
-- and the pairs the research does not settle. Two sets overlap rows written before them
-- (the Título IV set covers some explicit pairs; both exclusion sets produce 610/616):
-- ON CONFLICT DO NOTHING keeps the earlier, more specific row.
-- ---------------------------------------------------------------------------------------

-- Personas físicas: rules stated by the research.
INSERT INTO regime_compatibility (person_type, regime_a, regime_b, allowed, reason, legal_basis, source_url) VALUES
  ('fisica', '612', '626', false,
   'El RESICO sustituye al régimen de actividades empresariales y profesionales para esos ingresos; al salir del RESICO se tributa en 612 o 606.',
   'LISR art. 113-E', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('fisica', '606', '626', false,
   'Los ingresos por arrendamiento tributan dentro del RESICO; al salir del RESICO se tributa en 612 o 606.',
   'LISR art. 113-E', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('fisica', '611', '626', false,
   'Los socios o accionistas de personas morales no pueden tributar en el RESICO.',
   'LISR art. 113-E', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('fisica', '605', '626', true,
   'El RESICO admite ingresos por sueldos si el total del ejercicio anterior no excede $3,500,000 (los asimilados del art. 94 fr. III a VI no pueden tributar en el RESICO).',
   'LISR art. 113-E', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('fisica', '614', '626', true,
   'El RESICO admite ingresos por intereses si el total del ejercicio anterior no excede $3,500,000.',
   'LISR art. 113-E', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('fisica', '612', '625', true,
   'Sin la opción de pago definitivo, la retención de la plataforma es pago provisional que se acredita en el régimen 612.',
   'LISR arts. 113-A y 113-B', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('fisica', '625', '626', true,
   'Sin la opción de pago definitivo, la retención de la plataforma es pago provisional que se acredita en el RESICO.',
   'LISR arts. 113-A, 113-B y 113-E', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('fisica', '605', '625', true,
   'Los ingresos por sueldos no impiden la opción de pago definitivo de plataformas.',
   'LISR art. 113-B', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('fisica', '614', '625', true,
   'Los ingresos por intereses no impiden la opción de pago definitivo de plataformas.',
   'LISR art. 113-B', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('fisica', '605', '612', true,
   'Sueldos y actividades empresariales o profesionales coexisten; se acumulan en la declaración anual.',
   'LISR arts. 100 y 150', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('fisica', '605', '606', true,
   'Sueldos y arrendamiento coexisten; se acumulan en la declaración anual.',
   'LISR arts. 114 y 150', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('fisica', '606', '612', true,
   'Arrendamiento y actividades empresariales o profesionales coexisten como capítulos distintos del Título IV.',
   'LISR arts. 100 y 114', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf');

-- Personas físicas: the remaining pairs among the ordinary Título IV regimes coexist.
-- Each is a different chapter of Título IV, and their income is combined in the annual return.
INSERT INTO regime_compatibility (person_type, regime_a, regime_b, allowed, reason, legal_basis, source_url)
SELECT 'fisica', a.code, b.code, true,
       'Ingresos de capítulos distintos del Título IV de la LISR; coexisten y se acumulan en la declaración anual.',
       'LISR Título IV; art. 150', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'
FROM unnest(ARRAY['605', '606', '607', '608', '611', '612', '614', '615', '625']) AS a(code)
JOIN unnest(ARRAY['605', '606', '607', '608', '611', '612', '614', '615', '625']) AS b(code) ON a.code < b.code
ON CONFLICT DO NOTHING;

-- 616 means no tax obligations at all: it excludes every other regime.
INSERT INTO regime_compatibility (person_type, regime_a, regime_b, allowed, reason, legal_basis, source_url)
SELECT 'fisica', least(r.code, '616'), greatest(r.code, '616'), false,
       'Sin obligaciones fiscales no puede combinarse con un régimen que sí las tiene.',
       'Catálogo c_RegimenFiscal (SAT, CFDI 4.0)',
       'http://omawww.sat.gob.mx/tramitesyservicios/Paginas/documentos/catCFDI_V_4_20261001.xls'
FROM regimes r
WHERE r.person_type = 'fisica' AND r.code <> '616';

-- 610 (both person types) is for non-residents; every other regime is for residents.
INSERT INTO regime_compatibility (person_type, regime_a, regime_b, allowed, reason, legal_basis, source_url)
SELECT r.person_type, least(r.code, '610'), greatest(r.code, '610'), false,
       'Un residente en el extranjero sin establecimiento permanente no tributa en los regímenes de residentes.',
       'LISR Título V', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'
FROM regimes r
WHERE r.code <> '610'
ON CONFLICT DO NOTHING;

-- Personas físicas: pairs the research does not settle (undetermined, CPA review).
INSERT INTO regime_compatibility (person_type, regime_a, regime_b, allowed, reason, legal_basis, source_url,
                                  needs_cpa_review, review_note)
SELECT 'fisica', least(c.code, '621'), greatest(c.code, '621'), NULL,
       'Combinación por confirmar con un contador.',
       'LISR art. 111 (texto vigente hasta el 31-12-2021)', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf',
       true, 'La investigación no cubre qué otros ingresos podía tener un contribuyente del RIF (art. 111 derogado).'
FROM unnest(ARRAY['605', '606', '607', '608', '611', '612', '614', '615', '625', '626']) AS c(code);

INSERT INTO regime_compatibility (person_type, regime_a, regime_b, allowed, reason, legal_basis, source_url,
                                  needs_cpa_review, review_note)
SELECT 'fisica', c.code, '626', NULL,
       'Combinación por confirmar con un contador.',
       'LISR art. 113-E', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf',
       true, 'La investigación sólo admite sueldos e intereses junto al RESICO; no dice nada de este capítulo.'
FROM unnest(ARRAY['607', '608', '615']) AS c(code);

-- Personas morales: RESICO PM exclusions (research §B.2.2).
INSERT INTO regime_compatibility (person_type, regime_a, regime_b, allowed, reason, legal_basis, source_url) VALUES
  ('moral', '601', '626', false,
   'Una persona moral del RESICO que deja de cumplir los requisitos pasa al régimen general; no tributa en ambos.',
   'LISR arts. 206 y 214', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('moral', '603', '626', false,
   'Las personas morales del Título III (fines no lucrativos) no pueden tributar en el RESICO.',
   'LISR art. 206', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('moral', '622', '626', false,
   'Las personas morales del Capítulo VIII del Título II (AGAPES) no pueden tributar en el RESICO.',
   'LISR art. 206', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('moral', '623', '626', false,
   'Las personas morales del Capítulo VI del Título II (grupos de sociedades) no pueden tributar en el RESICO.',
   'LISR art. 206', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
  ('moral', '624', '626', false,
   'Las personas morales del Capítulo VII del Título II (coordinados) no pueden tributar en el RESICO.',
   'LISR art. 206', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf');

-- Personas morales: pairs the research does not settle (undetermined, CPA review).
INSERT INTO regime_compatibility (person_type, regime_a, regime_b, allowed, reason, legal_basis, source_url,
                                  needs_cpa_review, review_note)
SELECT 'moral', a.code, b.code, NULL,
       'Combinación por confirmar con un contador.',
       'LISR Títulos II, III y VII', 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf',
       true, 'La investigación no establece si estos regímenes de personas morales son excluyentes entre sí.'
FROM unnest(ARRAY['601', '603', '620', '622', '623', '624']) AS a(code)
JOIN unnest(ARRAY['601', '603', '620', '622', '623', '624', '626']) AS b(code) ON a.code < b.code
WHERE NOT (b.code = '626' AND a.code <> '620');

-- +goose Down
DROP TABLE regime_compatibility;
DROP TABLE regimes;
DROP FUNCTION regime_required_inputs_valid(jsonb);
DROP FUNCTION regime_option_defs_valid(jsonb);
