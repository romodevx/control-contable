package infra

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
)

// SQLSTATEs the stores translate.
const (
	sqlstateUniqueViolation     = "23505"
	sqlstateForeignKeyViolation = "23503"
)

// constraintErrors maps "SQLSTATE/constraint" to the use case error it means
// (migration 20261013211000_expenses). A constraint not listed stays an
// infrastructure failure: the use cases validate first, so reaching it is a
// bug, not a client error.
var constraintErrors = map[string]error{
	sqlstateUniqueViolation + "/expenses_pkey":                   usecase.ErrIDConflict,
	sqlstateForeignKeyViolation + "/expenses_category_code_fkey": usecase.ErrCategoryNotFound,
}

// translate maps a database refusal to its use case error, keeping the
// driver error for the logs; any other error is wrapped with op.
func translate(op string, err error) error {
	if err == nil {
		return nil
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if target, known := constraintErrors[pgErr.Code+"/"+pgErr.ConstraintName]; known {
			return fmt.Errorf("%w (expenses infra: %s: %w)", target, op, err)
		}
	}
	return fmt.Errorf("expenses infra: %s: %w", op, err)
}
