package domain

// EngineVersion is the version of the computation rules, shown in the
// breakdown header ("Motor 1.1.1", US-X2) and hashed into the records
// fingerprint: bump it whenever a strategy changes a figure for the same
// input, so M8 marks the months as changed.
const EngineVersion = "1.1.1"

// Legend is the mandatory notice of every estimate and breakdown (PRD §6,
// contract Legend const), rendered from this server field, never a client
// copy.
const Legend = "Esto es un aproximado. No todos los gastos aplican para tu giro comercial."
