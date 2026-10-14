package annualgolden_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/taxengine/annualgolden"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

func file(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }

func withID(id string) string { return strings.Replace(validCase, "id: AN-02", "id: "+id, 1) }

func TestLoadReadsEveryCaseSortedByPath(t *testing.T) {
	fsys := fstest.MapFS{
		"root/README.md":     file("# not a case"),
		"root/pf/AN-02.yaml": file(validCase),
		"root/pf/AN-01.yaml": file(withID("AN-01")),
		"root/pf/README.md":  file("notes"),
	}

	cases, err := annualgolden.Load(fsys, "root")

	require.NoError(t, err)
	require.Len(t, cases, 2)
	assert.Equal(t, "AN-01", cases[0].ID)
	assert.Equal(t, "pf/AN-01.yaml", cases[0].File)
	assert.Equal(t, "AN-02", cases[1].ID)
}

func TestLoadRefusesABadLayout(t *testing.T) {
	tests := []struct {
		name    string
		fsys    fstest.MapFS
		wantErr error
	}{
		{"no case", fstest.MapFS{"root/README.md": file("x")}, golden.ErrNoCases},
		{"file at the root", fstest.MapFS{"root/AN-02.yaml": file(validCase)}, golden.ErrLayout},
		{"wrong extension", fstest.MapFS{"root/pf/AN-02.yml": file(validCase)}, golden.ErrLayout},
		{"file name is not the id", fstest.MapFS{"root/pf/AN-09.yaml": file(validCase)}, golden.ErrLayout},
		{"PF case under pm", fstest.MapFS{"root/pm/AN-02.yaml": file(validCase)}, golden.ErrLayout},
		{"repeated id", fstest.MapFS{
			"root/pf/AN-02.yaml":     file(validCase),
			"root/pf/sub/AN-02.yaml": file(validCase),
		}, golden.ErrLayout},
		{"broken case", fstest.MapFS{"root/pf/AN-02.yaml": file("# generated\n")}, golden.ErrNotHandAuthored},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := annualgolden.Load(tt.fsys, "root")

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
