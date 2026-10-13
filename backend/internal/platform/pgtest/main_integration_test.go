//go:build integration

package pgtest_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// The throwaway schema these tests prove the helpers against. It stands in for the
// real migrations (P0-10) and is applied through the same migrator hook they will use.
//
//   - notes:        owner-scoped, ENABLE + FORCE RLS, fail-closed helper (the data-model pattern).
//   - leaky_notes:  same shape, no RLS. Every isolation check must catch it.
//   - strict_notes: RLS whose policy errors (instead of matching nothing) on an unset context.
const fixtureDDL = `
CREATE FUNCTION app_user_id() RETURNS uuid STABLE LANGUAGE sql
AS $$ SELECT nullif(current_setting('app.current_user_id', true), '')::uuid $$;
GRANT EXECUTE ON FUNCTION app_user_id() TO cc_app;

CREATE TABLE notes (
  id      bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id uuid NOT NULL,
  body    text NOT NULL
);
ALTER TABLE notes ENABLE ROW LEVEL SECURITY;
ALTER TABLE notes FORCE ROW LEVEL SECURITY;
CREATE POLICY notes_owner ON notes
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
GRANT SELECT, INSERT, UPDATE, DELETE ON notes TO cc_app;

CREATE TABLE leaky_notes (
  id      bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id uuid NOT NULL,
  body    text NOT NULL
);
GRANT SELECT, INSERT, UPDATE, DELETE ON leaky_notes TO cc_app;

CREATE TABLE strict_notes (
  id      bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id uuid NOT NULL,
  body    text NOT NULL
);
ALTER TABLE strict_notes ENABLE ROW LEVEL SECURITY;
ALTER TABLE strict_notes FORCE ROW LEVEL SECURITY;
CREATE POLICY strict_notes_owner ON strict_notes
  USING (user_id = current_setting('app.current_user_id')::uuid)
  WITH CHECK (user_id = current_setting('app.current_user_id')::uuid);
GRANT SELECT, INSERT, UPDATE, DELETE ON strict_notes TO cc_app;
`

// migratedAs records the role the migrator ran as, so a test can prove it is cc_owner.
var migratedAs string

func fixtureSchema(ctx context.Context, db *sql.DB) error {
	if err := db.QueryRowContext(ctx, "SELECT current_user").Scan(&migratedAs); err != nil {
		return fmt.Errorf("read migrator role: %w", err)
	}
	if _, err := db.ExecContext(ctx, fixtureDDL); err != nil {
		return fmt.Errorf("apply fixture schema: %w", err)
	}
	return nil
}

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(fixtureSchema)))
}
