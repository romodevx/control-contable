package params

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// dataDir holds the embedded parameter files: data/<fiscal_year>/v<N>.yaml.
const dataDir = "data"

// readme is the one non-parameter file allowed, at the root only.
const readme = "README.md"

//go:embed data
var embedded embed.FS

var (
	yearDir = regexp.MustCompile(`^[0-9]{4}$`)
	// setFile caps the version at six digits, so it always fits an int.
	setFile = regexp.MustCompile(`^([0-9]{4})/v([1-9][0-9]{0,5})\.yaml$`)
)

// Embedded loads the parameter sets compiled into the binary. The composition
// root calls it once at startup and refuses to boot on an error.
func Embedded() (*taxparams.Catalog, error) {
	return Load(embedded, dataDir)
}

// Load reads every <fiscal_year>/v<N>.yaml under root in fsys and returns
// their catalog. Every bad file is reported, joined, with its path; an I/O
// error stops the walk.
func Load(fsys fs.FS, root string) (*taxparams.Catalog, error) {
	var (
		sets     []*taxparams.Params
		problems []error
	)
	walkErr := fs.WalkDir(fsys, root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := relative(root, name)
		switch {
		case entry.IsDir():
			if rel != "." && !yearDir.MatchString(rel) {
				problems = append(problems, fmt.Errorf("%s: %w", rel, ErrUnexpectedFile))
				return fs.SkipDir
			}
			return nil
		case rel == readme:
			return nil
		}
		set, err := loadFile(fsys, name, rel)
		if errors.Is(err, errIO) {
			return err
		}
		if err != nil {
			problems = append(problems, err)
			return nil
		}
		sets = append(sets, set)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("params: read %s: %w", root, walkErr)
	}
	if len(problems) > 0 {
		return nil, errors.Join(problems...)
	}
	return taxparams.NewCatalog(sets...)
}

// errIO marks a read failure, which aborts the walk instead of being
// collected with the content problems.
var errIO = errors.New("params: read failed")

func relative(root, name string) string {
	if root == "." {
		return name
	}
	if name == root {
		return "."
	}
	return strings.TrimPrefix(name, root+"/")
}

func loadFile(fsys fs.FS, name, rel string) (*taxparams.Params, error) {
	m := setFile.FindStringSubmatch(rel)
	if m == nil {
		return nil, fmt.Errorf("%s: want <YYYY>/v<N>.yaml: %w", rel, ErrUnexpectedFile)
	}
	// setFile admits only short digit runs, so neither conversion can fail.
	year, _ := strconv.Atoi(m[1])
	version, _ := strconv.Atoi(m[2])
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %w", rel, errIO, err)
	}
	spec, err := decode(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", rel, err)
	}
	if spec.FiscalYear != year || spec.Version != version {
		return nil, fmt.Errorf("%s: file says %d/v%d: %w", rel, spec.FiscalYear, spec.Version, ErrPathMismatch)
	}
	set, err := taxparams.NewParams(spec)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", rel, err)
	}
	return set, nil
}

// decode parses exactly one strict YAML document into a Spec.
func decode(data []byte) (taxparams.Spec, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var doc fileDoc
	if err := dec.Decode(&doc); err != nil {
		return taxparams.Spec{}, fmt.Errorf("%w: %w", ErrSyntax, err)
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return taxparams.Spec{}, fmt.Errorf("%w: more than one document", ErrSyntax)
	}
	return (&converter{}).spec(doc)
}
