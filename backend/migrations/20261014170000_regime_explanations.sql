-- Regime explanations (M5-19): the plain-language es-MX text of regimes.explanation for
-- every regime the tax engine does not fully estimate (MODULE-tax-engine §2.2, §2.4):
--   * explanation only (informational): 605, 607, 608, 610 (PF and PM), 611, 614, 615,
--     616, 620, 623, 624 — no figure, only this text;
--   * partial: 621 (RIF, bimonthly, unverified reduction), 622 (AGAPES PM, IVA only) and
--     603 (fines no lucrativos, IVA only) — the text says what is not estimated.
-- The engine returns the label key regime_explanation:<code>/<person_type>; the adapter
-- renders it from this column. 625 (partial without the definitive option) is left to its
-- own strategy PR.
--
-- Each text is short, describes who calculates or withholds the tax and when the annual
-- return matters, cites the article, and gives no advice. Sources: research §A.1, §B.1,
-- Part C; the catalog's legal_basis column.
--
-- The column exists since 20261013180000_regimes_catalog.sql (explanation text, NULL
-- until this migration, CHECK non-blank); that applied migration is never edited.
-- RLS: regimes is global config (ENABLE, SELECT-only policy for cc_app, no write grant);
-- this migration runs as the owner and changes no policy or grant.

-- +goose Up

-- +goose StatementBegin
DO $$
DECLARE
  updated integer;
BEGIN
  UPDATE regimes AS r
     SET explanation = t.explanation
    FROM (VALUES
      ('605', 'fisica',
       'Tu patrón calcula y te retiene el ISR cada vez que te paga (LISR art. 96). La app no estima este impuesto. La declaración anual importa si, por ejemplo, tuviste dos patrones a la vez, dejaste de trabajar antes del 31 de diciembre, tuviste otros ingresos o tus sueldos del año superaron $400,000 (LISR art. 98).'),
      ('607', 'fisica',
       'El ISR se paga por cada venta, no cada mes: en general hay un pago provisional dentro de los 15 días siguientes a la operación y, en la venta de un inmueble, el notario lo calcula y lo entera (LISR art. 126). La ganancia también cuenta en la declaración anual. La app no estima este impuesto.'),
      ('608', 'fisica',
       'Son ingresos que la ley no ubica en otro capítulo (LISR art. 141 y siguientes). Según el tipo de ingreso, quien te paga puede retenerte ISR o te corresponde un pago provisional, y estos ingresos se acumulan en la declaración anual. La app no estima este impuesto.'),
      ('610', 'fisica',
       'Para quien reside en el extranjero sin establecimiento permanente en México, quien paga desde México calcula y retiene el ISR sobre el ingreso de fuente mexicana, normalmente como pago definitivo (LISR art. 153). Por eso, en general, no hay declaración anual en México por estos ingresos. La app no estima este impuesto.'),
      ('610', 'moral',
       'Para quien reside en el extranjero sin establecimiento permanente en México, quien paga desde México calcula y retiene el ISR sobre el ingreso de fuente mexicana, normalmente como pago definitivo (LISR art. 153). Por eso, en general, no hay declaración anual en México por estos ingresos. La app no estima este impuesto.'),
      ('611', 'fisica',
       'La empresa que te paga dividendos te retiene 10 % como pago definitivo (LISR art. 140). Además, el dividendo, junto con el ISR que pagó la empresa, se incluye en la declaración anual, donde ese ISR se puede acreditar. La app no estima este impuesto.'),
      ('614', 'fisica',
       'El banco o la institución que te paga intereses te retiene ISR durante el año (LISR art. 54; LIF 2026 art. 24). Si sólo tienes intereses y los reales del año no superan $100,000, esa retención es definitiva; si los superan o tienes otros ingresos, se acumulan en la declaración anual (LISR art. 135). La app no estima este impuesto.'),
      ('615', 'fisica',
       'Quien entrega el premio calcula y te retiene el ISR (LISR art. 138). En general esa retención es pago definitivo, así que estos ingresos no se incluyen en la declaración anual. La app no estima este impuesto.'),
      ('616', 'fisica',
       'Este régimen es para quien está inscrito en el RFC sin actividades por las que deba pagar impuestos: no hay impuesto que estimar ni declaraciones periódicas. Si empiezas a recibir ingresos, el cambio de régimen se avisa al SAT (CFF art. 27).'),
      ('620', 'moral',
       'La cooperativa calcula el ISR de sus socios y lo difiere hasta que les distribuye los rendimientos (LISR arts. 194 a 199). Los pagos y la declaración anual los prepara la propia cooperativa. La app no estima este impuesto.'),
      ('623', 'moral',
       'La sociedad integradora y sus integradas calculan su ISR con las reglas del régimen opcional para grupos, que permite diferir parte del impuesto (LISR arts. 59 a 71). Cada sociedad presenta sus pagos provisionales y su declaración anual. La app no estima este impuesto.'),
      ('624', 'moral',
       'El coordinado calcula y paga el ISR por cuenta de sus integrantes dedicados al autotransporte terrestre y presenta su declaración anual (LISR arts. 72 y 73). La app no estima este impuesto.'),
      ('621', 'fisica',
       'El RIF se paga por bimestre: la app estima el ISR y el IVA en el segundo mes de cada bimestre; en el primero no muestra cifra. Los pagos son definitivos y el ISR se reduce según tu año en el RIF (LISR art. 111, texto vigente hasta 2021). Un contador aún no valida esa reducción: toma la cifra con reserva.'),
      ('622', 'moral',
       'La app estima el IVA del mes, no el ISR: la exención y la reducción para actividades agrícolas, ganaderas, silvícolas y pesqueras (LISR art. 74) dependen de datos por socio que la app aún no pide. El ISR de los pagos provisionales y de la declaración anual se calcula fuera de la app.'),
      ('603', 'moral',
       'La app estima el IVA del mes cuando tus actividades lo causan. En general no hay pagos provisionales de ISR: la persona moral paga ISR sólo en casos como gastos no deducibles o ingresos ajenos a su objeto, y presenta su declaración anual (LISR arts. 79 y 80). La app no estima ese ISR.')
    ) AS t(code, person_type, explanation)
   WHERE r.code = t.code AND r.person_type = t.person_type;
  GET DIAGNOSTICS updated = ROW_COUNT;
  IF updated <> 15 THEN
    RAISE EXCEPTION 'regime explanations: expected 15 catalog rows, updated %', updated;
  END IF;
END
$$;
-- +goose StatementEnd

-- A regime the engine does not fully estimate has no figure to show without its text, so
-- the text is mandatory for it, for these rows and for any regime added later.
ALTER TABLE regimes
  ADD CONSTRAINT regimes_explanation_when_not_full
  CHECK (treatment = 'full' OR explanation IS NOT NULL);

-- +goose Down

ALTER TABLE regimes DROP CONSTRAINT regimes_explanation_when_not_full;

UPDATE regimes
   SET explanation = NULL
 WHERE (code, person_type) IN (
   ('605', 'fisica'), ('607', 'fisica'), ('608', 'fisica'), ('610', 'fisica'), ('610', 'moral'),
   ('611', 'fisica'), ('614', 'fisica'), ('615', 'fisica'), ('616', 'fisica'), ('620', 'moral'),
   ('623', 'moral'), ('624', 'moral'), ('621', 'fisica'), ('622', 'moral'), ('603', 'moral'));
