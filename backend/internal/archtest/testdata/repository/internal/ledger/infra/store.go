package infra

// Every statement below breaks the repository rule (P0-30 D-3).

func (s store) begin()     { s.db.Begin() }
func (s store) beginTx()   { s.pool.BeginTx(nil, nil) }
func (s store) commit()    { s.db.Commit() }
func (s store) rollback()  { s.db.Rollback() }
func (s store) savepoint() { s.db.SavePoint("a"); s.db.RollbackTo("a") }
func (s store) inTx()      { s.db.Transaction(nil) }

const setConfig = "SELECT set_config('app.current_user_id', $1, true)"

const setLocal = "set local app.current_user_id = 'x'"

const reset = `RESET ROLE`

const role = "SET ROLE cc_owner"

const chained = "SELECT 1; set search_path to evil"
