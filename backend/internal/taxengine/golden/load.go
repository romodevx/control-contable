package golden

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

const (
	caseExt    = ".yaml"
	readmeName = "README.md"
)

// personDir is the directory suffix of each person type.
var personDir = map[fiscal.PersonType]string{fiscal.PersonFisica: "pf", fiscal.PersonMoral: "pm"}

// Load reads every case under root in fsys, laid out as
// <regime>-<pf|pm>/[...]/<id>.yaml, sorted by path. README.md files are
// allowed anywhere. Every bad file is reported in one error; no case at all
// is ErrNoCases. fsys is read-only: the harness never writes a golden.
func Load(fsys fs.FS, root string) ([]Case, error) {
	var (
		cases    []Case
		problems []error
		seen     = map[string]string{}
	)
	walkErr := fs.WalkDir(fsys, root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(name, root), "/")
		if entry.IsDir() || path.Base(rel) == readmeName {
			return nil
		}
		c, err := loadCase(fsys, name, rel)
		if err != nil {
			problems = append(problems, err)
			return nil
		}
		if first, dup := seen[c.ID]; dup {
			problems = append(problems, fmt.Errorf("%s: id %s already used by %s: %w", rel, c.ID, first, ErrLayout))
			return nil
		}
		seen[c.ID] = rel
		cases = append(cases, c)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("golden: walk %s: %w", root, walkErr)
	}
	if len(problems) > 0 {
		return nil, errors.Join(problems...)
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("%s: %w", root, ErrNoCases)
	}
	slices.SortFunc(cases, func(a, b Case) int { return strings.Compare(a.File, b.File) })
	return cases, nil
}

func loadCase(fsys fs.FS, name, rel string) (Case, error) {
	dir, _, nested := strings.Cut(rel, "/")
	if !nested || path.Ext(rel) != caseExt {
		return Case{}, fmt.Errorf("%s: want <regime>-<pf|pm>/<id>%s: %w", rel, caseExt, ErrLayout)
	}
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return Case{}, fmt.Errorf("%s: %w", rel, err)
	}
	c, err := Decode(data, rel)
	if err != nil {
		return Case{}, err
	}
	if stem := strings.TrimSuffix(path.Base(rel), caseExt); stem != c.ID {
		return Case{}, fmt.Errorf("%s: file name must be the id %s: %w", rel, c.ID, ErrLayout)
	}
	if want := string(c.Input.Key.Regime) + "-" + personDir[c.Input.Key.PersonType]; dir != want {
		return Case{}, fmt.Errorf("%s: case of %s belongs under %s/: %w", rel, c.Input.Key, want, ErrLayout)
	}
	return c, nil
}
