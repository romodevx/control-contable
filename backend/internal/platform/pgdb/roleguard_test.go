package pgdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleProfileViolations(t *testing.T) {
	tests := []struct {
		name    string
		profile roleProfile
		want    []string
	}{
		{"plain runtime login", roleProfile{name: "cc_app"}, nil},
		{"superuser", roleProfile{name: "postgres", superuser: true}, []string{"is SUPERUSER"}},
		{"bypassrls", roleProfile{name: "cc_app", bypassRLS: true}, []string{"has BYPASSRLS"}},
		{"owner", roleProfile{name: "cc_owner", ownsObjects: true},
			[]string{"owns (or can act as the owner of) the database, a schema or a table"}},
		{"member of a privileged role", roleProfile{name: "cc_app", canBecomePrivileged: true},
			[]string{"is a member of a SUPERUSER or BYPASSRLS role"}},
		{"every problem is reported", roleProfile{name: "x", superuser: true, bypassRLS: true, ownsObjects: true, canBecomePrivileged: true},
			[]string{
				"is SUPERUSER", "has BYPASSRLS",
				"owns (or can act as the owner of) the database, a schema or a table",
				"is a member of a SUPERUSER or BYPASSRLS role",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.profile.violations())
		})
	}
}
