# Tax parameter sets (embedded)

One file per fiscal year and version: `<fiscal_year>/v<N>.yaml`, for example
`2026/v1.yaml`. Every value is transcribed verbatim from its official source and
cites it. The schema, validation rules and selection rules are in
`docs/architecture/tax-params-schema.md` (ADR-0002 §3).

Nothing else may live here: the loader rejects any other file or directory, and
a bad file fails the boot and CI.

## Published fixed fees are transcribed as printed

Progressive tariffs (art. 96, 106, 116, 152, RIF) are not chained to the
centavo: a row's fixed fee can differ from `round(fee + (upper − lower) × rate)`
of the row before it, by one centavo on a monthly or annual table and by up to
_n_ centavos on the _n_-th accumulated art. 106 table (Anexo 8 publishes those
as the January table × _n_). Transcribe the DOF value; never "correct" it to
make the tariff continuous. `params/tariff_continuity_2026_test.go` bounds the
drift, so a real slip (pesos, not centavos) still fails.
