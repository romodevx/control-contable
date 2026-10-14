package infra

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// SQLSTATEs the stores translate.
const (
	sqlstateUniqueViolation     = "23505"
	sqlstateForeignKeyViolation = "23503"
	sqlstateCheckViolation      = "23514"
	sqlstateExclusionViolation  = "23P01"
)

// constraintErrors maps "SQLSTATE/constraint" to the use case or domain
// error it means (migration 20261013190000_taxpayers). A constraint not
// listed stays an infrastructure failure: the use cases validate first, so
// reaching it is a bug, not a client error.
var constraintErrors = map[string]error{
	sqlstateUniqueViolation + "/taxpayers_user_rfc_hash_key":               usecase.ErrRFCAlreadyRegistered,
	sqlstateUniqueViolation + "/taxpayers_pkey":                            usecase.ErrIDConflict,
	sqlstateUniqueViolation + "/taxpayer_regimes_pkey":                     usecase.ErrRegimeAlreadyActive,
	sqlstateUniqueViolation + "/taxpayer_regimes_user_taxpayer_regime_key": usecase.ErrRegimeAlreadyActive,
	sqlstateExclusionViolation + "/taxpayer_regimes_compatible_periods":    usecase.ErrIncompatibleRegimes,
	sqlstateForeignKeyViolation + "/taxpayer_regimes_regime_fkey":          usecase.ErrRegimeNotForPersonType,
	sqlstateForeignKeyViolation + "/taxpayer_regimes_taxpayer_fkey":        usecase.ErrTaxpayerNotFound,
	sqlstateCheckViolation + "/taxpayers_display_name_valid":               domain.ErrInvalidDisplayName,
	sqlstateCheckViolation + "/taxpayer_regimes_start_period_valid":        usecase.ErrPeriodOutOfRange,
	sqlstateCheckViolation + "/taxpayer_regimes_end_period_valid":          usecase.ErrPeriodOutOfRange,
}

// translate maps a database refusal to its use case error, keeping the
// driver error for the logs; any other error is wrapped with op.
func translate(op string, err error) error {
	if err == nil {
		return nil
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if target, known := constraintErrors[pgErr.Code+"/"+pgErr.ConstraintName]; known {
			return fmt.Errorf("%w (taxpayers infra: %s: %w)", target, op, err)
		}
	}
	return fmt.Errorf("taxpayers infra: %s: %w", op, err)
}

// bytea passes a []byte as one bytea parameter: GORM expands a bare []byte
// argument of Raw/Exec/Updates into one placeholder per byte.
type bytea []byte

// Value implements driver.Valuer.
func (b bytea) Value() (driver.Value, error) { return []byte(b), nil }
