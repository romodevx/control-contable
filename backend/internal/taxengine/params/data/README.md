# Tax parameter sets (embedded)

One file per fiscal year and version: `<fiscal_year>/v<N>.yaml`, for example
`2026/v1.yaml`. Every value is transcribed verbatim from its official source and
cites it. The schema, validation rules and selection rules are in
`docs/architecture/tax-params-schema.md` (ADR-0002 §3).

Nothing else may live here: the loader rejects any other file or directory, and
a bad file fails the boot and CI.
