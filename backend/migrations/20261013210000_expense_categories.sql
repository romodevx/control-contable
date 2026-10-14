-- Expense category catalog (M4-02): the categories a user picks for an expense and, per
-- (category, regime, person type), the hint the app shows and the default it records.
-- Sources: docs/project-specs/MODULE-expenses.md §2.2 and §2.5 (DECIDED OQ-E3, OQ-E4),
-- docs/research/tax-regimes-mx-2026.md §A.2.3-§A.2.5, §A.3, §B.2, Part C,
-- docs/architecture/data-model.md (expense_categories, category_regime_rules).
--
-- ⚠️ Every row waits for the CPA (OQ-E3): needs_cpa_review = true with a review_note, and
-- the open items are listed in docs/reviews/CPA-REVIEW.md. The legal basis is article
-- level on purpose: the research does not verify fractions.
--
-- Rule columns:
--   * deductibility: the hint shown to the user (deductible | partially_deductible |
--     not_deductible | isr_not_applicable);
--   * default_deductible: the boolean an expense starts with. It differs from the hint on
--     purpose for partial cases (conservative estimate: investments, restaurants, personal
--     and "otros" start non-deductible); it is never true for not_deductible or
--     isr_not_applicable (CHECK);
--   * iva_creditable_default: whether the IVA paid is normally creditable (LIVA art. 5
--     fr. I: the expense is ISR-deductible, or would be, for RESICO PF). A hint for M5,
--     which owns the creditable-IVA rule; false for informational regimes (no IVA return);
--   * cash_payment_rule: how a cash payment affects the default (LISR art. 27 fr. III):
--     over_limit = cash above the limit is not deductible; any_cash = any cash payment is
--     not deductible (combustibles); not_applicable = the row is never deductible anyway.
--     The limit itself ($2,000.00) is dated parameter data served by RateCatalog
--     (deduction.cash_payment_limit, ADR-0002 §3), never stored or hard-coded here.
-- Regime options (606 blind_35, 625 definitive_option) change the effective hint in the
-- domain (DeductibilityResolver, §2.3), not in this table.
--
-- Global configuration, not user data (the P0-10 / M2-02 pattern):
--   * no user_id to key a policy on, and FORCE would lock out cc_owner, the only writer;
--   * cc_app gets SELECT only (the categories endpoint, the expense use cases); cc_worker
--     and cc_webhook get nothing;
--   * RLS is ENABLED (not forced) with one SELECT policy for cc_app and no write policy,
--     so a write granted by mistake later is still refused (INSERT) or matches no row;
--   * writes happen only here and in later migrations, as cc_owner.

-- +goose Up

CREATE TABLE expense_categories (
  code             text        PRIMARY KEY,
  name             text        NOT NULL,  -- es-MX, shown in the category picker
  sort_order       integer     NOT NULL,  -- spec order; the client sorts by use on top of it
  legal_basis      text        NOT NULL,  -- article level (MODULE-expenses §2.2)
  source_url       text        NOT NULL,
  needs_cpa_review boolean     NOT NULL DEFAULT false,
  review_note      text,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  -- docs/api/openapi.yaml ExpenseFields.category_code: maxLength 40; snake_case keys.
  CONSTRAINT expense_categories_code_format CHECK (code ~ '^[a-z][a-z0-9_]{0,39}$'),
  CONSTRAINT expense_categories_name_valid CHECK (
    name <> '' AND name = btrim(name) AND name IS NFC NORMALIZED),
  CONSTRAINT expense_categories_sort_order_positive CHECK (sort_order >= 1),
  CONSTRAINT expense_categories_sort_order_key UNIQUE (sort_order),
  CONSTRAINT expense_categories_legal_basis_present CHECK (btrim(legal_basis) <> ''),
  CONSTRAINT expense_categories_source_url_format CHECK (source_url ~ '^https?://[^[:space:]]+$'),
  -- A flagged row says why; an unflagged row carries no stale note.
  CONSTRAINT expense_categories_review_note_iff_flagged CHECK (
    needs_cpa_review = (review_note IS NOT NULL) AND (review_note IS NULL OR btrim(review_note) <> ''))
);
COMMENT ON TABLE expense_categories IS
  'Global expense category catalog (M4-02). Read by cc_app, written only by migrations.';

CREATE TRIGGER expense_categories_set_updated_at
  BEFORE UPDATE ON expense_categories
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE expense_categories ENABLE ROW LEVEL SECURITY;
-- Intent: cc_app reads the whole catalog (global config); no write policy for any runtime role.
CREATE POLICY expense_categories_app_read ON expense_categories FOR SELECT TO cc_app USING (true);
GRANT SELECT ON expense_categories TO cc_app;

-- category_regime_rules: one row per (category, catalog regime, person type), so a lookup
-- never misses (a catalog test asserts the matrix is complete). The primary key leads with
-- (regime_code, person_type): the categories endpoint reads every rule of one regime, and
-- the same btree serves the foreign key to regimes. Catalog rows change only through
-- migrations, so the category side needs no index of its own.
CREATE TABLE category_regime_rules (
  regime_code            text        NOT NULL,
  person_type            text        NOT NULL,
  category_code          text        NOT NULL REFERENCES expense_categories (code),
  deductibility          text        NOT NULL,
  default_deductible     boolean     NOT NULL,
  iva_creditable_default boolean     NOT NULL,
  cash_payment_rule      text        NOT NULL,
  note                   text,                  -- es-MX, user-facing; NULL = no note
  legal_basis            text        NOT NULL,
  source_url             text        NOT NULL,
  needs_cpa_review       boolean     NOT NULL DEFAULT false,
  review_note            text,
  created_at             timestamptz NOT NULL DEFAULT now(),
  updated_at             timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (regime_code, person_type, category_code),
  -- 626 PF and 626 PM are different rows (RESICO PF never deducts, RESICO PM does).
  FOREIGN KEY (regime_code, person_type) REFERENCES regimes (code, person_type),
  CONSTRAINT category_regime_rules_deductibility_valid CHECK (deductibility IN (
    'deductible', 'partially_deductible', 'not_deductible', 'isr_not_applicable')),
  -- MODULE-expenses §2.2: a row that never lowers ISR never starts as deductible.
  CONSTRAINT category_regime_rules_default_follows_hint CHECK (
    NOT default_deductible OR deductibility IN ('deductible', 'partially_deductible')),
  CONSTRAINT category_regime_rules_cash_payment_rule_valid CHECK (
    cash_payment_rule IN ('over_limit', 'any_cash', 'not_applicable')),
  -- The cash rule matters exactly where the expense can lower ISR.
  CONSTRAINT category_regime_rules_cash_rule_follows_hint CHECK (
    (cash_payment_rule = 'not_applicable') = (deductibility IN ('not_deductible', 'isr_not_applicable'))),
  CONSTRAINT category_regime_rules_note_valid CHECK (
    note IS NULL OR (note <> '' AND note = btrim(note) AND note IS NFC NORMALIZED)),
  CONSTRAINT category_regime_rules_legal_basis_present CHECK (btrim(legal_basis) <> ''),
  CONSTRAINT category_regime_rules_source_url_format CHECK (source_url ~ '^https?://[^[:space:]]+$'),
  CONSTRAINT category_regime_rules_review_note_iff_flagged CHECK (
    needs_cpa_review = (review_note IS NOT NULL) AND (review_note IS NULL OR btrim(review_note) <> ''))
);
COMMENT ON TABLE category_regime_rules IS
  'Deductibility hint and defaults per (regime, person type, expense category) (M4-02). Complete matrix; CPA review pending.';
COMMENT ON COLUMN category_regime_rules.default_deductible IS
  'The deductible value an expense starts with (before has_cfdi and the cash rule). Conservative for partial cases.';
COMMENT ON COLUMN category_regime_rules.iva_creditable_default IS
  'IVA paid is normally creditable (LIVA 5 fr. I). Hint for M5; false for informational regimes.';
COMMENT ON COLUMN category_regime_rules.cash_payment_rule IS
  'LISR 27 fr. III: over_limit | any_cash (combustibles) | not_applicable. The limit is RateCatalog data.';

CREATE TRIGGER category_regime_rules_set_updated_at
  BEFORE UPDATE ON category_regime_rules
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE category_regime_rules ENABLE ROW LEVEL SECURITY;
-- Intent: cc_app reads every rule (global config); no write policy for any runtime role.
CREATE POLICY category_regime_rules_app_read ON category_regime_rules FOR SELECT TO cc_app USING (true);
GRANT SELECT ON category_regime_rules TO cc_app;

-- ---------------------------------------------------------------------------------------
-- Seeds: categories, MODULE-expenses §2.2 in its order. Source: LISR, current text
-- (research Part D #2).
-- ---------------------------------------------------------------------------------------
INSERT INTO expense_categories (code, name, sort_order, legal_basis, source_url, needs_cpa_review, review_note)
SELECT c.code, c.name, c.sort_order, c.legal_basis, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf', true,
       'Seeded from MODULE-expenses 2.2 (OQ-E3) without CPA sign-off; article-level basis, fractions not verified.'
FROM (VALUES
  ('mercancia',               'Mercancía y materias primas',                          1, 'LISR arts. 25, 103, 208'),
  ('servicios_profesionales', 'Honorarios y servicios profesionales',                 2, 'LISR arts. 25, 103, 115'),
  ('renta_local',             'Renta de local, oficina o bodega',                     3, 'LISR arts. 25, 103'),
  ('servicios_basicos',       'Luz, agua, teléfono e internet',                       4, 'LISR arts. 103, 115'),
  ('software',                'Software y suscripciones',                             5, 'LISR art. 103'),
  ('papeleria',               'Papelería y artículos de oficina',                     6, 'LISR art. 103'),
  ('publicidad',              'Publicidad y mercadotecnia',                           7, 'LISR art. 103'),
  ('combustible',             'Combustibles',                                         8, 'LISR art. 27'),
  ('vehiculo_mantenimiento',  'Mantenimiento y refacciones de vehículo',              9, 'LISR arts. 28, 36'),
  ('viaticos',                'Viáticos y gastos de viaje',                          10, 'LISR art. 28'),
  ('restaurantes',            'Consumo en restaurantes',                             11, 'LISR art. 28'),
  ('equipo_computo',          'Equipo de cómputo (inversión)',                       12, 'LISR arts. 31, 34'),
  ('mobiliario',              'Mobiliario y equipo (inversión)',                     13, 'LISR arts. 31, 34'),
  ('vehiculo_compra',         'Compra de automóvil (inversión)',                     14, 'LISR arts. 34, 36'),
  ('construcciones',          'Construcciones y mejoras al inmueble',                15, 'LISR arts. 34, 115'),
  ('sueldos',                 'Sueldos y salarios pagados',                          16, 'LISR arts. 28, 115'),
  ('cuotas_imss',             'Cuotas IMSS e INFONAVIT patronales',                  17, 'LISR arts. 25, 103, 115'),
  ('intereses',               'Intereses pagados',                                   18, 'LISR arts. 25, 103, 115'),
  ('seguros',                 'Seguros y fianzas',                                   19, 'LISR arts. 25, 103, 115'),
  ('mantenimiento',           'Mantenimiento y reparaciones',                        20, 'LISR arts. 103, 115'),
  ('predial',                 'Impuesto predial',                                    21, 'LISR arts. 103, 115'),
  ('capacitacion',            'Capacitación y cursos',                               22, 'LISR art. 103'),
  ('comisiones',              'Comisiones bancarias y de plataformas',               23, 'LISR art. 103'),
  ('fletes',                  'Fletes y paquetería',                                 24, 'LISR art. 103'),
  ('gastos_personales',       'Gastos personales (médicos, colegiaturas, donativos)', 25, 'LISR art. 151'),
  ('otros',                   'Otros gastos',                                        26, 'LISR art. 27')
) AS c (code, name, sort_order, legal_basis);

-- ---------------------------------------------------------------------------------------
-- Seeds: rules. The spec table (§2.2) gives, per category, the hint of the "Business"
-- column (612, 621, 625, 601, 620, 622, 626 PM, 603: same values, one row per regime), the
-- 606 column (arrendamiento, deduction_mode = actual) and the defaults of both; 626 PF is
-- always "ISR not applicable" and every other regime is informational. The cross join
-- with regimes writes the complete matrix; the class of each regime is listed here,
-- explicitly, because it is a fiscal judgment, not a property of the regime row.
--   D = deductible, P = partially_deductible, N = not_deductible.
-- Notes (es-MX, user-facing):
--   * business and 606 rows carry the category note; the regimes whose provisional ISR
--     uses the profit coefficient (required input coeficiente_utilidad: 601, 622) add the
--     coefficient note;
--   * 626 PF rows carry the RESICO note of PRD US-E1; informational rows say so.
-- Row 13's "Igual que el anterior" is spelled out: the picker is searchable and sorted by
-- use, so a note may never point at its neighbor.
-- iva_creditable_default: true where the expense can be ISR-deductible for the regime
-- (D or P), or would be under the business column for 626 PF (LIVA 5 fr. I); false for
-- not_deductible and for informational regimes.
-- ---------------------------------------------------------------------------------------
INSERT INTO category_regime_rules (regime_code, person_type, category_code, deductibility, default_deductible,
                                   iva_creditable_default, cash_payment_rule, note, legal_basis, source_url,
                                   needs_cpa_review, review_note)
WITH spec (category_code, business, lessor, default_business, default_lessor, note) AS (VALUES
  ('mercancia',               'D', 'N', true,  false, NULL),
  ('servicios_profesionales', 'D', 'D', true,  true,  'Si eres persona moral, revisa las retenciones'),
  ('renta_local',             'D', 'N', true,  false, 'Si eres persona moral, revisa las retenciones'),
  ('servicios_basicos',       'D', 'P', true,  true,  'Arrendamiento: sólo el agua del inmueble rentado'),
  ('software',                'D', 'N', true,  false, NULL),
  ('papeleria',               'D', 'N', true,  false, NULL),
  ('publicidad',              'D', 'N', true,  false, NULL),
  ('combustible',             'D', 'N', true,  false, 'Sólo deducible si lo pagas con tarjeta, transferencia o monedero, sin importar el monto'),
  ('vehiculo_mantenimiento',  'P', 'N', true,  false, 'Tiene topes de deducción'),
  ('viaticos',                'P', 'N', true,  false, 'Tiene topes y requisitos'),
  ('restaurantes',            'P', 'N', false, false, 'Sólo una parte es deducible'),
  ('equipo_computo',          'P', 'N', false, false, 'Se deduce por depreciación anual, no completo en el mes'),
  ('mobiliario',              'P', 'N', false, false, 'Se deduce por depreciación anual, no completo en el mes'),
  ('vehiculo_compra',         'P', 'N', false, false, 'Depreciación con tope'),
  ('construcciones',          'P', 'P', false, false, 'Se deduce por depreciación'),
  ('sueldos',                 'P', 'D', true,  true,  'Parte de las prestaciones exentas no es deducible'),
  ('cuotas_imss',             'D', 'D', true,  true,  NULL),
  ('intereses',               'D', 'D', true,  true,  'Sólo los intereses reales'),
  ('seguros',                 'D', 'D', true,  true,  NULL),
  ('mantenimiento',           'D', 'D', true,  true,  NULL),
  ('predial',                 'D', 'D', true,  true,  'En arrendamiento se suma aun con la deducción ciega'),
  ('capacitacion',            'D', 'N', true,  false, NULL),
  ('comisiones',              'D', 'N', true,  false, NULL),
  ('fletes',                  'D', 'N', true,  false, 'Si eres persona moral, revisa la retención de IVA (4 %)'),
  ('gastos_personales',       'N', 'N', false, false, 'Se aplican en tu declaración anual, no en el pago mensual'),
  ('otros',                   'P', 'N', false, false, 'Verifica que cumpla los requisitos')
),
regime_class (regime_code, person_type, class) AS (VALUES
  ('612', 'fisica', 'business'), ('621', 'fisica', 'business'), ('625', 'fisica', 'business'),
  ('601', 'moral', 'business'), ('603', 'moral', 'business'), ('620', 'moral', 'business'),
  ('622', 'moral', 'business'), ('626', 'moral', 'business'),
  ('606', 'fisica', 'lessor'),
  ('626', 'fisica', 'resico_pf')
),
classified AS (
  SELECT r.code AS regime_code, r.person_type, coalesce(k.class, 'informational') AS class,
         r.required_inputs @> '[{"key": "coeficiente_utilidad"}]' AS uses_coefficient,
         r.legal_basis AS regime_basis
  FROM regimes r
  LEFT JOIN regime_class k ON k.regime_code = r.code AND k.person_type = r.person_type
),
resolved AS (
  SELECT g.regime_code, g.person_type, g.class, g.uses_coefficient, g.regime_basis, s.category_code, s.note,
         e.legal_basis AS category_basis, s.business,
         CASE g.class WHEN 'business' THEN s.business WHEN 'lessor' THEN s.lessor ELSE 'X' END AS letter,
         CASE g.class WHEN 'business' THEN s.default_business WHEN 'lessor' THEN s.default_lessor ELSE false END
           AS default_deductible
  FROM classified g
  CROSS JOIN spec s
  JOIN expense_categories e ON e.code = s.category_code
)
SELECT regime_code, person_type, category_code,
       CASE letter
         WHEN 'D' THEN 'deductible'
         WHEN 'P' THEN 'partially_deductible'
         WHEN 'N' THEN 'not_deductible'
         ELSE 'isr_not_applicable'
       END,
       default_deductible,
       CASE class
         WHEN 'informational' THEN false
         WHEN 'resico_pf' THEN business <> 'N'
         ELSE letter <> 'N'
       END,
       CASE
         WHEN letter NOT IN ('D', 'P') THEN 'not_applicable'
         WHEN category_code = 'combustible' THEN 'any_cash'
         ELSE 'over_limit'
       END,
       CASE
         WHEN class = 'resico_pf' THEN 'En RESICO tus gastos no reducen tu ISR, pero el IVA que pagaste sí puede acreditarse'
         WHEN class = 'informational' THEN 'Este régimen es informativo'
         WHEN uses_coefficient THEN concat_ws('. ', note,
           'En los pagos provisionales del 601 el ISR se calcula con el coeficiente de utilidad; este gasto cuenta para tu IVA y tu declaración anual')
         ELSE note
       END,
       CASE class
         WHEN 'business' THEN category_basis
         WHEN 'lessor' THEN 'LISR art. 115'
         WHEN 'resico_pf' THEN 'LISR art. 113-E'
         ELSE regime_basis
       END,
       'https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf',
       true,
       concat_ws(' ',
         'Seeded from MODULE-expenses 2.2 (OQ-E3) without CPA sign-off.',
         CASE WHEN category_code = 'combustible' OR letter IN ('D', 'P')
           THEN 'Cash rule LISR 27 fr. III ($2,000 on the total with IVA; fuel never in cash) is not in the research (OQ-E4).' END,
         CASE WHEN regime_code = '620'
           THEN '620 is informational in the regime catalog, but the spec gives it the business rows.' END,
         CASE WHEN class = 'resico_pf'
           THEN 'iva_creditable_default assumes LIVA 5 fr. I "would be deductible" uses the business column.' END)
FROM resolved;

-- +goose Down
DROP TABLE category_regime_rules;
DROP TABLE expense_categories;
