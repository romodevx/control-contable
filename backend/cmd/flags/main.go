// Command flags is the owner's module feature-flag admin CLI (P0-20; backs
// the /flags skill). It connects as cc_owner, the only role allowed to write
// feature_flags, through the owner URL in DATABASE_URL_OWNER:
//
//	flags list                   every flag, ON/off/hidden, its prerequisites
//	                             and a summary
//	flags enable <module_key>    turn one module ON
//	flags disable <module_key>   turn one module off
//	flags apply-preset <name>    turn the preset's modules ON and the rest off
//	flags presets                the presets and their modules
//
// A module is effectively ON only when its prerequisites (flags.Prerequisites)
// are ON too; a flag that is ON while one is OFF shows as "hidden".
//
// Running API instances cache flags for flags.DefaultTTL, so a change is
// visible to every instance within that time. Enabling a module in a shared
// environment requires its security and tech-lead gates (feature-flags rule).
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagsgorm"
	"github.com/romodevx/control-contable/backend/migrations"
)

const (
	envDatabaseURL = "DATABASE_URL_OWNER"
	driverName     = "pgx"
	tablePadding   = 2
	// stateHidden is a flag that is ON while a prerequisite is OFF.
	stateHidden = "hidden"
)

type verb string

const (
	verbList        verb = "list"
	verbEnable      verb = "enable"
	verbDisable     verb = "disable"
	verbApplyPreset verb = "apply-preset"
	verbPresets     verb = "presets"
)

// command is a parsed invocation: a verb and its single argument, if any.
type command struct {
	verb verb
	arg  string
}

var errUsage = errors.New("usage: flags list | enable <module_key> | disable <module_key> | " +
	"apply-preset <name> | presets (owner URL in " + envDatabaseURL + ")")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, os.Args[1:], os.LookupEnv, os.Stdout)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "flags:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, lookup config.Lookup, out io.Writer) error {
	cmd, err := parseCommand(args)
	if err != nil {
		return err
	}
	dsn, err := config.NewEnvSecrets(lookup).Secret(ctx, envDatabaseURL)
	if err != nil {
		return fmt.Errorf("%s: %w", envDatabaseURL, err)
	}
	db, err := sql.Open(driverName, dsn.Reveal())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	return errors.Join(execute(ctx, cmd, db, out), db.Close())
}

func parseCommand(args []string) (command, error) {
	if len(args) == 0 {
		return command{}, errUsage
	}
	v, rest := verb(args[0]), args[1:]
	switch v {
	case verbList, verbPresets:
		if len(rest) != 0 {
			return command{}, fmt.Errorf("%w: %s takes no argument", errUsage, v)
		}
		return command{verb: v}, nil
	case verbEnable, verbDisable, verbApplyPreset:
		if len(rest) != 1 {
			return command{}, fmt.Errorf("%w: %s takes exactly one argument", errUsage, v)
		}
		return command{verb: v, arg: rest[0]}, nil
	default:
		return command{}, fmt.Errorf("%w: unknown command %q", errUsage, args[0])
	}
}

// execute runs cmd against db after checking that db is logged in as the owner.
func execute(ctx context.Context, cmd command, db *sql.DB, out io.Writer) error {
	if err := migrations.RequireOwner(ctx, db); err != nil {
		return err
	}
	g, err := flagsgorm.Open(db)
	if err != nil {
		return err
	}
	return perform(ctx, cmd, flags.NewAdmin(flagsgorm.NewStore(g)), out)
}

func perform(ctx context.Context, cmd command, admin flags.Admin, out io.Writer) error {
	switch cmd.verb {
	case verbEnable, verbDisable:
		return setEnabled(ctx, admin, cmd.arg, cmd.verb == verbEnable, out)
	case verbApplyPreset:
		result, err := admin.ApplyPreset(ctx, cmd.arg)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "applied preset %s\n", cmd.arg)
		return printFlags(out, result)
	case verbPresets:
		presets, err := admin.Presets(ctx)
		if err != nil {
			return err
		}
		printPresets(out, presets)
		return nil
	default:
		list, err := admin.List(ctx)
		if err != nil {
			return err
		}
		return printFlags(out, list)
	}
}

func setEnabled(ctx context.Context, admin flags.Admin, key string, enabled bool, out io.Writer) error {
	previous, err := admin.SetEnabled(ctx, key, enabled)
	if err != nil {
		return err
	}
	if previous == enabled {
		fmt.Fprintf(out, "%s: already %s (no change)\n", key, state(enabled))
		return nil
	}
	fmt.Fprintf(out, "%s: %s -> %s\n", key, state(previous), state(enabled))
	fmt.Fprintf(out, "running API instances pick this up within %s (flag cache TTL)\n", flags.DefaultTTL)
	list, err := admin.List(ctx)
	if err != nil {
		return err
	}
	printHidden(out, effectiveStates(list))
	return nil
}

// flagState is one flag with its effective state: a flag that is ON while a
// prerequisite (flags.Prerequisites) is OFF is hidden by the API (404 and
// absent from GET /v1/flags).
type flagState struct {
	flags.Flag
	requires []string
	missing  []string
}

func (f flagState) state() string {
	if f.Enabled && len(f.missing) > 0 {
		return stateHidden
	}
	return state(f.Enabled)
}

func effectiveStates(list []flags.Flag) []flagState {
	graph := flags.Prerequisites()
	on := make(map[string]bool, len(list))
	for _, f := range list {
		on[f.Key] = f.Enabled
	}
	isOn := func(k string) bool { return on[k] }
	out := make([]flagState, 0, len(list))
	for _, f := range list {
		out = append(out, flagState{Flag: f, requires: graph[f.Key], missing: flags.MissingPrerequisites(graph, f.Key, isOn)})
	}
	return out
}

func printFlags(out io.Writer, list []flags.Flag) error {
	states := effectiveStates(list)
	tw := tabwriter.NewWriter(out, 0, 0, tablePadding, ' ', 0)
	fmt.Fprintln(tw, "KEY\tSTATE\tREQUIRES\tDESCRIPTION")
	on := 0
	for _, f := range states {
		if f.state() == state(true) {
			on++
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", f.Key, f.state(), requiresColumn(f.requires), f.Description)
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write table: %w", err)
	}
	fmt.Fprintf(out, "%d of %d modules ON\n", on, len(list))
	printHidden(out, states)
	return nil
}

// printHidden says, for each hidden flag, which prerequisites to turn ON.
func printHidden(out io.Writer, states []flagState) {
	for _, f := range states {
		if f.state() != stateHidden {
			continue
		}
		verb := "is"
		if len(f.missing) > 1 {
			verb = "are"
		}
		fmt.Fprintf(out, "%s is hidden until %s %s ON\n", f.Key, strings.Join(f.missing, ", "), verb)
	}
}

func requiresColumn(requires []string) string {
	if len(requires) == 0 {
		return "-"
	}
	return strings.Join(requires, ",")
}

func printPresets(out io.Writer, presets []flags.Preset) {
	for _, p := range presets {
		fmt.Fprintf(out, "%s  %s\n  %s\n", p.Name, p.Description, strings.Join(p.Flags, ", "))
	}
}

// state is upper-case ON so enabled modules stand out in the table.
func state(on bool) string {
	if on {
		return "ON"
	}
	return "off"
}
