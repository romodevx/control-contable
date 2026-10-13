package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrPrivilegedRole reports a runtime pool logged in as a role that Row Level
// Security does not bind: a superuser, a BYPASSRLS role, an owner (who may
// disable RLS on what it owns) or a member of any of those.
var ErrPrivilegedRole = errors.New("pgdb: the runtime database role is privileged")

// runtimeRoleQuery profiles current_user. pg_has_role(..., 'MEMBER') counts any
// role current_user is or can SET ROLE to, so ownership or privilege reached
// through membership counts too. System schemas are left out.
const runtimeRoleQuery = `
SELECT r.rolname, r.rolsuper, r.rolbypassrls,
  EXISTS (SELECT 1 FROM pg_database d
          WHERE d.datname = current_database() AND pg_has_role(d.datdba, 'MEMBER'))
  OR EXISTS (SELECT 1 FROM pg_namespace n
             WHERE n.nspname NOT LIKE 'pg\_%' AND n.nspname <> 'information_schema'
               AND pg_has_role(n.nspowner, 'MEMBER'))
  OR EXISTS (SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
             WHERE n.nspname NOT LIKE 'pg\_%' AND n.nspname <> 'information_schema'
               AND pg_has_role(c.relowner, 'MEMBER')),
  EXISTS (SELECT 1 FROM pg_roles p
          WHERE p.oid <> r.oid AND (p.rolsuper OR p.rolbypassrls) AND pg_has_role(p.oid, 'MEMBER'))
FROM pg_roles r
WHERE r.rolname = current_user`

// roleProfile is what runtimeRoleQuery finds out about current_user.
type roleProfile struct {
	name                string
	superuser           bool
	bypassRLS           bool
	ownsObjects         bool
	canBecomePrivileged bool
}

// violations lists every reason the role must not serve requests.
func (p roleProfile) violations() []string {
	var found []string
	if p.superuser {
		found = append(found, "is SUPERUSER")
	}
	if p.bypassRLS {
		found = append(found, "has BYPASSRLS")
	}
	if p.ownsObjects {
		found = append(found, "owns (or can act as the owner of) the database, a schema or a table")
	}
	if p.canBecomePrivileged {
		found = append(found, "is a member of a SUPERUSER or BYPASSRLS role")
	}
	return found
}

// CheckRuntimeRole verifies that db is logged in as a plain runtime role (cc_app,
// cc_webhook, cc_worker): RLS binds it. A privileged role is ErrPrivilegedRole;
// a role that cannot be verified (database unreachable) is an error too.
func CheckRuntimeRole(ctx context.Context, db *sql.DB) error {
	var p roleProfile
	err := db.QueryRowContext(ctx, runtimeRoleQuery).
		Scan(&p.name, &p.superuser, &p.bypassRLS, &p.ownsObjects, &p.canBecomePrivileged)
	if err != nil {
		return fmt.Errorf("pgdb: verify runtime role: %w", err)
	}
	if found := p.violations(); len(found) > 0 {
		return fmt.Errorf("%w: %s %s", ErrPrivilegedRole, p.name, strings.Join(found, ", "))
	}
	return nil
}
