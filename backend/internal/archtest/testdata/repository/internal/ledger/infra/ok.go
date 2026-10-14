package infra

// Nothing here breaks the repository rule: an UPDATE's SET clause, a column
// named like a setting and ordinary method calls are fine.

const update = `UPDATE entries
SET amount = $1
WHERE id = $2`

const column = "SELECT reset_at, settings FROM entries"

const label = "set archived_at" // an operation name for an error message

func (s store) create() { s.db.Create(nil); s.db.Exec(update) }
