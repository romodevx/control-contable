package infra

// Test files may drive transactions themselves (fixtures, probes).

func helper() { db.Begin(); db.Commit() }

const fixture = "SELECT set_config('app.current_user_id', $1, true)"
