package main

import (
	"log/slog"

	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/auth/infra"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// newAccountHandler wires the account use cases (M1-09): the profile,
// account deletion and the legal documents, over the cc_app unit of work.
func newAccountHandler(txm *pgdb.TxManager, keys authCrypto, logger *slog.Logger) *authhttp.AccountHandler {
	accounts := infra.NewAccountStore(keys.emails)
	return authhttp.NewAccountHandler(authhttp.AccountUseCases{
		Profile: usecase.NewProfile(txm, accounts),
		Delete: usecase.NewDeleteAccount(usecase.DeleteAccountDeps{
			Tx: txm, Passwords: keys.passwords, Accounts: accounts, Logger: logger,
		}),
		Legal: usecase.NewListLegalDocuments(txm, infra.NewLegalDocumentStore()),
	})
}
