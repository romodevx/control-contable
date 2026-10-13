package flags

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

// Errors of the flag administration use case.
var (
	ErrInvalidKey        = errors.New("flags: key must be snake_case module_<name>")
	ErrInvalidPresetName = errors.New("flags: preset name must be snake_case")
	ErrUnknownFlag       = errors.New("flags: no such flag")
	ErrUnknownPreset     = errors.New("flags: no such preset")
)

// The same formats the feature_flags and flag_presets CHECK constraints enforce.
var (
	keyPattern    = regexp.MustCompile(`^module_[a-z][a-z0-9]*(_[a-z0-9]+)*$`)
	presetPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)
)

// ValidKey reports whether key is a well-formed flag key ("module_<name>").
func ValidKey(key string) bool { return keyPattern.MatchString(key) }

// ValidPresetName reports whether name is a well-formed preset name.
func ValidPresetName(name string) bool { return presetPattern.MatchString(name) }

// Flag is one module flag as the owner administers it.
type Flag struct {
	Key         string
	Enabled     bool
	Description string
}

// Preset is a named set of flags: applying it turns those ON and every other
// flag OFF.
type Preset struct {
	Name        string
	Description string
	Flags       []string
}

// AdminStore is the port to the flag storage with owner privileges.
type AdminStore interface {
	// List returns every flag, ordered by key.
	List(ctx context.Context) ([]Flag, error)
	// SetEnabled turns key ON or OFF and returns its previous state, or
	// ErrUnknownFlag.
	SetEnabled(ctx context.Context, key string, enabled bool) (previous bool, err error)
	// ApplyPreset atomically turns the preset's flags ON and the rest OFF and
	// returns the resulting flags, or ErrUnknownPreset.
	ApplyPreset(ctx context.Context, name string) ([]Flag, error)
	// Presets returns every preset with its flags, ordered by name.
	Presets(ctx context.Context) ([]Preset, error)
}

// Admin is the flag administration use case behind cmd/flags (P0-20). It
// validates input before it reaches the store.
type Admin struct {
	store AdminStore
}

// NewAdmin returns an Admin over store.
func NewAdmin(store AdminStore) Admin { return Admin{store: store} }

// List returns every flag, ordered by key.
func (a Admin) List(ctx context.Context) ([]Flag, error) {
	list, err := a.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list flags: %w", err)
	}
	return list, nil
}

// SetEnabled turns key ON or OFF and returns its previous state.
func (a Admin) SetEnabled(ctx context.Context, key string, enabled bool) (bool, error) {
	if !ValidKey(key) {
		return false, fmt.Errorf("%w: %q", ErrInvalidKey, key)
	}
	previous, err := a.store.SetEnabled(ctx, key, enabled)
	if err != nil {
		return false, fmt.Errorf("set %s: %w", key, err)
	}
	return previous, nil
}

// ApplyPreset applies the named preset and returns the resulting flags.
func (a Admin) ApplyPreset(ctx context.Context, name string) ([]Flag, error) {
	if !ValidPresetName(name) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidPresetName, name)
	}
	result, err := a.store.ApplyPreset(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("apply preset %s: %w", name, err)
	}
	return result, nil
}

// Presets returns every preset with its flags.
func (a Admin) Presets(ctx context.Context) ([]Preset, error) {
	presets, err := a.store.Presets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list presets: %w", err)
	}
	return presets, nil
}
