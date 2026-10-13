// Package clock is the one place the backend reads the wall clock (ADR-0002
// §4). "Today" and the current tax period are Mexico City calendar facts, so
// every Clock returns times in America/Mexico_City. Use cases take a Clock;
// tests inject Fixed. The zone database is embedded (time/tzdata), so the
// binary does not depend on the host's zoneinfo.
package clock

import (
	"fmt"
	"time"
	_ "time/tzdata" // embed the IANA zone database: distroless images have none
)

// ZoneName is the IANA zone of every time a Clock returns.
const ZoneName = "America/Mexico_City"

var mexicoCity = mustLoadLocation(ZoneName)

// Clock is the port for "now". Now returns the current instant in Location().
type Clock interface {
	Now() time.Time
}

// Location returns the America/Mexico_City location.
func Location() *time.Location { return mexicoCity }

// System reads the wall clock.
type System struct{}

// NewSystem returns the production Clock.
func NewSystem() System { return System{} }

// Now returns the current instant in Mexico City.
func (System) Now() time.Time { return time.Now().In(mexicoCity) }

// Fixed always returns the same instant: the deterministic Clock for tests.
type Fixed struct {
	instant time.Time
}

// NewFixed returns a Clock frozen at instant, expressed in Mexico City.
func NewFixed(instant time.Time) Fixed { return Fixed{instant: instant.In(mexicoCity)} }

// Now returns the frozen instant.
func (f Fixed) Now() time.Time { return f.instant }

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(fmt.Errorf("clock: load time zone %q: %w", name, err))
	}
	return loc
}
