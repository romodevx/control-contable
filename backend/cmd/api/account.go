package main

import (
	"log/slog"
	"time"

	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/auth/infra"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// Account deletion's password attempts per account (MODULE-auth §5): a
// burst of deleteAttemptBurst, then one more every deleteAttemptInterval
// (5 per 15 minutes once spent), so a stolen access token is no password
// oracle. Keyed by user id, not by client address.
const (
	deleteAttemptBurst    = 5
	deleteAttemptInterval = 3 * time.Minute
)

// newAccountHandler wires the account use cases (M1-09): the profile,
// account deletion and the legal documents, over the cc_app unit of work.
func newAccountHandler(txm *pgdb.TxManager, keys authCrypto, now func() time.Time, logger *slog.Logger) *authhttp.AccountHandler {
	accounts := infra.NewAccountStore(keys.emails)
	return authhttp.NewAccountHandler(authhttp.AccountUseCases{
		Profile: usecase.NewProfile(txm, accounts),
		Delete: usecase.NewDeleteAccount(usecase.DeleteAccountDeps{
			Tx: txm, Passwords: keys.passwords, Accounts: accounts,
			Attempts: ratelimit.NewEvery(deleteAttemptInterval, deleteAttemptBurst, now), Logger: logger,
		}),
		Legal: usecase.NewListLegalDocuments(txm, infra.NewLegalDocumentStore()),
	})
}
