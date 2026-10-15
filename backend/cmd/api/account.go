package main

import (
	"log/slog"

	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/auth/infra"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
)

// newAccountHandler wires the account use cases (M1-09): the profile,
// account deletion (its password attempts limited per account, so a stolen
// access token is no password oracle) and the legal documents, over the
// cc_app unit of work.
func newAccountHandler(txm *pgdb.TxManager, keys authCrypto, limits *ratelimitpg.Store, logger *slog.Logger) *authhttp.AccountHandler {
	accounts := infra.NewAccountStore(keys.emails)
	return authhttp.NewAccountHandler(authhttp.AccountUseCases{
		// getMe carries the access the dashboard shows (contract 0.16.0), from
		// the same policy every module uses.
		Profile: usecase.NewProfile(txm, accounts,
			access.NewPolicy(txm, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem())),
		Delete: usecase.NewDeleteAccount(usecase.DeleteAccountDeps{
			Tx: txm, Passwords: keys.passwords, Accounts: accounts,
			Attempts: limits.Bucket(usecase.DeleteAttemptsPolicy), Logger: logger,
		}),
		Legal: usecase.NewListLegalDocuments(txm, infra.NewLegalDocumentStore()),
	})
}
