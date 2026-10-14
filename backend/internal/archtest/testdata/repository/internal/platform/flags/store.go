package flags

// Platform stores that own their pool (flagsgorm) are outside the rule: it
// covers module repositories, which run in the use case's unit of work.

func (s store) load() { s.db.Transaction(nil) }
