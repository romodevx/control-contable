package golden_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

// caseFile returns validCase renamed to id, for regime dir "612-pf".
func caseFile(id string) *fstest.MapFile {
	doc := strings.Replace(validCase, "id: PF-612-ISR-01", "id: "+id, 1)
	return &fstest.MapFile{Data: []byte(doc)}
}

func TestLoadReadsEveryCaseSortedByPath(t *testing.T) {
	fsys := fstest.MapFS{
		"golden/README.md":                   {Data: []byte("# format")},
		"golden/612-pf/PF-612-ISR-02.yaml":   caseFile("PF-612-ISR-02"),
		"golden/612-pf/PF-612-ISR-01.yaml":   caseFile("PF-612-ISR-01"),
		"golden/612-pf/nested/PF-612-X.yaml": caseFile("PF-612-X"),
	}

	cases, err := golden.Load(fsys, "golden")

	require.NoError(t, err)
	ids := make([]string, 0, len(cases))
	for _, c := range cases {
		ids = append(ids, c.ID)
	}
	assert.Equal(t, []string{"PF-612-ISR-01", "PF-612-ISR-02", "PF-612-X"}, ids)
	assert.Equal(t, "612-pf/PF-612-ISR-01.yaml", cases[0].File)
	assert.Equal(t, "612-pf/nested/PF-612-X.yaml", cases[2].File)
}

func TestLoadRejectsABadLayout(t *testing.T) {
	tests := []struct {
		name string
		fsys fstest.MapFS
		want error
	}{
		{"no cases at all", fstest.MapFS{"golden/README.md": {Data: []byte("x")}}, golden.ErrNoCases},
		{"file stem is not the id", fstest.MapFS{"golden/612-pf/other.yaml": caseFile("PF-612-ISR-01")}, golden.ErrLayout},
		{"directory is not the regime", fstest.MapFS{"golden/606-pf/PF-612-ISR-01.yaml": caseFile("PF-612-ISR-01")}, golden.ErrLayout},
		{"directory has the wrong person type", fstest.MapFS{"golden/612-pm/PF-612-ISR-01.yaml": caseFile("PF-612-ISR-01")}, golden.ErrLayout},
		{"case at the root", fstest.MapFS{"golden/PF-612-ISR-01.yaml": caseFile("PF-612-ISR-01")}, golden.ErrLayout},
		{"a stray file", fstest.MapFS{
			"golden/612-pf/PF-612-ISR-01.yaml": caseFile("PF-612-ISR-01"),
			"golden/612-pf/PF-612-ISR-01.yml":  caseFile("PF-612-ISR-01"),
		}, golden.ErrLayout},
		{"a repeated id", fstest.MapFS{
			"golden/612-pf/PF-612-ISR-01.yaml":   caseFile("PF-612-ISR-01"),
			"golden/612-pf/a/PF-612-ISR-01.yaml": caseFile("PF-612-ISR-01"),
		}, golden.ErrLayout},
		{"a bad case", fstest.MapFS{
			"golden/612-pf/PF-612-ISR-01.yaml": {Data: []byte("id: x\n")},
		}, golden.ErrNotHandAuthored},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := golden.Load(tt.fsys, "golden")

			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestLoadReportsEveryBadFileAtOnce(t *testing.T) {
	fsys := fstest.MapFS{
		"golden/612-pf/PF-612-ISR-01.yaml": {Data: []byte("id: x\n")},
		"golden/612-pf/PF-612-ISR-02.yaml": {Data: []byte("# golden: hand-authored\nid: [\n")},
	}

	_, err := golden.Load(fsys, "golden")

	require.ErrorIs(t, err, golden.ErrNotHandAuthored)
	require.ErrorIs(t, err, golden.ErrSyntax)
}
