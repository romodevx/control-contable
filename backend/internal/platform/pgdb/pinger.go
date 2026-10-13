package pgdb

import (
	"context"
	"database/sql"
	"fmt"
)

// Pinger is the readiness probe of the pool (health.Pinger).
type Pinger struct {
	db *sql.DB
}

// NewPinger returns a Pinger over db.
func NewPinger(db *sql.DB) Pinger { return Pinger{db: db} }

// Ping runs `SELECT 1`: a real round trip on a pooled connection.
// sql.DB.PingContext is not used because the pgx driver skips the round trip
// for a connection that was used within the last second.
func (p Pinger) Ping(ctx context.Context) error {
	if _, err := p.db.ExecContext(ctx, "SELECT 1"); err != nil {
		return fmt.Errorf("pgdb: ping: %w", err)
	}
	return nil
}
