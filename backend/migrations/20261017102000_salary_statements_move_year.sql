-- Salary constancias may move to another fiscal year (ANN-13). The frozen contract
-- (0.11.x, ANN-02 / PR #134) lets updateSalaryStatement change `fiscal_year`: "moving
-- fiscal_year applies that year's list bound (422 limit_reached; the row itself is not
-- counted twice)". ANN-03 (20261016121000_annual_inputs.sql) had kept the column
-- immutable for cc_app, so the edit the contract promises would fail with 42501.
--
-- Only the column grant changes: the CHECK (2026..2100), the RLS policies (UPDATE USING +
-- WITH CHECK on user_id) and the year index stay as they are; the use case checks the
-- year floor, the future year and the 20-per-year bound under its list lock before the
-- UPDATE. id, user_id, taxpayer_id and person_type stay immutable.

-- +goose Up
GRANT UPDATE (fiscal_year) ON salary_statements TO cc_app;

-- +goose Down
REVOKE UPDATE (fiscal_year) ON salary_statements FROM cc_app;
