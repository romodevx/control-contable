// Command flags is the owner's module feature-flag admin CLI (P0-20; backs
// the /flags skill). It connects as cc_owner, the only role allowed to write
// feature_flags, through the owner URL in DATABASE_URL_OWNER:
//
//	flags list                   every flag, ON/off, with a summary
//	flags enable <module_key>    turn one module ON
//	flags disable <module_key>   turn one module off
//	flags apply-preset <name>    turn the preset's modules ON and the rest off
//	flags presets                the presets and their modules
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
	return nil
}

func printFlags(out io.Writer, list []flags.Flag) error {
	tw := tabwriter.NewWriter(out, 0, 0, tablePadding, ' ', 0)
	fmt.Fprintln(tw, "KEY\tSTATE\tDESCRIPTION")
	on := 0
	for _, f := range list {
		if f.Enabled {
			on++
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", f.Key, state(f.Enabled), f.Description)
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write table: %w", err)
	}
	fmt.Fprintf(out, "%d of %d modules ON\n", on, len(list))
	return nil
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
